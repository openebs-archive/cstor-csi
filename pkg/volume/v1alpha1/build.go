/*
Copyright 2019 The OpenEBS Authors

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
	apis "github.com/openebs/csi/pkg/apis/openebs.io/core/v1alpha1"
	errors "github.com/openebs/csi/pkg/generated/maya/errors/v1alpha1"
)

// Builder is the builder object for CSIVolume
type Builder struct {
	volume *CSIVolume
	errs   []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		volume: &CSIVolume{
			Object: &apis.CSIVolume{},
		},
	}
}

// BuilderFrom returns new instance of Builder
// from the provided api instance
func BuilderFrom(volume *apis.CSIVolume) *Builder {
	return &Builder{
		volume: &CSIVolume{
			Object: volume,
		},
	}
}

// WithName sets the name of csi volume
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build csi volume object: missing volume name"))
		return b
	}

	b.volume.Object.Name = name
	return b
}

// Build returns csi volume API object
func (b *Builder) Build() (*apis.CSIVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.volume.Object, nil
}
