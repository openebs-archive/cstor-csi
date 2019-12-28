/*
Copyright © 2018-2019 The OpenEBS Authors

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=csivolume

// CSIVolume represents a CSI based volume
type CSIVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIVolumeSpec   `json:"spec"`
	Status CSIVolumeStatus `json:"status"`
}

// CSIVolumeSpec is the spec for a CStorVolume resource
type CSIVolumeSpec struct {
	// Volume specific info
	Volume VolumeInfo `json:"volume"`

	// ISCSIInfo specific to ISCSI protocol,
	// this is filled only if the volume type
	// is iSCSI
	ISCSI ISCSIInfo `json:"iscsi"`
}

// VolumeInfo contains the volume related info
// for all types of volumes in CSIVolumeSpec
type VolumeInfo struct {
	// Name of the CSI volume
	Name string `json:"name"`

	// Capacity of the volume
	Capacity string `json:"capacity,omitempty"`

	// TODO
	// Below fields might be moved to a separate
	// sub resource e.g. CSIVolumeContext

	// OwnerNodeID is the Node ID which
	// is also the owner of this Volume
	OwnerNodeID string `json:"ownerNodeID"`

	// FSType of a volume will specify the
	// format type - ext4(default), xfs of PV
	FSType string `json:"fsType,omitempty"`

	// AccessMode of a volume will hold the
	// access mode of the volume
	AccessModes []string `json:"accessModes,omitempty"`

	// AccessType of a volume will indicate if the volume will be used as a
	// block device or mounted on a path
	AccessType string `json:"accessType,omitempty"`

	// StagingPath of the volume will hold the
	// path on which the volume is mounted
	// on that node
	StagingTargetPath string `json:"stagingTargetPath,omitempty"`

	// TargetPath of the volume will hold the
	// path on which the volume is bind mounted
	// on that node
	TargetPath string `json:"targetPath,omitempty"`

	// ReadOnly specifies if the volume needs
	// to be mounted in ReadOnly mode
	ReadOnly bool `json:"readOnly,omitempty"`

	// MountOptions specifies the options with
	// which mount needs to be attempted
	MountOptions []string `json:"mountOptions,omitempty"`

	// Device Path specifies the device path
	// which is returned when the iSCSI
	// login is successful
	DevicePath string `json:"devicePath,omitempty"`
}

// ISCSIInfo has ISCSI protocol specific info,
// this can be used only if the volume type exposed
// by the vendor is iSCSI
type ISCSIInfo struct {
	// Iqn of this volume
	Iqn string `json:"iqn"`

	// TargetPortal holds the target portal
	// of this volume
	TargetPortal string `json:"targetPortal"`

	// IscsiInterface of this volume
	IscsiInterface string `json:"iscsiInterface"`

	// Lun specify the lun number 0, 1.. on
	// iSCSI Volume. (default: 0)
	Lun string `json:"lun"`
}

// CSIVolumeStatus status represents the current mount status of the volume
type CSIVolumeStatus string

// CSIVolumeStatusMounting indicated that a mount operation has been triggered
// on the volume and is under progress
const (
	// CSIVolumeStatusUninitialized indicates that no operation has been
	// performed on the volume yet on this node
	CSIVolumeStatusUninitialized CSIVolumeStatus = ""
	// CSIVolumeStatusMountUnderProgress indicates that the volume is busy and
	// unavailable for use by other goroutines, an iSCSI login followed by mount
	// is under progress on this volume
	CSIVolumeStatusMountUnderProgress CSIVolumeStatus = "MountUnderProgress"
	// CSIVolumeStatusMounteid indicated that the volume has been successfulled
	// mounted on the node
	CSIVolumeStatusMounted CSIVolumeStatus = "Mounted"
	// CSIVolumeStatusUnMounted indicated that the volume has been successfuly
	// unmounted and logged out of the node
	CSIVolumeStatusUnmounted CSIVolumeStatus = "Unmounted"
	// CSIVolumeStatusRaw indicates that the volume is being used in raw format
	// by the application, therefore CSI has only performed iSCSI login
	// operation on this volume and avoided filesystem creation and mount.
	CSIVolumeStatusRaw CSIVolumeStatus = "Raw"
	// CSIVolumeStatusResizeInProgress indicates that the volume is being
	// resized
	CSIVolumeStatusResizeInProgress CSIVolumeStatus = "ResizeInProgress"
	// CSIVolumeStatusMountFailed indicates that login and mount process from
	// the volume has bben started but failed kubernetes needs to retry sending
	// nodepublish
	CSIVolumeStatusMountFailed CSIVolumeStatus = "MountFailed"
	// CSIVolumeStatusUnmountInProgress indicates that the volume is busy and
	// unavailable for use by other goroutines, an unmount operation on volume
	// is under progress
	CSIVolumeStatusUnmountUnderProgress CSIVolumeStatus = "UnmountUnderProgress"
	// CSIVolumeStatusWaitingForCVCBound indicates that the volume components
	// are still being created
	CSIVolumeStatusWaitingForCVCBound CSIVolumeStatus = "WaitingForCVCBound"
	// CSIVolumeStatusWaitingForVolumeToBeReady indicates that the replicas are
	// yet to connect to target
	CSIVolumeStatusWaitingForVolumeToBeReady CSIVolumeStatus = "WaitingForVolumeToBeReady"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=csivolumes

// CSIVolumeList is a list of CSIVolume resources
type CSIVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CSIVolume `json:"items"`
}
