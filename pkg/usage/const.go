/*
Copyright 2020 The OpenEBS Authors.

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

package usage

const (
	// GAclientID is the unique code of OpenEBS project in Google Analytics
	GAclientID string = "UA-127388617-1"

	// supported events categories

	// InstallEvent event is sent on pod starts
	InstallEvent string = "install"
	// Ping event is sent periodically
	Ping string = "cstor-csi-ping"
	// VolumeProvision event is sent when a volume is created
	VolumeProvision string = "volume-provision"
	// VolumeDeprovision event is sent when a volume is deleted
	VolumeDeprovision string = "volume-deprovision"
	// AppName the application name
	AppName string = "OpenEBS"

	// RunningStatus status is running
	RunningStatus string = "running"
	// EventLabelNode holds the string label "nodes"
	EventLabelNode string = "nodes"
	// EventLabelCapacity holds the string label "capacity"
	EventLabelCapacity string = "capacity"

	// Replica Event replication
	Replica string = "replica:"
	// DefaultReplicaCount holds the replica count string
	DefaultReplicaCount string = "replica:3"

	// DefaultCASType Event application name constant for volume event
	DefaultCASType string = "cstor"
)
