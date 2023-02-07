/*
Copyright © 2023 Threeport admin@threeport.io
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

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
		// get threeport config
		threeportConfig := &config.ThreeportConfig{}
		if err := viper.Unmarshal(threeportConfig); err != nil {
			qout.Error("failed to get Threeport config", err)
			os.Exit(1)
		}

		// check threeport config for exisiting instance
		// find the threeport instance by name
		threeportInstanceConfigExists := false
		var instanceConfig config.Instance
		for _, instance := range threeportConfig.Instances {
			if instance.Name == deleteThreeportInstanceName {
				instanceConfig = instance
				threeportInstanceConfigExists = true
			}
		}
		if !threeportInstanceConfigExists {
			qout.Error("failed to find threeport instance config",
				errors.New(fmt.Sprintf(
					"config for threeport instance with name %s not found", deleteThreeportInstanceName)))
			os.Exit(1)
		}
		//fmt.Printf("%+v\n", instanceConfig)
		//fmt.Println("all good")
		//os.Exit(0)

		// the control plane object provides the config for installing on the
		// provider
		controlPlane := provider.ControlPlane{InstanceName: deleteThreeportInstanceName}

		// determine infra provider
		switch instanceConfig.Provider {
		case "kind":
			if err := controlPlane.DeleteControlPlaneOnKind(); err != nil {
				qout.Error("failed to delete threeport control plane on kind", err)
				os.Exit(1)
			}
		case "eks":
			if err := controlPlane.DeleteControlPlaneOnEKS(providerConfigDir); err != nil {
				qout.Error("failed to delete threeport control plane on EKS", err)
				os.Exit(1)
			}
		default:
			qout.Error("unrecognized infra provider",
				errors.New(fmt.Sprintf("infra provider %s not supported", infraProvider)))
			os.Exit(1)
		}
		///////////////////////////////////////////////////////////////////////

		//// delete kind cluster
		//fmt.Println("deleting kind cluster...")
		//kindDelete := exec.Command(
		//	"kind",
		//	"delete",
		//	"cluster",
		//	"--name",
		//	controlPlane.ThreeportClusterName(),
		//)
		//if err := kindDelete.Run(); err != nil {
		//	qout.Error("failed to delete kind cluster", err)
		//	os.Exit(1)
		//}
		//qout.Info("kind cluster deleted")

		//// get threeport config
		//threeportConfig := &config.ThreeportConfig{}
		//if err := viper.Unmarshal(threeportConfig); err != nil {
		//	qout.Error("failed to get Threeport config", err)
		//}

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
