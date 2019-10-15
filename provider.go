package terranova

import (
	"fmt"
	"sync"

	"github.com/hashicorp/terraform/configs/hcl2shim"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Provider implements the provider.Interface wrapping the legacy
// ResourceProvider but not using gRPC like terraform does
type Provider struct {
	provider *schema.Provider
	mu       sync.Mutex
	schemas  providers.GetSchemaResponse
}

// NewProvider creates a Terranova Provider to wrap the given legacy ResourceProvider
func NewProvider(provider terraform.ResourceProvider) *Provider {
	sp, ok := provider.(*schema.Provider)
	if !ok {
		return nil
	}

	return &Provider{
		provider: sp,
	}
}

// GetSchema implements the GetSchema from providers.Interface. Returns the
// complete schema for the provider.
func (p *Provider) GetSchema() (resp providers.GetSchemaResponse) {
	if p.schemas.Provider.Block != nil {
		return p.schemas
	}

	resp = providers.GetSchemaResponse{
		ResourceTypes: make(map[string]providers.Schema),
		DataSources:   make(map[string]providers.Schema),
	}

	resp.Provider = providers.Schema{
		// Version: p.provider.Schema.Version,
		Block: schema.InternalMap(p.provider.Schema).CoreConfigSchema(),
	}

	for name, resource := range p.provider.ResourcesMap {
		resp.ResourceTypes[name] = providers.Schema{
			Version: int64(resource.SchemaVersion),
			Block:   resource.CoreConfigSchema(),
		}
	}

	for name, data := range p.provider.DataSourcesMap {
		resp.DataSources[name] = providers.Schema{
			Version: int64(data.SchemaVersion),
			Block:   data.CoreConfigSchema(),
		}
	}

	p.schemas = resp

	return resp
}

// PrepareProviderConfig implements the PrepareProviderConfig from
// providers.Interface. Allows the provider to validate the configuration
// values, and set or override any values with defaults.
func (p *Provider) PrepareProviderConfig(req providers.PrepareProviderConfigRequest) (resp providers.PrepareProviderConfigResponse) {
	schema := p.getSchema()

	// lookup any required, top-level attributes that are Null, and see if we
	// have a Default value available.
	configVal, err := cty.Transform(req.Config, func(path cty.Path, val cty.Value) (cty.Value, error) {
		// we're only looking for top-level attributes
		if len(path) != 1 {
			return val, nil
		}

		// nothing to do if we already have a value
		if !val.IsNull() {
			return val, nil
		}

		// get the Schema definition for this attribute
		getAttr, ok := path[0].(cty.GetAttrStep)
		// these should all exist, but just ignore anything strange
		if !ok {
			return val, nil
		}

		attrSchema := p.provider.Schema[getAttr.Name]
		// continue to ignore anything that doesn't match
		if attrSchema == nil {
			return val, nil
		}

		// this is deprecated, so don't set it
		if attrSchema.Deprecated != "" || attrSchema.Removed != "" {
			return val, nil
		}

		// find a default value if it exists
		def, err := attrSchema.DefaultValue()
		if err != nil {
			resp.Diagnostics = resp.Diagnostics.Append(fmt.Errorf("error getting default for %q: %s", getAttr.Name, err))
			return val, err
		}

		// no default
		if def == nil {
			return val, nil
		}

		// create a cty.Value and make sure it's the correct type
		tmpVal := hcl2shim.HCL2ValueFromConfigValue(def)

		// helper/schema used to allow setting "" to a bool
		if val.Type() == cty.Bool && tmpVal.RawEquals(cty.StringVal("")) {
			// return a warning about the conversion
			resp.Diagnostics = resp.Diagnostics.Append("provider set empty string as default value for bool " + getAttr.Name)
			tmpVal = cty.False
		}

		val, err = convert.Convert(tmpVal, val.Type())
		if err != nil {
			resp.Diagnostics = resp.Diagnostics.Append(fmt.Errorf("error setting default for %q: %s", getAttr.Name, err))
		}

		return val, err
	})
	if err != nil {
		// any error here was already added to the diagnostics
		return resp
	}

	configVal, err = schema.Provider.Block.CoerceValue(configVal)
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(configVal, nil); err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	resp.PreparedConfig = configVal

	return resp
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

// getSchema is used internally to get the saved provider schema.  The schema
// should have already been fetched from the provider, but we have to
// synchronize access to avoid being called concurrently with GetSchema.
func (p *Provider) getSchema() providers.GetSchemaResponse {
	p.mu.Lock()
	// unlock inline in case GetSchema needs to be called
	if p.schemas.Provider.Block != nil {
		p.mu.Unlock()
		return p.schemas
	}
	p.mu.Unlock()

	schemas := p.GetSchema()
	if schemas.Diagnostics.HasErrors() {
		panic(schemas.Diagnostics.Err())
	}

	return schemas
}

// validateConfigNulls checks a config value for unsupported nulls before
// attempting to shim the value. While null values can mostly be ignored in the
// configuration, since they're not supported in HCL1, the case where a null
// appears in a list-like attribute (list, set, tuple) will present a nil value
// to helper/schema which can panic. Return an error to the user in this case,
// indicating the attribute with the null value.
func validateConfigNulls(v cty.Value, path cty.Path) []tfdiags.Diagnostic {
	var diags []tfdiags.Diagnostic
	if v.IsNull() || !v.IsKnown() {
		return diags
	}

	switch {
	case v.Type().IsListType() || v.Type().IsSetType() || v.Type().IsTupleType():
		it := v.ElementIterator()
		for it.Next() {
			kv, ev := it.Element()
			if ev.IsNull() {
				// if this is a set, the kv is also going to be null which
				// isn't a valid path element, so we can't append it to the
				// diagnostic.
				p := path
				if !kv.IsNull() {
					p = append(p, cty.IndexStep{Key: kv})
				}

				newDiag := tfdiags.AttributeValue(
					tfdiags.Error,
					"Null value found in list",
					"Null values are not allowed for this attribute value.",
					p,
				)

				diags = append(diags, newDiag)
				continue
			}

			d := validateConfigNulls(ev, append(path, cty.IndexStep{Key: kv}))
			diags = append(diags, d...)
		}

	case v.Type().IsMapType() || v.Type().IsObjectType():
		it := v.ElementIterator()
		for it.Next() {
			kv, ev := it.Element()
			var step cty.PathStep
			switch {
			case v.Type().IsMapType():
				step = cty.IndexStep{Key: kv}
			case v.Type().IsObjectType():
				step = cty.GetAttrStep{Name: kv.AsString()}
			}
			d := validateConfigNulls(ev, append(path, step))
			diags = append(diags, d...)
		}
	}

	return diags
}
