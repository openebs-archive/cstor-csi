# OpenEBS cStor CSI Driver
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fcsi.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fcsi?ref=badge_shield)
[![Build Status](https://github.com/openebs/cstor-csi/actions/workflows/build.yml/badge.svg)](https://github.com/openebs/cstor-csi/actions/workflows/build.yml)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/cstor-csi)](https://goreportcard.com/report/github.com/openebs/cstor-csi)
[![Slack](https://img.shields.io/badge/JOIN-SLACK-blue)](https://kubernetes.slack.com/messages/openebs/)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://openebs.io/community)

CSI driver implementation for OpenEBS cStor storage engine.

## Project Status: Beta

The current implementation supports the following for cStor Volumes:
1. Provisioning and De-provisioning with ext4,xfs filesystems
2. Snapshots and clones
3. Volume Expansion
4. Volume Metrics

### Setup OpenEBS cStor CSI Driver

OpenEBS cStor CSI driver components can be installed by using the helm chart or operator yaml.
Refer to [cstor-operators](https://github.com/openebs/cstor-operators) for detailed usage instructions.

### How does it work?

OpenEBS cStor CSI driver comprises of 2 components:

- A controller component launched as a StatefulSet, implementing the CSI controller
  services. The Control Plane services are responsible for creating/deleting the required
  OpenEBS Volume.
- A node component that runs as a DaemonSet,implementing the CSI node services. 
  The node component is responsible for performing the iSCSI connection management and
  connecting to the OpenEBS Volume.

The following steps indicate the PV provisioning workflow as it passes
through various components.

1. Create PVC with Storage Class referring to OpenEBS cStor CSI Driver.

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

## Contributing

OpenEBS welcomes your feedback and contributions in any form possible.

- [Join OpenEBS community on Kubernetes Slack](https://kubernetes.slack.com)
  - Already signed up? Head to our discussions at [#openebs](https://kubernetes.slack.com/messages/openebs/)
- Want to raise an issue or help with fixes and features?
  - See [open issues](https://github.com/openebs/openebs/issues)
  - See [contributing guide](./CONTRIBUTING.md)
  - See [Project Roadmap](https://github.com/orgs/openebs/projects/9)
  - Want to join our contributor community meetings, [check this out](https://hackmd.io/mfG78r7MS86oMx8oyaV8Iw?view).
- Join our OpenEBS CNCF Mailing lists
  - For OpenEBS project updates, subscribe to [OpenEBS Announcements](https://lists.cncf.io/g/cncf-openebs-announcements)
  - For interacting with other OpenEBS users, subscribe to [OpenEBS Users](https://lists.cncf.io/g/cncf-openebs-users)


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Fcsi.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Fcsi?ref=badge_large)
