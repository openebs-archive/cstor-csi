package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kubernetes-csi/csi-lib-iscsi/iscsi"
)

var (
	portals   = flag.String("portals", "192.168.1.112:3260", "Comma delimited.  Eg: 1.1.1.1,2.2.2.2")
	iqn       = flag.String("iqn", "iqn.2010-10.org.openstack:volume-95739000-1557-44f8-9f40-e9d29fe6ec47", "")
	multipath = flag.Bool("multipath", false, "")
	username  = flag.String("username", "3aX7EEf3CEgvESQG75qh", "")
	password  = flag.String("password", "eJBDC7Bt7WE3XFDq", "")
	lun       = flag.Int("lun", 1, "")
	debug     = flag.Bool("debug", false, "enable logging")
)

func main() {
	flag.Parse()
	tgtp := strings.Split(*portals, ",")
	if *debug {
		iscsi.EnableDebugLogging(os.Stdout)
	}

	// You can utilize the iscsiadm calls directly if you wish, but by creating a Connector
	// you can simplify interactions to simple calls like "Connect" and "Disconnect"
	c := iscsi.Connector{
		// Our example uses chap
		AuthType: "chap",
		// Specify the target iqn we're dealing with
		TargetIqn: *iqn,
		// List of portals must be >= 1 (>1 signals multipath/mpio)
		TargetPortals: tgtp,
		// CHAP can be setup up for discovery as well as sessions, our example
		// device only uses CHAP security for sessions, for those that use Discovery
		// as well, we'd add a DiscoverySecrets entry the same way
		SessionSecrets: iscsi.Secrets{
			UserName:    *username,
			Password:    *password,
			SecretsType: "chap"},
		// Lun is the lun number the devices uses for exports
		Lun: int32(*lun),
		// Multipath indicates that we want to configure this connection as a multipath device
		Multipath: *multipath,
		// Number of times we check for device path, waiting for CheckInterval seconds inbetween each check (defaults to 10 if omitted)
		RetryCount: 11,
		// CheckInterval is the time in seconds to wait inbetween device path checks when logging in to a target
		CheckInterval: 1,
	}

	// Now we can just issue a connection request using our Connector
	// A succesful connection will include the device path to access our iscsi volume
	path, err := iscsi.Connect(c)
	if err != nil {
		log.Printf("Error returned from iscsi.Connect: %s", err.Error())
		os.Exit(1)
	}

	if path == "" {
		log.Printf("Failed to connect, didn't receive a path, but also no error!")
		os.Exit(1)
	}

	log.Printf("Connected device at path: %s\n", path)
	time.Sleep(3 * time.Second)

	// Disconnect is easy as well, we don't need the full Connector any more, just the Target IQN and the Portals
	/// this should disconnect the volume as well as clear out the iscsi DB entries associated with it
	iscsi.Disconnect(c.TargetIqn, c.TargetPortals)
}
