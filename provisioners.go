/*
Copyright The Terranova Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package terranova

import (
	"fmt"
	"sync"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/provisioners"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Provisioner implements the provisioners.Interface wrapping the legacy
// terraform.ResourceProvisioner but not using gRPC like terraform does
type Provisioner struct {
	provisioner *schema.Provisioner
	mu          sync.Mutex
	schema      *configschema.Block
}

// NewProvisioner creates a Terranova Provisioner to wrap the given legacy ResourceProvisioner
func NewProvisioner(provisioner terraform.ResourceProvisioner) *Provisioner {
	sp, ok := provisioner.(*schema.Provisioner)
	if !ok {
		return nil
	}
	return &Provisioner{
		provisioner: sp,
	}
}

func provisionersFactory(rp terraform.ResourceProvisioner) provisioners.Factory {
	p := NewProvisioner(rp)
	return provisioners.FactoryFixed(p)
}

// GetSchema implements GetSchema from provisioners.Interface. It returns the
// schema for the provisioner configuration.
func (p *Provisioner) GetSchema() (resp provisioners.GetSchemaResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.schema != nil {
		return provisioners.GetSchemaResponse{
			Provisioner: p.schema,
		}
	}

	resp.Provisioner = schema.InternalMap(p.provisioner.Schema).CoreConfigSchema()

	if resp.Provisioner == nil {
		resp.Diagnostics = resp.Diagnostics.Append(fmt.Errorf("missing provisioner schema"))
		return resp
	}

	p.schema = resp.Provisioner

	return resp
}

// ValidateProvisionerConfig implements ValidateProvisionerConfig from
// provisioners.Interface. It allows the provisioner to validate the
// configuration values.
func (p *Provisioner) ValidateProvisionerConfig(req provisioners.ValidateProvisionerConfigRequest) (resp provisioners.ValidateProvisionerConfigResponse) {
	cfgSchema := schema.InternalMap(p.provisioner.Schema).CoreConfigSchema()
	config := terraform.NewResourceConfigShimmed(req.Config, cfgSchema)

	warns, errs := p.provisioner.Validate(config)
	resp.Diagnostics = appendWarnsAndErrsToDiags(warns, errs)

	return resp
}

// ProvisionResource implements ProvisionResource from provisioners.Interface.
// It runs the provisioner with provided configuration.
// ProvisionResource blocks until the execution is complete.
// If the returned diagnostics contain any errors, the resource will be
// left in a tainted state.
func (p *Provisioner) ProvisionResource(req provisioners.ProvisionResourceRequest) (resp provisioners.ProvisionResourceResponse) {
	cfgSchema := schema.InternalMap(p.provisioner.Schema).CoreConfigSchema()
	resourceConfig := terraform.NewResourceConfigShimmed(req.Config, cfgSchema)

	conn := stringMapFromValue(req.Connection)

	instanceState := &terraform.InstanceState{
		Ephemeral: terraform.EphemeralState{
			ConnInfo: conn,
		},
		Meta: make(map[string]interface{}),
	}

	if err := p.provisioner.Apply(req.UIOutput, instanceState, resourceConfig); err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
	}

	return resp
}

// Stop implements Stop from provisioners.Interface. It is called to interrupt
// the provisioner.
//
// Stop should not block waiting for in-flight actions to complete. It
// should take any action it wants and return immediately acknowledging it
// has received the stop request. Terraform will not make any further API
// calls to the provisioner after Stop is called.
//
// The error returned, if non-nil, is assumed to mean that signaling the
// stop somehow failed and that the user should expect potentially waiting
// a longer period of time.
func (p *Provisioner) Stop() error {
	return p.provisioner.Stop()
}

// Close implements Close from provisioners.Interface. It shuts down the plugin
// process if applicable.
func (p *Provisioner) Close() error {
	return nil
}

// stringMapFromValue converts a cty.Value to a map[stirng]string.
// This will panic if the val is not a cty.Map(cty.String).
func stringMapFromValue(val cty.Value) map[string]string {
	m := map[string]string{}
	if val.IsNull() || !val.IsKnown() {
		return m
	}

	for it := val.ElementIterator(); it.Next(); {
		ak, av := it.Element()
		name := ak.AsString()

		if !av.IsKnown() || av.IsNull() {
			continue
		}

		av, _ = convert.Convert(av, cty.String)
		m[name] = av.AsString()
	}

	return m
}
