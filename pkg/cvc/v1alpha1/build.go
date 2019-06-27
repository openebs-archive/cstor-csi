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
	apismaya "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	metav1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

// Builder is the builder object for CStorVolume
type Builder struct {
	cvc  *CStorVolumeClaim
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{cvc: &CStorVolumeClaim{object: &apismaya.CStorVolumeClaim{}}}
}

// WithName sets the Name field of CStorVolumeClaim with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing name"),
		)
		return b
	}
	b.cvc.object.Name = name
	return b
}

// WithGenerateName sets the GenerateName field of
// CStorVolumeClaim with provided value
func (b *Builder) WithGenerateName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing generateName"),
		)
		return b
	}

	b.cvc.object.GenerateName = name
	return b
}

// WithNamespace resets the Namespace field of CStorVolumeClaim with provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing namespace"),
		)
		return b
	}
	b.cvc.object.Namespace = namespace
	return b
}

// WithStatus updates the status field of CStorVolumeClaim with provided arguments
func (b *Builder) WithStatus(status apismaya.CStorVolumeClaimStatus) *Builder {
	if status.Phase == "" {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing phase"),
		)
		return b
	}
	b.cvc.object.Status.Phase = status.Phase
	if status.Conditions != nil {
		b.cvc.object.Status.Conditions = append(b.cvc.object.Status.Conditions,
			status.Conditions...)
	}
	return b
}

// WithStatusNew sets the status field of CStorVolumeClaim with provided arguments
func (b *Builder) WithStatusNew(status apismaya.CStorVolumeClaimStatus) *Builder {
	if status.Phase == "" {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing phase"),
		)
		return b
	}
	b.cvc.object.Status.Phase = status.Phase
	b.cvc.object.Status.Conditions = status.Conditions
	return b
}

// WithAnnotations merges existing annotations if any
// with the ones that are provided here
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing annotations"),
		)
		return b
	}

	if b.cvc.object.Annotations == nil {
		return b.WithAnnotationsNew(annotations)
	}

	for key, value := range annotations {
		b.cvc.object.Annotations[key] = value
	}
	return b
}

// WithAnnotationsNew resets existing annotations if any with
// ones that are provided here
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: no new annotations"),
		)
		return b
	}

	// copy of original map
	newannotations := map[string]string{}
	for key, value := range annotations {
		newannotations[key] = value
	}

	// override
	b.cvc.object.Annotations = newannotations
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing labels"),
		)
		return b
	}

	if b.cvc.object.Labels == nil {
		return b.WithLabelsNew(labels)
	}

	for key, value := range labels {
		b.cvc.object.Labels[key] = value
	}
	return b
}

// WithLabelsNew resets existing labels if any with
// ones that are provided here
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: no new labels"),
		)
		return b
	}

	// copy of original map
	newlbls := map[string]string{}
	for key, value := range labels {
		newlbls[key] = value
	}

	// override
	b.cvc.object.Labels = newlbls
	return b
}

// WithFinalizers merges existing finalizers if any
// with the ones that are provided here
func (b *Builder) WithFinalizers(finalizers []string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: missing finalizers"),
		)
		return b
	}

	if b.cvc.object.Finalizers == nil {
		return b.WithFinalizersNew(finalizers)
	}

	b.cvc.object.Finalizers = append(b.cvc.object.Finalizers, finalizers...)
	return b
}

// WithFinalizersNew resets existing finalizers if any with
// ones that are provided here
func (b *Builder) WithFinalizersNew(finalizers []string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build cstorvolumeclaim object: no new finalizers"),
		)
		return b
	}

	// override
	b.cvc.object.Finalizers = finalizers
	return b
}

// WithCapacity sets the Capacity field in CstorVOlumeClaim by converting string
// capacity into Quantity
func (b *Builder) WithCapacity(capacity string) *Builder {
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.errs = append(b.errs, errors.Wrapf(err, "failed to build PV object: failed to parse capacity {%s}", capacity))
		return b
	}
	return b.WithCapacityQty(resCapacity)
}

// WithCapacityQty sets the Capacity field in PV with provided arguments
func (b *Builder) WithCapacityQty(resCapacity resource.Quantity) *Builder {
	resourceList := metav1.ResourceList{
		metav1.ResourceName(metav1.ResourceStorage): resCapacity,
	}
	b.cvc.object.Spec.Capacity = resourceList
	return b
}

// Build returns the CStorVolumeClaim API instance
func (b *Builder) Build() (*apismaya.CStorVolumeClaim, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.cvc.object, nil
}
