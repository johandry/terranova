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
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/provisioners"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/hashicorp/terraform/terraform"
	"github.com/johandry/terranova/logger"
	"github.com/terraform-providers/terraform-provider-null/null"
)

// Platform is the platform to be managed by Terraform
type Platform struct {
	Code          string
	Providers     map[string]providers.Factory
	Provisioners  map[string]provisioners.Factory
	Vars          map[string]interface{}
	State         *State
	Hooks         []terraform.Hook
	LogMiddleware *logger.Middleware
}

// State is an alias for terraform.State
type State = states.State

// NewPlatform return an instance of Platform with default values
func NewPlatform(code string, hooks ...terraform.Hook) *Platform {
	platform := &Platform{
		Code:  code,
		Hooks: hooks,
	}
	platform.addDefaultProviders()
	platform.addDefaultProvisioners()

	platform.State = states.NewState()

	return platform
}

func (p *Platform) addDefaultProviders() {
	p.Providers = map[string]providers.Factory{}
	p.AddProvider("null", null.Provider())
}

// AddProvider adds a new provider to the providers list
func (p *Platform) AddProvider(name string, provider terraform.ResourceProvider) *Platform {
	p.Providers[name] = providersFactory(provider)
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

// WriteState takes a io.Writer as input to write the Terraform state
func (p *Platform) WriteState(w io.Writer) (*Platform, error) {
	sf := statefile.New(p.State, "", 0)
	return p, statefile.Write(sf, w)
}

// ReadState takes a io.Reader as input to read from it the Terraform state
func (p *Platform) ReadState(r io.Reader) (*Platform, error) {
	sf, err := statefile.Read(r)
	if err != nil {
		return p, err
	}
	p.State = sf.State
	return p, nil
}

// WriteStateToFile save the state of the Terraform state to a file
func (p *Platform) WriteStateToFile(filename string) (*Platform, error) {
	var state bytes.Buffer
	if _, err := p.WriteState(&state); err != nil {
		return p, err
	}
	return p, ioutil.WriteFile(filename, state.Bytes(), 0644)
}

// ReadStateFromFile will load the Terraform state from a file and assign it to the
// Platform state.
func (p *Platform) ReadStateFromFile(filename string) (*Platform, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return p, err
	}
	return p.ReadState(file)
}

// AddMiddleware adds the given log middleware into the Platform
func (p *Platform) AddMiddleware(lm *logger.Middleware) *Platform {
	p.LogMiddleware = lm
	return p
}
