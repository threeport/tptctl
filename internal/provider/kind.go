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
	ThreeportKindConfigPath  = "/tmp/threeport-kind-config.yaml"
	KindThreeportAPIProtocol = "http"
	KindThreeportAPIHostname = "localhost"
	KindThreeportAPIPort     = "1323"
)

// KindConfig returns the content of a kind config file used when installing
// threeport locally.
// https://kind.sigs.k8s.io/
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
`, c.ThreeportClusterName(), KindThreeportAPIPort)
}

// CreateControlPlaneOnKind creates a kind cluster and installs the threeport
// control plane.
// https://kind.sigs.k8s.io/
func (c *ControlPlane) CreateControlPlaneOnKind(providerConfigDir string) error {
	// write kind config file to /tmp directory
	configFile, err := os.Create(ThreeportKindConfigPath)
	if err != nil {
		return fmt.Errorf("failed to write kind config file to disk: %w", err)
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
	kindCreateOut, err := kindCreate.CombinedOutput()
	if err != nil {
		qout.Error(fmt.Sprintf("kind error: %s", kindCreateOut), nil)
		return fmt.Errorf("failed to create new kind cluster: %w", err)
	}
	qout.Info("kind cluster created")

	// write kubeconfig
	kubeconfigFilePath := filepath.Join(providerConfigDir,
		fmt.Sprintf("kubeconfig-%s", c.ThreeportClusterName()))
	kindKubeconfig := exec.Command(
		"kind",
		"get",
		"kubeconfig",
		"--name",
		c.ThreeportClusterName(),
	)
	kindKubeconfigOut, err := kindKubeconfig.CombinedOutput()
	if err != nil {
		qout.Error(fmt.Sprintf("kind error: %s", kindKubeconfigOut), nil)
		return fmt.Errorf("failed to get kubeconfig for kind cluster: %w", err)
	}
	ioutil.WriteFile(kubeconfigFilePath, []byte(kindKubeconfigOut), 0644)
	qout.Info(fmt.Sprintf("kubeconfig for kind cluster written to %s", kubeconfigFilePath))

	// install threeport API
	if err := install.InstallAPI(kubeconfigFilePath, "", "", ""); err != nil {
		return fmt.Errorf("failed to install threeport API on kind cluster: %w", err)
	}

	// install workload controller
	if err := install.InstallWorkloadController(kubeconfigFilePath); err != nil {
		return fmt.Errorf("failed to install workload controller on kind cluster: %w", err)
	}

	// wait a few seconds for everything to come up
	qout.Info("waiting for control plane components to spin up...")
	time.Sleep(time.Second * 200)

	// get kubeconfig
	defaultLoadRules := kubeclient.NewDefaultClientConfigLoadingRules()

	clientConfigLoadRules, err := defaultLoadRules.Load()
	if err != nil {
		return fmt.Errorf("failed to load default kubeconfig rules: %w", err)
	}

	clientConfig := kubeclient.NewDefaultClientConfig(*clientConfigLoadRules, &kubeclient.ConfigOverrides{})
	kubeConfig, err := clientConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
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
	}
	wc, err := tpclient.CreateWorkloadCluster(wcJSON, install.GetThreeportAPIEndpoint(), "")
	if err != nil {
		return fmt.Errorf("failed to create workload cluster in Threeport API: %w", err)
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
	}
	fpwd, err := tpclient.CreateWorkloadDefinition(fpwdJSON, install.GetThreeportAPIEndpoint(), "")
	if err != nil {
		return fmt.Errorf("failed to create forward proxy workload definition in Threeport API: %w", err)
	}
	qout.Info(fmt.Sprintf("forward proxy workload definition %s added", *fpwd.Name))

	return nil
}

// DeleteControlPlaneOnKind deletes a kind cluster used for a threeport instance
// to completely remove threeport.
func (c *ControlPlane) DeleteControlPlaneOnKind() error {
	fmt.Println("deleting kind cluster...")
	kindDelete := exec.Command(
		"kind",
		"delete",
		"cluster",
		"--name",
		c.ThreeportClusterName(),
	)
	kindDeleteOut, err := kindDelete.CombinedOutput()
	if err != nil {
		qout.Error(fmt.Sprintf("kind error: %s", kindDeleteOut), nil)
		return fmt.Errorf("failed to delete kind cluster: %w", err)
	}
	qout.Info("kind cluster deleted")

	return nil
}
