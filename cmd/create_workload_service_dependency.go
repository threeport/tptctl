/*
Copyright © 2023 Threeport admin@threeport.io
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/threeport/tptctl/internal/api"
	qout "github.com/threeport/tptctl/internal/output"
)

var createWorkloadServiceDependencyConfigPath string

// CreateWorkloadServiceDependencyCmd represents the workload-service-dependency command
var CreateWorkloadServiceDependencyCmd = &cobra.Command{
	Use:          "workload-service-dependency",
	Example:      "tptctl create workload-servicde-dependency -c /path/to/config.yaml",
	Short:        "Create a new workload service dependency",
	Long:         `Create a new workload service dependency.`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// load config
		configContent, err := ioutil.ReadFile(createWorkloadServiceDependencyConfigPath)
		if err != nil {
			qout.Error("failed to read config file", err)
			os.Exit(1)
		}
		var workloadServiceDependency api.WorkloadServiceDependencyConfig
		if err := yaml.Unmarshal(configContent, &workloadServiceDependency); err != nil {
			qout.Error("failed to unmarshal config file yaml content", err)
			os.Exit(1)
		}

		// create workload service dependency
		wsd, err := workloadServiceDependency.Create()
		if err != nil {
			qout.Error("failed to create workload", err)
			os.Exit(1)
		}

		qout.Complete(fmt.Sprintf("workload service dependency %s created\n", *wsd.Name))
	},
}

func init() {
	createCmd.AddCommand(CreateWorkloadServiceDependencyCmd)

	CreateWorkloadServiceDependencyCmd.Flags().StringVarP(&createWorkloadServiceDependencyConfigPath, "config", "c", "", "path to file with workload service dependency config")
	CreateWorkloadServiceDependencyCmd.MarkFlagRequired("config")
}
