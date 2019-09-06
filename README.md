# OpenEBS CSI Driver

CSI driver implementation for OpenEBS storage engines. 

## Project Status

This project is under active development and considered to be in Alpha state.

The current implementation only supports provisioning, de-provisioning and expansion of cStor Volumes. 

## Usage

### Prerequisites

Before setting up OpenEBS CSI driver make sure your Kubernetes Cluster 
meets the following prerequisites:

1. You will need to have Kubernetes version 1.14 or higher
2. You will need to have OpenEBS Version 1.2 or higher installed. 
   The steps to install OpenEBS are [here](https://docs.openebs.io/docs/next/quickstart.html)
3. iSCSI initiator utils installed on all the worker nodes
4. You have access to install RBAC components into kube-system namespace.
   The OpenEBS CSI driver components are installed in kube-system 
   namespace to allow them to be flagged as system critical components.
5. You will need to turn on  ExpandCSIVolumes and ExpandInUsePersistentVolumes feature gates on  kubelets and kube-apiserver:

### Setup OpenEBS CSI Driver

OpenEBS CSI driver comprises of 2 components:
- A controller component launched as a StatefulSet, 
  implementing the CSI controller services. The Control Plane
  services are responsible for creating/deleting the required 
  OpenEBS Volume.
- A node component that runs as a DaemonSet, 
  implementing the CSI node services. The node component is 
  responsible for performing the iSCSI connection management and
  connecting to the OpenEBS Volume.

OpenEBS CSI driver components can be installed by running the 
following command. 

The node components make use of the host iSCSI binaries for iSCSI 
connection management. Depending on the OS, the spec will have to 
be modified to load the required iSCSI files into the node pods. 

Depending on the OS select the appropriate deployment file.

- For Ubuntu 16.04 and CentOS.
  ```
  kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/csi-operator.yaml
  ```

- For Ubuntu 18.04 
  ```
  kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/csi-operator-ubuntu-18.04.yaml
  ```

Verify that the OpenEBS CSI Components are installed. 

```
$ kubectl get pods -n kube-system -l role=openebs-csi
NAME                       READY   STATUS    RESTARTS   AGE
openebs-csi-controller-0   4/4     Running   0          6m14s
openebs-csi-node-56t5g     2/2     Running   0          6m13s

```

### Provision a cStor volume using OpenEBS CSI driver

1. Make sure you already have a cStor Pool Created or you can 
   create one using the below command. In the below cspc.yaml make sure 
   that the specified pools list should be greater than or equal to
   the number of replicas required for the volume. Update `kubernetes.io/hostname`
   and `blockDeviceName` in the below yaml before applying the same.

   The following command will create the specified cStor Pools in the cspc yaml:

   ```
   kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/cspc.yaml
   ```

2. Create a Storage Class to dynamically provision volumes 
   using OpenEBS CSI provisioner. A sample storage class looks like:
   ```
   kind: StorageClass
   apiVersion: storage.k8s.io/v1
   metadata:
     name: openebs-csi-cstor-sparse
   provisioner: openebs-csi.openebs.io
   allowVolumeExpansion: true
   parameters:
     cas-type: cstor
     cstorPoolCluster: cstor-sparse-cspc
     replicaCount: "1"
   ```
   You will need to specify the correct cStor CSPC from your cluster 
   and specify the desired `replicaCount` for the volume. The `replicaCount`
   should be less than or equal to the max pools available.  

   The following file helps you to create a Storage Class
   using the cStor sparse pool created in the previous step. 
   ```
   kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/sc.yaml
   ```

3. Run your application by specifying the above Storage Class for 
   the PVCs. 

   The following example launches a busybox pod using a cStor Volume 
   provisioned via CSI Provisioner. 
   ```
   kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/busybox-csi-cstor-sparse.yaml
   ```

   Verify that the pods is running and is able to write the data. 
   ```
   $ kubectl get pods
   NAME      READY   STATUS    RESTARTS   AGE
   busybox   1/1     Running   0          97s
   ```

   The busybox is instructed to write the date when it starts into the 
   mounted path at `/mnt/openebs-csi/date.txt`

   ```
   $ kubectl exec -it busybox -- cat /mnt/openebs-csi/date.txt
   Wed Jul 31 04:56:26 UTC 2019
   ```
   

### How does it work?

The following steps indicate the PV provisioning workflow as it passes
through various components. 

1. Create PVC with Storage Class referring to OpenEBS CSI Driver.

2. Kubernetes will pass the PV creation request to the OpenEBS
   CSI Controller service via `CreateVolume()`, as this controller
   registered with Kubernetes for receiving any requests related to
   `openebs-csi.openebs.io`  

3. OpenEBS CSI Controller will create a custom resource called 
   `CStorVolumeClaim(CVC)` and returns the details of the newly 
   created object back to Kubernetes. The `CVC`s will be
   monitored by the cstor-operator (embedded in m-apiserver). The
   cstor-operator will wait to proceed with provisioning a `CStorVolume`
   for a given `CVC` until the Kubernetes has scheduled the application 
   using the PVC/CVC to a node in the cluster. 

   This is in effect working like `waitforFirstConsumer`.

4. When the node is assigned for the application, Kubernetes will 
   invoke the `NodePublishVolume()` request with the node and the 
   volume details - which includes the identifier of the CVC. 

   This API will then specify the node details in the CVC. 

   After updating the node id, the OpenEBS CSI Driver - Node
   Service will wait for the CVC to be bound to an actual cStor Volume.

5. The cstor-operator checks that node details are available on CVC, 
   and proceeds with the cStor Volume Creation. Once the cStor Volume 
   is created, the CVC is updated with the reference to the cStor Volume
   and change the status on CVC to bound.

6. Node Component which was waiting on the CVC status will proceed
   to connect to the cStor volume. 


Note: While the asynchronous handling of the Volume provisioning is 
in progress, the application pod may throw some errors like:

- `Waiting for CVC to be bound`: Implies volume components are still being created
- `Volume is not ready: Replicas yet to connect to controller`: 
   Implies volume components are already created but yet to interact with each other.

On successful completion of the above steps the application pod can 
be seen in running state.

### Expand a cStor volume using OpenEBS CSI driver

#### Notes:
- Only dynamically provisioned volumes can be resized.
- You can only resize volumes containing a file system if the file system is XFS, Ext3, or Ext4.
- Make sure that the storage class has the `allowVolumeExpansion` field set to `true` when the volume is provisioned.

#### Steps:
1. Update the increased pvc size in the pvc spec section (pvc.spec.resources.requests.storage).
2. Wait for the updated capacity to reflect in PVC status (pvc.status.capacity.storage).

It is internally a two step process for volumes containing a file system:
1. Volume expansion
2. FileSystem expansion
