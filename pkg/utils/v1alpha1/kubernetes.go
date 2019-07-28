// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	apis "github.com/openebs/csi/pkg/apis/openebs.io/core/v1alpha1"
	csv "github.com/openebs/csi/pkg/maya/cstorvolume/v1alpha1"
	errors "github.com/openebs/csi/pkg/maya/errors/v1alpha1"
	node "github.com/openebs/csi/pkg/maya/kubernetes/node/v1alpha1"
	pv "github.com/openebs/csi/pkg/maya/kubernetes/persistentvolume/v1alpha1"
	csivolume "github.com/openebs/csi/pkg/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getNodeDetails fetches the nodeInfo for the current node
func getNodeDetails(name string) (*corev1.Node, error) {
	return node.NewKubeClient().Get(name, metav1.GetOptions{})
}

// FetchPVDetails gets the PV related to this VolumeID
func FetchPVDetails(name string) (*corev1.PersistentVolume, error) {
	return pv.NewKubeClient().Get(name, metav1.GetOptions{})
}

// getVolStatus fetches the current VolumeStatus which specifies if the volume
// is ready to serve IOs
func getVolStatus(volumeID string) (string, error) {
	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + volumeID,
	}

	volumeList, err := csv.NewKubeclient().WithNamespace(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return "", err
	}

	if len(volumeList.Items) != 1 {
		return "", errors.Errorf(
			"expected single volume got {%d} for selector {%v}",
			len(volumeList.Items),
			listOptions,
		)
	}

	return string(volumeList.Items[0].Status.Phase), nil
}

// CreateCSIVolumeCR creates a new CSIVolume CR with this nodeID
func CreateCSIVolumeCR(csivol *apis.CSIVolume, nodeID, mountPath string) (err error) {

	csivol.Name = csivol.Spec.Volume.Name + "-" + nodeID
	csivol.Labels = make(map[string]string)
	csivol.Spec.Volume.OwnerNodeID = nodeID
	csivol.Labels["Volname"] = csivol.Spec.Volume.Name
	csivol.Labels["nodeID"] = nodeID
	nodeInfo, err := getNodeDetails(nodeID)
	if err != nil {
		return
	}

	csivol.OwnerReferences = []v1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       nodeInfo.Name,
			UID:        nodeInfo.UID,
		},
	}
	csivol.Finalizers = []string{nodeID}

	_, err = csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).Create(csivol)
	return
}

// UpdateCSIVolumeCR updates CSIVolume CR related to current nodeID
func UpdateCSIVolumeCR(csivol *apis.CSIVolume) error {

	oldcsivol, err := csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).Get(csivol.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	oldcsivol.Spec.Volume.DevicePath = csivol.Spec.Volume.DevicePath
	oldcsivol.Status = csivol.Status

	_, err = csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(oldcsivol)
	return err
}

// DeleteOldCSIVolumeCR deletes all CSIVolumes
// related to this volume so that a new one
// can be created with node as current nodeID
func DeleteOldCSIVolumeCR(vol *apis.CSIVolume, nodeID string) (err error) {
	listOptions := v1.ListOptions{
		// TODO Update this label selector name as per naming standards
		LabelSelector: "Volname=" + vol.Name,
	}

	csivols, err := csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return
	}

	// If a node goes down and kubernetes is unable to send an Unpublish request
	// to this node, the CR is marked for deletion but finalizer is not removed
	// and a new CR is created for current node. When the degraded node comes up
	// it removes the finalizer and the CR is deleted.
	for _, csivol := range csivols.Items {
		if csivol.Labels["nodeID"] == nodeID {
			csivol.Finalizers = nil
			_, err = csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(&csivol)
			if err != nil {
				return
			}
		}

		err = csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).Delete(csivol.Name)
		if err != nil {
			return
		}
	}
	return
}

// TODO Explain when a create of csi volume happens & when it
// gets deleted or replaced or updated
//
// DeleteCSIVolumeCR removes the CSIVolume with this nodeID as
// labelSelector from the list
func DeleteCSIVolumeCR(vol *apis.CSIVolume) (err error) {
	var csivols *apis.CSIVolumeList
	listOptions := v1.ListOptions{
		// TODO use label as per standards
		LabelSelector: "Volname=" + vol.Spec.Volume.Name,
	}

	csivols, err = csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return
	}

	for _, csivol := range csivols.Items {
		if csivol.Spec.Volume.OwnerNodeID == vol.Spec.Volume.OwnerNodeID {
			csivol.Finalizers = nil
			_, err = csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).Update(&csivol)
			if err != nil {
				return
			}

			err = csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).Delete(csivol.Name)
			if err != nil {
				return
			}
		}
	}
	return
}

// FetchAndUpdateVolInfos gets the list of CSIVolInfos
// that are supposed to be mounted on this node and
// stores the info in memory. This is required when the
// CSI driver gets restarted & hence start monitoring all
// the existing volumes and at the same time use this
// logic to reject duplicate volume creation requests
func FetchAndUpdateVolInfos(nodeID string) (err error) {
	var listOptions v1.ListOptions

	if nodeID != "" {
		listOptions = v1.ListOptions{
			LabelSelector: "nodeID=" + nodeID,
		}
	}

	csivols, err := csivolume.NewKubeclient().WithNamespace(OpenEBSNamespace).List(listOptions)
	if err != nil {
		return
	}

	for _, csivol := range csivols.Items {
		if !csivol.DeletionTimestamp.IsZero() {
			continue
		}
		vol := csivol
		if vol.Status == apis.CSIVolumeStatusMountUnderProgress {
			vol.Status = apis.CSIVolumeStatusUninitialized
		}
		Volumes[csivol.Spec.Volume.Name] = &vol
	}

	return
}
