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
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	apis "github.com/openebs/cstor-csi/pkg/apis/openebs.io/core/v1alpha1"
	iscsiutils "github.com/openebs/cstor-csi/pkg/iscsi/v1alpha1"
	utils "github.com/openebs/cstor-csi/pkg/utils/v1alpha1"
	"golang.org/x/net/context"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// node is the server implementation
// for CSI NodeServer
type node struct {
	driver       *CSIDriver
	capabilities []*csi.NodeServiceCapability
	mounter      *utils.NodeMounter
}

// VolumeStatistics represents statistics information of a volume
type VolumeStatistics struct {
	availableBytes, totalBytes, usedBytes    int64
	availableInodes, totalInodes, usedInodes int64
}

// NewNode returns a new instance
// of CSI NodeServer
func NewNode(d *CSIDriver) csi.NodeServer {
	return &node{
		driver:       d,
		capabilities: newNodeCapabilities(),
		mounter:      utils.NewNodeMounter(),
	}
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

	resp := &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.capabilities,
	}

	return resp, nil
}

// NodeStageVolume mounts the volume on the staging path
func (ns *node) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {
	var (
		err             error
		vol             *apis.CSIVolume
		isMountRequired bool
	)

	if err = ns.validateNodeStageReq(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()

	err = addVolumeToTransitionList(volumeID, apis.CSIVolumeStatusUninitialized)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer removeVolumeFromTransitionList(volumeID)

	if vol, err = utils.GetCSIVolume(volumeID + "-" + utils.NodeIDENV); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err = utils.WaitForVolumeReadyAndReachable(vol); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	isMountRequired, err = VerifyIfMountRequired(volumeID, stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if isMountRequired {
		vol.Spec.Volume.StagingTargetPath = stagingTargetPath
		vol.Finalizers = []string{utils.NodeIDENV}
		vol.Spec.Volume.DevicePath = strings.Join([]string{
			"/dev/disk/by-path/ip", vol.Spec.ISCSI.TargetPortal,
			"iscsi", vol.Spec.ISCSI.Iqn, "lun", fmt.Sprint(0)}, "-",
		)
		err = utils.UpdateCSIVolumeCR(vol)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// Permission is changed for the local directory before the volume is
		// mounted on the node. This helps to resolve cases when the CSI driver
		// Unmounts the volume to remount again in required mount mode(ro/rw),
		// the app starts writing directly in the local directory.
		// As soon as the volume is mounted the permissions of this directory are
		// automatically changed to allow Reads and writes.
		// And as soon as it is unmounted permissions change
		// back to what we are setting over here.
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusMountUnderProgress
		// Login to the volume and attempt mount operation on the requested path
		devicePath, err := ns.attachDisk(vol)
		if err != nil {
			logrus.Errorf("NodeStageVolume: failed to attachDisk for volume %v, err: %v", volumeID, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		// If the access type is block, do nothing for stage
		switch req.GetVolumeCapability().GetAccessType().(type) {
		case *csi.VolumeCapability_Block:
			return &csi.NodeStageVolumeResponse{}, nil
		}

		if err := os.MkdirAll(stagingTargetPath, 0750); err != nil {
			logrus.Errorf("failed to mkdir %s, error: %v", stagingTargetPath, err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		logrus.Info("NodeStageVolume: start format and mount operation")
		if err := ns.formatAndMount(req, devicePath); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusMounted
	}

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
	var (
		err             error
		vol             *apis.CSIVolume
		unmountRequired bool
	)

	if err = ns.validateNodeUnStageReq(req); err != nil {
		return nil, err
	}

	stagingTargetPath := req.GetStagingTargetPath()
	volumeID := req.GetVolumeId()

	err = addVolumeToTransitionList(volumeID, apis.CSIVolumeStatusUninitialized)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer removeVolumeFromTransitionList(volumeID)

	if vol, err = utils.GetCSIVolume(volumeID + "-" + utils.NodeIDENV); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if vol.Spec.Volume.StagingTargetPath == "" {
		return &csi.NodeUnstageVolumeResponse{}, nil

	}
	unmountRequired, err = IsUnmountRequired(volumeID, stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if unmountRequired {
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
		if err = iscsiutils.UnmountAndDetachDisk(vol, stagingTargetPath); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		utils.TransitionVolList[volumeID] = apis.CSIVolumeStatusUnmounted
	}
	// It is safe to delete the CSIVolume CR now since the volume has already
	// been unmounted and logged out

	vol.Finalizers = nil
	vol.Spec.Volume.StagingTargetPath = ""
	if err = utils.UpdateCSIVolumeCR(vol); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	logrus.Infof("hostpath: volume %s path: %s has been unmounted.",
		volumeID, stagingTargetPath)

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume publishes (mounts) the volume
// at the corresponding node at a given path
//
// This implements csi.NodeServer
func (ns *node) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	volumeID := req.GetVolumeId()
	err := addVolumeToTransitionList(volumeID, apis.CSIVolumeStatusUninitialized)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer removeVolumeFromTransitionList(volumeID)

	mountOptions := []string{"bind"}
	if req.GetReadonly() {
		mountOptions = append(mountOptions, "ro")
	}
	vol, err := utils.GetCSIVolume(volumeID + "-" + utils.NodeIDENV)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	vol.Spec.Volume.TargetPath = req.GetTargetPath()
	if err = utils.UpdateCSIVolumeCR(vol); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	switch mode := req.GetVolumeCapability().GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		if err := ns.nodePublishVolumeForBlock(req, mountOptions); err != nil {
			return nil, err
		}
	case *csi.VolumeCapability_Mount:
		if err := ns.nodePublishVolumeForFileSystem(req, mountOptions, mode); err != nil {
			return nil, err
		}
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *node) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	volumeID := req.GetVolumeId()
	target := req.GetTargetPath()

	err := addVolumeToTransitionList(volumeID, apis.CSIVolumeStatusUninitialized)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer removeVolumeFromTransitionList(volumeID)

	notMnt, err := ns.mounter.IsLikelyNotMountPoint(target)
	if (err == nil && notMnt) || os.IsNotExist(err) {
		logrus.Warningf("NodeUnpublishVolume: %s is not mounted, err: %v", target, err)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	logrus.Infof("NodeUnpublishVolume: unmounting %s", target)
	if err := ns.mounter.Unmount(target); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not unmount %q: %v", target, err)
	}
	vol, err := utils.GetCSIVolume(volumeID + "-" + utils.NodeIDENV)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	vol.Spec.Volume.TargetPath = ""
	if err = utils.UpdateCSIVolumeCR(vol); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
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
	volumeID := req.GetVolumeId()
	err := addVolumeToTransitionList(volumeID, apis.CSIVolumeStatusResizeInProgress)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle NodeExpandVolumeRequest for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}
	defer removeVolumeFromTransitionList(volumeID)

	vol, err := utils.GetCSIVolume(volumeID + "-" + utils.NodeIDENV)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = iscsiutils.ResizeVolume(req.GetVolumePath(), vol.Spec.Volume.FSType); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle NodeExpandVolumeRequest for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}

	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
	}, nil
}

// NodeGetVolumeStats returns statistics for the given volume
func (ns *node) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {

	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume ID must be provided")
	}

	volumePath := req.GetVolumePath()
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume Path must be provided")
	}

	mounted, err := ns.mounter.ExistsPath(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if volume path %q is mounted: %s", volumePath, err)
	}

	if !mounted {
		return nil, status.Errorf(codes.NotFound, "volume path %q is not mounted", volumePath)
	}

	stats, err := ns.GetStatistics(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve capacity statistics for volume path %q: %s", volumePath, err)
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			&csi.VolumeUsage{
				Available: stats.availableBytes,
				Total:     stats.totalBytes,
				Used:      stats.usedBytes,
				Unit:      csi.VolumeUsage_BYTES,
			},
			&csi.VolumeUsage{
				Available: stats.availableInodes,
				Total:     stats.totalInodes,
				Used:      stats.usedInodes,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

// GetStatistics get the statistics for a given volume path
func (ns *node) GetStatistics(volumePath string) (VolumeStatistics, error) {
	var statfs unix.Statfs_t
	// See http://man7.org/linux/man-pages/man2/statfs.2.html for details.
	err := unix.Statfs(volumePath, &statfs)
	if err != nil {
		return VolumeStatistics{}, err
	}

	volStats := VolumeStatistics{
		availableBytes: int64(statfs.Bavail) * int64(statfs.Bsize),
		totalBytes:     int64(statfs.Blocks) * int64(statfs.Bsize),
		usedBytes:      (int64(statfs.Blocks) - int64(statfs.Bfree)) * int64(statfs.Bsize),

		availableInodes: int64(statfs.Ffree),
		totalInodes:     int64(statfs.Files),
		usedInodes:      int64(statfs.Files) - int64(statfs.Ffree),
	}

	return volStats, nil
}
