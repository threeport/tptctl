package provider

import (
	"fmt"

	"github.com/threeport/tptctl/internal/install"
)

const (
	ThreeportKindConfigPath = "/tmp/threeport-kind-config.yaml"
)

func GetThreeportKindClusterName(threeportInstanceName string) string {
	return fmt.Sprintf("threeport-%s", threeportInstanceName)
}

func KindConfig(threeportInstanceName string) string {
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
`, GetThreeportKindClusterName(threeportInstanceName), install.ThreeportAPIPort)
}
