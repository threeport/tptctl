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
	"github.com/threeport/tptctl/internal/install"
	qout "github.com/threeport/tptctl/internal/output"
	"github.com/threeport/tptctl/internal/provider"
)

var (
	createThreeportInstanceName string
	forceOverwriteConfig        bool
	infraProvider               string
)

// CreateControlPlaneCmd represents the create threeport command
var CreateControlPlaneCmd = &cobra.Command{
	Use:          "control-plane",
	Example:      "tptctl create control-plane",
	Short:        "Create a new instance of the Threeport control plane",
	Long:         `Create a new instance of the Threeport control plane.`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// get threeport config
		threeportConfig := &config.ThreeportConfig{}
		if err := viper.Unmarshal(threeportConfig); err != nil {
			qout.Error("failed to get Threeport config", err)
			os.Exit(1)
		}

		// check threeport config for exisiting instance
		threeportInstanceConfigExists := false
		for _, instance := range threeportConfig.Instances {
			if instance.Name == createThreeportInstanceName {
				threeportInstanceConfigExists = true
				if !forceOverwriteConfig {
					qout.Error(
						"interupted creation of Threeport instance",
						errors.New(fmt.Sprintf("instance of Threeport with name %s already exists", instance.Name)),
					)
					qout.Info("if you wish to overwrite the existing config use --force-overwrite-config flag")
					qout.Warning("you will lose the ability to connect to the existing Threeport instance if it still exists")
					os.Exit(1)
				}
			}
		}

		// flag validation
		if err := validateCreateControlPlaneFlags(infraProvider); err != nil {
			qout.Error("flag validation failed", err)
			os.Exit(1)
		}

		// the control plane object provides the config for installing on the
		// provider
		controlPlane := provider.ControlPlane{InstanceName: createThreeportInstanceName}

		// determine infra provider
		switch infraProvider {
		case "kind":
			if err := controlPlane.CreateControlPlaneOnKind(providerConfigDir); err != nil {
				qout.Error("failed to install control plane on kind", err)
				os.Exit(1)
			}
		case "eks":
			if err := controlPlane.CreateControlPlaneOnEKS(providerConfigDir); err != nil {
				qout.Error("failed to install control plane on EKS", err)
				os.Exit(1)
			}
		default:
			qout.Error("unrecognized infra provider",
				errors.New(fmt.Sprintf("infra provider %s not supported", infraProvider)))
			os.Exit(1)
		}

		// create threeport config for new instance
		newThreeportInstance := &config.Instance{
			Name:      createThreeportInstanceName,
			Provider:  infraProvider,
			APIServer: install.GetThreeportAPIEndpoint(),
		}

		// update threeport config to add the new instance and set as current instance
		if threeportInstanceConfigExists {
			for n, instance := range threeportConfig.Instances {
				if instance.Name == createThreeportInstanceName {
					threeportConfig.Instances[n] = *newThreeportInstance
				}
			}
		} else {
			threeportConfig.Instances = append(threeportConfig.Instances, *newThreeportInstance)
		}
		viper.Set("Instances", threeportConfig.Instances)
		viper.Set("CurrentInstance", createThreeportInstanceName)
		viper.WriteConfig()
		qout.Info("Threeport config updated")

		qout.Complete("Threeport instance created")
	},
}

func init() {
	createCmd.AddCommand(CreateControlPlaneCmd)
	CreateControlPlaneCmd.Flags().StringVarP(&infraProvider,
		"provider", "p", "kind", "the infrasture provider to install upon")
	CreateControlPlaneCmd.Flags().StringVarP(&createThreeportInstanceName,
		"name", "n", "", "name of control plane instance")
	CreateControlPlaneCmd.MarkFlagRequired("name")
	CreateControlPlaneCmd.Flags().BoolVar(
		&forceOverwriteConfig, "force-overwrite-config", false,
		"force the overwrite of an existing Threeport instance config.  Warning: this will erase the connection info for the existing instance.  Only do this if the existing instance has already been deleted and is no longer in use.")
}

// validateCreateControlPlaneFlags validates flag inputs as needed
func validateCreateControlPlaneFlags(infraProvider string) error {
	allowedInfraProviders := []string{"kind", "eks"}
	matched := false
	for _, prov := range allowedInfraProviders {
		if infraProvider == prov {
			matched = true
			break
		}
	}
	if !matched {
		return errors.New(fmt.Sprintf("invalid provider value '%s' - must be one of %s",
			infraProvider, allowedInfraProviders))
	}

	return nil
}
