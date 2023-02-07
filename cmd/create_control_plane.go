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

		//configFile, err := os.Create(provider.ThreeportKindConfigPath)
		//if err != nil {
		//	qout.Error("failed to write kind config file to disk", err)
		//	os.Exit(1)
		//}
		//defer configFile.Close()
		//configFile.WriteString(provider.KindConfig(createThreeportInstanceName))
		//qout.Info("kind config written to /tmp directory")

		//// start kind cluster
		//qout.Info("creating kind cluster... (this could take a few minutes)")
		//kindCreate := exec.Command(
		//	"kind",
		//	"create",
		//	"cluster",
		//	"--config",
		//	provider.ThreeportKindConfigPath,
		//)
		//if err := kindCreate.Run(); err != nil {
		//	qout.Error("failed to create new kind cluster", err)
		//	os.Exit(1)
		//}
		//qout.Info("kind cluster created")

		//// write API dependencies manifest to /tmp directory
		//apiDepsManifest, err := os.Create(install.APIDepsManifestPath)
		//if err != nil {
		//	qout.Error("failed to write API dependency manifests to disk", err)
		//	os.Exit(1)
		//}
		//defer apiDepsManifest.Close()
		//apiDepsManifest.WriteString(install.APIDepsManifest())
		//qout.Info("Threeport API dependencies manifest written to /tmp directory")

		//// install API dependencies on kind cluster
		//qout.Info("installing Threeport API dependencies")
		//apiDepsCreate := exec.Command(
		//	"kubectl",
		//	"apply",
		//	"-f",
		//	install.APIDepsManifestPath,
		//)
		//if err := apiDepsCreate.Run(); err != nil {
		//	qout.Error("failed to install API dependencies to kind cluster", err)
		//	os.Exit(1)
		//}
		//psqlConfigCreate := exec.Command(
		//	"kubectl",
		//	"create",
		//	"configmap",
		//	"postgres-config-data",
		//	"-n",
		//	install.ThreeportControlPlaneNs,
		//)
		//if err := psqlConfigCreate.Run(); err != nil {
		//	qout.Error("failed to create API database config", err)
		//	os.Exit(1)
		//}

		//qout.Info("Threeport API dependencies created")

		//// write API server manifest to /tmp directory
		//apiServerManifest, err := os.Create(install.APIServerManifestPath)
		//if err != nil {
		//	qout.Error("failed to write API manifest to disk", err)
		//	os.Exit(1)
		//}
		//defer apiServerManifest.Close()
		//apiServerManifest.WriteString(install.APIServerManifest())
		//qout.Info("Threeport API server manifest written to /tmp directory")

		//// install Threeport API
		//qout.Info("installing Threeport API server")
		//apiServerCreate := exec.Command(
		//	"kubectl",
		//	"apply",
		//	"-f",
		//	install.APIServerManifestPath,
		//)
		//if err := apiServerCreate.Run(); err != nil {
		//	qout.Error("failed to create API server", err)
		//	os.Exit(1)
		//}

		//qout.Info("Threeport API server created")

		//// write workload controller manifest to /tmp directory
		//workloadControllerManifest, err := os.Create(install.WorkloadControllerManifestPath)
		//if err != nil {
		//	qout.Error("failed to write workload controller manifest to disk", err)
		//	os.Exit(1)
		//}
		//defer workloadControllerManifest.Close()
		//workloadControllerManifest.WriteString(install.WorkloadControllerManifest())
		//qout.Info("Threeport workload controller manifest written to /tmp directory")

		//// install workload controller
		//qout.Info("installing Threeport workload controller")
		//workloadControllerCreate := exec.Command(
		//	"kubectl",
		//	"apply",
		//	"-f",
		//	install.WorkloadControllerManifestPath,
		//)
		//if err := workloadControllerCreate.Run(); err != nil {
		//	qout.Error("failed to create workload controller", err)
		//	os.Exit(1)
		//}

		//qout.Info("Threeport workload controller created")

		//// wait a few seconds for everything to come up
		//qout.Info("waiting for control plane components to spin up...")
		//time.Sleep(time.Second * 200)

		//// get kubeconfig
		//defaultLoadRules := kubeclient.NewDefaultClientConfigLoadingRules()

		//clientConfigLoadRules, err := defaultLoadRules.Load()
		//if err != nil {
		//	qout.Error("failed to load default kubeconfig rules", err)
		//	os.Exit(1)
		//}

		//clientConfig := kubeclient.NewDefaultClientConfig(*clientConfigLoadRules, &kubeclient.ConfigOverrides{})
		//kubeConfig, err := clientConfig.RawConfig()
		//if err != nil {
		//	qout.Error("failed to load kubeconfig", err)
		//	os.Exit(1)
		//}

		//// get cluster CA and server endpoint
		//var caCert string
		//clusterFound := false
		//for clusterName, cluster := range kubeConfig.Clusters {
		//	if clusterName == kubeConfig.CurrentContext {
		//		caCert = string(cluster.CertificateAuthorityData)
		//		clusterFound = true
		//	}
		//}
		//if !clusterFound {
		//	qout.Error(
		//		"failed to get Kubernetes cluster CA and endpoint",
		//		errors.New("cluster config not found in kubeconfig"),
		//	)
		//	os.Exit(1)
		//}

		//// get client certificate and key
		//var cert string
		//var key string
		//userFound := false
		//for userName, user := range kubeConfig.AuthInfos {
		//	if userName == kubeConfig.CurrentContext {
		//		cert = string(user.ClientCertificateData)
		//		key = string(user.ClientKeyData)
		//		userFound = true
		//	}
		//}
		//if !userFound {
		//	qout.Error(
		//		"failed to get user credentials to Kubernetes cluster",
		//		errors.New("kubeconfig user for threeport cluster not found"),
		//	)
		//	os.Exit(1)
		//}

		//// setup default compute space cluster
		//defaultClusterName := defaultComputeClusterName
		//defaultClusterRegion := defaultComputeClusterRegion
		//defaultClusterProvider := defaultComputeClusterProvider
		//defaultClusterAPIEndpoint := defaultComputeClusterAPIEndpoint
		//workloadCluster := tpapi.WorkloadCluster{
		//	Name:          &defaultClusterName,
		//	Region:        &defaultClusterRegion,
		//	Provider:      &defaultClusterProvider,
		//	APIEndpoint:   &defaultClusterAPIEndpoint,
		//	CACertificate: &caCert,
		//	Certificate:   &cert,
		//	Key:           &key,
		//}
		//wcJSON, err := json.Marshal(&workloadCluster)
		//if err != nil {
		//	qout.Error("failed to marshal workload cluster to json", err)
		//	os.Exit(1)
		//}
		//wc, err := tpclient.CreateWorkloadCluster(wcJSON, install.GetThreeportAPIEndpoint(), "")
		//if err != nil {
		//	qout.Error("failed to create workload cluster in Threeport API", err)
		//	os.Exit(1)
		//}
		//qout.Info(fmt.Sprintf("default workload cluster %s for compute space set up", *wc.Name))

		//// TODO: add superuser
		//superuserID := uint(1)

		//// add forward proxy definition
		//fwdProxyDefName := forwardProxyWorkloadDefinitionName
		//fwdProxyYAML := install.ForwardProxyManifest()
		//fwdProxyWorkloadDefinition := tpapi.WorkloadDefinition{
		//	Name:         &fwdProxyDefName,
		//	YAMLDocument: &fwdProxyYAML,
		//	UserID:       &superuserID,
		//}
		//fpwdJSON, err := json.Marshal(&fwdProxyWorkloadDefinition)
		//if err != nil {
		//	qout.Error("failed to marshal forward proxy workload definition to json", err)
		//	os.Exit(1)
		//}
		//fpwd, err := tpclient.CreateWorkloadDefinition(fpwdJSON, install.GetThreeportAPIEndpoint(), "")
		//if err != nil {
		//	qout.Error("failed to create forward proxy workload definition in Threeport API", err)
		//	os.Exit(1)
		//}
		//qout.Info(fmt.Sprintf("forward proxy workload definition %s added", *fpwd.Name))

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
