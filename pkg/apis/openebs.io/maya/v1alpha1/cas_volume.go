/*
Copyright 2017 The OpenEBS Authors.
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

// CASVolumeKey is a typed string to represent
// CAS Volume related annotations' or labels' keys
type CASVolumeKey string

const (
	// CASTemplateKeyForVolumeCreate is the key to fetch name of CASTemplate
	// to create a CAS Volume
	CASTemplateKeyForVolumeCreate CASVolumeKey = "cas.openebs.io/create-volume-template"

	// CASTemplateKeyForVolumeRead is the key to fetch name of CASTemplate
	// to read a CAS Volume
	CASTemplateKeyForVolumeRead CASVolumeKey = "cas.openebs.io/read-volume-template"

	// CASTemplateKeyForVolumeDelete is the key to fetch name of CASTemplate
	// to delete a CAS Volume
	CASTemplateKeyForVolumeDelete CASVolumeKey = "cas.openebs.io/delete-volume-template"

	// CASTemplateKeyForVolumeList is the key to fetch name of CASTemplate
	// to list CAS Volumes
	CASTemplateKeyForVolumeList CASVolumeKey = "cas.openebs.io/list-volume-template"
)

// CASVolume represents a cas volume
type CASVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec i.e. specifications of this cas volume
	Spec CASVolumeSpec `json:"spec"`

	// CloneSpec contains the required information related to volume clone
	CloneSpec VolumeCloneSpec `json:"cloneSpec,omitempty"`

	// Status of this cas volume
	Status CASVolumeStatus `json:"status"`
}

// CASVolumeSpec has the properties of a cas volume
type CASVolumeSpec struct {
	// Capacity of a volume will hold the capacity of the Volume
	Capacity string `json:"capacity"`

	// Iqn of a volume will hold the iqn value of the Volume
	Iqn string `json:"iqn"`

	// TargetPortal of a volume will hold the target portal of the volume
	TargetPortal string `json:"targetPortal"`

	// TargetIP of a volume will hold the targetIP of the Volume
	TargetIP string `json:"targetIP"`

	// TargetPort of a volume will hold the targetIP of the Volume
	TargetPort string `json:"targetPort"`

	// Replicas of a volume will hold the replica count of the volume
	Replicas string `json:"replicas"`

	// CasType of a volume will hold the storage engine used to provision the volume
	CasType string `json:"casType"`

	// FSType of a volume will specify the format type - ext4(default), xfs of PV
	FSType string `json:"fsType"`

	// Lun of volume will specify the lun number 0, 1.. on iSCSI Volume. (default: 0)
	Lun int32 `json:"lun"`

	// AccessMode of a volume will hold the access mode of the volume
	AccessMode string `json:"accessMode"`
}

// CASVolumeStatus provides status of a cas volume
type CASVolumeStatus struct {
	// Phase indicates if a volume is available, pending or failed
	Phase VolumePhase

	// A human-readable message indicating details about why the volume
	// is in this state
	Message string

	// Reason is a brief CamelCase string that describes any failure and is meant
	// for machine parsing and tidy display in the CLI
	Reason string
}

// VolumeCloneSpec contains the required information which enable volume to cloned
type VolumeCloneSpec struct {
	// Defaults to false, true will enable the volume to be created as a clone
	IsClone bool `json:"isClone,omitempty"`

	// SourceVolume is snapshotted volume
	SourceVolume string `json:"sourceVolume,omitempty"`

	// SourceVolumeTargetIP is the source controller IP
	// which will be used to make a sync and rebuild
	// request from the new clone replica.
	SourceVolumeTargetIP string `json:"sourceTargetIP,omitempty"`

	// SnapshotName name of snapshot which is getting
	// promoted as persistent volume(this snapshot will
	// be cloned to new volume).
	SnapshotName string `json:"snapshotName,omitempty"`
}

// VolumePhase defines phase of a volume
type VolumePhase string

const (
	// VolumePending - used for Volumes that are not available
	VolumePending VolumePhase = "Pending"

	// VolumeAvailable - used for Volumes that are available
	VolumeAvailable VolumePhase = "Available"

	// VolumeFailed - used for Volumes that failed for some reason
	VolumeFailed VolumePhase = "Failed"
)

// CASVolumeList is a list of CASVolume resources
type CASVolumeList struct {
	metav1.ListOptions `json:",inline"`
	metav1.ObjectMeta  `json:"metadata,omitempty"`
	metav1.ListMeta    `json:"metalist"`

	// Items are the list of volumes
	Items []CASVolume `json:"items"`
}
