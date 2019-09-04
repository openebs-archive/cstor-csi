package v1alpha1

import (
	"encoding/json"
	"fmt"

	apismaya "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// JVCKey returns an unique key of a JVC object,
func JVCKey(cvc *apismaya.JivaVolumeClaim) string {
	return fmt.Sprintf("%s/%s", cvc.Namespace, cvc.Name)
}

func getPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("mashal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}
