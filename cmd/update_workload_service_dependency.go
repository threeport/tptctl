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

var updateWorkloadServiceDependencyConfigPath string

// UpdateWorkloadServiceDependencyCmd represents the workload-service-dependency command
var UpdateWorkloadServiceDependencyCmd = &cobra.Command{
	Use:          "workload-service-dependency",
	Example:      "tptctl update workload-servicde-dependency -c /path/to/config.yaml",
	Short:        "Update an existing workload service dependency",
	Long:         `Update an existing workload service dependency.`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// load config
		configContent, err := ioutil.ReadFile(updateWorkloadServiceDependencyConfigPath)
		if err != nil {
			qout.Error("failed to read config file", err)
			os.Exit(1)
		}
		var workloadServiceDependency api.WorkloadServiceDependencyConfig
		if err := yaml.Unmarshal(configContent, &workloadServiceDependency); err != nil {
			qout.Error("failed to unmarshal config file yaml content", err)
			os.Exit(1)
		}

		// update workload service dependency
		wsd, err := workloadServiceDependency.Update()
		if err != nil {
			qout.Error("failed to update workload", err)
			os.Exit(1)
		}

		qout.Complete(fmt.Sprintf("workload service dependency %s updated\n", *wsd.Name))
	},
}

func init() {
	updateCmd.AddCommand(UpdateWorkloadServiceDependencyCmd)

	UpdateWorkloadServiceDependencyCmd.Flags().StringVarP(&updateWorkloadServiceDependencyConfigPath, "config", "c", "", "path to file with workload service dependency config")
	UpdateWorkloadServiceDependencyCmd.MarkFlagRequired("config")
}
