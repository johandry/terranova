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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/config/module"
	"github.com/hashicorp/terraform/terraform"
)

// Apply brings the platform to the desired state. It'll destroy the platform
// when `destroy` is `true`.
func (p *Platform) Apply(destroy bool) error {
	ctx, err := p.newContext(destroy)
	if err != nil {
		return err
	}

	// state := ctx.State()

	if _, err := ctx.Refresh(); err != nil {
		return err
	}

	if _, err := ctx.Plan(); err != nil {
		return err
	}

	_, err = ctx.Apply()
	p.State = ctx.State()

	return err
}

// Plan returns execution plan for an existing configuration to apply to the
// platform.
func (p *Platform) Plan(destroy bool) (*terraform.Plan, error) {
	ctx, err := p.newContext(destroy)
	if err != nil {
		return nil, err
	}

	if _, err := ctx.Refresh(); err != nil {
		return nil, err
	}

	plan, err := ctx.Plan()
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// newContext creates the Terraform context or configuration
func (p *Platform) newContext(destroy bool) (*terraform.Context, error) {
	module, err := p.module()
	if err != nil {
		return nil, err
	}

	providerResolver := p.getProviderResolver()
	provisioners := p.getProvisioners()

	// Create ContextOpts with the current state and variables to apply
	ctxOpts := &terraform.ContextOpts{
		Destroy:          destroy,
		State:            p.State,
		Variables:        p.Vars,
		Module:           module,
		ProviderResolver: providerResolver,
		Provisioners:     provisioners,
	}

	ctx, err := terraform.NewContext(ctxOpts)
	if err != nil {
		return nil, err
	}

	// TODO: Validate the context

	return ctx, nil
}

func (p *Platform) module() (*module.Tree, error) {
	if len(p.Code) == 0 {
		return nil, fmt.Errorf("no code to apply")
	}

	// Get a temporal directory to save the infrastructure code
	cfgPath, err := ioutil.TempDir("", ".terranova")
	if err != nil {
		return nil, err
	}
	// This defer is executed second
	defer os.RemoveAll(cfgPath)

	// Save the infrastructure code
	cfgFileName := filepath.Join(cfgPath, "main.tf")
	cfgFile, err := os.Create(cfgFileName)
	if err != nil {
		return nil, err
	}
	// This defer is executed first
	defer cfgFile.Close()
	if _, err = io.Copy(cfgFile, strings.NewReader(p.Code)); err != nil {
		return nil, err
	}

	mod, err := module.NewTreeModule("", cfgPath)
	if err != nil {
		return nil, err
	}

	s := module.NewStorage(filepath.Join(cfgPath, "modules"), nil)
	s.Mode = module.GetModeNone // or module.GetModeGet?

	if err := mod.Load(s); err != nil {
		return nil, fmt.Errorf("failed to load the modules. %s", err)
	}

	if err := mod.Validate().Err(); err != nil {
		return nil, fmt.Errorf("failed Terraform code validation. %s", err)
	}

	return mod, nil
}

func (p *Platform) getProviderResolver() terraform.ResourceProviderResolver {
	ctxProviders := make(map[string]terraform.ResourceProviderFactory)

	for name, provider := range p.Providers {
		ctxProviders[name] = terraform.ResourceProviderFactoryFixed(provider)
	}

	providerResolver := terraform.ResourceProviderResolverFixed(ctxProviders)

	// TODO: Reset the providers?

	return providerResolver
}

func (p *Platform) getProvisioners() map[string]terraform.ResourceProvisionerFactory {
	provisioners := make(map[string]terraform.ResourceProvisionerFactory)

	for name, provisioner := range p.Provisioners {
		provisioners[name] = func() (terraform.ResourceProvisioner, error) {
			return provisioner, nil
		}
	}

	return provisioners
}
