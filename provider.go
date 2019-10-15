package terranova

import (
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
)

// Provider implements the provider.Interface wrapping the legacy
// ResourceProvider but not using gRPC like terraform does
type Provider struct {
	provider terraform.ResourceProvider
}

// NewProvider creates a Terranova Provider to wrap the given legacy ResourceProvider
func NewProvider(provider terraform.ResourceProvider) *Provider {
	return &Provider{
		provider: provider,
	}
}

// GetSchema implements the GetSchema from providers.Interface. Returns the
// complete schema for the provider.
func (p *Provider) GetSchema() (resp providers.GetSchemaResponse) {
	return providers.GetSchemaResponse{}
}

// PrepareProviderConfig implements the PrepareProviderConfig from
// providers.Interface. Allows the provider to validate the configuration
// values, and set or override any values with defaults.
func (p *Provider) PrepareProviderConfig(req providers.PrepareProviderConfigRequest) (resp providers.PrepareProviderConfigResponse) {
	return providers.PrepareProviderConfigResponse{}
}

// ValidateResourceTypeConfig implements the ValidateResourceTypeConfig from
// providers.Interface. Allows the provider to validate the resource
// configuration values.
func (p *Provider) ValidateResourceTypeConfig(req providers.ValidateResourceTypeConfigRequest) (resp providers.ValidateResourceTypeConfigResponse) {
	return providers.ValidateResourceTypeConfigResponse{}
}

// ValidateDataSourceConfig implements the ValidateDataSourceConfig from providers.Interface.
// Allows the provider to validate the data source configuration values.
func (p *Provider) ValidateDataSourceConfig(req providers.ValidateDataSourceConfigRequest) (resp providers.ValidateDataSourceConfigResponse) {
	return providers.ValidateDataSourceConfigResponse{}
}

// UpgradeResourceState implements the UpgradeResourceState from providers.Interface.
// It is called when the state loader encounters an instance state whose schema
// version is less than the one reported by the currently-used version of the
// corresponding provider, and the upgraded result is used for any further processing.
func (p *Provider) UpgradeResourceState(req providers.UpgradeResourceStateRequest) (resp providers.UpgradeResourceStateResponse) {
	return providers.UpgradeResourceStateResponse{}
}

// Configure implements the Configure from providers.Interface. Configures and
// initialized the provider.
func (p *Provider) Configure(req providers.ConfigureRequest) (resp providers.ConfigureResponse) {
	return providers.ConfigureResponse{}
}

// Stop implements the Stop from providers.Interface. It is called when the
// provider should halt any in-flight actions.
//
// Stop should not block waiting for in-flight actions to complete. It
// should take any action it wants and return immediately acknowledging it
// has received the stop request. Terraform will not make any further API
// calls to the provider after Stop is called.
//
// The error returned, if non-nil, is assumed to mean that signaling the
// stop somehow failed and that the user should expect potentially waiting
// a longer period of time.
func (p *Provider) Stop() error {
	return nil
}

// ReadResource implements the ReadResource from providers.Interface. Refreshes
// a resource and returns its current state.
func (p *Provider) ReadResource(req providers.ReadResourceRequest) (resp providers.ReadResourceResponse) {
	return providers.ReadResourceResponse{}
}

// PlanResourceChange implements the PlanResourceChange from providers.Interface.
// Takes the current state and proposed state of a resource, and returns the
// planned final state.
func (p *Provider) PlanResourceChange(req providers.PlanResourceChangeRequest) (resp providers.PlanResourceChangeResponse) {
	return providers.PlanResourceChangeResponse{}
}

// ApplyResourceChange implements the ApplyResourceChange from providers.Interface.
// Takes the planned state for a resource, which may yet contain unknown computed
// values, and applies the changes returning the final state.
func (p *Provider) ApplyResourceChange(req providers.ApplyResourceChangeRequest) (resp providers.ApplyResourceChangeResponse) {
	return providers.ApplyResourceChangeResponse{}
}

// ImportResourceState implements the ImportResourceState from providers.Interface.
// Requests that the given resource be imported.
func (p *Provider) ImportResourceState(req providers.ImportResourceStateRequest) (resp providers.ImportResourceStateResponse) {
	return providers.ImportResourceStateResponse{}
}

// ReadDataSource implements the ReadDataSource from providers.Interface.
// Returns the data source's current state.
func (p *Provider) ReadDataSource(req providers.ReadDataSourceRequest) (resp providers.ReadDataSourceResponse) {
	return providers.ReadDataSourceResponse{}
}

// Close implements the Close from providers.Interface. Shuts down the plugin
// process if applicable.
func (p *Provider) Close() error {
	return nil
}
