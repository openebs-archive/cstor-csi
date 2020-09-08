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

package driver

import (
	"fmt"

	apisv1 "github.com/openebs/api/pkg/apis/cstor/v1"
	apis "github.com/openebs/cstor-csi/pkg/apis/cstor/v1"
	utils "github.com/openebs/cstor-csi/pkg/utils"
	corev1 "k8s.io/api/core/v1"
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

	// TopologyNodeKey is a key of topology that represents node name.
	TopologyNodeKey = "topology.cstor.openebs.io/nodeName"

	// pvcNameKey holds the name of the PVC which is passed as a parameter
	// in CreateVolume request
	pvcNameKey = "csi.storage.k8s.io/pvc/name"

	// pvcNamespaceKey holds the namespace of the PVC which is passed parameter
	// in CreateVolume request
	pvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"

	// pvNameKey holds the name of the PV which is passed as a parameter
	// in CreateVolume request
	pvNameKey = "csi.storage.k8s.io/pv/name"
)

var (
	ValidFSTypes = []string{FSTypeExt4, FSTypeXfs}
)

func isValidFStype(fstype string) bool {
	for _, fs := range ValidFSTypes {
		if fs == fstype {
			return true
		}
	}
	return false
}

func removeVolumeFromTransitionList(volumeID string) {
	utils.TransitionVolListLock.Lock()
	defer utils.TransitionVolListLock.Unlock()
	delete(utils.TransitionVolList, volumeID)
}

func addVolumeToTransitionList(volumeID string, status apis.CStorVolumeAttachmentStatus) error {
	utils.TransitionVolListLock.Lock()
	defer utils.TransitionVolListLock.Unlock()

	if _, ok := utils.TransitionVolList[volumeID]; ok {
		return fmt.Errorf("Volume Busy, status: %v",
			utils.TransitionVolList[volumeID])
	}
	utils.TransitionVolList[volumeID] = status
	return nil
}

// getCapacity converts capacity as string
func getCapacity(cvc *apisv1.CStorVolumeConfig) string {
	qCap := cvc.Spec.Capacity[corev1.ResourceStorage]
	return qCap.String()
}
