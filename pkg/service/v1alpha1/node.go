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
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	apis "github.com/openebs/csi/pkg/apis/openebs.io/core/v1alpha1"
	iscsi "github.com/openebs/csi/pkg/iscsi/v1alpha1"
	"github.com/openebs/csi/pkg/utils/v1alpha1"
	csivol "github.com/openebs/csi/pkg/volume/v1alpha1"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// node is the server implementation
// for CSI NodeServer
type node struct {
	driver *CSIDriver
}

// NewNode returns a new instance
// of CSI NodeServer
func NewNode(d *CSIDriver) csi.NodeServer {
	return &node{
		driver: d,
	}
}

func prepareVolSpecAndWaitForVolumeReady(
	req *csi.NodePublishVolumeRequest,
	nodeID string,
) (*apis.CSIVolume, error) {
	volumeID := req.GetVolumeId()
	labels := map[string]string{
		"nodeID": nodeID,
	}

	vol, err := csivol.NewBuilder().
		WithName(req.GetVolumeId()).
		WithLabels(labels).
		WithVolName(req.GetVolumeId()).
		WithMountPath(req.GetTargetPath()).
		WithFSType(req.GetVolumeCapability().GetMount().GetFsType()).
		WithReadOnly(req.GetReadonly()).Build()
	if err != nil {
		return nil, err
	}

	if err = utils.PatchCVCNodeID(volumeID, nodeID); err != nil {
		return nil, err
	}

	if isCVCBound, err := utils.IsCVCBound(volumeID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else if !isCVCBound {
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusWaitingForCVCBound
		time.Sleep(10 * time.Second)
		return nil, errors.New("Waiting for CVC to be bound")
	}

	if err = utils.FetchAndUpdateISCSIDetails(volumeID, vol); err != nil {
		return nil, err
	}

	//Check if volume is ready to serve IOs,
	//info is fetched from the cstorvolume CR
	if err := utils.WaitForVolumeToBeReady(volumeID); err != nil {
		return nil, err
	}

	// A temporary TCP connection is made to the volume to check if its
	// reachable
	if err := utils.WaitForVolumeToBeReachable(
		vol.Spec.ISCSI.TargetPortal,
	); err != nil {
		return nil,
			status.Error(codes.Internal, err.Error())
	}
	return vol, nil
}

func removeVolumeFromTransitionList(volumeID string) {
	utils.TransitionVolListLock.Lock()
	defer utils.TransitionVolListLock.Unlock()
	delete(utils.TransitionVolList, volumeID)
}

func addVolumeToTransitionList(volumeID string, status apis.CSIVolumeStatus) error {
	utils.TransitionVolListLock.Lock()
	defer utils.TransitionVolListLock.Unlock()

	if _, ok := utils.TransitionVolList[volumeID]; ok {
		return fmt.Errorf("Volume Busy, status: %v",
			utils.TransitionVolList[volumeID])
	}
	utils.TransitionVolList[volumeID] = status
	return nil
}

// VerifyIfMountRequired returns true if volume is already mounted on targetPath
// and unmounts if it is mounted on a different path
func VerifyIfMountRequired(volumeID, targetPath string) (bool, error) {
	var (
		currentMounts []string
		err           error
	)
	currentMounts, err = utils.GetMounts(volumeID)
	if err != nil {
		return false, err
	}
	if len(currentMounts) > 1 {
		logrus.Fatalf(
			"More than one mounts for volume:%s mounts: %v",
			volumeID, currentMounts,
		)
	}
	if len(currentMounts) == 1 {
		if currentMounts[0] == targetPath {
			return false, nil
		}
		if err = iscsi.Unmount(currentMounts[0]); err != nil {
			return false, err
		}
	}
	return true, nil
}

// NodePublishVolume publishes (mounts) the volume
// at the corresponding node at a given path
//
// This implements csi.NodeServer
func (ns *node) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	var (
		err             error
		devicePath      string
		vol             *apis.CSIVolume
		isMountRequired bool
	)

	if err = ns.validateNodePublishReq(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	nodeID := ns.driver.config.NodeID

	err = addVolumeToTransitionList(volumeID, apis.CSIVolumeStatusUninitialized)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer removeVolumeFromTransitionList(volumeID)

	if vol, err = prepareVolSpecAndWaitForVolumeReady(req, nodeID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	isMountRequired, err = VerifyIfMountRequired(volumeID, targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if isMountRequired {
		// Permission is changed for the local directory before the volume is
		// mounted on the node. This helps to resolve cases when the CSI driver
		// Unmounts the volume to remount again in required mount mode(ro/rw),
		// the app starts writing directly in the local directory.
		// As soon as the volume is mounted the permissions of this directory are
		// automatically changed to allow Reads and writes.
		// And as soon as it is unmounted permissions change
		// back to what we are setting over here.
		if err = utils.ChmodMountPath(vol.Spec.Volume.MountPath); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusMountUnderProgress
		// Login to the volume and attempt mount operation on the requested path
		if devicePath, err = iscsi.AttachAndMountDisk(vol); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// TODO update status also in below function
		vol.Spec.Volume.DevicePath = devicePath
		vol.Status = apis.CSIVolumeStatusMounted
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusMounted
	}

	err = utils.CreateOrUpdateCSIVolumeCR(vol)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

// IsUnmountRequired returns true if the volume needs to be unmounted
func IsUnmountRequired(volumeID, targetPath string) (bool, error) {
	var (
		currentMounts []string
		err           error
	)
	currentMounts, err = utils.GetMounts(volumeID)
	if err != nil {
		return false, err
	}
	if len(currentMounts) > 1 {
		logrus.Fatalf(
			"More than one mounts for volume:%s mounts: %v",
			volumeID, currentMounts,
		)
	}
	if len(currentMounts) == 0 {
		return false, nil
	}
	if currentMounts[0] != targetPath {
		return false, err
	}
	return true, nil
}

// NodeUnpublishVolume unpublishes (unmounts) the volume
// from the corresponding node from the given path
//
// This implements csi.NodeServer
func (ns *node) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	var (
		err             error
		vol             *apis.CSIVolume
		unmountRequired bool
	)

	if err = ns.validateNodeUnpublishReq(req); err != nil {
		return nil, err
	}

	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	err = addVolumeToTransitionList(volumeID, apis.CSIVolumeStatusUninitialized)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer removeVolumeFromTransitionList(volumeID)

	unmountRequired, err = IsUnmountRequired(volumeID, targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if unmountRequired {
		if vol, err = utils.GetCSIVolume(volumeID); (err != nil) || (vol == nil) {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// if node driver restarts before this step Kubelet will trigger the
		// NodeUnpublish command again so there is no need to worry that when this
		// driver restarts it will pick up the CSIVolume CR and start monitoring
		// mount point again.
		// If the node is down for some time, other node driver will first delete
		// this node's CSIVolume CR and then only will start its mount process.
		// If there is a case that this node comes up and CSIVolume CR is picked and
		// this node starts monitoring the mount point while the other node is also
		// trying to mount which appears to be a race condition but is not since
		// first of  all this CR will be marked for deletion when the other node
		// starts mounting. But lets say this node started monitoring and
		// immediately other node deleted this node's CR, in that case iSCSI
		// target(istgt) will pick up the new one and allow only that node to login,
		// so all the cases are handled
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusUnmountUnderProgress
		if err = iscsi.UnmountAndDetachDisk(vol, req.GetTargetPath()); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusUnmounted
	}
	// It is safe to delete the CSIVolume CR now since the volume has already
	// been unmounted and logged out
	if err = utils.DeleteCSIVolumeCRForPath(volumeID, targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	logrus.Infof("hostpath: volume %s path: %s has been unmounted.",
		volumeID, targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetInfo returns node details
//
// This implements csi.NodeServer
func (ns *node) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {

	return &csi.NodeGetInfoResponse{
		NodeId: ns.driver.config.NodeID,
	}, nil
}

// NodeGetCapabilities returns capabilities supported
// by this node service
//
// This implements csi.NodeServer
func (ns *node) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

// TODO
// This needs to be implemented
//
// NodeStageVolume mounts the volume on the staging
// path
//
// This implements csi.NodeServer
func (ns *node) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {

	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unmounts the volume from
// the staging path
//
// This implements csi.NodeServer
func (ns *node) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// TODO
// Verify if this needs to be implemented
//
// NodeExpandVolume resizes the filesystem if required
//
// If ControllerExpandVolumeResponse returns true in
// node_expansion_required then FileSystemResizePending
// condition will be added to PVC and NodeExpandVolume
// operation will be queued on kubelet
//
// This implements csi.NodeServer
func (ns *node) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {

	return nil, nil
}

// NodeGetVolumeStats returns statistics for the
// given volume
//
// This implements csi.NodeServer
func (ns *node) NodeGetVolumeStats(
	ctx context.Context,
	in *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *node) validateNodePublishReq(
	req *csi.NodePublishVolumeRequest,
) error {
	if req.GetVolumeCapability() == nil {
		return status.Error(codes.InvalidArgument,
			"Volume capability missing in request")
	}

	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}
	return nil
}

func (ns *node) validateNodeUnpublishReq(
	req *csi.NodeUnpublishVolumeRequest,
) error {
	if req.GetVolumeId() == "" {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}
	return nil
}
