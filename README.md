# OpenEBS CSI Driver

CSI driver implementation for OpenEBS storage engines. 

## Project Status

This project is under active development and considered to be in Alpha state.

The current implementation only supports provisioning and de-provisioning of cStor Volumes. 

## Usage

### Prerequisites

Before setting up OpenEBS CSI driver make sure your Kubernetes Cluster 
meets the following prerequisities:

1. You will need to have Kubernetes version 1.14 or higher
2. You will need to have OpenEBS Version 1.1 or higher installed. 
   The steps to install OpenEBS are [here](https://docs.openebs.io/docs/next/quickstart.html)
3. iSCSI initiator utils installed on all the worker nodes
4. You have access to install RBAC components into kube-system namespace.
   The OpenEBS CSI driver components are installed in kube-system 
   namespace to allow them to be flagged as system critical components. 

### Setup OpenEBS CSI Driver

OpenEBS CSI driver comprises of 2 components:
- A controller component launched as a Stateful set, 
  implementing the CSI controller services. The Control Plane
  services are responsible for creating/deleting the required 
  OpenEBS Volume.
- A node component that runs as a DaemonSet, 
  implmenting the CSI node services. The node component is 
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
   create one using the below command. In the below spc.yaml make sure 
   that maxPools should be greater than or equal to the number of 
   replicas required for the volume.

   The following command will create the specified number of cStor Pools
   using the Sparse files. 

   ```
   kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/spc.yaml
   ```

2. Create a Storage Class to dynamically provision volumes 
   using OpenEBS CSI provisioner. A sample storage class looks like:
   ```
   kind: StorageClass
   apiVersion: storage.k8s.io/v1
   metadata:
     name: openebs-csi-cstor-sparse
     namespace: kube-system
     annotations:
   provisioner: openebs-csi.openebs.io
   allowVolumeExpansion: true
   parameters:
     storagePoolClaim: cstor-sparse-pool
     replicaCount: "1"
   ```
   You will need to specificy the correct cStor SPC from your cluster 
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

4. Create PVC with above Storage Class:
```
kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/pvc.yaml
```
Since the provisioner specified in the SC is openebs-csi.openebs.io,
CSI Controller service recieves a Volume creation request CreateVolume() via grpc. 
which in turn creates a CStorVolumeClaim(CVC) CR. This will be created with empty nodeID and status marked as pending. 
And a succees response is sent immediately after creating CVC. Once this API responds success kubernetes creates a PV object and binds it to PVC

The watcher for CVC CR in m-apiserver waits for node-id to be filled to provision the volume.
```
# Sample CVC CR on receiving VolumeCreate Request
apiVersion: openebs.io/v1alpha1
kind: CStorVolumeClaim
metadata:
  annotations:
    openebs.io/volumeID: pvc-*
  finalizers:
  - cvc.openebs.io/finalizer
  labels:
    openebs.io/storage-pool-claim: cstor-sparse-pool
  name: pvc-*
  namespace: openebs
spec:
  capacity:
    storage: 
status: 
  phase: Pending
```

5. Deploy a sample app with the above PVC:
```
kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/percona.yaml
kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/sqltest_configmap.yaml
```

On deploying the app CSI Node Service receives a NodePublishVolume() request via grpc,
which in turn patches nodeID to the previously created CVC CR and waits for the  status to be updated to bound by CVC watcher. 
The bound status implies that the following required volume components have been created by CVC watcher:
- Target service
- Target deployment
- CstorVolume CR
- CstoVolumeReplica CR
```
# Sample CVC CR after NodePublish is successful
apiVersion: openebs.io/v1alpha1
kind: CStorVolumeClaim
metadata:
  annotations:
    openebs.io/volumeID: pvc-*
  finalizers:
  - cvc.openebs.io/finalizer
  labels:
    openebs.io/storage-pool-claim: cstor-sparse-pool
  name: pvc-*
  namespace: openebs
publish:
  nodeId: csi-node-2
spec:
  capacity:
    storage:
  cstorVolumeRef:
  replicaCount: 
status:
  phase: Bound
```
Once the status is changed to bound, steps to mount the volume are processed.
While these steps are in progress, there might be some intermittent errors seen on describing the application pod:
- `Waiting for CVC to be bound`: Implies volume components are still being created
- `Volume is not ready: Replicas yet to connect to controller`: Implies volume components are already created but yet to interact with each other.

On successful completion of these steps the application pod can be seen in running state.
