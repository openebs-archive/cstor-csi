package utils

import (
	"fmt"
	"time"

	apis "github.com/openebs/cstor-csi/pkg/apis/openebs.io/core/v1alpha1"
	apismaya "github.com/openebs/cstor-csi/pkg/apis/openebs.io/maya/v1alpha1"
	cv "github.com/openebs/cstor-csi/pkg/cstor/volume/v1alpha1"
	cvc "github.com/openebs/cstor-csi/pkg/cvc/v1alpha1"
	"github.com/openebs/cstor-csi/pkg/version"
	csivol "github.com/openebs/cstor-csi/pkg/volume/v1alpha1"

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
	// OpenebsConfigClass is the config class name passed to CSI from the
	// storage class parameters
	OpenebsConfigClass = "openebs.io/config-class"
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
	cspcName string,
) error {

	annotations := map[string]string{
		OpenebsVolumeID: volName,
	}

	labels := map[string]string{
		OpenebsCSPCName: cspcName,
	}

	finalizers := []string{
		CVCFinalizer,
	}

	sSize := ByteCount(uint64(size))
	cvcObj, err := cvc.NewBuilder().
		WithName(volName).
		WithNamespace(OpenEBSNamespace).
		WithAnnotations(annotations).
		WithLabelsNew(labels).
		WithFinalizers(finalizers).
		WithCapacity(sSize).
		WithReplicaCount(replicaCount).
		WithNewVersion(version.Current()).
		WithDependentsUpgraded().
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

// ResizeVolume updates the CstorVolumeClaim(cvc) CR,
// watcher for cvc is present in maya-apiserver
func ResizeVolume(
	volumeID string,
	size int64,
) error {

	sSize := ByteCount(uint64(size))
	cvc, err := getCVC(volumeID)
	if err != nil {
		return err
	}
	desiredSize, _ := resource.ParseQuantity(sSize)
	cvcDesiredSize := cvc.Spec.Capacity[corev1.ResourceStorage]

	if (desiredSize).Cmp(cvcDesiredSize) < 0 {
		return fmt.Errorf("Volume shrink not supported from: %v to: %v",
			cvc.Status.Capacity, cvc.Spec.Capacity)
	}

	if cvc.Status.Phase == apismaya.CStorVolumeClaimPhasePending {
		return handleResize(cvc, sSize)
	}
	cvcActualSize := cvc.Status.Capacity[corev1.ResourceStorage]

	if cvcDesiredSize.Cmp(cvcActualSize) > 0 {
		return fmt.Errorf("ResizeInProgress from: %v to: %v",
			cvcActualSize, cvcDesiredSize)
	}

	if (desiredSize).Cmp(cvcActualSize) == 0 {
		return nil
	}
	return handleResize(cvc, sSize)

}

func handleResize(
	cvc *apismaya.CStorVolumeClaim, sSize string,
) error {
	if err := updateCVCSize(cvc, sSize); err != nil {
		return err
	}
	if cvc.Publish.NodeID == "" {
		return nil
	}
	return waitAndReverifyResizeStatus(cvc.Name, sSize)
}

func waitAndReverifyResizeStatus(cvcName, sSize string) error {

	time.Sleep(5 * time.Second)
	cvcObj, err := getCVC(cvcName)
	if err != nil {
		return err
	}
	desiredSize, _ := resource.ParseQuantity(sSize)
	cvcActualSize := cvcObj.Status.Capacity[corev1.ResourceStorage]
	if (desiredSize).Cmp(cvcActualSize) != 0 {
		return fmt.Errorf("ResizeInProgress from: %v to: %v",
			cvcActualSize, desiredSize)
	}
	return nil
}
func updateCVCSize(oldCVCObj *apismaya.CStorVolumeClaim, sSize string) error {
	newCVCObj, err := cvc.BuildFrom(oldCVCObj.DeepCopy()).
		WithCapacity(sSize).Build()
	if err != nil {
		return err
	}
	_, err = cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Patch(oldCVCObj, newCVCObj)
	return err
}

func getCVC(cvcName string) (*apismaya.CStorVolumeClaim, error) {
	return cvc.NewKubeclient().
		WithNamespace(OpenEBSNamespace).
		Get(cvcName, metav1.GetOptions{})
}
