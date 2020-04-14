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
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/backend/local"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

// Apply brings the platform to the desired state. It'll destroy the platform
// when `destroy` is `true`.
func (p *Platform) Apply(destroy bool) error {
	p.startMiddleware()

	p.countHook = new(local.CountHook)
	stateHook := new(local.StateHook)

	if p.Hooks == nil {
		p.Hooks = []terraform.Hook{}
	}
	p.Hooks = append(p.Hooks, p.countHook, stateHook)

	ctx, err := p.newContext(destroy)
	if err != nil {
		return err
	}

	// state := ctx.State()

	if _, diag := ctx.Refresh(); diag.HasErrors() {
		return diag.Err()
	}

	plan, diag := ctx.Plan()
	if diag.HasErrors() {
		return diag.Err()
	}
	p.ExpectedStats = NewStats().FromPlan(plan)

	stateHook.StateMgr = p.stateMgr

	sts, diag := ctx.Apply()
	p.State = sts
	// p.State = ctx.State()

	if diag.HasErrors() {
		return diag.Err()
	}
	return nil
}

// Plan returns execution plan for an existing configuration to apply to the
// platform.
func (p *Platform) Plan(destroy bool) (*plans.Plan, error) {
	p.startMiddleware()

	ctx, err := p.newContext(destroy)
	if err != nil {
		return nil, err
	}

	if _, diag := ctx.Refresh(); diag.HasErrors() {
		return nil, diag.Err()
	}

	plan, diag := ctx.Plan()
	if diag.HasErrors() {
		return nil, diag.Err()
	}

	return plan, nil
}

// startMiddleware starts the Log Middleware to intercept the logs if it has not
// been already started
func (p *Platform) startMiddleware() {
	if p.LogMiddleware == nil {
		return
	}
	if !p.LogMiddleware.IsEnabled() {
		p.LogMiddleware.Start()
	}
}

// newContext creates the Terraform context or configuration
func (p *Platform) newContext(destroy bool) (*terraform.Context, error) {
	cfg, err := p.config()
	if err != nil {
		return nil, err
	}

	vars, err := p.variables(cfg.Module.Variables)
	if err != nil {
		return nil, err
	}

	// providerResolver := providers.ResolverFixed(p.Providers)
	// provisioners := p.Provisioners

	// Create ContextOpts with the current state and variables to apply
	ctxOpts := terraform.ContextOpts{
		Config:           cfg,
		Destroy:          destroy,
		State:            p.State,
		Variables:        vars,
		ProviderResolver: providers.ResolverFixed(p.Providers),
		Provisioners:     p.Provisioners,
		Hooks:            p.Hooks,
	}

	ctx, diags := terraform.NewContext(&ctxOpts)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	// Validate the context
	if diags = ctx.Validate(); diags.HasErrors() {
		return nil, diags.Err()
	}

	return ctx, nil
}

func (p *Platform) config() (*configs.Config, error) {
	if len(p.Code) == 0 {
		return nil, fmt.Errorf("no code to apply")
	}

	// Get a temporal directory to save the infrastructure code
	cfgPath, err := ioutil.TempDir("", ".terranova")
	if err != nil {
		return nil, err
	}
	// defer os.RemoveAll(cfgPath)

	if err := p.saveCode(cfgPath); err != nil {
		return nil, err
	}

	loader, err := configload.NewLoader(&configload.Config{
		ModulesDir: filepath.Join(cfgPath, "modules"),
	})
	if err != nil {
		return nil, err
	}

	config, diags := loader.LoadConfig(cfgPath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to load the configuration. %s", diags.Error())
	}

	return config, nil
}

// Export save all the code to the given directory. The directory must exists
// and there should be code to export.
func (p *Platform) Export(dir string) error {
	if len(p.Code) == 0 {
		return fmt.Errorf("no code to export")
	}

	if err := p.saveCode(dir); err != nil {
		return err
	}

	if len(p.Vars) == 0 {
		return nil
	}

	tfvarsFile, err := os.Create(filepath.Join(dir, "terraform.tfvars"))
	if err != nil {
		return err
	}
	defer tfvarsFile.Close()

	var totalErr string
	w := bufio.NewWriter(tfvarsFile)
	for name, value := range p.Vars {
		_, err := w.WriteString(fmt.Sprintf("%s = \"%s\"\n", name, value))
		if err != nil {
			totalErr = fmt.Sprintf("%s\n\t%s", totalErr, err)
		}
	}
	if err := w.Flush(); err != nil {
		totalErr = fmt.Sprintf("%s\n\t%s", totalErr, err)
	}

	if len(totalErr) != 0 {
		return fmt.Errorf("Failed to create the terraform.tfvars file. Errors:%s", totalErr)
	}

	return nil
}

func (p *Platform) saveCode(cfgPath string) error {
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return fmt.Errorf("not found the directory to save the code. %s", err)
	}

	for filename, content := range p.Code {
		cfgFileName := filepath.Join(cfgPath, filename)

		cfgFileDir := filepath.Dir(cfgFileName)
		if _, err := os.Stat(cfgFileDir); os.IsNotExist(err) {
			os.MkdirAll(cfgFileDir, 0700)
		}

		cfgFile, err := os.Create(cfgFileName)
		if err != nil {
			return err
		}
		defer cfgFile.Close()

		if _, err = io.Copy(cfgFile, strings.NewReader(content)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Platform) variables(v map[string]*configs.Variable) (terraform.InputValues, error) {
	iv := make(terraform.InputValues)
	for name, value := range p.Vars {
		if _, declared := v[name]; !declared {
			return iv, fmt.Errorf("variable %q is not declared in the code", name)
		}

		val := &terraform.InputValue{
			Value:      cty.StringVal(fmt.Sprintf("%v", value)),
			SourceType: terraform.ValueFromCaller,
		}

		iv[name] = val
	}

	return iv, nil
}
