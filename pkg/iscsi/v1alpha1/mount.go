package iscsi

import (
	apis "github.com/openebs/cstor-csi/pkg/apis/openebs.io/core/v1alpha1"
	"k8s.io/kubernetes/pkg/util/mount"
)

// UnmountAndDetachDisk unmounts the disk from the specified path
// and logs out of the iSCSI Volume
func UnmountAndDetachDisk(vol *apis.CSIVolume, path string) error {
	iscsiInfo := &iscsiDisk{
		VolName: vol.Spec.Volume.Name,
		Portals: []string{vol.Spec.ISCSI.TargetPortal},
		Iqn:     vol.Spec.ISCSI.Iqn,
		lun:     vol.Spec.ISCSI.Lun,
		Iface:   vol.Spec.ISCSI.IscsiInterface,
	}

	diskUnmounter := &iscsiDiskUnmounter{
		iscsiDisk: iscsiInfo,
		mounter:   &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()},
		exec:      mount.NewOsExec(),
	}
	util := &ISCSIUtil{}
	return util.DetachDisk(*diskUnmounter, path)
}

// Unmount unmounts the path provided
func Unmount(path string) error {
	diskUnmounter := &iscsiDiskUnmounter{
		mounter: &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()},
		exec:    mount.NewOsExec(),
	}
	util := &ISCSIUtil{}
	return util.UnmountDisk(*diskUnmounter, path)
}

// ResizeVolume rescans the iSCSI session and runs the resize to filesystem
// command on that particular device
func ResizeVolume(volumePath string) error {
	mounter := mount.New("")
	list, _ := mounter.List()
	for _, mpt := range list {
		if mpt.Path == volumePath {
			util := &ISCSIUtil{}
			if err := util.ReScan(); err != nil {
				return err
			}
			if err := util.ReSizeFS(mpt.Device); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
