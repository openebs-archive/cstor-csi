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
	"strconv"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	apisv1 "github.com/openebs/api/v3/pkg/apis/cstor/v1"
	"github.com/openebs/cstor-csi/pkg/env"
	k8snode "github.com/openebs/cstor-csi/pkg/kubernetes/node"
	csipayload "github.com/openebs/cstor-csi/pkg/payload"
	utils "github.com/openebs/cstor-csi/pkg/utils"
	analytics "github.com/openebs/google-analytics-4/usage"
	errors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Ping event is sent periodically
	Ping string = "ping"
	// DefaultCASType Event application name constant for volume event
	DefaultCASType string = "cstor"
	// Replica Event replication
	Replica string = "replica:"
	// DefaultReplicaCount holds the replica count string
	DefaultReplicaCount string = "replica:3"
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

// TODO Implementation will be taken up later

// ValidateVolumeCapabilities validates the capabilities
// required to create a new volume
// This implements csi.ControllerServer
func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
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
	var (
		err        error
		nodeID     string
		snapshotID string
	)
	logrus.Infof("received request to create volume {%s}", req.GetName())

	if err = cs.validateVolumeCreateReq(req); err != nil {
		return nil, err
	}

	volName := req.GetName()
	size := req.GetCapacityRange().RequiredBytes
	rCount := req.GetParameters()["replicaCount"]
	cspcName := req.GetParameters()["cstorPoolCluster"]
	policyName := req.GetParameters()["cstorVolumePolicy"]
	VolumeContext := map[string]string{
		"openebs.io/cas-type": req.GetParameters()["cas-type"],
	}
	pvcName := req.GetParameters()[pvcNameKey]
	pvcNamespace := req.GetParameters()[pvcNamespaceKey]

	nodeID, err = getAccessibilityRequirements(req.GetAccessibilityRequirements())
	if err != nil {
		return nil, err
	}

	contentSource := req.GetVolumeContentSource()
	if contentSource != nil && contentSource.GetSnapshot() != nil {
		snapshotID = contentSource.GetSnapshot().GetSnapshotId()
		if snapshotID == "" {
			return nil, status.Error(codes.InvalidArgument, "snapshot ID is empty")
		}
		if isValidSrc, _ := utils.IsSourceAvailable(snapshotID); !isValidSrc {
			return nil, status.Error(
				codes.InvalidArgument,
				"VolumeSrc Not Available")
		}
	}

	// verify if the volume has already been created
	cvc, err := utils.GetVolume(volName)
	if err == nil && cvc != nil && cvc.DeletionTimestamp == nil {
		goto createVolumeResponse
	}

	err = utils.ProvisionVolume(size, volName, rCount,
		cspcName, snapshotID,
		nodeID, policyName, pvcName, pvcNamespace)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sendEventOrIgnore(pvcName, volName, strconv.FormatInt(int64(size), 10),
		rCount,
		"cstor-csi",
		analytics.VolumeProvision,
	)

createVolumeResponse:
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volName,
			CapacityBytes: size,
			VolumeContext: VolumeContext,
			ContentSource: contentSource,
		},
	}, nil
}

// DeleteVolume deletes the specified volume
func (cs *controller) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	logrus.Infof("received request to delete volume {%s}", req.VolumeId)

	var (
		err error
		cvc *apisv1.CStorVolumeConfig
	)

	if err = cs.validateDeleteVolumeReq(req); err != nil {
		return nil, err
	}

	volumeID := req.GetVolumeId()

	// verify if the volume has already been deleted
	cvc, err = utils.GetVolume(volumeID)
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

	sendEventOrIgnore(cvc.GetAnnotations()[utils.OpenebsPVC], volumeID, getCapacity(cvc),
		strconv.Itoa(cvc.Spec.Provision.ReplicaCount),
		"cstor-csi",
		analytics.VolumeDeprovision,
	)

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

// ControllerExpandVolume resizes previously provisioned volume
//
// This implements csi.ControllerServer
func (cs *controller) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
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

	snapTimeStamp := time.Now().Unix()
	if err := utils.CreateSnapshot(req.SourceVolumeId, req.Name); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle CreateSnapshotRequest for %s: %s, {%s}",
			req.SourceVolumeId, req.Name,
			err.Error(),
		)
	}
	return csipayload.NewCreateSnapshotResponseBuilder().
		WithSourceVolumeID(req.SourceVolumeId).
		WithSnapshotID(req.SourceVolumeId+"@"+req.Name).
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

	snapshotID := strings.Split(req.SnapshotId, "@")
	if len(snapshotID) != 2 {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle DeleteSnapshotRequest for %s, {%s}",
			req.SnapshotId,
			"Manual intervention required",
		)
	}
	if err := utils.DeleteSnapshot(snapshotID[0], snapshotID[1]); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle DeleteSnapshotRequest for %s, {%s}",
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

func getAccessibilityRequirements(requirement *csi.TopologyRequirement) (string, error) {
	if requirement == nil {
		return "", status.Error(codes.Internal, "accessibility_requirements not found")
	}

	node, err := getNode(requirement)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to get the accessibility_requirements node %v", err)
	}

	if len(node) == 0 {
		return "", status.Error(codes.Internal, "can not find any node")
	}
	return node, nil
}

// sendEventOrIgnore sends anonymous cstor provision/delete events
func sendEventOrIgnore(pvcName, pvName, capacity, replicaCount, stgType, method string) {
	if env.Truthy(env.OpenEBSEnableAnalytics) {
		analytics.New().CommonBuild(GetVolumeType(stgType, method)).ApplicationBuilder().
			SetVolumeName(pvName).
			SetVolumeClaimName(pvcName).
			SetLabel(analytics.EventLabelCapacity).
			SetAction(GetReplicaCount(stgType, method)).
			SetCategory(method).
			SetVolumeCapacity(capacity).Send()
	}
}

// getNode gets the node which satisfies the topology info
func getNode(topo *csi.TopologyRequirement) (string, error) {

	list, err := k8snode.NewKubeClient().List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, prf := range topo.Preferred {
		for _, node := range list.Items {
			nodeFiltered := false
			for key, value := range prf.Segments {
				if node.Labels[key] != value {
					nodeFiltered = true
					break
				}
			}
			if nodeFiltered == false {
				return node.Name, nil
			}
		}
	}
	return "", nil
}

func (cs *controller) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {

	volume, err := utils.GetCStorVolume(req.GetVolumeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get the cstorvolume %v", err)
	}

	logrus.Infof("Healthy state: %s Volume: %s", volume.Name, volume.Status.Phase)
	return &csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volume.Name,
			CapacityBytes: volume.Status.Capacity.MilliValue(),
		},
		Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
			VolumeCondition: getVolumeCondition(volume),
		},
	}, nil
}

// SetVolumeType Wrapper for setting the default storage-engine for volume-provision event
func GetVolumeType(volType, method string) string {
	if method == analytics.VolumeProvision && volType == "" {
		// Set the default storage engine, if not specified in the request
		return DefaultCASType
	} else {
		return volType
	}
}

// GetReplicaCount Wrapper for setting replica count for volume events
func GetReplicaCount(count, method string) string {
	if method == analytics.VolumeProvision && count == "" {
		// Case: When volume-provision the replica count isn't specified
		// it is set to three by default by the cstor-operators
		return DefaultReplicaCount
	} else {
		// Catch all case for volume-deprovision event and
		// volume-provision event with an overridden replica-count
		return Replica + count
	}
}
