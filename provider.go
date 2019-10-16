package terranova

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/configs/hcl2shim"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/plans/objchange"
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
	p.mu.Lock()
	defer p.mu.Unlock()

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
	// Other way to get the schema block is from the schema:
	// resourceSchema := p.getResourceSchema(req.TypeName)
	// schemaBlock := resourceSchema.Block
	schemaBlock := p.getResourceSchemaBlock(req.TypeName)

	config := terraform.NewResourceConfigShimmed(req.Config, schemaBlock)

	warns, errs := p.provider.ValidateResource(req.TypeName, config)
	resp.Diagnostics = appendWarnsAndErrsToDiags(warns, errs)

	return resp
}

// ValidateDataSourceConfig implements the ValidateDataSourceConfig from providers.Interface.
// Allows the provider to validate the data source configuration values.
func (p *Provider) ValidateDataSourceConfig(req providers.ValidateDataSourceConfigRequest) (resp providers.ValidateDataSourceConfigResponse) {
	// Other way is to get the data schema block from the schedule
	// dataSchema := p.getDatasourceSchema(req.TypeName)
	// schemaBlock := dataSchema.Block

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(req.Config, nil); err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	schemaBlock := p.getDatasourceSchemaBlock(req.TypeName)
	config := terraform.NewResourceConfigShimmed(req.Config, schemaBlock)

	warns, errs := p.provider.ValidateDataSource(req.TypeName, config)
	resp.Diagnostics = appendWarnsAndErrsToDiags(warns, errs)

	return resp
}

// UpgradeResourceState implements the UpgradeResourceState from providers.Interface.
// It is called when the state loader encounters an instance state whose schema
// version is less than the one reported by the currently-used version of the
// corresponding provider, and the upgraded result is used for any further processing.
func (p *Provider) UpgradeResourceState(req providers.UpgradeResourceStateRequest) (resp providers.UpgradeResourceStateResponse) {
	res := p.provider.ResourcesMap[req.TypeName]
	schemaBlock := p.getResourceSchemaBlock(req.TypeName)

	version := int(req.Version)

	jsonMap := map[string]interface{}{}
	var err error

	switch {
	// We first need to upgrade a flatmap state if it exists.
	// There should never be both a JSON and Flatmap state in the request.
	case len(req.RawStateFlatmap) > 0:
		jsonMap, version, err = p.upgradeFlatmapState(version, req.RawStateFlatmap, res)
		if err != nil {
			resp.Diagnostics = resp.Diagnostics.Append(err)
			return resp
		}
	// if there's a JSON state, we need to decode it.
	case len(req.RawStateJSON) > 0:
		err = json.Unmarshal(req.RawStateJSON, &jsonMap)
		if err != nil {
			resp.Diagnostics = resp.Diagnostics.Append(err)
			return resp
		}
	default:
		// log.Println("[DEBUG] no state provided to upgrade")
		return resp
	}

	// complete the upgrade of the JSON states
	jsonMap, err = p.upgradeJSONState(version, jsonMap, res)
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	// The provider isn't required to clean out removed fields
	p.removeAttributes(jsonMap, schemaBlock.ImpliedType())

	// now we need to turn the state into the default json representation, so
	// that it can be re-decoded using the actual schema.
	val, err := schema.JSONMapToStateValue(jsonMap, schemaBlock)
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	// Now we need to make sure blocks are represented correctly, which means
	// that missing blocks are empty collections, rather than null.
	// First we need to CoerceValue to ensure that all object types match.
	val, err = schemaBlock.CoerceValue(val)
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}
	// Normalize the value and fill in any missing blocks.
	val = objchange.NormalizeObjectFromLegacySDK(val, schemaBlock)

	resp.UpgradedState = val

	return resp
}

// Configure implements the Configure from providers.Interface. Configures and
// initialized the provider.
func (p *Provider) Configure(req providers.ConfigureRequest) (resp providers.ConfigureResponse) {
	p.provider.TerraformVersion = req.TerraformVersion

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(req.Config, nil); err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	schemaBlock := schema.InternalMap(p.provider.Schema).CoreConfigSchema()
	config := terraform.NewResourceConfigShimmed(req.Config, schemaBlock)
	err := p.provider.Configure(config)
	resp.Diagnostics = resp.Diagnostics.Append(err)

	return resp
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
	return p.provider.Stop()
}

// ReadResource implements the ReadResource from providers.Interface. Refreshes
// a resource and returns its current state.
func (p *Provider) ReadResource(req providers.ReadResourceRequest) (resp providers.ReadResourceResponse) {
	res := p.provider.ResourcesMap[req.TypeName]
	schemaBlock := p.getResourceSchemaBlock(req.TypeName)

	stateVal := req.PriorState

	instanceState, err := res.ShimInstanceStateFromValue(stateVal)
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	resp.Private = req.Private
	private := make(map[string]interface{})
	if len(req.Private) > 0 {
		if err := json.Unmarshal(req.Private, &private); err != nil {
			resp.Diagnostics = resp.Diagnostics.Append(err)
			return resp
		}
	}
	instanceState.Meta = private

	newInstanceState, err := res.RefreshWithoutUpgrade(instanceState, p.provider.Meta())
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	if newInstanceState == nil || newInstanceState.ID == "" {
		newStateVal := cty.NullVal(schemaBlock.ImpliedType())
		resp.NewState = newStateVal

		return resp
	}

	// helper/schema should always copy the ID over, but do it again just to be safe
	newInstanceState.Attributes["id"] = newInstanceState.ID

	newStateVal, err := hcl2shim.HCL2ValueFromFlatmap(newInstanceState.Attributes, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	newStateVal = normalizeNullValues(newStateVal, stateVal, false)
	newStateVal = copyTimeoutValues(newStateVal, stateVal)

	resp.NewState = newStateVal

	return resp
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

// Close implements the Close from providers.Interface. It's to shuts down the
// plugin process but this is not a plugin, so nothing is done here
func (p *Provider) Close() error {
	return nil
}

// Internal methods
// =============================================================================

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

// getResourceSchema is a helper to extract the schema for a resource, and
// panics if the schema is not available.
// DELETE?
func (p *Provider) getResourceSchema(name string) providers.Schema {
	schema := p.getSchema()
	resSchema, ok := schema.ResourceTypes[name]
	if !ok {
		panic("unknown resource type " + name)
	}
	return resSchema
}

func (p *Provider) getResourceSchemaBlock(name string) *configschema.Block {
	res := p.provider.ResourcesMap[name]
	return res.CoreConfigSchema()
}

func appendWarnsAndErrsToDiags(warns []string, errs []error) (diags []tfdiags.Diagnostic) {
	for _, w := range warns {
		newDiag := tfdiags.WholeContainingBody(tfdiags.Warning, w, w)
		diags = append(diags, newDiag)
	}

	for _, e := range errs {
		newDiag := tfdiags.WholeContainingBody(tfdiags.Error, e.Error(), e.Error())
		diags = append(diags, newDiag)
	}

	return diags
}

// getDatasourceSchema is a helper to extract the schema for a datasource, and
// panics if that schema is not available.
// DELETE?
func (p *Provider) getDatasourceSchema(name string) providers.Schema {
	schema := p.getSchema()
	dataSchema, ok := schema.DataSources[name]
	if !ok {
		panic("unknown data source " + name)
	}
	return dataSchema
}

func (p *Provider) getDatasourceSchemaBlock(name string) *configschema.Block {
	dat := p.provider.DataSourcesMap[name]
	return dat.CoreConfigSchema()
}

// upgradeFlatmapState takes a legacy flatmap state, upgrades it using Migrate
// state if necessary, and converts it to the new JSON state format decoded as a
// map[string]interface{}.
// upgradeFlatmapState returns the json map along with the corresponding schema
// version.
func (p *Provider) upgradeFlatmapState(version int, m map[string]string, res *schema.Resource) (map[string]interface{}, int, error) {
	// this will be the version we've upgraded so, defaulting to the given
	// version in case no migration was called.
	upgradedVersion := version

	// first determine if we need to call the legacy MigrateState func
	requiresMigrate := version < res.SchemaVersion

	schemaType := res.CoreConfigSchema().ImpliedType()

	// if there are any StateUpgraders, then we need to only compare
	// against the first version there
	if len(res.StateUpgraders) > 0 {
		requiresMigrate = version < res.StateUpgraders[0].Version
	}

	if requiresMigrate && res.MigrateState == nil {
		// Providers were previously allowed to bump the version
		// without declaring MigrateState.
		// If there are further upgraders, then we've only updated that far.
		if len(res.StateUpgraders) > 0 {
			schemaType = res.StateUpgraders[0].Type
			upgradedVersion = res.StateUpgraders[0].Version
		}
	} else if requiresMigrate {
		is := &terraform.InstanceState{
			ID:         m["id"],
			Attributes: m,
			Meta: map[string]interface{}{
				"schema_version": strconv.Itoa(version),
			},
		}

		is, err := res.MigrateState(version, is, p.provider.Meta())
		if err != nil {
			return nil, 0, err
		}

		// re-assign the map in case there was a copy made, making sure to keep
		// the ID
		m := is.Attributes
		m["id"] = is.ID

		// if there are further upgraders, then we've only updated that far
		if len(res.StateUpgraders) > 0 {
			schemaType = res.StateUpgraders[0].Type
			upgradedVersion = res.StateUpgraders[0].Version
		}
	} else {
		// the schema version may be newer than the MigrateState functions
		// handled and older than the current, but still stored in the flatmap
		// form. If that's the case, we need to find the correct schema type to
		// convert the state.
		for _, upgrader := range res.StateUpgraders {
			if upgrader.Version == version {
				schemaType = upgrader.Type
				break
			}
		}
	}

	// now we know the state is up to the latest version that handled the
	// flatmap format state. Now we can upgrade the format and continue from
	// there.
	newConfigVal, err := hcl2shim.HCL2ValueFromFlatmap(m, schemaType)
	if err != nil {
		return nil, 0, err
	}

	jsonMap, err := schema.StateValueToJSONMap(newConfigVal, schemaType)
	return jsonMap, upgradedVersion, err
}

func (p *Provider) upgradeJSONState(version int, m map[string]interface{}, res *schema.Resource) (map[string]interface{}, error) {
	var err error

	for _, upgrader := range res.StateUpgraders {
		if version != upgrader.Version {
			continue
		}

		m, err = upgrader.Upgrade(m, p.provider.Meta())
		if err != nil {
			return nil, err
		}
		version++
	}

	return m, nil
}

// Remove any attributes no longer present in the schema, so that the json can
// be correctly decoded.
func (p *Provider) removeAttributes(v interface{}, ty cty.Type) {
	// we're only concerned with finding maps that corespond to object
	// attributes
	switch v := v.(type) {
	case []interface{}:
		// If these aren't blocks the next call will be a noop
		if ty.IsListType() || ty.IsSetType() {
			eTy := ty.ElementType()
			for _, eV := range v {
				p.removeAttributes(eV, eTy)
			}
		}
		return
	case map[string]interface{}:
		// map blocks aren't yet supported, but handle this just in case
		if ty.IsMapType() {
			eTy := ty.ElementType()
			for _, eV := range v {
				p.removeAttributes(eV, eTy)
			}
			return
		}

		if ty == cty.DynamicPseudoType {
			log.Printf("[DEBUG] ignoring dynamic block: %#v\n", v)
			return
		}

		if !ty.IsObjectType() {
			// This shouldn't happen, and will fail to decode further on, so
			// there's no need to handle it here.
			log.Printf("[WARN] unexpected type %#v for map in json state", ty)
			return
		}

		attrTypes := ty.AttributeTypes()
		for attr, attrV := range v {
			attrTy, ok := attrTypes[attr]
			if !ok {
				log.Printf("[DEBUG] attribute %q no longer present in schema", attr)
				delete(v, attr)
				continue
			}

			p.removeAttributes(attrV, attrTy)
		}
	}
}

// Zero values and empty containers may be interchanged by the apply process.
// When there is a discrepency between src and dst value being null or empty,
// prefer the src value. This takes a little more liberty with set types, since
// we can't correlate modified set values. In the case of sets, if the src set
// was wholly known we assume the value was correctly applied and copy that
// entirely to the new value.
// While apply prefers the src value, during plan we prefer dst whenever there
// is an unknown or a set is involved, since the plan can alter the value
// however it sees fit. This however means that a CustomizeDiffFunction may not
// be able to change a null to an empty value or vice versa, but that should be
// very uncommon nor was it reliable before 0.12 either.
func normalizeNullValues(dst, src cty.Value, apply bool) cty.Value {
	ty := dst.Type()
	if !src.IsNull() && !src.IsKnown() {
		// Return src during plan to retain unknown interpolated placeholders,
		// which could be lost if we're only updating a resource. If this is a
		// read scenario, then there shouldn't be any unknowns at all.
		if dst.IsNull() && !apply {
			return src
		}
		return dst
	}

	// Handle null/empty changes for collections during apply.
	// A change between null and empty values prefers src to make sure the state
	// is consistent between plan and apply.
	if ty.IsCollectionType() && apply {
		dstEmpty := !dst.IsNull() && dst.IsKnown() && dst.LengthInt() == 0
		srcEmpty := !src.IsNull() && src.IsKnown() && src.LengthInt() == 0

		if (src.IsNull() && dstEmpty) || (srcEmpty && dst.IsNull()) {
			return src
		}
	}

	// check the invariants that we need below, to ensure we are working with
	// non-null and known values.
	if src.IsNull() || !src.IsKnown() || !dst.IsKnown() {
		return dst
	}

	switch {
	case ty.IsMapType(), ty.IsObjectType():
		var dstMap map[string]cty.Value
		if !dst.IsNull() {
			dstMap = dst.AsValueMap()
		}
		if dstMap == nil {
			dstMap = map[string]cty.Value{}
		}

		srcMap := src.AsValueMap()
		for key, v := range srcMap {
			dstVal, ok := dstMap[key]
			if !ok && apply && ty.IsMapType() {
				// don't transfer old map values to dst during apply
				continue
			}

			if dstVal == cty.NilVal {
				if !apply && ty.IsMapType() {
					// let plan shape this map however it wants
					continue
				}
				dstVal = cty.NullVal(v.Type())
			}

			dstMap[key] = normalizeNullValues(dstVal, v, apply)
		}

		// you can't call MapVal/ObjectVal with empty maps, but nothing was
		// copied in anyway. If the dst is nil, and the src is known, assume the
		// src is correct.
		if len(dstMap) == 0 {
			if dst.IsNull() && src.IsWhollyKnown() && apply {
				return src
			}
			return dst
		}

		if ty.IsMapType() {
			// helper/schema will populate an optional+computed map with
			// unknowns which we have to fixup here.
			// It would be preferable to simply prevent any known value from
			// becoming unknown, but concessions have to be made to retain the
			// broken legacy behavior when possible.
			for k, srcVal := range srcMap {
				if !srcVal.IsNull() && srcVal.IsKnown() {
					dstVal, ok := dstMap[k]
					if !ok {
						continue
					}

					if !dstVal.IsNull() && !dstVal.IsKnown() {
						dstMap[k] = srcVal
					}
				}
			}

			return cty.MapVal(dstMap)
		}

		return cty.ObjectVal(dstMap)

	case ty.IsSetType():
		// If the original was wholly known, then we expect that is what the
		// provider applied. The apply process loses too much information to
		// reliably re-create the set.
		if src.IsWhollyKnown() && apply {
			return src
		}

	case ty.IsListType(), ty.IsTupleType():
		// If the dst is null, and the src is known, then we lost an empty value
		// so take the original.
		if dst.IsNull() {
			if src.IsWhollyKnown() && src.LengthInt() == 0 && apply {
				return src
			}

			// if dst is null and src only contains unknown values, then we lost
			// those during a read or plan.
			if !apply && !src.IsNull() {
				allUnknown := true
				for _, v := range src.AsValueSlice() {
					if v.IsKnown() {
						allUnknown = false
						break
					}
				}
				if allUnknown {
					return src
				}
			}

			return dst
		}

		// if the lengths are identical, then iterate over each element in succession.
		srcLen := src.LengthInt()
		dstLen := dst.LengthInt()
		if srcLen == dstLen && srcLen > 0 {
			srcs := src.AsValueSlice()
			dsts := dst.AsValueSlice()

			for i := 0; i < srcLen; i++ {
				dsts[i] = normalizeNullValues(dsts[i], srcs[i], apply)
			}

			if ty.IsTupleType() {
				return cty.TupleVal(dsts)
			}
			return cty.ListVal(dsts)
		}

	case ty == cty.String:
		// The legacy SDK should not be able to remove a value during plan or
		// apply, however we are only going to overwrite this if the source was
		// an empty string, since that is what is often equated with unset and
		// lost in the diff process.
		if dst.IsNull() && src.AsString() == "" {
			return src
		}
	}

	return dst
}

// helper/schema throws away timeout values from the config and stores them in
// the Private/Meta fields. we need to copy those values into the planned state
// so that core doesn't see a perpetual diff with the timeout block.
func copyTimeoutValues(to cty.Value, from cty.Value) cty.Value {
	// if `to` is null we are planning to remove it altogether.
	if to.IsNull() {
		return to
	}
	toAttrs := to.AsValueMap()
	// We need to remove the key since the hcl2shims will add a non-null block
	// because we can't determine if a single block was null from the flatmapped
	// values. This needs to conform to the correct schema for marshaling, so
	// change the value to null rather than deleting it from the object map.
	timeouts, ok := toAttrs[schema.TimeoutsConfigKey]
	if ok {
		toAttrs[schema.TimeoutsConfigKey] = cty.NullVal(timeouts.Type())
	}

	// if from is null then there are no timeouts to copy
	if from.IsNull() {
		return cty.ObjectVal(toAttrs)
	}

	fromAttrs := from.AsValueMap()
	timeouts, ok = fromAttrs[schema.TimeoutsConfigKey]

	// timeouts shouldn't be unknown, but don't copy possibly invalid values either
	if !ok || timeouts.IsNull() || !timeouts.IsWhollyKnown() {
		// no timeouts block to copy
		return cty.ObjectVal(toAttrs)
	}

	toAttrs[schema.TimeoutsConfigKey] = timeouts

	return cty.ObjectVal(toAttrs)
}
