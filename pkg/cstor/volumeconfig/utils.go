package volumeconfig

import (
	"encoding/json"
	"fmt"

	apisv1 "github.com/openebs/api/pkg/apis/cstor/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// CVCKey returns an unique key of a CVC object,
func CVCKey(cvc *apisv1.CStorVolumeConfig) string {
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
