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
// +resource:path=jivavolumeclaim

// JivaVolumeClaim describes a jiva volume claim resource created as
// custom resource. JivaVolumeClaim is a request for creating jiva volume
// related resources like deployment, svc etc.
type JivaVolumeClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines a specification of a jiva volume claim required
	// to provision jiva volume resources
	Spec JivaVolumeClaimSpec `json:"spec"`

	// NodeId indicates where the volume is needed to be mounted, i.e the node
	// where the app is scheduled
	NodeID string `json:"nodeId,omitempty"`

	// Status represents the current information/status for the jiva volume
	// claim, populated by the controller.
	Status JivaVolumeClaimStatus `json:"status"`
}

// JivaVolumeClaimSpec is the spec for a JivaVolumeClaim resource
type JivaVolumeClaimSpec struct {
	// Capacity represents the actual resources of the underlying
	// jiva volume.
	Capacity int64 `json:"capacity"`
	// ReplicaCount represents the actual replica count for the underlying
	// jiva volume
	ReplicaCount int `json:"replicaCount"`
	// JivaVolumeRef contains the reference to JIVAVolume i.e. JIVAVolume Name
	// This field will be updated by maya after jiva Volume has been
	// provisioned
	JivaVolumeRef *corev1.ObjectReference `json:"jivaVolumeRef,omitempty"`
	TargetIP      string                  `json:"targetIP"`
	TargetPort    string                  `json:"targetPort"`
	Iqn           string                  `json:"iqn"`
	TargetPortal  string                  `json:"targetPortal"`
	Status        string                  `json:"status"`
}

// JivaVolumeClaimPhase represents the current phase of CStorVolumeClaim.
type JivaVolumeClaimPhase string

const (
	//JivaVolumeClaimPhasePending indicates that the jvc is still waiting for
	//the jivavolume to be created and bound
	JivaVolumeClaimPhasePending JivaVolumeClaimPhase = "Pending"

	//JivaVolumeClaimPhaseBound indiacates that the jivavolume has been
	//provisioned and bound to the jiva volume claim
	JivaVolumeClaimPhaseBound JivaVolumeClaimPhase = "Bound"

	//JivaVolumeClaimPhaseFailed indiacates that the jivavolume provisioning
	//has failed
	JivaVolumeClaimPhaseFailed JivaVolumeClaimPhase = "Failed"
)

// JivaVolumeClaimStatus is for handling status of jiva volume claim.
// defines the observed state of JivaVolumeClaim
type JivaVolumeClaimStatus struct {
	Name            string      `json:"name"`
	Status          string      `json:"status"`
	ReplicaStatuses []RepStatus `json:"replicaStatus"`
	// Phase represents the current phase of JivaVolumeClaim.
	Phase JivaVolumeClaimPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// JivaVolumeClaimList is a list of JivaVolumeClaim resources
type JivaVolumeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []JivaVolumeClaim `json:"items"`
}

// RepStatus stores the status of replicas
type RepStatus struct {
	Address string `json:"address"`
	Mode    string `json:"mode"`
}
