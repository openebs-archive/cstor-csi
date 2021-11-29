/*
 Copyright © 2020 The OpenEBS Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package iscsi

import (
	apis "github.com/openebs/api/v3/pkg/apis/cstor/v1"
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
