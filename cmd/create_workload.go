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

var createWorkloadConfigPath string

// CreateWorkloadCmd represents the workload command
var CreateWorkloadCmd = &cobra.Command{
	Use:          "workload",
	Example:      "tptctl create workload -c /path/to/config.yaml",
	Short:        "Create a new workload",
	Long:         `Create a new workload.`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// load config
		configContent, err := ioutil.ReadFile(createWorkloadConfigPath)
		if err != nil {
			qout.Error("failed to read config file", err)
			os.Exit(1)
		}
		var workloadConfig api.WorkloadConfig
		if err := yaml.Unmarshal(configContent, &workloadConfig); err != nil {
			qout.Error("failed to unmarshal config file yaml content", err)
			os.Exit(1)
		}

		// create workload
		if err := workloadConfig.Create(); err != nil {
			qout.Error("failed to create workload", err)
			os.Exit(1)
		}

		qout.Complete(fmt.Sprintf("workload %s created\n", workloadConfig.Name))
	},
}

func init() {
	createCmd.AddCommand(CreateWorkloadCmd)

	CreateWorkloadCmd.Flags().StringVarP(&createWorkloadConfigPath, "config", "c", "", "path to file with workload config")
	CreateWorkloadCmd.MarkFlagRequired("config")
}
