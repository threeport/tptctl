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

var createWorkloadDefinitionConfigPath string

// CreateWorkloadDefinitionCmd represents the workload-definition command
var CreateWorkloadDefinitionCmd = &cobra.Command{
	Use:          "workload-definition",
	Example:      "tptctl create workload-definition -c /path/to/config.yaml",
	Short:        "Create a new workload definition",
	Long:         `Create a new workload definition.`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// load config
		configContent, err := ioutil.ReadFile(createWorkloadDefinitionConfigPath)
		if err != nil {
			qout.Error("failed to read config file", err)
			os.Exit(1)
		}
		var workloadDefinition api.WorkloadDefinitionConfig
		if err := yaml.Unmarshal(configContent, &workloadDefinition); err != nil {
			qout.Error("failed to unmarshal config file yaml content", err)
			os.Exit(1)
		}

		// create workload definition
		wd, err := workloadDefinition.Create()
		if err != nil {
			qout.Error("failed to create workload definition", err)
			os.Exit(1)
		}

		qout.Complete(fmt.Sprintf("workload definition %s created\n", *wd.Name))
	},
}

func init() {
	createCmd.AddCommand(CreateWorkloadDefinitionCmd)

	CreateWorkloadDefinitionCmd.Flags().StringVarP(&createWorkloadDefinitionConfigPath, "config", "c", "", "path to file with workload definition config")
	CreateWorkloadDefinitionCmd.MarkFlagRequired("config")
}
