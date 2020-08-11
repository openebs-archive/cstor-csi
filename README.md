# OpenEBS CStor CSI Driver

CSI driver implementation for OpenEBS CStor storage engine.

## Project Status

This project is under active development and considered to be in Alpha state.

The current implementation supports the following for CStor Volumes:
1. Provisioning and De-provisioning with ext4,xfs filesystems
2. Snapshots and clones
3. Volume Expansion
4. Volume Metrics

## Usage

### Prerequisites

Before setting up OpenEBS CStor CSI driver make sure your Kubernetes Cluster
meets the following prerequisites:

1. You will need to have Kubernetes version 1.17 or higher
2. You will need to have OpenEBS Version 1.12 or higher installed.
   The steps to install OpenEBS are [here](https://docs.openebs.io/docs/next/quickstart.html)
3. CStor CSI driver operates on the cStor Pools provisioned using the new schema called CSPC.
   Steps to provision the pools using the same are [here](https://docs.google.com/document/d/1CZLBqGgBy7mTAMth4G2kR_jBn5bOV0th03ce3idX8KA/edit?usp=sharing)
4. iSCSI initiator utils installed on all the worker nodes
5. You have access to install RBAC components into openebs namespace.
   The OpenEBS CStor CSI driver components are installed in openebs
   namespace to allow them to be flagged as system critical components.

Note: if older k8s version has been used, it requires to enable ExpandCSIVolumes and ExpandInUsePersistentVolumes, VolumeSnapshotDataSource feature gates on  kubelets and kube-apiserver

### Setup OpenEBS CStor CSI Driver

OpenEBS CStor CSI driver comprises of 2 components:
- A controller component launched as a StatefulSet,
  implementing the CSI controller services. The Control Plane
  services are responsible for creating/deleting the required
  OpenEBS Volume.
- A node component that runs as a DaemonSet,
  implementing the CSI node services. The node component is
  responsible for performing the iSCSI connection management and
  connecting to the OpenEBS Volume.

OpenEBS CStor CSI driver components can be installed by running the
following command.

The node components make use of the host iSCSI binaries for iSCSI
connection management. Depending on the OS, iscsi utils needs to be 
installed and loaded in k8s workers nodes

- Install CStor CSI driver using below command.
  ```
  kubectl apply -f https://raw.githubusercontent.com/openebs/cstor-csi/master/deploy/csi-operator.yaml
  ```

- Verify that the OpenEBS CSI Components are installed.

  ```
  $ kubectl get pods -n openebs -l role=openebs-cstor-csi
  NAME                       READY   STATUS    RESTARTS   AGE
  openebs-csi-controller-0   4/4     Running   0          6m14s
  openebs-csi-node-56t5g     2/2     Running   0          6m13s
  ```

### Provision a cStor volume using OpenEBS CStor CSI driver

1. Make sure you already have a cStor Pool Created or you can
   create one using the below command. In the below cspc.yaml make sure
   that the specified pools list should be greater than or equal to
   the number of replicas required for the volume. Update `kubernetes.io/hostname`
   and `blockDeviceName` in the below yaml before applying the same.

   The following command will create the specified cStor Pools in the cspc yaml:

   ```
   kubectl apply -f https://raw.githubusercontent.com/openebs/cstor-csi/master/examples/cspc.yaml
   ```

2. Create a Storage Class to dynamically provision volumes
   using OpenEBS CSI provisioner. A sample storage class looks like:
   ```
   kind: StorageClass
   apiVersion: storage.k8s.io/v1
   metadata:
     name: openebs-csi-cstor-sparse
   provisioner: cstor.csi.openebs.io
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
   kubectl apply -f https://raw.githubusercontent.com/openebs/cstor-csi/master/examples/csi-storageclass.yaml
   ```

3. Run your application by specifying the above Storage Class for
   the PVCs.

   The following example launches a busybox pod using a cStor Volume
   provisioned via CSI Provisioner.
   ```
   kubectl apply -f https://raw.githubusercontent.com/openebs/cstor-csi/master/examples/busybox-csi-cstor-sparse.yaml
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

1. Create PVC with Storage Class referring to OpenEBS CStor CSI Driver.

2. Kubernetes will pass the PV creation request to the OpenEBS
   CSI Controller service via `CreateVolume()`, as this controller
   registered with Kubernetes for receiving any requests related to
   `cstor.csi.openebs.io`

3. OpenEBS CSI Controller will create a custom resource called
   `CStorVolumeConfig(CVC)` and returns the details of the newly
   created object back to Kubernetes. The `CVC`s will be
   monitored by the cvc-operator. The cvc-operator watches the CVC
   resources and creates the respected volume specific resources like
   CStorVolume(CV), Target deployement, CStorVolumeReplicas and K8s service.

   Once the cStor Volume is created, the CVC is updated with the reference to
   the cStor Volume and change the status on CVC to bound.

4. Node Component which was waiting on the CVC status to be `Bound` will proceed
   to connect to the cStor volume.

Note: While the asynchronous handling of the Volume provisioning is
in progress, the application pod may throw some errors like:

- `Waiting for CVC to be bound`: Implies volume components are still being created
- `Volume is not ready: Replicas yet to connect to controller`:
   Implies volume components are already created but yet to interact with each other.

On successful completion of the above steps the application pod can
be seen in running state.

### Expand a cStor volume using OpenEBS CStor CSI driver

#### Notes:
- Only dynamically provisioned volumes can be resized.
- You can only resize volumes containing a file system if the file system is ext4.
- Make sure that the storage class has the `allowVolumeExpansion` field set to `true` when the volume is provisioned.

#### Steps:
1. Update the increased pvc size in the pvc spec section (pvc.spec.resources.requests.storage).
2. Wait for the updated capacity to reflect in PVC status (pvc.status.capacity.storage).

It is internally a two step process for volumes containing a file system:
1. Volume expansion
2. FileSystem expansion

### Snapshot And Clone CStor Volume using OpenEBS CStor CSI Driver

#### Notes:
-  `VolumeSnapshotDataSource` feature gate needs to be enabled at kubelet and kube-apiserver,
    from k8s version 1.17.0 onwards `VolumeSnapshotDataSource` feature enable by default

#### Steps:
1. Create snapshot class pointing to cstor csi driver:
```
kubectl apply -f https://raw.githubusercontent.com/openebs/cstor-csi/master/deploy/snapshot-class.yaml
```
2. Create a snapshot after updating the PVC and snapshot name in the following yaml:
```
kubectl apply -f https://raw.githubusercontent.com/openebs/cstor-csi/master/examples/csi-snapshot.yaml
```
3. Verify that the snapshot has been created successfully:
```
kubectl get volumesnapshots.snapshot
NAME            AGE
demo-snapshot   3d1h
```
4. Create the volume clone using the above Snapshot by updating and modifying the following yaml:
```
kubectl apply -f https://raw.githubusercontent.com/openebs/cstor-csi/master/examples/csi-pvc-clone.yaml
```
5. Verify that the PVC has been successfully created:
```
kubectl get pvc
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS              AGE
demo-csivol-claim       Bound    pvc-52d88903-0518-11ea-b887-42010a80006c   5Gi        RWO            openebs-csi-cstor-sparse  3d1h
pvc-clone               Bound    pvc-2f2d65fc-0784-11ea-b887-42010a80006c   5Gi        RWO            openebs-csi-cstor-sparse  3s
```
