package v1alpha1

import (
	"fmt"

	apis "github.com/openebs/cstor-csi/pkg/apis/openebs.io/core/v1alpha1"
	utils "github.com/openebs/cstor-csi/pkg/utils/v1alpha1"
)

const (
	// FSTypeExt2 represents the ext2 filesystem type
	FSTypeExt2 = "ext2"
	// FSTypeExt3 represents the ext3 filesystem type
	FSTypeExt3 = "ext3"
	// FSTypeExt4 represents the ext4 filesystem type
	FSTypeExt4 = "ext4"
	// FSTypeXfs represents te xfs filesystem type
	FSTypeXfs = "xfs"

	defaultFsType = FSTypeExt4

	MaxRetryCount = 10

	defaultISCSILUN       = int32(0)
	defaultISCSIInterface = "default"
)

var (
	ValidFSTypes = []string{FSTypeExt2, FSTypeExt3, FSTypeExt4, FSTypeXfs}
)

func removeVolumeFromTransitionList(volumeID string) {
	utils.TransitionVolListLock.Lock()
	defer utils.TransitionVolListLock.Unlock()
	delete(utils.TransitionVolList, volumeID)
}

func addVolumeToTransitionList(volumeID string, status apis.CSIVolumeStatus) error {
	utils.TransitionVolListLock.Lock()
	defer utils.TransitionVolListLock.Unlock()

	if _, ok := utils.TransitionVolList[volumeID]; ok {
		return fmt.Errorf("Volume Busy, status: %v",
			utils.TransitionVolList[volumeID])
	}
	utils.TransitionVolList[volumeID] = status
	return nil
}
