package v1alpha1

import (
	"fmt"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	apis "github.com/openebs/cstor-csi/pkg/apis/openebs.io/core/v1alpha1"
	errors "github.com/openebs/cstor-csi/pkg/maya/errors/v1alpha1"
	utils "github.com/openebs/cstor-csi/pkg/utils/v1alpha1"
	csivol "github.com/openebs/cstor-csi/pkg/volume/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

// SupportedVolumeCapabilityAccessModes contains the list of supported access
// modes for the volume
var SupportedVolumeCapabilityAccessModes = []*csi.VolumeCapability_AccessMode{
	&csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	},
}

// newControllerCapabilities returns a list
// of this controller's capabilities
func newControllerCapabilities() []*csi.ControllerServiceCapability {
	fromType := func(
		cap csi.ControllerServiceCapability_RPC_Type,
	) *csi.ControllerServiceCapability {
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
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
}

// IsSupportedVolumeCapabilityAccessMode valides the requested access mode
func IsSupportedVolumeCapabilityAccessMode(
	accessMode csi.VolumeCapability_AccessMode_Mode,
) bool {

	for _, access := range SupportedVolumeCapabilityAccessModes {
		if accessMode == access.Mode {
			return true
		}
	}
	return false
}

// validateCapabilities validates if provided capabilities
// are supported by this driver
func validateCapabilities(caps []*csi.VolumeCapability) bool {

	for _, cap := range caps {
		if !IsSupportedVolumeCapabilityAccessMode(cap.AccessMode.Mode) {
			return false
		}
	}
	return true
}

// validateRequest validates if the requested service is
// supported by the driver
func (cs *controller) validateRequest(
	c csi.ControllerServiceCapability_RPC_Type,
) error {

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

func (cs *controller) validateVolumeCreateReq(req *csi.CreateVolumeRequest) error {
	err := cs.validateRequest(
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to handle create volume request for {%s}",
			req.GetName(),
		)
	}

	if req.GetName() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume name",
		)
	}

	if req.GetParameters()["cstorPoolCluster"] == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing storage class parameter cstorPoolCluster",
		)
	}
	if req.GetParameters()["replicaCount"] == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing storage class parameter replicaCount",
		)
	}

	if req.GetParameters()["cas-type"] == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing storage class parameter cas-type",
		)
	}

	volCapabilities := req.GetVolumeCapabilities()
	if volCapabilities == nil {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create volume request: missing volume capabilities",
		)
	}
	for _, volcap := range volCapabilities {
		mount := volcap.GetMount()
		if mount != nil {
			if !isValidFStype(mount.GetFsType()) {
				return status.Errorf(
					codes.InvalidArgument,
					"failed to handle create volume request, invalid fsType : %s",
					req.GetParameters()["fsType"],
				)
			}
		}
	}
	return nil
}

func (cs *controller) validateDeleteVolumeReq(req *csi.DeleteVolumeRequest) error {
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle delete volume request: missing volume id",
		)
	}

	err := cs.validateRequest(
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to handle delete volume request for {%s}",
			volumeID,
		)
	}
	return nil
}

func (cs *controller) validateVolumePublishReq(req *csi.NodePublishVolumeRequest) error {
	err := cs.validateRequest(
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to handle publish volume request for {%s}",
			req.GetVolumeId(),
		)
	}

	if req.GetVolumeId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing volume ID",
		)
	}

	if req.GetTargetPath() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing targetPath",
		)
	}

	if req.GetStagingTargetPath() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing stagingTargetPath",
		)
	}
	return nil
}

func prepareVolumeForNode(
	req *csi.ControllerPublishVolumeRequest,
) error {
	volumeID := req.GetVolumeId()
	nodeID := req.GetNodeId()

	if err := utils.PatchCVCNodeID(volumeID, nodeID); err != nil {
		return err
	}

	labels := map[string]string{
		"nodeID":  nodeID,
		"Volname": volumeID,
	}
	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	if fsType == "" {
		fsType = "ext4"
	}

	// If the access type is block, do nothing for stage
	var accessType string
	switch req.GetVolumeCapability().GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		accessType = "block"
	case *csi.VolumeCapability_Mount:
		accessType = "mount"
	}

	vol, err := csivol.NewBuilder().
		WithName(volumeID + "-" + nodeID).
		WithLabels(labels).
		WithVolName(req.GetVolumeId()).
		WithAccessType(accessType).
		WithFSType(fsType).
		WithReadOnly(req.GetReadonly()).Build()
	if err != nil {
		return err
	}
	retryCount := 0
retry:
	if isCVCBound, err := utils.IsCVCBound(volumeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	} else if !isCVCBound {
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusWaitingForCVCBound
		time.Sleep(5 * time.Second)
		retryCount++
		if retryCount == 5 {
			return errors.New("Waiting for CVC to be bound")
		}
		goto retry
	}

	if err = utils.FetchAndUpdateISCSIDetails(volumeID, vol); err != nil {
		return err
	}
	if err = utils.WaitForVolumeToBeReady(vol.Spec.Volume.Name); err != nil {
		return err
	}

	oldvol, err := utils.GetCSIVolume(vol.Name)
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	} else if err == nil && oldvol != nil {
		if oldvol.DeletionTimestamp != nil {
			return errors.Errorf("Volume still mounted on node: %s", nodeID)
		}
		return nil
	}

	if err = utils.DeleteOldCSIVolumeCRs(volumeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if err = utils.CreateCSIVolumeCR(vol, nodeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (cs *controller) validateExpandVolumeReq(req *csi.ControllerExpandVolumeRequest) error {
	if req.GetCapacityRange() == nil {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle expand volume request: missing capacity",
		)
	}

	if req.GetVolumeId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle expand volume request: missing volume ID",
		)
	}
	return nil
}

func (cs *controller) validateCreateSnapshotReq(req *csi.CreateSnapshotRequest) error {
	if req.GetName() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle create snapshot request: missing snapshotName",
		)
	}

	if req.GetSourceVolumeId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle  create snapshot request: missing SourceVolumeID",
		)
	}

	return nil
}

func (cs *controller) validateDeleteSnapshotReq(req *csi.DeleteSnapshotRequest) error {
	if req.GetSnapshotId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle delete snapshot request: missing snapshotID",
		)
	}

	return nil
}

func (cs *controller) validateControllerPublishVolumeReq(req *csi.ControllerPublishVolumeRequest) error {

	if req.GetVolumeId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing volume ID",
		)
	}

	if req.GetVolumeCapability() == nil {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing volume ID",
		)
	}

	_, err := utils.GetVolume(req.GetVolumeId())
	if err != nil {
		if k8serror.IsNotFound(err) {
			return status.Error(
				codes.NotFound,
				"failed to handle publish volume request: volumeNotPresent",
			)
		}
		return status.Error(
			codes.Internal, err.Error(),
		)
	}
	if req.GetNodeId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing targetPath",
		)
	}
	_, err = utils.GetNode(req.GetNodeId())
	if err != nil {
		if k8serror.IsNotFound(err) {
			return status.Error(
				codes.NotFound,
				"failed to handle publish volume request: volumeNotPresent",
			)
		}
		return status.Error(
			codes.Internal, err.Error(),
		)
	}

	return nil
}

func (ns *controller) validateControllerUnpublishVolumeReq(req *csi.ControllerUnpublishVolumeRequest) error {

	if req.GetVolumeId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing volume ID",
		)
	}

	if req.GetNodeId() == "" {
		return status.Error(
			codes.InvalidArgument,
			"failed to handle publish volume request: missing targetPath",
		)
	}

	return nil
}
