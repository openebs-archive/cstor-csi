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
	apis "github.com/openebs/cstor-csi/pkg/apis/openebs.io/core/v1alpha1"
	csv "github.com/openebs/cstor-csi/pkg/maya/cstorvolume/v1alpha1"
	errors "github.com/openebs/cstor-csi/pkg/maya/errors/v1alpha1"
	node "github.com/openebs/cstor-csi/pkg/maya/kubernetes/node/v1alpha1"
	pv "github.com/openebs/cstor-csi/pkg/maya/kubernetes/persistentvolume/v1alpha1"
	csivolume "github.com/openebs/cstor-csi/pkg/volume/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TODO csi.openebs.io/nodeid

	// NODEID is the node name on which this pod is currently scheduled
	NODEID = "nodeID"
	// TODO csi.openebs.io/nodeid

	// VOLNAME is the name of the provisioned volume
	VOLNAME = "Volname"
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
	listOptions := metav1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + volumeID,
	}

	volumeList, err := csv.NewKubeclient().
		WithNamespace(OpenEBSNamespace).List(listOptions)
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

// GetVolListForNode fetches the current Published Volume list
func GetVolListForNode() (*apis.CSIVolumeList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: NODEID + "=" + NodeIDENV,
	}

	return csivolume.NewKubeclient().
		WithNamespace(OpenEBSNamespace).List(listOptions)

}

// GetVolList fetches the current Published Volume list
func GetVolList(volume string) (*apis.CSIVolumeList, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: VOLNAME + "=" + volume,
	}
	return csivolume.NewKubeclient().
		WithNamespace(OpenEBSNamespace).List(listOptions)

}

// GetCSIVolume fetches the current Published csi Volume
func GetCSIVolume(csivol string) (*apis.CSIVolume, error) {
	return csivolume.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Get(csivol, metav1.GetOptions{})
}

// GetVolumeIP fetches the cstor target IP Address
func GetVolumeIP(volumeID string) (string, error) {
	cstorvolume, err := csv.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Get(volumeID, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if cstorvolume == nil {
		return "", nil
	}
	return cstorvolume.Spec.TargetIP, nil
}

// CreateCSIVolumeCR creates a CSI VOlume CR
func CreateCSIVolumeCR(csivol *apis.CSIVolume, nodeID string) error {
	csivol.Spec.Volume.OwnerNodeID = nodeID
	nodeInfo, err := getNodeDetails(nodeID)
	if err != nil {
		return err
	}

	csivol.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       nodeInfo.Name,
			UID:        nodeInfo.UID,
		},
	}
	_, err = csivolume.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Create(csivol)
	return err

}

// UpdateCSIVolumeCR updates CSIVolume CR related to current nodeID
func UpdateCSIVolumeCR(csivol *apis.CSIVolume) error {

	_, err := csivolume.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Update(csivol)
	return err
}

// TODO Explain when a create of csi volume happens & when it
// gets deleted or replaced or updated

// DeleteOldCSIVolumeCRs removes the CSIVolumeCR for the specified path
func DeleteOldCSIVolumeCRs(volumeID string) error {
	csivols, err := GetVolList(volumeID)
	if err != nil {
		return err
	}

	for _, csivol := range csivols.Items {
		err = csivolume.NewKubeclient().
			WithNamespace(OpenEBSNamespace).Delete(csivol.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO Explain when a create of csi volume happens & when it
// gets deleted or replaced or updated

// DeleteCSIVolumeCR removes the CSIVolumeCR for the specified path
func DeleteCSIVolumeCR(csivolName string) error {
	return csivolume.NewKubeclient().
		WithNamespace(OpenEBSNamespace).Delete(csivolName)
}
