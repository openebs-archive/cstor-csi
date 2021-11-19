// Copyright © 2020 The OpenEBS Authors
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

package volumeattachment

import (
	apis "github.com/openebs/api/v3/pkg/apis/cstor/v1"
)

// CStorVolumeAttachment is a wrapper over
// CStorVolumeAttachment API instance
type CStorVolumeAttachment struct {
	Object *apis.CStorVolumeAttachment
}

// From returns a new instance of
// csi volume
func From(vol *apis.CStorVolumeAttachment) *CStorVolumeAttachment {
	return &CStorVolumeAttachment{
		Object: vol,
	}
}

// FromCASVolume returns a new instance
// of csi volume from given cas volume
// instance
//func FromCASVolume(vol *apis.CASVolume) *CStorVolumeAttachment {
//	return &CStorVolumeAttachment{
//		Object: &apis.CStorVolumeAttachment{
//			Spec: apis.CStorVolumeAttachmentSpec{
//				Volume: apis.VolumeInfo{
//					Name:     vol.Name,
//					Capacity: vol.Spec.Capacity,
//				},
//				ISCSI: apis.ISCSIInfo{
//					Iqn:          vol.Spec.Iqn,
//					TargetPortal: vol.Spec.TargetPortal,
//					Lun:          strconv.FormatInt(int64(vol.Spec.Lun), 10),
//				},
//			},
//		},
//	}
//}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type Predicate func(*CStorVolumeAttachment) bool

// PredicateList holds a list of predicate
type predicateList []Predicate

// CStorVolumeAttachmentList holds the list
// of csi volume instances
type CStorVolumeAttachmentList struct {
	List apis.CStorVolumeAttachmentList
}

// Len returns the number of items present
// in the CStorVolumeAttachmentList
func (p *CStorVolumeAttachmentList) Len() int {
	return len(p.List.Items)
}

// all returns true if all the predicates
// succeed against the provided CStorVolumeAttachment
// instance
func (l predicateList) all(p *CStorVolumeAttachment) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// HasLabels returns true if provided labels
// are present in the provided CStorVolumeAttachment instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(p *CStorVolumeAttachment) bool {
		for key, value := range keyValuePair {
			if !p.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel returns true if provided label
// is present in the provided CStorVolumeAttachment instance
func (p *CStorVolumeAttachment) HasLabel(key, value string) bool {
	val, ok := p.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel returns true if provided label
// is present in the provided CStorVolumeAttachment instance
func HasLabel(key, value string) Predicate {
	return func(p *CStorVolumeAttachment) bool {
		return p.HasLabel(key, value)
	}
}

// IsNil returns true if the csi volume instance
// is nil
func (p *CStorVolumeAttachment) IsNil() bool {
	return p.Object == nil
}

// IsNil is predicate to filter out nil csi volume
// instances
func IsNil() Predicate {
	return func(p *CStorVolumeAttachment) bool {
		return p.IsNil()
	}
}

// GetAPIObject returns csi volume's API instance
func (p *CStorVolumeAttachment) GetAPIObject() *apis.CStorVolumeAttachment {
	return p.Object
}
