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

package volumeattachment

import (
	apis "github.com/openebs/api/v2/pkg/apis/cstor/v1"
	"github.com/pkg/errors"
)

// Builder is the builder object for CStorVolumeAttachment
type Builder struct {
	volume *CStorVolumeAttachment
	errs   []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		volume: &CStorVolumeAttachment{
			Object: &apis.CStorVolumeAttachment{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(volume *apis.CStorVolumeAttachment) *Builder {
	if volume == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil volume"),
		)
		return b
	}
	return &Builder{
		volume: &CStorVolumeAttachment{
			Object: volume,
		},
	}
}

// WithNamespace sets the namespace of csi volume
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing namespace",
			),
		)
		return b
	}
	b.volume.Object.Namespace = namespace
	return b
}

// WithName sets the name of csi volume
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing name",
			),
		)
		return b
	}
	b.volume.Object.Name = name
	return b
}

// WithVolName sets the VolName of csi volume
func (b *Builder) WithVolName(volName string) *Builder {
	if volName == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing volume name",
			),
		)
		return b
	}
	b.volume.Object.Spec.Volume.Name = volName
	return b
}

// WithAccessType sets the accessType of csi volume i.e. block or mount
func (b *Builder) WithAccessType(accessType string) *Builder {
	if accessType == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing accessType",
			),
		)
		return b
	}
	b.volume.Object.Spec.Volume.AccessType = accessType
	return b
}

// WithCapacity sets the Capacity of csi volume by converting string
// capacity into Quantity
func (b *Builder) WithCapacity(capacity string) *Builder {
	if capacity == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing capacity",
			),
		)
		return b
	}
	b.volume.Object.Spec.Volume.Capacity = capacity
	return b
}

// WithFSType sets the fstype of csi volume
func (b *Builder) WithFSType(fstype string) *Builder {
	if fstype == "" && b.volume.Object.Spec.Volume.AccessType == "mount" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing fstype",
			),
		)
		return b
	}
	b.volume.Object.Spec.Volume.FSType = fstype
	return b
}

// WithStagingTargetPath sets the stagingTargetpath of csi volume
func (b *Builder) WithStagingTargetPath(stagingTargetPath string) *Builder {
	if stagingTargetPath == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing mountPath",
			),
		)
		return b
	}
	b.volume.Object.Spec.Volume.StagingTargetPath = stagingTargetPath
	return b
}

// WithMountOptions sets the mountoptions of csi volume
func (b *Builder) WithMountOptions(mountOptions []string) *Builder {
	//Error is not being retured over here since this is an optional field
	if len(mountOptions) == 0 {
		return b
	}
	if b.volume.Object.Spec.Volume.MountOptions == nil {
		return b.WithMountOptionsNew(mountOptions)
	}
	b.volume.Object.Spec.Volume.MountOptions = append(
		b.volume.Object.Spec.Volume.MountOptions, mountOptions...)
	return b
}

// WithMountOptionsNew sets the mountoptions of csi volume
func (b *Builder) WithMountOptionsNew(mountOptions []string) *Builder {
	//Error is not being retured over here since this is an optional field
	if len(mountOptions) == 0 {
		return b
	}
	b.volume.Object.Spec.Volume.MountOptions = nil
	b.volume.Object.Spec.Volume.MountOptions = append(
		b.volume.Object.Spec.Volume.MountOptions, mountOptions...)
	return b
}

// WithDevicePath sets the devicePath of csi volume
func (b *Builder) WithDevicePath(devicePath string) *Builder {
	if devicePath == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing devicePath",
			),
		)
		return b
	}
	b.volume.Object.Spec.Volume.DevicePath = devicePath
	return b
}

// WithOwnerNodeID sets the ownerNodeID of csi volume
func (b *Builder) WithOwnerNodeID(ownerNodeID string) *Builder {
	if ownerNodeID == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing ownerNodeID",
			),
		)
		return b
	}
	b.volume.Object.Spec.Volume.OwnerNodeID = ownerNodeID
	return b
}

// WithIQN sets the IQN of csi volume
func (b *Builder) WithIQN(iqn string) *Builder {
	if iqn == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing IQN",
			),
		)
		return b
	}
	b.volume.Object.Spec.ISCSI.Iqn = iqn
	return b
}

// WithTargetPortal sets the TargetPortal of csi volume
func (b *Builder) WithTargetPortal(targetPortal string) *Builder {
	if targetPortal == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing targetPortal",
			),
		)
		return b
	}
	b.volume.Object.Spec.ISCSI.TargetPortal = targetPortal
	return b
}

// WithIscsiInterface sets the iscsiInterface of csi volume
func (b *Builder) WithIscsiInterface(iscsiInterface string) *Builder {
	if iscsiInterface == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing iscsiInterface",
			),
		)
		return b
	}
	b.volume.Object.Spec.ISCSI.IscsiInterface = iscsiInterface
	return b
}

// WithLun sets the lunID of csi volume
func (b *Builder) WithLun(lun string) *Builder {
	if lun == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing lun",
			),
		)
		return b
	}
	b.volume.Object.Spec.ISCSI.Lun = lun
	return b
}

// WithReadOnly sets the readOnly property of csi volume
func (b *Builder) WithReadOnly(readOnly bool) *Builder {
	b.volume.Object.Spec.Volume.ReadOnly = readOnly
	return b
}

// WithLabels merges existing labels of csi volume if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: missing labels",
			),
		)
		return b
	}

	if b.volume.Object.Labels == nil {
		return b.WithLabelsNew(labels)
	}

	for key, value := range labels {
		b.volume.Object.Labels[key] = value
	}
	return b
}

// WithLabelsNew resets existing labels of csi volume if any with
// ones that are provided here
func (b *Builder) WithLabelsNew(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi volume object: no new labels",
			),
		)
		return b
	}

	// copy of original map
	newlbls := map[string]string{}
	for key, value := range labels {
		newlbls[key] = value
	}

	// override
	b.volume.Object.Labels = newlbls
	return b
}

// Build returns csi volume API object
func (b *Builder) Build() (*apis.CStorVolumeAttachment, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.volume.Object, nil
}
