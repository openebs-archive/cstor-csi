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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/openebs/cstor-csi/pkg/config"
	"github.com/openebs/cstor-csi/pkg/driver"
	"github.com/openebs/cstor-csi/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	_ = flag.CommandLine.Parse([]string{})
	var config = config.Default()

	cmd := &cobra.Command{
		Use:   "openebs-csi-driver",
		Short: "openebs-csi-driver",
		Run: func(cmd *cobra.Command, args []string) {
			run(config)
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(
		&config.RestURL, "url", "", "REST URL that points to maya api server",
	)

	cmd.PersistentFlags().StringVar(
		&config.NodeID, "nodeid", "node1", "NodeID to identify the node running this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.Version, "version", "", "Displays driver version",
	)

	cmd.PersistentFlags().StringVar(
		&config.Endpoint, "endpoint", "unix://csi/csi.sock", "CSI endpoint",
	)

	cmd.PersistentFlags().StringVar(
		&config.DriverName, "name", "cstor.csi.openebs.io", "Name of this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.PluginType, "plugin", "csi-plugin", "Type of this driver i.e. controller or node",
	)

	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func run(config *config.Config) {
	if config.Version == "" {
		config.Version = version.Current()
	}

	logrus.Infof("%s - %s", version.Current(), version.GetGitCommit())
	logrus.Infof(
		"DriverName: %s Plugin: %s EndPoint: %s URL: %s NodeID: %s",
		config.DriverName,
		config.PluginType,
		config.Endpoint,
		config.RestURL,
		config.NodeID,
	)

	err := driver.New(config).Run()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
