package install

import (
	"fmt"
	"os"
	"os/exec"

	qout "github.com/threeport/tptctl/internal/output"
)

const (
	WorkloadControllerManifestPath = "/tmp/threeport-workload-controller.yaml"
	WorkloadControllerImage        = "ghcr.io/threeport/threeport-workload-controller:v0.1.3"
)

func InstallWorkloadController(kubeconfig string) error {
	// write workload controller manifest to /tmp directory
	workloadControllerManifest, err := os.Create(WorkloadControllerManifestPath)
	if err != nil {
		return fmt.Errorf("failed to write workload controller manifest to disk", err)
		//qout.Error("failed to write workload controller manifest to disk", err)
		//os.Exit(1)
	}
	defer workloadControllerManifest.Close()
	workloadControllerManifest.WriteString(WorkloadControllerManifest())
	qout.Info("Threeport workload controller manifest written to /tmp directory")

	// install workload controller
	qout.Info("installing Threeport workload controller")
	workloadControllerCreate := exec.Command(
		"kubectl",
		"--kubeconfig",
		kubeconfig,
		"apply",
		"-f",
		WorkloadControllerManifestPath,
	)
	if err := workloadControllerCreate.Run(); err != nil {
		return fmt.Errorf("failed to create workload controller", err)
		//qout.Error("failed to create workload controller", err)
		//os.Exit(1)
	}

	qout.Info("Threeport workload controller created")

	return nil
}

// WorkloadControllerManifest returns a yaml manifest for the workload controller
// with the namespace included.
func WorkloadControllerManifest() string {
	return fmt.Sprintf(`---
apiVersion: v1
kind: Secret
metadata:
  name: workload-controller-config
  namespace: %[1]s
type: Opaque
stringData:
  API_SERVER: http://threeport-api-server
  MSG_BROKER_HOST: threeport-message-broker
  MSG_BROKER_PORT: "4222"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: threeport-workload-controller
  namespace: %[1]s
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: threeport-workload-controller
  template:
    metadata:
      labels:
        app.kubernetes.io/name: threeport-workload-controller
    spec:
      containers:
      - name: workload-controller
        image: %[2]s
        imagePullPolicy: IfNotPresent
        envFrom:
          - secretRef:
              name: workload-controller-config
`, ThreeportControlPlaneNs, WorkloadControllerImage)
}
