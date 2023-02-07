package provider

import "fmt"

type ControlPlane struct {
	InstanceName string
}

func (c *ControlPlane) ThreeportClusterName() string {
	return fmt.Sprintf("threeport-%s", c.InstanceName)
}
