package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tpclient "github.com/threeport/threeport-go-client"
	tpapi "github.com/threeport/threeport-rest-api/pkg/api/v0"
	kubeclient "k8s.io/client-go/tools/clientcmd"

	"github.com/threeport/tptctl/internal/install"
	qout "github.com/threeport/tptctl/internal/output"
	"github.com/threeport/tptctl/internal/threeport"
)

const (
	ThreeportKindConfigPath = "/tmp/threeport-kind-config.yaml"
)

func (c *ControlPlane) KindConfig() string {
	return fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: %[1]s
nodes:
- role: control-plane
- role: worker
  extraPortMappings:
    - containerPort: %[2]s
      hostPort: %[2]s
      protocol: TCP
`, c.ThreeportClusterName(), install.ThreeportAPIPort)
}

func (c *ControlPlane) CreateControlPlaneOnKind(providerConfigDir string) error {
	// write kind config file to /tmp directory
	configFile, err := os.Create(ThreeportKindConfigPath)
	if err != nil {
		return fmt.Errorf("failed to write kind config file to disk: %w", err)
		//qout.Error("failed to write kind config file to disk", err)
		//os.Exit(1)
	}
	defer configFile.Close()
	configFile.WriteString(c.KindConfig())
	qout.Info("kind config written to /tmp directory")

	// start kind cluster
	qout.Info("creating kind cluster... (this could take a few minutes)")
	kindCreate := exec.Command(
		"kind",
		"create",
		"cluster",
		"--config",
		ThreeportKindConfigPath,
	)
	if err := kindCreate.Run(); err != nil {
		return fmt.Errorf("failed to create new kind cluster: %w", err)
		//qout.Error("failed to create new kind cluster", err)
		//os.Exit(1)
	}
	qout.Info("kind cluster created")

	// write kubeconfig
	kubeconfigFilePath := filepath.Join(providerConfigDir,
		fmt.Sprintf("kubeconfig-%s", c.ThreeportClusterName()))
	kindKubeconfig, err := exec.Command(
		"kind",
		"get",
		"kubeconfig",
		"--name",
		c.ThreeportClusterName(),
	).Output()
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig content for kind cluster: %w", err)
	}
	ioutil.WriteFile(kubeconfigFilePath, []byte(kindKubeconfig), 0644)

	// install threeport API
	if err := install.InstallAPI(kubeconfigFilePath); err != nil {
		return fmt.Errorf("failed to install threeport API on kind cluster: %w", err)
	}

	// install workload controller
	if err := install.InstallWorkloadController(kubeconfigFilePath); err != nil {
		return fmt.Errorf("failed to install workload controller on kind cluster: %w", err)
	}

	//// write API dependencies manifest to /tmp directory
	//apiDepsManifest, err := os.Create(install.APIDepsManifestPath)
	//if err != nil {
	//	return fmt.Errorf("failed to write API dependency manifests to disk", err)
	//	//qout.Error("failed to write API dependency manifests to disk", err)
	//	//os.Exit(1)
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
	//	return fmt.Errorf("failed to install API dependencies to kind cluster", err)
	//	//qout.Error("failed to install API dependencies to kind cluster", err)
	//	//os.Exit(1)
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
	//	return fmt.Errorf("failed to create API database config", err)
	//	//qout.Error("failed to create API database config", err)
	//	//os.Exit(1)
	//}

	//qout.Info("Threeport API dependencies created")

	//// write API server manifest to /tmp directory
	//apiServerManifest, err := os.Create(install.APIServerManifestPath)
	//if err != nil {
	//	return fmt.Errorf("failed to write API manifest to disk", err)
	//	//qout.Error("failed to write API manifest to disk", err)
	//	//os.Exit(1)
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
	//	return fmt.Errorf("failed to create API server", err)
	//	//qout.Error("failed to create API server", err)
	//	//os.Exit(1)
	//}

	//qout.Info("Threeport API server created")

	//// write workload controller manifest to /tmp directory
	//workloadControllerManifest, err := os.Create(install.WorkloadControllerManifestPath)
	//if err != nil {
	//	return fmt.Errorf("failed to write workload controller manifest to disk", err)
	//	//qout.Error("failed to write workload controller manifest to disk", err)
	//	//os.Exit(1)
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
	//	return fmt.Errorf("failed to create workload controller", err)
	//	//qout.Error("failed to create workload controller", err)
	//	//os.Exit(1)
	//}

	//qout.Info("Threeport workload controller created")

	// wait a few seconds for everything to come up
	qout.Info("waiting for control plane components to spin up...")
	time.Sleep(time.Second * 200)

	// get kubeconfig
	defaultLoadRules := kubeclient.NewDefaultClientConfigLoadingRules()

	clientConfigLoadRules, err := defaultLoadRules.Load()
	if err != nil {
		return fmt.Errorf("failed to load default kubeconfig rules: %w", err)
		//qout.Error("failed to load default kubeconfig rules", err)
		//os.Exit(1)
	}

	clientConfig := kubeclient.NewDefaultClientConfig(*clientConfigLoadRules, &kubeclient.ConfigOverrides{})
	kubeConfig, err := clientConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
		//qout.Error("failed to load kubeconfig", err)
		//os.Exit(1)
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
		return fmt.Errorf(
			"failed to get Kubernetes cluster CA and endpoint: %w",
			errors.New("cluster config not found in kubeconfig"),
		)
		//qout.Error(
		//	"failed to get Kubernetes cluster CA and endpoint",
		//	errors.New("cluster config not found in kubeconfig"),
		//)
		//os.Exit(1)
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
		return fmt.Errorf(
			"failed to get user credentials to Kubernetes cluster: %w",
			errors.New("kubeconfig user for threeport cluster not found"),
		)
		//qout.Error(
		//	"failed to get user credentials to Kubernetes cluster",
		//	errors.New("kubeconfig user for threeport cluster not found"),
		//)
		//os.Exit(1)
	}

	// setup default compute space cluster
	defaultClusterName := threeport.DefaultComputeClusterName
	defaultClusterRegion := threeport.DefaultComputeClusterRegion
	defaultClusterProvider := threeport.DefaultComputeClusterProvider
	defaultClusterAPIEndpoint := threeport.DefaultComputeClusterAPIEndpoint
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
		return fmt.Errorf("failed to marshal workload cluster to json: %w", err)
		//qout.Error("failed to marshal workload cluster to json", err)
		//os.Exit(1)
	}
	wc, err := tpclient.CreateWorkloadCluster(wcJSON, install.GetThreeportAPIEndpoint(), "")
	if err != nil {
		return fmt.Errorf("failed to create workload cluster in Threeport API: %w", err)
		//qout.Error("failed to create workload cluster in Threeport API", err)
		//os.Exit(1)
	}
	qout.Info(fmt.Sprintf("default workload cluster %s for compute space set up", *wc.Name))

	// TODO: add superuser
	superuserID := uint(1)

	// add forward proxy definition
	fwdProxyDefName := threeport.ForwardProxyWorkloadDefinitionName
	fwdProxyYAML := install.ForwardProxyManifest()
	fwdProxyWorkloadDefinition := tpapi.WorkloadDefinition{
		Name:         &fwdProxyDefName,
		YAMLDocument: &fwdProxyYAML,
		UserID:       &superuserID,
	}
	fpwdJSON, err := json.Marshal(&fwdProxyWorkloadDefinition)
	if err != nil {
		return fmt.Errorf("failed to marshal forward proxy workload definition to json: %w", err)
		//qout.Error("failed to marshal forward proxy workload definition to json", err)
		//os.Exit(1)
	}
	fpwd, err := tpclient.CreateWorkloadDefinition(fpwdJSON, install.GetThreeportAPIEndpoint(), "")
	if err != nil {
		return fmt.Errorf("failed to create forward proxy workload definition in Threeport API: %w", err)
		//qout.Error("failed to create forward proxy workload definition in Threeport API", err)
		//os.Exit(1)
	}
	qout.Info(fmt.Sprintf("forward proxy workload definition %s added", *fpwd.Name))

	return nil
}

func (c *ControlPlane) DeleteControlPlaneOnKind() error {
	fmt.Println("deleting kind cluster...")
	kindDelete := exec.Command(
		"kind",
		"delete",
		"cluster",
		"--name",
		c.ThreeportClusterName(),
	)
	if err := kindDelete.Run(); err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w", err)
		//qout.Error("failed to delete kind cluster", err)
		//os.Exit(1)
	}
	qout.Info("kind cluster deleted")

	return nil
}
