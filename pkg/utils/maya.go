package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	apisv1 "github.com/openebs/api/pkg/apis/cstor/v1"
	apis "github.com/openebs/cstor-csi/pkg/apis/cstor/v1"
	cv "github.com/openebs/cstor-csi/pkg/cstor/volume"
	csivol "github.com/openebs/cstor-csi/pkg/cstor/volumeattachment"
	cvc "github.com/openebs/cstor-csi/pkg/cstor/volumeconfig"
	"github.com/openebs/cstor-csi/pkg/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
	// OpenebsVolumePolicy is the config policy name passed to CSI from the
	// storage class parameters
	OpenebsVolumePolicy = "openebs.io/volume-policy"
	// OpenebsVolumeID is the PV name passed to CSI
	OpenebsVolumeID = "openebs.io/volumeID"
	// OpenebsCSPCName is the name of cstor storagepool cluster
	OpenebsCSPCName = "openebs.io/cstor-pool-cluster"
	// CVCFinalizer is used for CVC protection so that cvc is not deleted until
	// the underlying cv is deleted
	CVCFinalizer = "cvc.openebs.io/finalizer"
	// TargetLunID indicates the LUN ID at the target
	TargetLunID = "0"
	// DefaultIscsiInterface can be used when there is no specific
	// IscsiInterface set
	DefaultIscsiInterface = "default"
)

// ProvisionVolume creates a CstorVolumeClaim(cvc) CR,
// watcher for cvc is present in maya-apiserver
func ProvisionVolume(
	size int64,
	volName,
	replicaCount,
	cspcName,
	snapshotID,
	nodeID,
	policyName string,
) error {

	annotations := map[string]string{
		OpenebsVolumeID:     volName,
		OpenebsVolumePolicy: policyName,
	}

	labels := map[string]string{
		OpenebsCSPCName: cspcName,
	}

	if snapshotID != "" {
		srcVolName, _, _ := GetVolumeSourceDetails(snapshotID)
		labels["openebs.io/source-volume"] = srcVolName
	}

	finalizers := []string{
		CVCFinalizer,
	}

	requestGIB := RoundUpGiB(size)
	sSize := resource.MustParse(fmt.Sprintf("%dGi", requestGIB))

	cvcObj, err := cvc.NewBuilder().
		WithName(volName).
		WithNamespace(OpenEBSNamespace).
		WithAnnotations(annotations).
		WithLabelsNew(labels).
		WithFinalizers(finalizers).
		WithCapacityQty(sSize).
		WithSource(snapshotID).
		WithNodeID(nodeID).
		WithReplicaCount(replicaCount).
		WithProvisionCapacityQty(sSize).
		WithNewVersion(version.Current()).
		WithDependentsUpgraded().
		WithStatusPhase(apisv1.CStorVolumeConfigPhasePending).Build()
	if err != nil {
		return err
	}

	_, err = cvc.NewKubeclient().WithNamespace(OpenEBSNamespace).Create(cvcObj)
	return err
}

// GetVolume the corresponding CstorVolumeClaim(cvc) CR
func GetVolume(volumeID string) (*apisv1.CStorVolumeConfig, error) {
	return cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(volumeID, metav1.GetOptions{})
}

// IsSourceAvailable returns true if the source volume is available
func IsSourceAvailable(snapshotID string) (bool, error) {
	srcVolName, _, err := GetVolumeSourceDetails(snapshotID)
	if err != nil {
		return false, err
	}
	cvc, err := GetVolume(srcVolName)
	if cvc != nil {
		return true, nil
	}
	return false, err
}

// DeleteVolume deletes the corresponding CstorVolumeClaim(cvc) CR
func DeleteVolume(volumeID string) (err error) {
	err = cvc.NewKubeclient().WithNamespace(OpenEBSNamespace).Delete(volumeID)
	return
}

// IsCVCBound returns if the CV is bound to CVC or not
func IsCVCBound(volumeID string) (bool, error) {
	cvcObj, err := cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(volumeID, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if cvcObj.Status.Phase == apisv1.CStorVolumeConfigPhasePending {
		return false, nil
	}
	return true, nil
}

//PatchCVCNodeID patches the NodeID of CVC
func PatchCVCNodeID(volumeID, nodeID string) error {
	oldCVCObj, err := cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(volumeID, metav1.GetOptions{})
	if err != nil {
		return err
	}

	newCVCObj, err := cvc.BuildFrom(oldCVCObj.DeepCopy()).
		WithNodeID(nodeID).Build()
	_, err = cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Patch(oldCVCObj, newCVCObj)

	return err
}

// GetVolumeSourceDetails splits the volumeName and snapshot
func GetVolumeSourceDetails(snapshotID string) (string, string, error) {
	volSrc := strings.Split(snapshotID, "@")
	if len(volSrc) == 0 {
		return "", "", errors.New(
			"failed to get volumeSource",
		)
	}
	return volSrc[0], volSrc[1], nil
}

//FetchAndUpdateISCSIDetails fetches the iSCSI details from cstor volume
//resource and updates the corresponding csivolume resource
func FetchAndUpdateISCSIDetails(volumeID string, vol *apis.CStorVolumeAttachment) error {
	getOptions := metav1.GetOptions{}
	cstorVolume, err := cv.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(volumeID, getOptions)
	if err != nil {
		return err
	}
	_, err = csivol.BuildFrom(vol).
		WithIQN(cstorVolume.Spec.Iqn).
		WithTargetPortal(cstorVolume.Spec.TargetPortal).
		WithLun(TargetLunID).
		WithIscsiInterface(DefaultIscsiInterface).
		Build()
	return err
}

// ResizeVolume updates the CstorVolumeClaim(cvc) CR,
// watcher for cvc is present in maya-apiserver
func ResizeVolume(
	volumeID string,
	size int64,
) error {

	requestGIB := RoundUpGiB(size)
	desiredSize := resource.MustParse(fmt.Sprintf("%dGi", requestGIB))

	cvc, err := getCVC(volumeID)
	if err != nil {
		return err
	}

	cvcDesiredSize := cvc.Spec.Capacity[corev1.ResourceStorage]

	if (desiredSize).Cmp(cvcDesiredSize) < 0 {
		return fmt.Errorf("Volume shrink not supported from: %v to: %v",
			cvc.Status.Capacity, cvc.Spec.Capacity)
	}

	if cvc.Status.Phase == apisv1.CStorVolumeConfigPhasePending {
		return handleResize(cvc, desiredSize)
	}
	cvcActualSize := cvc.Status.Capacity[corev1.ResourceStorage]

	if cvcDesiredSize.Cmp(cvcActualSize) > 0 {
		return fmt.Errorf("ResizeInProgress from: %v to: %v",
			cvcActualSize, cvcDesiredSize)
	}

	if (desiredSize).Cmp(cvcActualSize) == 0 {
		return nil
	}
	return handleResize(cvc, desiredSize)

}

func handleResize(
	cvc *apisv1.CStorVolumeConfig, sSize resource.Quantity,
) error {
	if err := updateCVCSize(cvc, sSize); err != nil {
		return err
	}
	if cvc.Publish.NodeID == "" {
		return nil
	}
	return waitAndReverifyResizeStatus(cvc.Name, sSize)
}

func waitAndReverifyResizeStatus(cvcName string, sSize resource.Quantity) error {

	time.Sleep(5 * time.Second)
	cvcObj, err := getCVC(cvcName)
	if err != nil {
		return err
	}
	desiredSize := sSize
	cvcActualSize := cvcObj.Status.Capacity[corev1.ResourceStorage]
	if (desiredSize).Cmp(cvcActualSize) != 0 {
		return fmt.Errorf("ResizeInProgress from: %v to: %v",
			cvcActualSize, desiredSize)
	}
	return nil
}

func updateCVCSize(oldCVCObj *apisv1.CStorVolumeConfig, sSize resource.Quantity) error {
	newCVCObj, err := cvc.BuildFrom(oldCVCObj.DeepCopy()).
		WithCapacityQty(sSize).Build()
	if err != nil {
		return err
	}
	_, err = cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Patch(oldCVCObj, newCVCObj)
	return err
}

func getCVC(cvcName string) (*apisv1.CStorVolumeConfig, error) {
	return cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(cvcName, metav1.GetOptions{})
}
