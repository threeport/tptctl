/*
Copyright © 2023 Threeport admin@threeport.io
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tpclient "github.com/threeport/threeport-go-client"
	tpapi "github.com/threeport/threeport-rest-api/pkg/api/v0"
	kubeclient "k8s.io/client-go/tools/clientcmd"

	"github.com/threeport/tptctl/internal/config"
	"github.com/threeport/tptctl/internal/install"
	qout "github.com/threeport/tptctl/internal/output"
	"github.com/threeport/tptctl/internal/provider"
)

const (
	defaultComputeClusterName          string = "default-threeport-compute-space"
	defaultComputeClusterRegion               = "local"
	defaultComputeClusterProvider             = "kind"
	defaultComputeClusterAPIEndpoint          = "kubernetes.default"
	forwardProxyWorkloadDefinitionName        = "forwardProxy"
)

var (
	createThreeportInstanceName string
	forceOverwriteConfig        bool
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

		// write kind config file to /tmp directory
		configFile, err := os.Create(provider.ThreeportKindConfigPath)
		if err != nil {
			qout.Error("failed to write kind config file to disk", err)
			os.Exit(1)
		}
		defer configFile.Close()
		configFile.WriteString(provider.KindConfig(createThreeportInstanceName))
		qout.Info("kind config written to /tmp directory")

		// start kind cluster
		qout.Info("creating kind cluster... (this could take a few minutes)")
		kindCreate := exec.Command(
			"kind",
			"create",
			"cluster",
			"--config",
			provider.ThreeportKindConfigPath,
		)
		if err := kindCreate.Run(); err != nil {
			qout.Error("failed to create new kind cluster", err)
			os.Exit(1)
		}
		qout.Info("kind cluster created")

		// write API dependencies manifest to /tmp directory
		apiDepsManifest, err := os.Create(install.APIDepsManifestPath)
		if err != nil {
			qout.Error("failed to write API dependency manifests to disk", err)
			os.Exit(1)
		}
		defer apiDepsManifest.Close()
		apiDepsManifest.WriteString(install.APIDepsManifest())
		qout.Info("Threeport API dependencies manifest written to /tmp directory")

		// install API dependencies on kind cluster
		qout.Info("installing Threeport API dependencies")
		apiDepsCreate := exec.Command(
			"kubectl",
			"apply",
			"-f",
			install.APIDepsManifestPath,
		)
		if err := apiDepsCreate.Run(); err != nil {
			qout.Error("failed to install API dependencies to kind cluster", err)
			os.Exit(1)
		}
		psqlConfigCreate := exec.Command(
			"kubectl",
			"create",
			"configmap",
			"postgres-config-data",
			"-n",
			install.ThreeportControlPlaneNs,
		)
		if err := psqlConfigCreate.Run(); err != nil {
			qout.Error("failed to create API database config", err)
			os.Exit(1)
		}

		qout.Info("Threeport API dependencies created")

		// write API server manifest to /tmp directory
		apiServerManifest, err := os.Create(install.APIServerManifestPath)
		if err != nil {
			qout.Error("failed to write API manifest to disk", err)
			os.Exit(1)
		}
		defer apiServerManifest.Close()
		apiServerManifest.WriteString(install.APIServerManifest())
		qout.Info("Threeport API server manifest written to /tmp directory")

		// install Threeport API
		qout.Info("installing Threeport API server")
		apiServerCreate := exec.Command(
			"kubectl",
			"apply",
			"-f",
			install.APIServerManifestPath,
		)
		if err := apiServerCreate.Run(); err != nil {
			qout.Error("failed to create API server", err)
			os.Exit(1)
		}

		qout.Info("Threeport API server created")

		// write workload controller manifest to /tmp directory
		workloadControllerManifest, err := os.Create(install.WorkloadControllerManifestPath)
		if err != nil {
			qout.Error("failed to write workload controller manifest to disk", err)
			os.Exit(1)
		}
		defer workloadControllerManifest.Close()
		workloadControllerManifest.WriteString(install.WorkloadControllerManifest())
		qout.Info("Threeport workload controller manifest written to /tmp directory")

		// install workload controller
		qout.Info("installing Threeport workload controller")
		workloadControllerCreate := exec.Command(
			"kubectl",
			"apply",
			"-f",
			install.WorkloadControllerManifestPath,
		)
		if err := workloadControllerCreate.Run(); err != nil {
			qout.Error("failed to create workload controller", err)
			os.Exit(1)
		}

		qout.Info("Threeport workload controller created")

		// wait a few seconds for everything to come up
		qout.Info("waiting for control plane components to spin up...")
		time.Sleep(time.Second * 200)

		// get kubeconfig
		defaultLoadRules := kubeclient.NewDefaultClientConfigLoadingRules()

		clientConfigLoadRules, err := defaultLoadRules.Load()
		if err != nil {
			qout.Error("failed to load default kubeconfig rules", err)
			os.Exit(1)
		}

		clientConfig := kubeclient.NewDefaultClientConfig(*clientConfigLoadRules, &kubeclient.ConfigOverrides{})
		kubeConfig, err := clientConfig.RawConfig()
		if err != nil {
			qout.Error("failed to load kubeconfig", err)
			os.Exit(1)
		}

		// get cluster CA and server endpoint
		var caCert string
		clusterFound := false
		for clusterName, cluster := range kubeConfig.Clusters {
			if clusterName == kubeConfig.CurrentContext {
				caCert = string(cluster.CertificateAuthorityData)
				clusterFound = true
			}
		}
		if !clusterFound {
			qout.Error(
				"failed to get Kubernetes cluster CA and endpoint",
				errors.New("cluster config not found in kubeconfig"),
			)
			os.Exit(1)
		}

		// get client certificate and key
		var cert string
		var key string
		userFound := false
		for userName, user := range kubeConfig.AuthInfos {
			if userName == kubeConfig.CurrentContext {
				cert = string(user.ClientCertificateData)
				key = string(user.ClientKeyData)
				userFound = true
			}
		}
		if !userFound {
			qout.Error(
				"failed to get user credentials to Kubernetes cluster",
				errors.New("kubeconfig user for threeport cluster not found"),
			)
			os.Exit(1)
		}

		// setup default compute space cluster
		defaultClusterName := defaultComputeClusterName
		defaultClusterRegion := defaultComputeClusterRegion
		defaultClusterProvider := defaultComputeClusterProvider
		defaultClusterAPIEndpoint := defaultComputeClusterAPIEndpoint
		workloadCluster := tpapi.WorkloadCluster{
			Name:          &defaultClusterName,
			Region:        &defaultClusterRegion,
			Provider:      &defaultClusterProvider,
			APIEndpoint:   &defaultClusterAPIEndpoint,
			CACertificate: &caCert,
			Certificate:   &cert,
			Key:           &key,
		}
		wcJSON, err := json.Marshal(&workloadCluster)
		if err != nil {
			qout.Error("failed to marshal workload cluster to json", err)
			os.Exit(1)
		}
		wc, err := tpclient.CreateWorkloadCluster(wcJSON, install.GetThreeportAPIEndpoint(), "")
		if err != nil {
			qout.Error("failed to create workload cluster in Threeport API", err)
			os.Exit(1)
		}
		qout.Info(fmt.Sprintf("default workload cluster %s for compute space set up", *wc.Name))

		// TODO: add superuser
		superuserID := uint(1)

		// add forward proxy definition
		fwdProxyDefName := forwardProxyWorkloadDefinitionName
		fwdProxyYAML := install.ForwardProxyManifest()
		fwdProxyWorkloadDefinition := tpapi.WorkloadDefinition{
			Name:         &fwdProxyDefName,
			YAMLDocument: &fwdProxyYAML,
			UserID:       &superuserID,
		}
		fpwdJSON, err := json.Marshal(&fwdProxyWorkloadDefinition)
		if err != nil {
			qout.Error("failed to marshal forward proxy workload definition to json", err)
			os.Exit(1)
		}
		fpwd, err := tpclient.CreateWorkloadDefinition(fpwdJSON, install.GetThreeportAPIEndpoint(), "")
		if err != nil {
			qout.Error("failed to create forward proxy workload definition in Threeport API", err)
			os.Exit(1)
		}
		qout.Info(fmt.Sprintf("forward proxy workload definition %s added", *fpwd.Name))

		// create threeport config for new instance
		newThreeportInstance := &config.Instance{
			Name:      createThreeportInstanceName,
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

func AddControlPlaneCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(
		"provider", "p", "kind", "the infrasture provider to install upon")
	cmd.Flags().StringVarP(&createThreeportInstanceName,
		"name", "n", "", "name of control plane instance")
	cmd.MarkFlagRequired("name")
	cmd.Flags().BoolVar(
		&forceOverwriteConfig, "force-overwrite-config", false,
		"force the overwrite of an existing Threeport instance config.  Warning: this will erase the connection info for the existing instance.  Only do this if the existing instance has already been deleted and is no longer in use.")
}

func init() {
	createCmd.AddCommand(CreateControlPlaneCmd)
	//AddControlPlaneCmdFlags(CreateControlPlaneCmd)
	CreateControlPlaneCmd.Flags().StringP(
		"provider", "p", "kind", "the infrasture provider to install upon")
	CreateControlPlaneCmd.Flags().StringVarP(&createThreeportInstanceName,
		"name", "n", "", "name of control plane instance")
	CreateControlPlaneCmd.MarkFlagRequired("name")
	CreateControlPlaneCmd.Flags().BoolVar(
		&forceOverwriteConfig, "force-overwrite-config", false,
		"force the overwrite of an existing Threeport instance config.  Warning: this will erase the connection info for the existing instance.  Only do this if the existing instance has already been deleted and is no longer in use.")
}
