package driver

import (
	"fmt"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	apis "github.com/openebs/cstor-csi/pkg/apis/cstor/v1"
	"github.com/openebs/cstor-csi/pkg/cstor/volumeattachment"
	utils "github.com/openebs/cstor-csi/pkg/utils"
	"github.com/pkg/errors"
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
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
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
		if mode := volcap.GetAccessMode(); mode != nil {
			modeName := csi.VolumeCapability_AccessMode_Mode_name[int32(mode.GetMode())]
			// we only support SINGLE_NODE_WRITER
			if mode.GetMode() != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
				return status.Errorf(codes.InvalidArgument,
					"only SINGLE_NODE_WRITER supported, unsupported access mode requested: %s",
					modeName,
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

	// If the access type is block, do nothing for stage
	var accessType string
	switch req.GetVolumeCapability().GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		accessType = "block"
	case *csi.VolumeCapability_Mount:
		accessType = "mount"
	}

	vol, err := volumeattachment.NewBuilder().
		WithName(volumeID + "-" + nodeID).
		WithLabels(labels).
		WithVolName(req.GetVolumeId()).
		WithAccessType(accessType).
		WithFSType(req.GetVolumeCapability().GetMount().GetFsType()).
		WithReadOnly(req.GetReadonly()).Build()
	if err != nil {
		return err
	}
	if isCVCBound, err := utils.IsCVCBound(volumeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	} else if !isCVCBound {
		utils.TransitionVolList[volumeID] = apis.CStorVolumeAttachmentStatusWaitingForCVCBound
		time.Sleep(10 * time.Second)
		return errors.New("Waiting for CVC to be bound")
	}

	if err = utils.FetchAndUpdateISCSIDetails(volumeID, vol); err != nil {
		return err
	}

	oldvol, err := utils.GetCStorVolumeAttachment(vol.Name)
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	} else if err == nil && oldvol != nil {
		if oldvol.DeletionTimestamp != nil {
			return errors.Errorf("Volume still mounted on node: %s", nodeID)
		}
		return nil
	}

	if err = utils.DeleteOldCStorVolumeAttachmentCRs(volumeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if err = utils.CreateCStorVolumeAttachmentCR(vol, nodeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
