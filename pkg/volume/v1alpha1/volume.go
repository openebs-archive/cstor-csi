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

package v1alpha1

import (
	"strconv"

	apis "github.com/openebs/csi/pkg/apis/openebs.io/core/v1alpha1"
	apismaya "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
)

// CSIVolume is a wrapper over
// CSIVolume API instance
type CSIVolume struct {
	Object *apis.CSIVolume
}

// From returns a new instance of
// csi volume
func From(vol *apis.CSIVolume) *CSIVolume {
	return &CSIVolume{
		Object: vol,
	}
}

// FromCASVolume returns a new instance
// of csi volume from given cas volume
// instance
func FromCASVolume(vol *apismaya.CASVolume) *CSIVolume {
	return &CSIVolume{
		Object: &apis.CSIVolume{
			Spec: apis.CSIVolumeSpec{
				Volume: apis.CSIVolumeInfo{
					Name:     vol.Name,
					Capacity: vol.Spec.Capacity,
				},
				ISCSI: apis.CSIISCSIInfo{
					Iqn:          vol.Spec.Iqn,
					TargetPortal: vol.Spec.TargetPortal,
					Lun:          strconv.FormatInt(int64(vol.Spec.Lun), 10),
				},
			},
		},
	}
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pod instance
type Predicate func(*CSIVolume) bool

// PredicateList holds a list of predicate
type predicateList []Predicate

// CSIVolumeList holds the list
// of csi volume instances
type CSIVolumeList struct {
	List apis.CSIVolumeList
}

// Len returns the number of items present
// in the CSIVolumeList
func (p *CSIVolumeList) Len() int {
	return len(p.List.Items)
}

// all returns true if all the predicates
// succeed against the provided CSIVolume
// instance
func (l predicateList) all(p *CSIVolume) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// HasLabels returns true if provided labels
// are present in the provided CSIVolume instance
func HasLabels(keyValuePair map[string]string) Predicate {
	return func(p *CSIVolume) bool {
		for key, value := range keyValuePair {
			if !p.HasLabel(key, value) {
				return false
			}
		}
		return true
	}
}

// HasLabel returns true if provided label
// is present in the provided CSIVolume instance
func (p *CSIVolume) HasLabel(key, value string) bool {
	val, ok := p.Object.GetLabels()[key]
	if ok {
		return val == value
	}
	return false
}

// HasLabel returns true if provided label
// is present in the provided CSIVolume instance
func HasLabel(key, value string) Predicate {
	return func(p *CSIVolume) bool {
		return p.HasLabel(key, value)
	}
}

// IsNil returns true if the csi volume instance
// is nil
func (p *CSIVolume) IsNil() bool {
	return p.Object == nil
}

// IsNil is predicate to filter out nil csi volume
// instances
func IsNil() Predicate {
	return func(p *CSIVolume) bool {
		return p.IsNil()
	}
}

// GetAPIObject returns csi volume's API instance
func (p *CSIVolume) GetAPIObject() *apis.CSIVolume {
	return p.Object
}
