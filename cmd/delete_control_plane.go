/*
Copyright © 2023 Threeport admin@threeport.io
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/threeport/tptctl/internal/config"
	qout "github.com/threeport/tptctl/internal/output"
	"github.com/threeport/tptctl/internal/provider"
)

var deleteThreeportInstanceName string

// DeleteControlPlaneCmd represents the delete control-plane command
var DeleteControlPlaneCmd = &cobra.Command{
	Use:          "control-plane",
	Example:      "tptctl delete control-plane",
	Short:        "Delete an instance of the Threeport control plane",
	Long:         `Delete an instance of the Threeport control plane.`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// delete kind cluster
		fmt.Println("deleting kind cluster...")
		kindDelete := exec.Command(
			"kind",
			"delete",
			"cluster",
			"--name",
			provider.GetThreeportKindClusterName(deleteThreeportInstanceName),
		)
		if err := kindDelete.Run(); err != nil {
			qout.Error("failed to delete kind cluster", err)
			os.Exit(1)
		}
		qout.Info("kind cluster deleted")

		// get threeport config
		threeportConfig := &config.ThreeportConfig{}
		if err := viper.Unmarshal(threeportConfig); err != nil {
			qout.Error("failed to get Threeport config", err)
		}

		// update threeport config to remove the deleted threeport instance and
		// current instance
		updatedInstances := []config.Instance{}
		for _, instance := range threeportConfig.Instances {
			if instance.Name == deleteThreeportInstanceName {
				continue
			} else {
				updatedInstances = append(updatedInstances, instance)
			}
		}

		viper.Set("Instances", updatedInstances)
		viper.Set("CurrentInstance", "")
		viper.WriteConfig()
		qout.Info("Threeport config updated")

		qout.Complete("Threeport instance deleted")
	},
}

func init() {
	deleteCmd.AddCommand(DeleteControlPlaneCmd)

	DeleteControlPlaneCmd.Flags().StringVarP(&deleteThreeportInstanceName,
		"name", "n", "", "name of control plane instance")
	DeleteControlPlaneCmd.MarkFlagRequired("name")
}
