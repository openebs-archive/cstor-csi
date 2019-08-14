package utils

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	apis "github.com/openebs/csi/pkg/apis/openebs.io/core/v1alpha1"
	iscsi "github.com/openebs/csi/pkg/iscsi/v1alpha1"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/util/mount"
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
)

var (
	// OpenEBSNamespace is openebs system namespace
	OpenEBSNamespace string

	// NodeID is the NodeID of the node on which the pod is present
	NodeID string

	// TransitionVolList contains the list of volumes under transition
	// This list is protected by TransitionVolListLock
	TransitionVolList map[string]apis.CSIVolumeStatus

	// TransitionVolListLock is required to protect the above Volumes list
	TransitionVolListLock sync.RWMutex

	// ReqMountList contains the list of volumes which are required
	// to be remounted. This list is secured by ReqMountListLock
	ReqMountList map[string]apis.CSIVolumeStatus
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
	NodeID = os.Getenv("OPENEBS_NODE_ID")
	if NodeID == "" && os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		logrus.Fatalf("NodeID environment variable not set")
	}

	TransitionVolList = make(map[string]apis.CSIVolumeStatus)
	ReqMountList = make(map[string]apis.CSIVolumeStatus)

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
	logrus.Infof("GRPC call: %s", info.FullMethod)
	logrus.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
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
		time.Sleep(VolumeWaitTimeout * time.Second)
		retries++
		goto checkVolumeStatus
	}
	return nil
}

/*
// GetVolumeByName fetches the volume from Volumes list based on th input name
func GetVolumeByName(volName string) (*apis.CSIVolume, error) {
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
		csivolList *apis.CSIVolumeList
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
				break
			}
			if csivolList, err = GetVolList(NodeID); err != nil {
				break
			}
			for _, vol := range csivolList.Items {
				// Search the volume in the list of mounted volumes at the node
				// retrieved above
				mountPoint, exists := listContains(
					vol.Spec.Volume.MountPath, mountList,
				)
				// If the volume is present in the list verify its state
				if exists && verifyMountOpts(mountPoint.Opts, "rw") {
					// Continue with remaining volumes since this volume looks
					// to be in good shape
					continue
				}
				if _, ok := TransitionVolList[vol.Spec.Volume.Name]; !ok {
					TransitionVolList[vol.Spec.Volume.Name] = vol.Status
					ReqMountList[vol.Spec.Volume.Name] = vol.Status
					csivol := vol
					go func() {
						devicePath, err := RemountVolume(
							exists, &csivol, mountPoint,
							vol.Spec.Volume.MountPath,
						)
						logrus.Errorf(
							"Remount failed: DevPath: %v %v",
							devicePath, err,
						)
					}()
				}
			}
			TransitionVolListLock.Unlock()
		}
	}
}

// WaitForVolumeReadyAndReachable waits until the volume is ready to accept IOs
// and is reachable, this function will not come out until both the conditions
// are met. This function stops the driver from overloading the OS with iSCSI
// login commands.
func WaitForVolumeReadyAndReachable(vol *apis.CSIVolume) {
	var err error
	for {
		// This function return after 12s in case the volume is not ready
		if err = WaitForVolumeToBeReady(vol.Spec.Volume.Name); err != nil {
			logrus.Error(err)
			// Keep retrying until the volume is ready
			continue
		}
		// This function return after 12s in case the volume is not reachable
		err = WaitForVolumeToBeReachable(vol.Spec.ISCSI.TargetPortal)
		if err == nil {
			return
		}
		logrus.Error(err)
	}
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
	exists bool, vol *apis.CSIVolume,
	mountPoint *mount.MountPoint,
	desiredMountOpt string,
) (devicePath string, err error) {
	mounter := mount.New("")
	options := []string{"rw"}
	// Wait until it is possible to chhange the state of mountpoint or when
	// login to volume is possible
	WaitForVolumeReadyAndReachable(vol)
	if exists {
		logrus.Infof("MountPoint:%v IN RO MODE", mountPoint.Path)
		// Unmout and mount operation is performed instead of just remount since
		// the remount option didn't give the desired results
		mounter.Unmount(mountPoint.Path)
		err = mounter.Mount(mountPoint.Device,
			mountPoint.Path, "", options)
	} else {
		// A complete attach and mount is performed if for some reason disk is
		// not present in the mounted list with the OS.
		devicePath, err = iscsi.AttachAndMountDisk(vol)
	}
	TransitionVolListLock.Lock()
	// Remove the volume from ReqMountList once the remount operation is
	// complete
	delete(TransitionVolList, vol.Spec.Volume.Name)
	delete(ReqMountList, vol.Spec.Volume.Name)
	TransitionVolListLock.Unlock()
	return
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
