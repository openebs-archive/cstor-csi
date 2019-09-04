package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	config "github.com/openebs/csi/pkg/config/v1alpha1"
	service "github.com/openebs/csi/pkg/service"
	"github.com/openebs/csi/pkg/version"
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
		&config.DriverName, "name", "openebs-csi.openebs.io", "Name of this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.PluginType, "plugin", "csi-plugin", "Type of this driver i.e. controller or node",
	)

	cmd.Flags().StringVar(
		&config.CASEngine, "casengine", "cas-engine", "Type of engine i.e. cstor or jiva",
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

	if config.CASEngine == "" {
		config.CASEngine = "cstor"
	}
	logrus.Infof("%s - %s", version.Current(), version.GetGitCommit())
	logrus.Infof(
		"DriverName: %s Plugin: %s EndPoint: %s URL: %s NodeID: %s CASEngine: %s",
		config.DriverName,
		config.PluginType,
		config.Endpoint,
		config.RestURL,
		config.NodeID,
		config.CASEngine,
	)

	err := service.New(config).Run()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
