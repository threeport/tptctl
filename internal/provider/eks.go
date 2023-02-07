package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/nukleros/eks-cluster/pkg/resource"
	"github.com/threeport/tptctl/internal/install"
	qout "github.com/threeport/tptctl/internal/output"
)

func (c *ControlPlane) CreateControlPlaneOnEKS(providerConfigDir string) error {
	// create eks resource config
	resourceConfig := resource.NewResourceConfig()
	resourceConfig.Name = c.ThreeportClusterName()
	resourceConfig.InstanceTypes = []string{"t3.medium"}
	resourceConfig.Tags = map[string]string{"provisioner": "threeport"}

	// create eks resource client
	msgChan := make(chan string)
	go outputMessages(&msgChan)
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config for AWS: %w")
	}
	resourceClient := resource.ResourceClient{&msgChan, ctx, cfg}

	// create resources in aws
	qout.Info("Creating resources for EKS cluster...")
	inventory, createErr := resourceClient.CreateResourceStack(resourceConfig)

	// write inventory file
	// important: write file even if there was some error so we can clean up
	inventoryJSON, err := json.MarshalIndent(inventory, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal inventory to JSON: %w")
	}
	//inventoryFilePath := filepath.Join(providerConfigDir,
	//	fmt.Sprintf("eks-inventory-%s.json", c.ThreeportClusterName()))
	ioutil.WriteFile(c.inventoryFilePath(providerConfigDir), inventoryJSON, 0644)

	// handle any resource creation error
	if createErr != nil {
		qout.Error("Problem encountered creating resources: %w - deleting resources that were created...", err)
		if deleteErr := resourceClient.DeleteResourceStack(inventory); deleteErr != nil {
			return fmt.Errorf("\nerror creating resources: %w\nerror deleting resources: %w", err, deleteErr)
		}
		return fmt.Errorf("error creating resources: %w", createErr)
	}

	// update kubeconfig
	//kubeconfigFilePath := filepath.Join(providerConfigDir,
	//	fmt.Sprintf("kubeconfig-%s", c.ThreeportClusterName()))
	updateKubeconfig := exec.Command(
		"aws",
		"eks",
		"update-kubeconfig",
		"--name",
		c.ThreeportClusterName(),
		"--kubeconfig",
		c.kubeconfigFilePath(providerConfigDir),
	)
	if err := updateKubeconfig.Run(); err != nil {
		return fmt.Errorf("failed to update kubeconfig: %w", err)
		//qout.Error("failed to create new kind cluster", err)
		//os.Exit(1)
	}
	qout.Info("kubeconfig updated to include new EKS cluster")

	// install threeport API
	if err := install.InstallAPI(c.kubeconfigFilePath(providerConfigDir)); err != nil {
		return fmt.Errorf("failed to install threeport API on EKS cluster: %w", err)
	}

	// install workload controller
	if err := install.InstallWorkloadController(c.kubeconfigFilePath(providerConfigDir)); err != nil {
		return fmt.Errorf("failed to install workload controller on EKS cluster: %w", err)
	}

	return nil
}

func (c *ControlPlane) DeleteControlPlaneOnEKS(providerConfigDir string) error {
	// load inventory
	//inventoryFilePath := filepath.Join(providerConfigDir,
	//	fmt.Sprintf("eks-inventory-%s.json", c.ThreeportClusterName()))
	//inventoryFilePath, err := c.eksInventoryFile()
	//if err != nil {
	//	return fmt.Errorf("failed to build filepath to inventory file: %w", err)
	//}
	var resourceInventory resource.ResourceInventory
	inventoryJSON, err := ioutil.ReadFile(c.inventoryFilePath(providerConfigDir))
	if err != nil {
		return fmt.Errorf("failed to read inventory file: %w", err)
	}
	json.Unmarshal(inventoryJSON, &resourceInventory)

	// create eks resource client
	msgChan := make(chan string)
	go outputMessages(&msgChan)
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	resourceClient := resource.ResourceClient{&msgChan, ctx, cfg}

	// delete resources
	qout.Info("Deleting resources for EKS cluster...")
	if err := resourceClient.DeleteResourceStack(&resourceInventory); err != nil {
		return fmt.Errorf("failed to delete EKS resources: %w", err)
	}

	// remove inventory file
	if err := os.Remove(c.inventoryFilePath(providerConfigDir)); err != nil {
		fmt.Errorf("failed to remove inventory file: %w", err)
	}

	// remove kubeconfig
	//kubeconfigFilePath := filepath.Join(providerConfigDir,
	//	fmt.Sprintf("kubeconfig-%s", c.ThreeportClusterName()))
	if err := os.Remove(c.kubeconfigFilePath(providerConfigDir)); err != nil {
		fmt.Errorf("failed to remove kubeconfig file: %w", err)
	}

	// update kubeconfig

	//// update kubeconfig to delete cluster, user, context
	//deleteKubeconfigCluster := exec.Command(
	//	"kubectl",
	//	"config",
	//	"delete-cluster",
	//	resourceInventory.Cluster.ClusterARN,
	//)
	//if err := deleteKubeconfigCluster.Run(); err != nil {
	//	return fmt.Errorf("failed to delete cluster from kubeconfig: %w", err)
	//}
	//deleteKubeconfigUser := exec.Command(
	//	"kubectl",
	//	"config",
	//	"delete-user",
	//	resourceInventory.Cluster.ClusterARN,
	//)
	//if err := deleteKubeconfigUser.Run(); err != nil {
	//	return fmt.Errorf("failed to delete user from kubeconfig: %w", err)
	//}
	//unsetKubeconfigContext := exec.Command(
	//	"kubectl",
	//	"config",
	//	"unset",
	//	"current-context",
	//)
	//if err := unsetKubeconfigContext.Run(); err != nil {
	//	return fmt.Errorf("failed to unset current context in kubeconfig: %w", err)
	//}
	//deleteKubeconfigContext := exec.Command(
	//	"kubectl",
	//	"config",
	//	"delete-context",
	//	resourceInventory.Cluster.ClusterARN,
	//)
	//if err := deleteKubeconfigContext.Run(); err != nil {
	//	return fmt.Errorf("failed to delete context from kubeconfig: %w", err)
	//}

	return nil
}

func (c *ControlPlane) inventoryFilePath(providerConfigDir string) string {
	return filepath.Join(
		providerConfigDir,
		fmt.Sprintf("eks-inventory-%s.json", c.ThreeportClusterName()),
	)
}

func (c *ControlPlane) kubeconfigFilePath(providerConfigDir string) string {
	return filepath.Join(
		providerConfigDir,
		fmt.Sprintf("kubeconfig-%s", c.ThreeportClusterName()),
	)
}

//func (c *ControlPlane) eksInventoryFile() (string, error) {
//	homeDir, err := os.UserHomeDir()
//	if err != nil {
//		return "", fmt.Errorf("failed to get user's home directory: %w", err)
//	}
//	inventoryFilePath := filepath.Join(
//		homeDir, ".config", "threeport",
//		fmt.Sprintf("eks-inventory-%s.json", c.ThreeportClusterName()))
//
//	return inventoryFilePath, nil
//}

func outputMessages(msgChan *chan string) {
	for {
		msg := <-*msgChan
		qout.Info(msg)
	}
}
