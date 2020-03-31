package iscsi

import (
	apis "github.com/openebs/cstor-csi/pkg/apis/cstor/v1"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
)

// UnmountAndDetachDisk unmounts the disk from the specified path
// and logs out of the iSCSI Volume
func UnmountAndDetachDisk(vol *apis.CStorVolumeAttachment, path string) error {
	iscsiInfo := &iscsiDisk{
		VolName: vol.Spec.Volume.Name,
		Portals: []string{vol.Spec.ISCSI.TargetPortal},
		Iqn:     vol.Spec.ISCSI.Iqn,
		lun:     vol.Spec.ISCSI.Lun,
		Iface:   vol.Spec.ISCSI.IscsiInterface,
	}

	diskUnmounter := &iscsiDiskUnmounter{
		iscsiDisk: iscsiInfo,
		mounter:   &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()},
		//exec:      mount.NewOsExec(),
		exec: utilexec.New(),
	}
	util := &ISCSIUtil{}
	return util.DetachDisk(*diskUnmounter, path)
}

// Unmount unmounts the path provided
func Unmount(path string) error {
	diskUnmounter := &iscsiDiskUnmounter{
		mounter: &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()},
		//		exec:    mount.NewOsExec(),
		exec: utilexec.New(),
	}
	util := &ISCSIUtil{}
	return util.UnmountDisk(*diskUnmounter, path)
}

// ResizeVolume rescans the iSCSI session and runs the resize to filesystem
// command on that particular device
func ResizeVolume(volumePath string, vol *apis.CStorVolumeAttachment) error {
	var err error
	mounter := mount.New("")
	list, _ := mounter.List()
	for _, mpt := range list {
		if mpt.Path == volumePath {
			util := &ISCSIUtil{}
			if err := util.ReScan(vol.Spec.ISCSI.Iqn, vol.Spec.ISCSI.TargetPortal); err != nil {
				return err
			}
			switch vol.Spec.Volume.FSType {
			case "ext4":
				err = util.ResizeExt4(mpt.Device)
			case "xfs":
				err = util.ResizeXFS(volumePath)
			}
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}
