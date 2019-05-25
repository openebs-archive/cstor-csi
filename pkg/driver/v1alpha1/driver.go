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

package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	//"github.com/golang/glog"
	"github.com/Sirupsen/logrus"
	"github.com/openebs/csi/pkg/config/v1alpha1"
	"github.com/openebs/csi/pkg/utils/v1alpha1"
)

const (
	// TODO rename to Name
	//
	// DriverName defines the name that is used
	// in Kubernetes and the CSI system for the
	// canonical, official name of this plugin
	DriverName = "openebs-csi.openebs.io"
)

// TODO check if this can be renamed to Base
//
// CSIDriver defines a common data structure
// for drivers
type CSIDriver struct {
	// TODO change the field names to make it
	// readable
	config *config.Config
	ids    *IdentityServer
	ns     *NodeServer
	cs     *ControllerServer

	cap []*csi.VolumeCapability_AccessMode
}

// GetVolumeCapabilityAccessModes fetches the access
// modes on which the volume can be exposed
func GetVolumeCapabilityAccessModes() []*csi.VolumeCapability_AccessMode {
	supported := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}

	var vcams []*csi.VolumeCapability_AccessMode
	for _, vcam := range supported {
		logrus.Infof("enabling volume access mode: %s", vcam.String())
		vcams = append(vcams, newVolumeCapabilityAccessMode(vcam))
	}
	return vcams
}

func newVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

// New returns a new driver instance
func New(config *config.Config) *CSIDriver {
	driver := &CSIDriver{
		config: config,
		cap:    GetVolumeCapabilityAccessModes(),
	}

	switch config.PluginType {
	case "controller":
		driver.cs = NewControllerServer(driver)

	case "node":
		utils.FetchAndUpdateVolInfos(config.NodeID)

		// Start monitor goroutine to monitor the
		// mounted paths. If a path goes down or
		// becomes read only (in case of RW mount
		// points), this thread will fetch the path
		// and relogin or remount
		go utils.MonitorMounts()

		driver.ns = NewNodeServer(driver)
	}

	// Identity server is common to both node and
	// controller, it is required to register,
	// share capabilities and probe the corresponding
	// driver
	driver.ids = NewIdentityServer(driver)
	return driver
}

// Run starts the CSI plugin by communicating
// over the given endpoint
func (d *CSIDriver) Run() error {
	// Initialize and start listening on grpc server
	s := utils.NewNonBlockingGRPCServer()

	s.Start(d.config.Endpoint, d.ids, d.cs, d.ns)
	s.Wait()

	return nil
}
