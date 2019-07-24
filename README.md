### csi
CSI driver implementation for openebs storage engines.
Currently, Volumes can only be provisioned for openebs cstor engine.

### Overview
OpenEBS CSI driver implementation comprises of 2 components:
1) Controller service: Runs as stateful set
2) Node service: Runs as a daemonset

### Prerequisites
1) Kubernetes version 1.14+
2) OpenEBS Version 1.1+ ([openebs-operator](https://raw.githubusercontent.com/openebs/openebs/master/k8s/openebs-operator.yaml))

### Provision a volume using OpenEBS CSI driver

1. Modify and apply the OpenEBS CSI Operator based on the host OS, with the steps mentioned in the yaml itself:
```
kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/csi-operator.yaml
```
2. Create a storage pool claim(spc) where the volume can be provisioned. In the below spc.yaml make sure that maxPools should be greater than or equal to the number of replicas required for the volume.
This step can be avoided if volume needs to be created on already existing cstor pools. 
```
kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/spc.yaml
```
3. Create a Storage Class pointing to OpenEBS CSI provisioner after updating these values in sc.yaml:
1) replicaCount: Number of copies of the data required to be maintained
2) storagePoolClaim: Where the volumes need to be created
```
kubectl apply -f https://raw.githubusercontent.com/openebs/csi/master/deploy/sc.yaml
```
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
1) Target service
2) Target deployment
3) CstorVolume CR
4) CstoVolumeReplica CR
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
1) `Waiting for CVC to be bound`: Implies volume components are still being created
2) `Volume is not ready: Replicas yet to connect to controller`: Implies volume components are already created but yet to interact with each other.

On successful completion of these steps the application pod can be seen in running state.
### NOTE
This is very much a work in progress and is currently considered as `experimental`.
