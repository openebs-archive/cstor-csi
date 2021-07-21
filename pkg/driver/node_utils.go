/*
 Copyright © 2020 The OpenEBS Authors

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

package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-iscsi/iscsi"
	apis "github.com/openebs/api/v2/pkg/apis/cstor/v1"
	"github.com/openebs/cstor-csi/pkg/cstor/volumeattachment"
	iscsiutils "github.com/openebs/cstor-csi/pkg/iscsi"
	k8snode "github.com/openebs/cstor-csi/pkg/kubernetes/node"
	utils "github.com/openebs/cstor-csi/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

func getTargetIP(url string) string {
	s := strings.Split(url, ":")
	//ip, port := s[0], s[1]
	return s[0]
}

func (ns *node) attachDisk(vol *apis.CStorVolumeAttachment) (string, error) {
	connector := iscsi.Connector{
		VolumeName: vol.Spec.Volume.Name,
		Targets: []iscsi.TargetInfo{
			iscsi.TargetInfo{
				Iqn:    vol.Spec.ISCSI.Iqn,
				Portal: getTargetIP(vol.Spec.ISCSI.TargetPortal),
			},
		},
		Lun:         defaultISCSILUN,
		Interface:   defaultISCSIInterface,
		DoDiscovery: true,
	}

	logrus.Debugf("NodeStageVolume: attach disk with config: {%+v}", connector)
	devicePath, err := iscsi.Connect(connector)
	if err != nil {
		return "", err
	}

	if devicePath == "" {
		return "", fmt.Errorf("connect reported success, but no path returned")
	}
	return devicePath, err
}

func (ns *node) formatAndMount(req *csi.NodeStageVolumeRequest, devicePath string) error {
	// Mount device
	mntPath := req.GetStagingTargetPath()
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(mntPath)
	if err != nil && !os.IsNotExist(err) {
		if err := os.MkdirAll(mntPath, 0750); err != nil {
			logrus.Errorf("failed to mkdir %s, error", mntPath)
			return err
		}
	}

	if !notMnt {
		logrus.Infof("Volume %s has been mounted already at %v", req.GetVolumeId(), mntPath)
		return nil
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	options := []string{}
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
	options = append(options, mountFlags...)

	err = ns.mounter.FormatAndMount(devicePath, mntPath, fsType, options)
	if err != nil {
		logrus.Errorf(
			"Failed to mount iscsi volume %s [%s, %s] to %s, error %v",
			req.GetVolumeId(), devicePath, fsType, mntPath, err,
		)
		return err
	}
	return nil
}

func (ns *node) nodePublishVolumeForFileSystem(req *csi.NodePublishVolumeRequest, mountOptions []string, mode *csi.VolumeCapability_Mount) error {
	target := req.GetTargetPath()
	source := req.GetStagingTargetPath()
	if m := mode.Mount; m != nil {
		hasOption := func(options []string, opt string) bool {
			for _, o := range options {
				if o == opt {
					return true
				}
			}
			return false
		}
		for _, f := range m.MountFlags {
			if !hasOption(mountOptions, f) {
				mountOptions = append(mountOptions, f)
			}
		}
	}

	logrus.Infof("NodePublishVolume: creating dir %s", target)
	if err := os.MkdirAll(target, 0000); err != nil {
		return status.Errorf(codes.Internal, "Could not create dir {%q}, err: %v", target, err)
	}

	// in case if the dir already exists, above call returns nil
	// so permission needs to be updated
	if err := os.Chmod(target, 0000); err != nil {
		return status.Errorf(codes.Internal, "Could not change mode of dir {%q}, err: %v", target, err)
	}
	fsType := mode.Mount.GetFsType()
	if len(fsType) == 0 {
		fsType = defaultFsType
	}

	logrus.Infof("NodePublishVolume: mounting %s at %s with option %s as fstype %s", source, target, mountOptions, fsType)
	if err := ns.mounter.Mount(source, target, fsType, mountOptions); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, err)
		}
		return status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, target, err)
	}

	return nil
}

func (ns *node) nodePublishVolumeForBlock(req *csi.NodePublishVolumeRequest, mountOptions []string) error {
	target := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	source, err := ns.GetDevicePath(target, volumeID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to find device path %s. %v", target, err)
	}

	logrus.Debugf("NodePublishVolume [block]: find device path %s -> %s", source, source)

	globalMountPath := filepath.Dir(target)

	// create the global mount path if it is missing
	// Path in the form of /var/lib/kubelet/plugins/kubernetes.io/csi/volumeDevices/publish/{volumeName}
	exists, err := ns.mounter.ExistsPath(globalMountPath)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not check if path exists %q: %v", globalMountPath, err)
	}

	if !exists {
		if err := ns.mounter.MakeDir(globalMountPath); err != nil {
			return status.Errorf(codes.Internal, "Could not create dir %q: %v", globalMountPath, err)
		}
	}

	// Create the mount point as a file since bind mount device node requires it to be a file
	logrus.Debugf("NodePublishVolume [block]: making target file %s", target)
	err = ns.mounter.MakeFile(target)
	if err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, removeErr)
		}
		return status.Errorf(codes.Internal, "Could not create file %q: %v", target, err)
	}

	logrus.Debugf("NodePublishVolume [block]: mounting %s at %s", source, target)
	if err := ns.mounter.Mount(source, target, "", mountOptions); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, removeErr)
		}
		return status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, target, err)
	}
	return nil
}

// GetDevicePath get path of device and verifies its existence
func (ns *node) GetDevicePath(devicePath, volumeID string) (string, error) {

	vol, err := utils.GetCStorVolumeAttachment(volumeID + "-" + utils.NodeIDENV)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}
	return vol.Spec.Volume.DevicePath, nil
}

// newNodeCapabilities returns a list
// of this Node's capabilities
func newNodeCapabilities() []*csi.NodeServiceCapability {
	fromType := func(
		cap csi.NodeServiceCapability_RPC_Type,
	) *csi.NodeServiceCapability {
		return &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var capabilities []*csi.NodeServiceCapability
	for _, cap := range []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
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
	if len(currentMounts) > 2 {
		logrus.Warningf(
			"Unexpected mounts for volume:%s mounts: %v",
			volumeID, currentMounts,
		)
	}
	if len(currentMounts) == 0 {
		return false, nil
	}

	for _, mounts := range currentMounts {
		if strings.Contains(mounts, targetPath) {
			return true, nil
		}
	}
	return false, nil
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
	if len(currentMounts) > 2 {
		logrus.Warningf(
			"Unexpected mounts for volume:%s mounts: %v",
			volumeID, currentMounts,
		)
	}

	if len(currentMounts) >= 1 {
		for _, mounts := range currentMounts {
			if strings.Contains(mounts, targetPath) {
				return false, nil
			}
		}
		if err = iscsiutils.Unmount(currentMounts[0]); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (ns *node) validateNodeStageReq(
	req *csi.NodeStageVolumeRequest,
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

func (ns *node) validateNodeUnStageReq(
	req *csi.NodeUnstageVolumeRequest,
) error {
	if req.GetVolumeId() == "" {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if req.GetStagingTargetPath() == "" {
		return status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}
	return nil
}

func (ns *node) prepareVolumeForNode(
	req *csi.NodeStageVolumeRequest,
) error {
	volumeID := req.GetVolumeId()
	nodeID := ns.driver.config.NodeID

	// As of now Kubernetes will alow multiple pods to consume same
	// volume(from different nodes) eventhough volume is marked for
	// RWO this happens due to disabiling creation of VolumeAttachments.
	// To handle this case we added checks to restrict pods from different
	// nodes consuming same volume
	//
	// 1. List all the CStorVolumeAttachments(CVA) related to requested volume
	//    1.1 If CVA exist and corresponding node is stll running then return an error
	//
	//
	// LIMITATIONS:
	//				1. Canary deployments model & Rolling update strategy will not work pods
	//                 will reamin in pending state.
	//				2. Multiple pod instances of same (or) different deployments can run on same
	//				   node(If we add checks then rolling update strategy will never work).

	existingCSIVols, err := utils.GetVolList(volumeID)
	if err != nil {
		return err
	}
	for _, csiVol := range existingCSIVols.Items {
		if csiVol.Name == volumeID+"-"+nodeID {
			// In older Kubernetes version Kubelet will send NodeStage &
			// Unstage request even by deleting the pod to honor it we are
			// allowing login only after cleanup of old pod
			if csiVol.DeletionTimestamp != nil {
				return errors.Errorf("Volume %s still mounted on node: %s", volumeID, nodeID)
			}
			// This is a case where after creation of CVA if login/attachment/mount
			// operation failed during in next reconciliation things should work smooth
			continue
		}
		oldNodeName := csiVol.GetLabels()["nodeID"]

		if oldNodeName == nodeID {
			return errors.Errorf("Volume %s still mounted on node: %s", volumeID, oldNodeName)
		}

		isNodeReady, err := k8snode.IsNodeReady(oldNodeName)
		if err != nil && !k8serror.IsNotFound(err) {
			logrus.Errorf("failed to get the node %s details error: %s", oldNodeName, err.Error())
			return errors.Wrapf(err, "failed to get node %s details to know previous mounts", oldNodeName)
		} else if err == nil && isNodeReady {
			return errors.Errorf("Volume %s still mounted on node %s", volumeID, oldNodeName)
		}
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
		WithReadOnly(false).Build()
	if err != nil {
		return err
	}
	if isCVCBound, err := utils.IsCVCBound(volumeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	} else if !isCVCBound {
		utils.TransitionVolList[volumeID] = apis.CStorVolumeAttachmentStatusWaitingForCVCBound
		time.Sleep(10 * time.Second)
		return errors.Errorf("Waiting for %s's CVC to be bound", volumeID)
	}

	if err = utils.FetchAndUpdateISCSIDetails(volumeID, vol); err != nil {
		return err
	}

	if err = utils.DeleteOldCStorVolumeAttachmentCRs(volumeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if err = utils.CreateCStorVolumeAttachmentCR(vol, nodeID); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
