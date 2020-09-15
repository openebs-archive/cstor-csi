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
package utils

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	apis "github.com/openebs/cstor-csi/pkg/apis/cstor/v1"
	"github.com/openebs/cstor-csi/pkg/cstor/snapshot"
	iscsiutils "github.com/openebs/cstor-csi/pkg/iscsi"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/utils/mount"
)

const (
	// TODO make VolumeWaitTimeout as env

	// VolumeWaitTimeout indicates the timegap between two consecutive volume
	// status check attempts
	VolumeWaitTimeout = 2

	// TODO make VolumeWaitRetryCount as env

	// VolumeWaitRetryCount indicates the number of retries made to check the
	// status of volume before erroring out
	VolumeWaitRetryCount = 6

	// TODO make MonitorMountRetryTimeout as env

	// MonitorMountRetryTimeout indicates the time gap between two consecutive
	//monitoring attempts
	MonitorMountRetryTimeout = 5

	// This environment variable is set via env
	GoogleAnalyticsKey string = "OPENEBS_IO_ENABLE_ANALYTICS"
)

var (
	// OpenEBSNamespace is openebs system namespace
	OpenEBSNamespace string

	// NodeIDENV is the NodeID of the node on which the pod is present
	NodeIDENV string

	// TransitionVolList contains the list of volumes under transition
	// This list is protected by TransitionVolListLock
	TransitionVolList map[string]apis.CStorVolumeAttachmentStatus

	// TransitionVolListLock is required to protect the above Volumes list
	TransitionVolListLock sync.RWMutex

	// ReqMountList contains the list of volumes which are required
	// to be remounted. This list is secured by ReqMountListLock
	ReqMountList map[string]apis.CStorVolumeAttachmentStatus
)

const (
	// timmeout indiactes the REST call timeouts
	timeout = 60 * time.Second
)

func init() {

	OpenEBSNamespace = os.Getenv("OPENEBS_NAMESPACE")
	if OpenEBSNamespace == "" {
		logrus.Fatalf("OPENEBS_NAMESPACE environment variable not set")
	}
	NodeIDENV = os.Getenv("OPENEBS_NODE_ID")
	if NodeIDENV == "" && os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		logrus.Fatalf("OPENEBS_NODE_ID not set")
	}

	TransitionVolList = make(map[string]apis.CStorVolumeAttachmentStatus)
	ReqMountList = make(map[string]apis.CStorVolumeAttachmentStatus)

}

// parseEndpoint should have a valid prefix(unix/tcp)
// to return a valid endpoint parts
func parseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") ||
		strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("Invalid endpoint: %v", ep)
}

// logGRPC logs all the grpc related errors, i.e the final errors
// which are returned to the grpc clients
func logGRPC(
	ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	logrus.Debugf("GRPC call: %s", info.FullMethod)
	logrus.Debugf("GRPC request: %s", protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		logrus.Errorf("GRPC error: %v", err)
	} else {
		logrus.Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}

// ChmodMountPath removes all permission from the folder if volume is not
// mounted on it
func ChmodMountPath(mountPath string) error {
	return os.Chmod(mountPath, 0000)
}

// WaitForVolumeToBeReachable keeps the mounts on hold until the volume is
// reachable
func WaitForVolumeToBeReachable(targetPortal string) error {
	var (
		retries int
		err     error
		conn    net.Conn
	)

	for {
		// Create a connection to test if the iSCSI Portal is reachable,
		if conn, err = net.Dial("tcp", targetPortal); err == nil {
			conn.Close()
			logrus.Infof("Volume is reachable to create connections")
			return nil
		}
		// wait until the iSCSI targetPortal is reachable
		// There is no pointn of triggering iSCSIadm login commands
		// until the portal is reachable
		time.Sleep(VolumeWaitTimeout * time.Second)
		retries++
		if retries >= VolumeWaitRetryCount {
			// Let the caller function decide further if the volume is
			// not reachable even after 12 seconds ( This number was arrived at
			// based on the kubelets retrying logic. Kubelet retries to publish
			// volume after every 14s )
			return fmt.Errorf(
				"iSCSI Target not reachable, TargetPortal %v, err:%v",
				targetPortal, err)
		}
	}
}

// WaitForVolumeToBeReady retrieves the volume info from cstorVolume CR and
// waits until consistency factor is met for connected replicas
func WaitForVolumeToBeReady(volumeID string) error {
	var retries int
checkVolumeStatus:
	// Status is fetched from cstorVolume CR
	volStatus, err := getVolStatus(volumeID)
	if err != nil {
		return err
	} else if volStatus == "Healthy" || volStatus == "Degraded" {
		// In both healthy and degraded states the volume can serve IOs
		logrus.Infof("Volume is ready to accept IOs")
	} else if retries >= VolumeWaitRetryCount {
		// Let the caller function decide further if the volume is still not
		// ready to accdept IOs after 12 seconds ( This number was arrived at
		// based on the kubelets retrying logic. Kubelet retries to publish
		// volume after every 14s )
		return fmt.Errorf(
			"Volume is not ready: Replicas yet to connect to controller",
		)
	} else {
		TransitionVolList[volumeID] = apis.CStorVolumeAttachmentStatusWaitingForVolumeToBeReady
		time.Sleep(VolumeWaitTimeout * time.Second)
		retries++
		goto checkVolumeStatus
	}
	return nil
}

/*
// GetVolumeByName fetches the volume from Volumes list based on th input name
func GetVolumeByName(volName string) (*apis.CStorVolumeAttachment, error) {
	for _, Vol := range Volumes {
		if Vol.Spec.Volume.Name == volName {
			return Vol, nil
		}
	}
	return nil,
		fmt.Errorf("volume name %s does not exit in the volumes list", volName)
}
*/
func listContains(
	mountPath string, list []mount.MountPoint,
) (*mount.MountPoint, bool) {
	for _, info := range list {
		if info.Path == mountPath {
			mntInfo := info
			return &mntInfo, true
		}
	}
	return nil, false
}

// MonitorMounts makes sure that all the volumes present in the inmemory list
// with the driver are mounted with the original mount options
// This function runs a never ending loop therefore should be run as a goroutine
// Mounted list is fetched from the OS and the state of all the volumes is
// reverified after every 5 seconds. If the mountpoint is not present in the
// list or if it has been remounted with a different mount option by the OS, the
// volume is added to the ReqMountList which is removed as soon as the remount
// operation on the volume is complete
// For each remount operation a new goroutine is created, so that if multiple
// volumes have lost their original state they can all be remounted in parallel
func MonitorMounts() {
	var (
		err        error
		csivolList *apis.CStorVolumeAttachmentList
		mountList  []mount.MountPoint
	)
	mounter := mount.New("")
	ticker := time.NewTicker(MonitorMountRetryTimeout * time.Second)
	for {
		select {
		case <-ticker.C:
			// Get list of mounted paths present with the node
			TransitionVolListLock.Lock()
			if mountList, err = mounter.List(); err != nil {
				TransitionVolListLock.Unlock()
				break
			}
			if csivolList, err = GetVolListForNode(); err != nil {
				TransitionVolListLock.Unlock()
				break
			}
			for _, vol := range csivolList.Items {
				// ignore monitoring for volumes with deletion timestamp set
				if vol.DeletionTimestamp != nil {
					continue
				}
				// ignore monitoring the mount for a block device
				if vol.Spec.Volume.AccessType == "block" {
					continue
				}
				// This check is added to avoid monitoring volume if it has not
				// been mounted yet. Although CStorVolumeAttachment CR gets created at
				// ControllerPublish step.
				if (vol.Spec.Volume.StagingTargetPath == "") ||
					(vol.Spec.Volume.TargetPath == "") {
					continue
				}
				// Search the volume in the list of mounted volumes at the node
				// retrieved above
				stagingMountPoint, stagingPathExists := listContains(
					vol.Spec.Volume.StagingTargetPath, mountList,
				)
				_, targetPathExists := listContains(
					vol.Spec.Volume.TargetPath, mountList,
				)
				// If the volume is present in the list verify its state
				// If stagingPath is in rw then TargetPath will also be in rw
				// mode
				if stagingPathExists && targetPathExists &&
					verifyMountOpts(stagingMountPoint.Opts, "rw") {
					// Continue with remaining volumes since this volume looks
					// to be in good shape
					continue
				}
				if _, ok := TransitionVolList[vol.Spec.Volume.Name]; !ok {
					csivol := vol
					TransitionVolList[csivol.Spec.Volume.Name] = csivol.Status
					ReqMountList[csivol.Spec.Volume.Name] = csivol.Status
					go func(csivol apis.CStorVolumeAttachment) {
						logrus.Infof("Remounting vol: %s at %s and %s",
							csivol.Spec.Volume.Name, csivol.Spec.Volume.StagingTargetPath,
							csivol.Spec.Volume.TargetPath)
						defer func() {
							TransitionVolListLock.Lock()
							// Remove the volume from ReqMountList once the remount operation is
							// complete
							delete(TransitionVolList, csivol.Spec.Volume.Name)
							delete(ReqMountList, csivol.Spec.Volume.Name)
							TransitionVolListLock.Unlock()
						}()
						if err := RemountVolume(
							stagingPathExists, targetPathExists,
							&csivol,
						); err != nil {
							logrus.Errorf(
								"Remount failed for vol: %s : err: %v",
								csivol.Spec.Volume.Name, err,
							)
						} else {
							logrus.Infof(
								"Remount successful for vol: %s",
								csivol.Spec.Volume.Name,
							)
						}
					}(csivol)
				}
			}
			TransitionVolListLock.Unlock()
		}
	}
}

// CleanupOnRestart unmounts and detaches the volumes having
// DeletionTimestamp set and removes finalizers from the
// corresponding CStorVolumeAttachment CRs
func CleanupOnRestart() {
	var (
		err        error
		csivolList *apis.CStorVolumeAttachmentList
	)
	// Get list of mounted paths present with the node
	TransitionVolListLock.Lock()
	defer TransitionVolListLock.Unlock()
	if csivolList, err = GetVolListForNode(); err != nil {
		return
	}
	for _, Vol := range csivolList.Items {
		if Vol.DeletionTimestamp == nil {
			continue
		}
		vol := Vol
		TransitionVolList[vol.Spec.Volume.Name] = apis.CStorVolumeAttachmentStatusUnmountUnderProgress
		// This is being run in a go routine so that if unmount and detach
		// commands take time, the startup is not delayed
		go func(vol *apis.CStorVolumeAttachment) {
			if err := iscsiutils.UnmountAndDetachDisk(vol, vol.Spec.Volume.StagingTargetPath); err == nil {
				vol.Finalizers = nil
				if vol, err = UpdateCStorVolumeAttachmentCR(vol); err != nil {
					logrus.Errorf(err.Error())
				}
			} else {
				logrus.Errorf(err.Error())
			}

			TransitionVolListLock.Lock()
			TransitionVolList[vol.Spec.Volume.Name] = apis.CStorVolumeAttachmentStatusUnmounted
			delete(TransitionVolList, vol.Spec.Volume.Name)
			TransitionVolListLock.Unlock()
		}(&vol)
	}
}

// IsVolumeReachable makes a TCP connection to target
// and checks if volume is Reachable
func IsVolumeReachable(targetPortal string) (bool, error) {
	var (
		err  error
		conn net.Conn
	)

	// Create a connection to test if the iSCSI Portal is reachable,
	if conn, err = net.Dial("tcp", targetPortal); err == nil {
		conn.Close()
		logrus.Infof("Volume is reachable to create connections")
		return true, nil
	}
	logrus.Infof(
		"iSCSI Target not reachable, TargetPortal %v, err:%v",
		targetPortal, err,
	)
	return false, err
}

// IsVolumeReady retrieves the volume info from cstorVolume CR and
// verifies if consistency factor is met for connected replicas
func IsVolumeReady(volumeID string) (bool, error) {
	volStatus, err := getVolStatus(volumeID)
	if err != nil {
		return false, err
	}
	if volStatus == "Healthy" || volStatus == "Degraded" {
		// In both healthy and degraded states the volume can serve IOs
		logrus.Infof("Volume is ready to accept IOs")
		return true, nil
	}
	logrus.Infof("Volume is not ready: Replicas yet to connect to controller")
	return false, nil
}

// WaitForVolumeReadyAndReachable waits until the volume is ready to accept IOs
// and is reachable, this function will not come out until both the conditions
// are met. This function stops the driver from overloading the OS with iSCSI
// login commands.
func WaitForVolumeReadyAndReachable(vol *apis.CStorVolumeAttachment) error {
	var err error
	// This function return after 12s in case the volume is not ready
	if err = WaitForVolumeToBeReady(vol.Spec.Volume.Name); err != nil {
		logrus.Error(err)
		return err
	}
	// This function return after 12s in case the volume is not reachable
	err = WaitForVolumeToBeReachable(vol.Spec.ISCSI.TargetPortal)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func verifyMountOpts(opts []string, desiredOpt string) bool {
	for _, opt := range opts {
		if opt == desiredOpt {
			return true
		}
	}
	return false
}

// RemountVolume unmounts the volume if it is already mounted in an undesired
// state and then tries to mount again. If it is not mounted the volume, first
// the disk will be attached via iSCSI login and then it will be mounted
func RemountVolume(
	stagingPathExists bool, targetPathExists bool,
	vol *apis.CStorVolumeAttachment,
) error {
	mounter := mount.New("")
	options := []string{"rw"}

	if ready, err := IsVolumeReady(vol.Spec.Volume.Name); err != nil || !ready {
		return fmt.Errorf("Volume is not ready")
	}
	if reachable, err := IsVolumeReachable(vol.Spec.ISCSI.TargetPortal); err != nil || !reachable {
		return fmt.Errorf("Volume is not reachable")
	}
	if stagingPathExists {
		mounter.Unmount(vol.Spec.Volume.StagingTargetPath)
	}
	if targetPathExists {
		mounter.Unmount(vol.Spec.Volume.TargetPath)
	}

	// Unmount and mount operation is performed instead of just remount since
	// the remount option didn't give the desired results
	if err := mounter.Mount(vol.Spec.Volume.DevicePath,
		vol.Spec.Volume.StagingTargetPath, "", options,
	); err != nil {
		return err
	}
	options = []string{"bind"}
	err := mounter.Mount(vol.Spec.Volume.StagingTargetPath,
		vol.Spec.Volume.TargetPath, "", options)
	return err
}

// GetMounts gets mountpoints for the specified volume
func GetMounts(volumeID string) ([]string, error) {

	var (
		currentMounts []string
		err           error
		mountList     []mount.MountPoint
	)
	mounter := mount.New("")
	// Get list of mounted paths present with the node
	if mountList, err = mounter.List(); err != nil {
		return nil, err
	}
	for _, mntInfo := range mountList {
		if strings.Contains(mntInfo.Path, volumeID) {
			currentMounts = append(currentMounts, mntInfo.Path)
		}
	}
	return currentMounts, nil
}

// CreateSnapshot creates a snapshot of cstor volume
func CreateSnapshot(volumeName, snapName string) error {
	volIP, err := GetVolumeIP(volumeName)
	if err != nil {
		return err
	}
	_, err = snapshot.CreateSnapshot(volIP, volumeName, snapName)
	// If there is no err that means call was successful
	return err
}

// DeleteSnapshot deletes a snapshot of cstor volume
func DeleteSnapshot(volumeName, snapName string) error {
	volIP, err := GetVolumeIP(volumeName)
	if err != nil {
		return err
	}
	_, err = snapshot.DestroySnapshot(volIP, volumeName, snapName)
	return err
}
