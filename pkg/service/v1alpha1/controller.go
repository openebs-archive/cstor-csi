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
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	apismaya "github.com/openebs/cstor-csi/pkg/apis/openebs.io/maya/v1alpha1"
	errors "github.com/openebs/cstor-csi/pkg/maya/errors/v1alpha1"
	csipayload "github.com/openebs/cstor-csi/pkg/payload/v1alpha1"
	utils "github.com/openebs/cstor-csi/pkg/utils/v1alpha1"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

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

// ValidateVolumeCapabilities validates the capabilities
// required to create a new volume
// This implements csi.ControllerServer
func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	if req.GetVolumeId() == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"failed to handle ValidateVolumeCapabilities request: missing volume id",
		)
	}

	if len(req.GetVolumeCapabilities()) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"failed to handle ValidateVolumeCapabilities request: missing VolumeCapabilities",
		)
	}
	cvc, err := utils.GetVolume(req.GetVolumeId())
	if err == nil && cvc != nil && cvc.DeletionTimestamp != nil {
		return nil, status.Error(codes.NotFound, "Volume does not exist")
	}
	if err != nil {
		if k8serror.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "Volume does not exist")
		}
		return nil, err
	}
	return &csi.ValidateVolumeCapabilitiesResponse{}, nil
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

// CreateVolume provisions a volume
func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {

	logrus.Infof("received request to create volume {%s}", req.GetName())
	var (
		snapshotID string
		err        error
		nodeID     string
		size       int64
	)

	if err = cs.validateVolumeCreateReq(req); err != nil {
		return nil, err
	}

	volName := strings.ToLower(req.GetName())
	if req.GetCapacityRange() != nil {
		size = req.GetCapacityRange().RequiredBytes
	} else {
		size = 1024 * 1024 * 1024
	}
	rCount := req.GetParameters()["replicaCount"]
	cspcName := req.GetParameters()["cstorPoolCluster"]
	policyName := req.GetParameters()["cstorVolumePolicy"]
	VolumeContext := map[string]string{
		"openebs.io/cas-type": req.GetParameters()["cas-type"],
	}
	if req.GetAccessibilityRequirements() != nil && len(req.GetAccessibilityRequirements().
		GetPreferred()) != 0 {
		nodeID = req.GetAccessibilityRequirements().
			GetPreferred()[0].GetSegments()[HostTopologyKey]
	}

	contentSource := req.GetVolumeContentSource()
	if contentSource != nil && contentSource.GetSnapshot() != nil {
		snapshotID = contentSource.GetSnapshot().GetSnapshotId()
		if snapshotID == "" {
			return nil, status.Error(codes.InvalidArgument, "snapshot ID is empty")
		}
		if isValidSrc, _ := utils.IsSourceAvailable(snapshotID); !isValidSrc {
			return nil, status.Error(
				codes.NotFound,
				"VolumeSrc Not Available")
		}
	}

	// verify if the volume has already been created
	cvc, err := utils.GetVolume(volName)
	if err == nil && cvc != nil && cvc.DeletionTimestamp == nil {
		qcap := cvc.Spec.Capacity[corev1.ResourceStorage]
		cap, _ := qcap.AsInt64()
		if size == cap {
			goto createVolumeResponse
		}
		return nil, status.Error(codes.AlreadyExists, "Volume already exist with different size")
	}
	err = utils.ProvisionVolume(size, volName, rCount,
		cspcName, snapshotID,
		nodeID, policyName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

createVolumeResponse:
	return csipayload.NewCreateVolumeResponseBuilder().
		WithName(volName).
		WithCapacity(size).
		WithContext(VolumeContext).
		Build(), nil
}

// DeleteVolume deletes the specified volume
func (cs *controller) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	logrus.Infof("received request to delete volume {%s}", req.VolumeId)

	var (
		err error
		cvc *apismaya.CStorVolumeClaim
	)

	if err = cs.validateDeleteVolumeReq(req); err != nil {
		return nil, err
	}

	volumeID := strings.ToLower(req.GetVolumeId())

	// verify if the volume has already been deleted
	cvc, err = utils.GetVolume(volumeID)
	if k8serror.IsNotFound(err) {
		goto deleteResponse
	}
	if cvc != nil && cvc.DeletionTimestamp != nil {
		goto deleteResponse
	}

	// Delete the corresponding CVC
	err = utils.DeleteVolume(volumeID)
	if err != nil {
		if !k8serror.IsNotFound(err) {
			return nil, errors.Wrapf(
				err,
				"failed to handle delete volume request for {%s}",
				volumeID,
			)
		}
	}
deleteResponse:
	return csipayload.NewDeleteVolumeResponseBuilder().Build(), nil
}

// ControllerPublishVolume attaches given volume
// at the specified node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {

	if err := cs.validateControllerPublishVolumeReq(req); err != nil {
		return nil, err
	}
	if err := prepareVolumeForNode(req); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.ControllerPublishVolumeResponse{}, nil
}

// ControllerUnpublishVolume removes a previously
// attached volume from the given node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {
	if err := cs.validateControllerUnpublishVolumeReq(req); err != nil {
		return nil, err
	}
	if err := utils.DeleteCSIVolumeCR(req.GetVolumeId() + "-" + req.GetNodeId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ControllerExpandVolume resizes previously provisioned volume
//
// This implements csi.ControllerServer
func (cs *controller) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
	var err error
	if err = cs.validateExpandVolumeReq(req); err != nil {
		return nil, err
	}

	updatedSize := req.GetCapacityRange().GetRequiredBytes()
	if err := utils.ResizeVolume(req.VolumeId, updatedSize); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle ControllerExpandVolumeRequest for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}
	return csipayload.NewControllerExpandVolumeResponseBuilder().
		WithCapacityBytes(updatedSize).
		WithNodeExpansionRequired(true).
		Build(), nil
}

// CreateSnapshot creates a snapshot for given volume
//
// This implements csi.ControllerServer
func (cs *controller) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {

	var err error
	if err = cs.validateCreateSnapshotReq(req); err != nil {
		return nil, err
	}
	snapTimeStamp := time.Now().Unix()
	snapTimeStampStr := strconv.FormatInt(time.Now().Unix(), 10)
	srcVolID := strings.ToLower(req.SourceVolumeId)
	if err := utils.CreateSnapshot(srcVolID, req.Name+"-"+snapTimeStampStr); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle CreateSnapshotRequest for %s: %s, {%s}",
			req.SourceVolumeId, req.Name,
			err.Error(),
		)
	}
	return csipayload.NewCreateSnapshotResponseBuilder().
		WithSourceVolumeID(req.SourceVolumeId).
		WithSnapshotID(srcVolID+"@"+req.Name+"-"+snapTimeStampStr).
		WithCreationTime(snapTimeStamp, 0).
		WithReadyToUse(true).
		Build(), nil
}

// DeleteSnapshot deletes given snapshot
//
// This implements csi.ControllerServer
func (cs *controller) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {

	var err error
	if err = cs.validateDeleteSnapshotReq(req); err != nil {
		return nil, err
	}
	snapshotID := strings.Split(req.SnapshotId, "@")
	if len(snapshotID) != 2 {
		return &csi.DeleteSnapshotResponse{}, nil
	}
	if err := utils.DeleteSnapshot(snapshotID[0], snapshotID[1]); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle CreateSnapshotRequest for %s, {%s}",
			req.SnapshotId,
			err.Error(),
		)
	}
	return &csi.DeleteSnapshotResponse{}, nil
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
