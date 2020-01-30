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
	"sync"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/backend/local"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/provisioners"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statemgr"
	"github.com/hashicorp/terraform/terraform"
	"github.com/johandry/terranova/logger"
	"github.com/terraform-providers/terraform-provider-null/null"
)

// Platform is the platform to be managed by Terraform
type Platform struct {
	Code          map[string]string
	Providers     map[addrs.Provider]providers.Factory
	Provisioners  map[string]provisioners.Factory
	Vars          map[string]interface{}
	State         *State
	Hooks         []terraform.Hook
	LogMiddleware *logger.Middleware
	stateMgr      statemgr.Writer
	countHook     *local.CountHook
	ExpectedStats *Stats
	mu            sync.Mutex
}

// State is an alias for terraform.State
type State = states.State

// NewPlatform return an instance of Platform with default values
func NewPlatform(code string, hooks ...terraform.Hook) *Platform {
	platform := &Platform{
		Hooks: hooks,
	}
	platform.addDefaultProviders()
	platform.addDefaultProvisioners()

	platform.Code = map[string]string{}
	if len(code) != 0 {
		platform.Code["main.tf"] = code
	}

	platform.State = states.NewState()

	return platform
}

func (p *Platform) addDefaultProviders() {
	p.Providers = map[addrs.Provider]providers.Factory{}
	p.AddProvider("null", null.Provider())
}

// AddProvider adds a new provider to the providers list
func (p *Platform) AddProvider(name string, provider terraform.ResourceProvider) *Platform {
	p.Providers[addrs.NewLegacyProvider(name)] = providersFactory(provider)
	return p
}

func (p *Platform) addDefaultProvisioners() {
	p.Provisioners = map[string]provisioners.Factory{}
}

// AddProvisioner adds a new provisioner to the provisioner list
func (p *Platform) AddProvisioner(name string, provisioner terraform.ResourceProvisioner) *Platform {
	p.Provisioners[name] = provisionersFactory(provisioner)
	return p
}

// AddFile adds Terraform code into a file. Such file could be in a directory,
// use os.PathSeparator as path separator.
func (p *Platform) AddFile(filename, code string) *Platform {
	if filename == "" {
		filename = "main.tf"
	}
	p.Code[filename] = code
	return p
}

// BindVars binds the map of variables to the Platform variables, to be used
// by Terraform
func (p *Platform) BindVars(vars map[string]interface{}) *Platform {
	for name, value := range vars {
		p.Var(name, value)
	}

	return p
}

// Var set a variable with it's value
func (p *Platform) Var(name string, value interface{}) *Platform {
	if len(p.Vars) == 0 {
		p.Vars = make(map[string]interface{})
	}
	p.Vars[name] = value

	return p
}

// SetMiddleware assigns the given log middleware into the Platform
func (p *Platform) SetMiddleware(lm *logger.Middleware) *Platform {
	p.LogMiddleware = lm
	return p
}
