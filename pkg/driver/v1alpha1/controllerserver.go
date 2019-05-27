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
	csipayload "github.com/openebs/csi/pkg/payload/v1alpha1"
	"github.com/openebs/csi/pkg/utils/v1alpha1"
	csivolume "github.com/openebs/csi/pkg/volume/v1alpha1"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newControllerServiceCapabilities returns a list
// of this controller service's capabilities
func newControllerServiceCapabilities() []*csi.ControllerServiceCapability {
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

// ControllerServer defines a controller driver
type ControllerServer struct {
	driver       *CSIDriver
	capabilities []*csi.ControllerServiceCapability
}

// NewControllerServer returns a new instance
// of controller server
func NewControllerServer(d *CSIDriver) *ControllerServer {
	return &ControllerServer{
		driver:       d,
		capabilities: newControllerServiceCapabilities(),
	}
}

// validateRequest validates if the requested service is
// supported by the driver
func (cs *ControllerServer) validateRequest(c csi.ControllerServiceCapability_RPC_Type) error {
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

// CreateVolume dynamically provisions a volume
// on demand
func (cs *ControllerServer) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	logrus.Infof("received create volume request")
	err := cs.validateRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME)
	if err != nil {
		return nil, err
	}

	volName := req.GetName()
	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "failed to create volume: missing volume name")
	}

	volCapabilities := req.GetVolumeCapabilities()
	if volCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "failed to create volume: missing volume capabilities")
	}

	for _, cap := range volCapabilities {
		if cap.GetBlock() != nil {
			return nil, status.Error(codes.Unimplemented, "failed to create volume: block volume not supported")
		}
	}

	// verify if the volume has already been created
	_, err = utils.GetVolumeByName(volName)
	if err == nil {
		return nil,
			status.Error(
				codes.AlreadyExists,
				fmt.Sprintf("failed to create volume: volume {%s} already exists", volName),
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
func (cs *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	logrus.Infof("Delete Volume")

	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if err := cs.validateRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		logrus.Infof("invalid delete volume req: %v", req)
		return nil, err
	}
	volumeID := req.VolumeId

	// This call is made just to fetch pvc namespace
	pv, err := utils.FetchPVDetails(volumeID)
	if err != nil {
		logrus.Infof("fetch Volume Failed, volID:%v %v", volumeID, err)
		return nil, err
	}
	pvcNamespace := pv.Spec.ClaimRef.Namespace

	//Send delete request to m-apiserver
	if err := utils.DeleteVolume(volumeID, pvcNamespace); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Remove entry from the volume list maintained
	delete(utils.Volumes, volumeID)
	return &csi.DeleteVolumeResponse{}, nil
}

// ValidateVolumeCapabilities validates the capabilities required to create a
// new volume
func (cs *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return cs.ValidateVolumeCapabilities(ctx, req)
}

// ControllerGetCapabilities fetches the controller capabilities
func (cs *ControllerServer) ControllerGetCapabilities(ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.capabilities,
	}

	return resp, nil
}

// ControllerExpandVolume can be used to resize the previously provisioned
// volume
func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context,
	req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, nil
}

// validateCapabilities validates if the corresponding capability is supported
// by the driver
func validateCapabilities(caps []*csi.VolumeCapability) bool {
	vcaps := []*csi.VolumeCapability_AccessMode{supportedAccessMode}

	hasSupport := func(mode csi.VolumeCapability_AccessMode_Mode) bool {
		for _, m := range vcaps {
			if mode == m.Mode {
				return true
			}
		}
		return false
	}

	supported := false
	for _, cap := range caps {
		if hasSupport(cap.AccessMode.Mode) {
			supported = true
		} else {
			supported = false
		}
	}

	return supported
}

// CreateSnapshot can be used to create a snapsnhot for a particular volumeID
// provided
func (cs *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// DeleteSnapshot can be used to delete a particular snapshot of a specified
// volume
func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListSnapshots lists all the snapshots for the volume specified via VolumeID
func (cs *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume can be used to remove a previously attached volume
// from the specified node
func (cs *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerPublishVolume can be used to attach the volume at the specified
// node
func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity return the capacity of the the storage pool
func (cs *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes lists the info of all the OpenEBS volumes created via m-apiserver
func (cs *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
