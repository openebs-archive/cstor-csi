/*
Copyright © 2018-2019 The OpenEBS Authors

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
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	errors "github.com/openebs/csi/pkg/generated/maya/errors/v1alpha1"
	csipayload "github.com/openebs/csi/pkg/payload/v1alpha1"
	"github.com/openebs/csi/pkg/utils/v1alpha1"
	csivolume "github.com/openebs/csi/pkg/volume/v1alpha1"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var SupportedVolumeCapabilityAccessModes = []*csi.VolumeCapability_AccessMode{
	&csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	},
}

func IsSupportedVolumeCapabilityAccessMode(
	given csi.VolumeCapability_AccessMode_Mode,
) bool {

	for _, access := range SupportedVolumeCapabilityAccessModes {
		if given == access.Mode {
			return true
		}
	}
	return false
}

// newControllerCapabilities returns a list
// of this controller's capabilities
func newControllerCapabilities() []*csi.ControllerServiceCapability {
	fromType := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var capabilities []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
}

// controller is the server implementation
// for CSI Controller
type controller struct {
	driver       *CSIDriver
	capabilities []*csi.ControllerServiceCapability
}

// NewController returns a new instance
// of CSI controller
func NewController(d *CSIDriver) csi.ControllerServer {
	return &controller{
		driver:       d,
		capabilities: newControllerCapabilities(),
	}
}

// validateRequest validates if the requested service is
// supported by the driver
func (cs *controller) validateRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, cap := range cs.capabilities {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}

	return status.Error(
		codes.InvalidArgument,
		fmt.Sprintf("failed to validate request: {%s} is not supported", c),
	)
}

// CreateVolume provisions a volume
func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	logrus.Infof("received request to create volume {%s}", req.GetName())

	err := cs.validateRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to handle create volume request for {%s}",
			req.GetName(),
		)
	}

	volName := req.GetName()
	if len(volName) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume name",
		)
	}

	volCapabilities := req.GetVolumeCapabilities()
	if volCapabilities == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume capabilities",
		)
	}

	for _, cap := range volCapabilities {
		if cap.GetBlock() != nil {
			return nil, status.Error(
				codes.Unimplemented,
				"failed to handle create volume request: block volume is not supported",
			)
		}
	}

	// verify if the volume has already been created
	_, err = utils.GetVolumeByName(volName)
	if err == nil {
		return nil,
			status.Error(
				codes.AlreadyExists,
				fmt.Sprintf("failed to handle create volume request: volume {%s} already exists", volName),
			)
	}

	// TODO
	// This needs to be de-coupled. csi & maya api server
	// should deal with custom resources and hence
	// reconciliation
	//
	// Send volume creation request to maya apiserver
	casvol, err := utils.ProvisionVolume(req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Create a csi vol object from maya apiserver's
	// create volume request
	csivol := csivolume.FromCASVolume(casvol).Object

	//TODO take lock against this local map
	//
	// Keep a local copy of the volume info
	// to catch duplicate requests
	utils.Volumes[volName] = csivol

	return csipayload.NewCreateVolumeResponseBuilder().
		WithName(volName).
		WithCapacity(req.GetCapacityRange().GetRequiredBytes()).
		// VolumeContext is essential for publishing
		// volumes at nodes, for iscsi login, this
		// will be stored in PV CR
		WithContext(map[string]string{
			"volname":        volName,
			"iqn":            casvol.Spec.Iqn,
			"targetPortal":   casvol.Spec.TargetPortal,
			"lun":            "0",
			"iscsiInterface": "default",
			"portals":        casvol.Spec.TargetPortal,
		}).
		Build(), nil
}

// DeleteVolume deletes the specified volume
func (cs *controller) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	logrus.Infof("received request to delete volume {%s}", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"failed to handle delete volume request: missing volume id",
		)
	}

	err := cs.validateRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s}",
			req.VolumeId,
		)
	}

	// this call is made just to fetch pvc namespace
	pv, err := utils.FetchPVDetails(req.VolumeId)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s}",
			req.VolumeId,
		)
	}

	pvcNamespace := pv.Spec.ClaimRef.Namespace

	// send delete request to maya apiserver
	err = utils.DeleteVolume(req.VolumeId, pvcNamespace)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s}",
			req.VolumeId,
		)
	}

	// TODO
	// Use a lock to remove
	//
	// remove entry from the in-memory
	// maintained list
	delete(utils.Volumes, req.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

// TODO
// Verify if this is a never ending loop
//
// ValidateVolumeCapabilities validates the capabilities
// required to create a new volume
//
// This implements csi.ControllerServer
func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	return cs.ValidateVolumeCapabilities(ctx, req)
}

// ControllerGetCapabilities fetches controller capabilities
//
// This implements csi.ControllerServer
func (cs *controller) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.capabilities,
	}

	return resp, nil
}

// ControllerExpandVolume resizes previously provisioned volume
//
// This implements csi.ControllerServer
func (cs *controller) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// validateCapabilities validates if provided capabilities
// are supported by this driver
func validateCapabilities(caps []*csi.VolumeCapability) bool {

	var supported bool
	for _, cap := range caps {
		if IsSupportedVolumeCapabilityAccessMode(cap.AccessMode.Mode) {
			supported = true
		} else {
			supported = false
		}
	}

	return supported
}

// CreateSnapshot creates a snapshot for given volume
//
// This implements csi.ControllerServer
func (cs *controller) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// DeleteSnapshot deletes given snapshot
//
// This implements csi.ControllerServer
func (cs *controller) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ListSnapshots lists all snapshots for the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume removes a previously
// attached volume from the given node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerPublishVolume attaches given volume
// at the specified node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity return the capacity of the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest,
) (*csi.GetCapacityResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes lists all the volumes
//
// This implements csi.ControllerServer
func (cs *controller) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest,
) (*csi.ListVolumesResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}
