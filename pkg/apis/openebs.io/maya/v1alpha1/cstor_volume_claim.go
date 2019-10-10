/*
Copyright 2019 The OpenEBS Authors.
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// CStorVolumeClaim describes a cstor volume claim resource created as
// custom resource. CStorVolumeClaim is a request for creating cstor volume
// related resources like deployment, svc etc.
type CStorVolumeClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines a specification of a cstor volume claim required
	// to provision cstor volume resources
	Spec CStorVolumeClaimSpec `json:"spec"`

	// Publish contains info related to attachment of a volume to a node.
	// i.e. NodeId etc.
	Publish CStorVolumeClaimPublish `json:"publish,omitempty"`

	VersionDetails VersionDetails `json:"versionDetails"`
	// Status represents the current information/status for the cstor volume
	// claim, populated by the controller.
	Status CStorVolumeClaimStatus `json:"status"`
}

// CStorVolumeClaimSpec is the spec for a CStorVolumeClaim resource
type CStorVolumeClaimSpec struct {
	// Capacity represents the actual resources of the underlying
	// cstor volume.
	Capacity corev1.ResourceList `json:"capacity"`
	// ReplicaCount represents the actual replica count for the underlying
	// cstor volume
	ReplicaCount int `json:"replicaCount"`
	// CStorVolumeRef contains the reference to CStorVolume i.e. CstorVolume Name
	// This field will be updated by maya after cstor Volume has been
	// provisioned
	CStorVolumeRef *corev1.ObjectReference `json:"cstorVolumeRef,omitempty"`
}

// CStorVolumeClaimPublish will contain the publish details of the volume.
// eg. NodeId indicates where the volume is currently published.
// More fields will be added as and when required
type CStorVolumeClaimPublish struct {
	// NodeId indicates where the volume is needed to be mounted, i.e the node
	// where the app is scheduled
	NodeID string `json:"nodeId,omitempty"`
}

// CStorVolumeClaimPhase represents the current phase of CStorVolumeClaim.
type CStorVolumeClaimPhase string

const (
	//CStorVolumeClaimPhasePending indicates that the cvc is still waiting for
	//the cstorvolume to be created and bound
	CStorVolumeClaimPhasePending CStorVolumeClaimPhase = "Pending"

	//CStorVolumeClaimPhaseBound indiacates that the cstorvolume has been
	//provisioned and bound to the cstor volume claim
	CStorVolumeClaimPhaseBound CStorVolumeClaimPhase = "Bound"

	//CStorVolumeClaimPhaseFailed indiacates that the cstorvolume provisioning
	//has failed
	CStorVolumeClaimPhaseFailed CStorVolumeClaimPhase = "Failed"
)

// CStorVolumeClaimStatus is for handling status of CstorVolume Claim.
// defines the observed state of CStorVolumeClaim
type CStorVolumeClaimStatus struct {
	// Phase represents the current phase of CStorVolumeClaim.
	Phase CStorVolumeClaimPhase `json:"phase"`
	// Capacity the actual resources of the underlying volume.
	Capacity   corev1.ResourceList         `json:"capacity,omitempty"`
	Conditions []CStorVolumeClaimCondition `json:"condition,omitempty"`
}

// CStorVolumeClaimCondition contains details about state of cstor volume
type CStorVolumeClaimCondition struct {
	// Current Condition of cstor volume claim.
	// If underlying persistent volume is being resized
	// then the Condition will be set to 'ResizeStarted' etc
	Type CStorVolumeClaimConditionType `json:"type"`

	// Reason is a brief CamelCase string that describes any failure
	Reason string `json:"reason"`

	// Human-readable message indicating details about last transition.
	Message string `json:"message"`
}

// CStorVolumeClaimConditionType can indiacte the current condition of
// CstorVOlumeClaim depending on the operation being performed on Volume.
type CStorVolumeClaimConditionType string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// CStorVolumeClaimList is a list of CStorVolumeClaim resources
type CStorVolumeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CStorVolumeClaim `json:"items"`
}
