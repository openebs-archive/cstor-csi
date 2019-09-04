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
	"strconv"

	apismaya "github.com/openebs/csi/pkg/apis/openebs.io/maya/v1alpha1"
	errors "github.com/openebs/csi/pkg/maya/errors/v1alpha1"
)

// Builder is the builder object for JivaVolumeClaim
type Builder struct {
	jvc  *JivaVolumeClaim
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		jvc: &JivaVolumeClaim{
			object: &apismaya.JivaVolumeClaim{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(jvc *apismaya.JivaVolumeClaim) *Builder {
	if jvc == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolumeclaim object: nil jvc"),
		)
		return b
	}
	return &Builder{
		jvc: &JivaVolumeClaim{
			object: jvc,
		},
	}
}

// WithName sets the Name of JivaVolumeClaim
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolumeclaim object: missing name"),
		)
		return b
	}
	b.jvc.object.Name = name
	return b
}

// WithGenerateName sets the GenerateName field of
// CStorVolume with provided value
func (b *Builder) WithGenerateName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing generateName"),
		)
		return b
	}

	b.jvc.object.GenerateName = name
	return b
}

// WithNamespace sets the Namespace field of CStorVolume provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing namespace"),
		)
		return b
	}
	b.jvc.object.Namespace = namespace
	return b
}

// WithAnnotations merges existing annotations if any
// with the ones that are provided here
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing annotations"),
		)
		return b
	}

	if b.jvc.object.Annotations == nil {
		return b.WithAnnotationsNew(annotations)
	}

	for key, value := range annotations {
		b.jvc.object.Annotations[key] = value
	}
	return b
}

// WithAnnotationsNew resets existing annotations if any with
// ones that are provided here
func (b *Builder) WithAnnotationsNew(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: no new annotations"),
		)
		return b
	}

	// copy of original map
	newannotations := map[string]string{}
	for key, value := range annotations {
		newannotations[key] = value
	}

	// override
	b.jvc.object.Annotations = newannotations
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing labels"),
		)
		return b
	}

	if b.jvc.object.Labels == nil {
		return b.WithLabelsNew(labels)
	}

	for key, value := range labels {
		b.jvc.object.Labels[key] = value
	}
	return b
}

// WithLabelsNew resets existing labels if any with
// ones that are provided here
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: no new labels"),
		)
		return b
	}

	// copy of original map
	newlbls := map[string]string{}
	for key, value := range labels {
		newlbls[key] = value
	}

	// override
	b.jvc.object.Labels = newlbls
	return b
}

// WithTargetIP sets the target IP address field of
// CStorVolume with provided arguments
func (b *Builder) WithTargetIP(targetip string) *Builder {
	if len(targetip) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing targetip"),
		)
		return b
	}
	b.jvc.object.Spec.TargetIP = targetip
	return b
}

// WithCapacity sets the Capacity field of CStorVolume with provided arguments
func (b *Builder) WithCapacity(capacity int64) *Builder {
	if capacity == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing capacity"),
		)
		return b
	}
	b.jvc.object.Spec.Capacity = capacity
	return b
}

// WithIQN sets the IQN field of CStorVolume with provided arguments
func (b *Builder) WithIQN(iqn string) *Builder {
	if len(iqn) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing iqn"),
		)
		return b
	}
	b.jvc.object.Spec.Iqn = iqn
	return b
}

// WithIQN sets the IQN field of CStorVolume with provided arguments
func (b *Builder) WithStatus(status string) *Builder {
	if len(status) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing status"),
		)
		return b
	}
	b.jvc.object.Spec.Status = status
	return b
}

// WithTargetPort sets the TargetPort field of
// CStorVolume with provided arguments
func (b *Builder) WithTargetPort(targetport string) *Builder {
	if len(targetport) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing targetport"),
		)
		return b
	}
	b.jvc.object.Spec.TargetPort = targetport
	return b
}

// WithTargetPortal sets the TargetPortal field of
// CStorVolume with provided arguments
func (b *Builder) WithTargetPortal(targetportal string) *Builder {
	if len(targetportal) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build jivavolume object: missing targetportal"),
		)
		return b
	}
	b.jvc.object.Spec.TargetPortal = targetportal
	return b
}

// WithStatusPhase updates the phase of JivaVolumeClaim
func (b *Builder) WithStatusPhase(
	phase apismaya.JivaVolumeClaimPhase) *Builder {
	if phase == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build jivavolumeclaim object: missing phase",
			),
		)
		return b
	}
	b.jvc.object.Status.Phase = phase
	return b
}

// WithFinalizers merges existing finalizers of JivaVolumeClaim if any
// with the ones that are provided here
func (b *Builder) WithFinalizers(finalizers []string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build jivavolumeclaim object: missing finalizers",
			),
		)
		return b
	}

	if b.jvc.object.Finalizers == nil {
		return b.WithFinalizersNew(finalizers)
	}

	b.jvc.object.Finalizers = append(b.jvc.object.Finalizers, finalizers...)
	return b
}

// WithFinalizersNew resets existing finalizers of JivaVolumeClaim if any with
// ones that are provided here
func (b *Builder) WithFinalizersNew(finalizers []string) *Builder {
	if len(finalizers) == 0 {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build jivavolumeclaim object: no new finalizers",
			),
		)
		return b
	}

	// override
	b.jvc.object.Finalizers = nil
	b.jvc.object.Finalizers = append(b.jvc.object.Finalizers, finalizers...)
	return b
}

// WithReplicaCount sets replica count of JivaVolumeClaim
func (b *Builder) WithReplicaCount(count string) *Builder {

	replicaCount, err := strconv.Atoi(count)
	if err != nil {
		b.errs = append(
			b.errs,
			errors.Wrapf(
				err,
				"failed to build jivavolumeclaim object {%s}",
				count,
			),
		)
		return b
	}
	b.jvc.object.Spec.ReplicaCount = replicaCount
	return b
}

// WithNodeID sets NodeID details of JivaVolumeClaim
func (b *Builder) WithNodeID(nodeID string) *Builder {
	if nodeID == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build jivavolumeclaim object: missing nodeID",
			),
		)
		return b
	}
	b.jvc.object.NodeID = nodeID
	return b
}

// WithNodeID sets NodeID details of JivaVolumeClaim
func (b *Builder) WithPhase(phase apismaya.JivaVolumeClaimPhase) *Builder {
	if phase == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build jivavolumeclaim object: missing phase",
			),
		)
		return b
	}
	b.jvc.object.Status.Phase = phase
	return b
}

// Build returns the JivaVolumeClaim API instance
func (b *Builder) Build() (*apismaya.JivaVolumeClaim, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.jvc.object, nil
}
