package v1alpha1

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-iscsi/iscsi"
	apis "github.com/openebs/cstor-csi/pkg/apis/openebs.io/core/v1alpha1"
	iscsiutils "github.com/openebs/cstor-csi/pkg/iscsi/v1alpha1"
	utils "github.com/openebs/cstor-csi/pkg/utils/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (ns *node) attachDisk(vol *apis.CSIVolume) (string, error) {
	connector := iscsi.Connector{
		VolumeName:    vol.Spec.Volume.Name,
		TargetIqn:     vol.Spec.ISCSI.Iqn,
		Lun:           defaultISCSILUN,
		Interface:     defaultISCSIInterface,
		TargetPortals: []string{fmt.Sprintf("%v", vol.Spec.ISCSI.TargetPortal)},
		DoDiscovery:   true,
	}

	logrus.Debugf("NodeStageVolume: attach disk with config: {%+v}", connector)
	devicePath, err := iscsi.Connect(connector)
	if err != nil {
		return "", err
	}

	if devicePath == "" {
		return "", fmt.Errorf("connect reported success, but no path returned")
	}
	return devicePath, err
}

func (ns *node) formatAndMount(req *csi.NodeStageVolumeRequest, devicePath string) error {
	// Mount device
	mntPath := req.GetStagingTargetPath()
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(mntPath)
	if err != nil && !os.IsNotExist(err) {
		if err := os.MkdirAll(mntPath, 0750); err != nil {
			logrus.Errorf("failed to mkdir %s, error", mntPath)
			return err
		}
	}

	if !notMnt {
		logrus.Infof("Volume %s has been mounted already at %v", req.GetVolumeId(), mntPath)
		return nil
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	options := []string{}
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
	options = append(options, mountFlags...)

	err = ns.mounter.FormatAndMount(devicePath, mntPath, fsType, options)
	if err != nil {
		logrus.Errorf(
			"Failed to mount iscsi volume %s [%s, %s] to %s, error %v",
			req.GetVolumeId(), devicePath, fsType, mntPath, err,
		)
		return err
	}
	return nil
}

func (ns *node) nodePublishVolumeForFileSystem(req *csi.NodePublishVolumeRequest, mountOptions []string, mode *csi.VolumeCapability_Mount) error {
	target := req.GetTargetPath()
	source := req.GetStagingTargetPath()
	if m := mode.Mount; m != nil {
		hasOption := func(options []string, opt string) bool {
			for _, o := range options {
				if o == opt {
					return true
				}
			}
			return false
		}
		for _, f := range m.MountFlags {
			if !hasOption(mountOptions, f) {
				mountOptions = append(mountOptions, f)
			}
		}
	}

	logrus.Infof("NodePublishVolume: creating dir %s", target)
	if err := os.MkdirAll(target, 0000); err != nil {
		return status.Errorf(codes.Internal, "Could not create dir {%q}, err: %v", target, err)
	}

	// in case if the dir already exists, above call returns nil
	// so permission needs to be updated
	if err := os.Chmod(target, 0000); err != nil {
		return status.Errorf(codes.Internal, "Could not change mode of dir {%q}, err: %v", target, err)
	}
	fsType := mode.Mount.GetFsType()
	if len(fsType) == 0 {
		fsType = defaultFsType
	}

	logrus.Infof("NodePublishVolume: mounting %s at %s with option %s as fstype %s", source, target, mountOptions, fsType)
	if err := ns.mounter.Mount(source, target, fsType, mountOptions); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, err)
		}
		return status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, target, err)
	}

	return nil
}

// newNodeCapabilities returns a list
// of this Node's capabilities
func newNodeCapabilities() []*csi.NodeServiceCapability {
	fromType := func(
		cap csi.NodeServiceCapability_RPC_Type,
	) *csi.NodeServiceCapability {
		return &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var capabilities []*csi.NodeServiceCapability
	for _, cap := range []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
}

// IsUnmountRequired returns true if the volume needs to be unmounted
func IsUnmountRequired(volumeID, targetPath string) (bool, error) {
	var (
		currentMounts []string
		err           error
	)
	currentMounts, err = utils.GetMounts(volumeID)
	if err != nil {
		return false, err
	}
	if len(currentMounts) > 1 {
		logrus.Fatalf(
			"More than one mounts for volume:%s mounts: %v",
			volumeID, currentMounts,
		)
	}
	if len(currentMounts) == 0 {
		return false, nil
	}
	if currentMounts[0] != targetPath {
		return false, err
	}
	return true, nil
}

// VerifyIfMountRequired returns true if volume is already mounted on targetPath
// and unmounts if it is mounted on a different path
func VerifyIfMountRequired(volumeID, targetPath string) (bool, error) {
	var (
		currentMounts []string
		err           error
	)
	currentMounts, err = utils.GetMounts(volumeID)
	if err != nil {
		return false, err
	}
	if len(currentMounts) > 1 {
		logrus.Fatalf(
			"More than one mounts for volume:%s mounts: %v",
			volumeID, currentMounts,
		)
	}
	if len(currentMounts) == 1 {
		if currentMounts[0] == targetPath {
			return false, nil
		}
		if err = iscsiutils.Unmount(currentMounts[0]); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (ns *node) validateNodeStageReq(
	req *csi.NodeStageVolumeRequest,
) error {
	if req.GetVolumeCapability() == nil {
		return status.Error(codes.InvalidArgument,
			"Volume capability missing in request")
	}

	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}
	return nil
}

func (ns *node) validateNodeUnStageReq(
	req *csi.NodeUnstageVolumeRequest,
) error {
	if req.GetVolumeId() == "" {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if req.GetStagingTargetPath() == "" {
		return status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}
	return nil
}
