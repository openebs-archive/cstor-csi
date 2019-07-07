package utils

import (
	"strconv"

	apis "github.com/openebs/csi/pkg/apis/openebs.io/core/v1alpha1"
	apismaya "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
	cv "github.com/openebs/csi/pkg/cstor/volume/v1alpha1"
	cvc "github.com/openebs/csi/pkg/cvc/v1alpha1"
	csivol "github.com/openebs/csi/pkg/volume/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
	// OpenebsConfigClass is the config class name passed to CSI from the
	// storage class parameters
	OpenebsConfigClass = "openebs.io/config-class"
	// OpenebsVolumeID is the PV name passed to CSI
	OpenebsVolumeID = "openebs.io/volumeID"
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
func ProvisionVolume(size int64, volName, configclass string) error {

	annotations := map[string]string{
		OpenebsConfigClass: configclass,
		OpenebsVolumeID:    volName,
	}

	finalizers := []string{
		CVCFinalizer,
	}

	sSize := strconv.FormatInt(size, 10)
	cvcObj, err := cvc.NewBuilder().
		WithName(volName).
		WithNamespace(OpenEBSNamespace).
		WithAnnotations(annotations).
		WithFinalizers(finalizers).
		WithCapacity(sSize).
		WithStatusPhase(apismaya.CStorVolumeClaimPhasePending).Build()
	if err != nil {
		return err
	}

	_, err = cvc.NewKubeclient().WithNamespace(OpenEBSNamespace).Create(cvcObj)
	return err
}

// GetVolume the corresponding CstorVolumeClaim(cvc) CR
func GetVolume(volumeID string) (*apismaya.CStorVolumeClaim, error) {
	return cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(volumeID, metav1.GetOptions{})
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
	if cvcObj.Status.Phase == apismaya.CStorVolumeClaimPhasePending {
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

//FetchAndUpdateISCSIDetails fetches the iSCSI details from cstor volume
//resource and updates the corresponding csivolume resource
func FetchAndUpdateISCSIDetails(volumeID string, vol *apis.CSIVolume) error {
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
