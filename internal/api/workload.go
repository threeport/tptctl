package api

import (
	"encoding/json"
	"io/ioutil"

	tpclient "github.com/threeport/threeport-go-client"
	tpapi "github.com/threeport/threeport-rest-api/pkg/api/v0"

	"github.com/threeport/tptctl/internal/install"
)

// WorkloadConfig contains the attributes needed to manage a workload.
type WorkloadConfig struct {
	Name                      string                          `yaml:"Name"`
	WorkloadDefinition        WorkloadDefinitionConfig        `yaml:"WorkloadDefinition"`
	WorkloadInstance          WorkloadInstanceConfig          `yaml:"WorkloadInstance"`
	WorkloadServiceDependency WorkloadServiceDependencyConfig `yaml:"WorkloadServiceDependency"`
}

// WorkloadDefinitionConfig contains the attributes needed to manage a workload
// definition.
type WorkloadDefinitionConfig struct {
	Name         string `yaml:"Name"`
	YAMLDocument string `yaml:"YAMLDocument"`
	UserID       uint   `yaml:"UserID"`
}

// WorkloadInstanceConfig contains the attributes needed to manage a workload
// instance.
type WorkloadInstanceConfig struct {
	Name                   string `yaml:"Name"`
	WorkloadClusterName    string `yaml:"WorkloadClusterName"`
	WorkloadDefinitionName string `yaml:"WorkloadDefinitionName"`
}

// WorkloadServiceDependencyConfig contains the attributes needed to manage a
// workload service dependency.
type WorkloadServiceDependencyConfig struct {
	Name                 string `yaml:"Name"`
	UpstreamHost         string `yaml:"UpstreamHost"`
	UpstreamPath         string `yaml:"UpstreamPath"`
	WorkloadInstanceName string `yaml:"WorkloadInstanceName"`
}

// Create creates a workload in the Threeport API.
func (wc *WorkloadConfig) Create() error {
	// create the definition
	_, aerr := wc.WorkloadDefinition.Create()
	if aerr != nil {
		return aerr
	}

	// create the instance
	_, berr := wc.WorkloadInstance.Create()
	if berr != nil {
		return berr
	}

	// create the service dependency
	_, cerr := wc.WorkloadServiceDependency.Create()
	if cerr != nil {
		return cerr
	}

	return nil
}

// Create creates a workload definition in the Threeport API.
func (wdc *WorkloadDefinitionConfig) Create() (*tpapi.WorkloadDefinition, error) {
	// get the content of the yaml document
	definitionContent, err := ioutil.ReadFile(wdc.YAMLDocument)
	if err != nil {
		return nil, err
	}
	stringContent := string(definitionContent)

	// construct workload definition object
	workloadDefinition := &tpapi.WorkloadDefinition{
		Name:         &wdc.Name,
		YAMLDocument: &stringContent,
		UserID:       &wdc.UserID,
	}

	// create workload definition in API
	wdJSON, err := json.Marshal(&workloadDefinition)
	if err != nil {
		return nil, err
	}
	wd, err := tpclient.CreateWorkloadDefinition(wdJSON, install.GetThreeportAPIEndpoint(), "")
	if err != nil {
		return nil, err
	}

	return wd, nil
}

// Create creates a workload instance in the Threeport API.
func (wic *WorkloadInstanceConfig) Create() (*tpapi.WorkloadInstance, error) {
	// get workload cluster by name
	workloadCluster, err := tpclient.GetWorkloadClusterByName(
		wic.WorkloadClusterName,
		install.GetThreeportAPIEndpoint(), "",
	)
	if err != nil {
		return nil, err
	}

	// get workload definition by name
	workloadDefinition, err := tpclient.GetWorkloadDefinitionByName(
		wic.WorkloadDefinitionName,
		install.GetThreeportAPIEndpoint(), "",
	)
	if err != nil {
		return nil, err
	}

	// construct workload instance object
	workloadInstance := &tpapi.WorkloadInstance{
		Name:                 &wic.Name,
		WorkloadClusterID:    workloadCluster.ID,
		WorkloadDefinitionID: workloadDefinition.ID,
	}

	// create workload instance in API
	wiJSON, err := json.Marshal(&workloadInstance)
	if err != nil {
		return nil, err
	}
	wi, err := tpclient.CreateWorkloadInstance(wiJSON, install.GetThreeportAPIEndpoint(), "")
	if err != nil {
		return nil, err
	}

	return wi, nil
}

// Create creates a workload service dependency in the Threeport API.
func (wsdc *WorkloadServiceDependencyConfig) Create() (*tpapi.WorkloadServiceDependency, error) {
	// get workload instance by name
	workloadInstance, err := tpclient.GetWorkloadInstanceByName(
		wsdc.WorkloadInstanceName,
		install.GetThreeportAPIEndpoint(), "",
	)
	if err != nil {
		return nil, err
	}

	// construct workload service dependency object
	workloadServiceDependency := &tpapi.WorkloadServiceDependency{
		Name:               &wsdc.Name,
		UpstreamHost:       &wsdc.UpstreamHost,
		UpstreamPath:       &wsdc.UpstreamPath,
		WorkloadInstanceID: workloadInstance.ID,
	}

	// create workload instance in API
	wsdJSON, err := json.Marshal(&workloadServiceDependency)
	if err != nil {
		return nil, err
	}
	wsd, err := tpclient.CreateWorkloadServiceDependency(wsdJSON, install.GetThreeportAPIEndpoint(), "")
	if err != nil {
		return nil, err
	}

	return wsd, nil
}

// Update updates a workload service dependency in the Threeport API.
func (wsdc *WorkloadServiceDependencyConfig) Update() (*tpapi.WorkloadServiceDependency, error) {
	// get workload instance by name
	workloadInstance, err := tpclient.GetWorkloadInstanceByName(
		wsdc.WorkloadInstanceName,
		install.GetThreeportAPIEndpoint(), "",
	)
	if err != nil {
		return nil, err
	}

	// construct workload service dependency object
	workloadServiceDependency := &tpapi.WorkloadServiceDependency{
		Name:               &wsdc.Name,
		UpstreamHost:       &wsdc.UpstreamHost,
		UpstreamPath:       &wsdc.UpstreamPath,
		WorkloadInstanceID: workloadInstance.ID,
	}

	// get existing workload service dependency by name to retrieve its ID
	existingWSD, err := tpclient.GetWorkloadServiceDependencyByName(
		wsdc.Name,
		install.GetThreeportAPIEndpoint(), "",
	)
	if err != nil {
		return nil, err
	}

	// update workload instance in API
	wsdJSON, err := json.Marshal(&workloadServiceDependency)
	if err != nil {
		return nil, err
	}
	wsd, err := tpclient.UpdateWorkloadServiceDependency(
		*existingWSD.ID,
		wsdJSON,
		install.GetThreeportAPIEndpoint(), "",
	)
	if err != nil {
		return nil, err
	}

	return wsd, nil
}
