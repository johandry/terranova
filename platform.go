package terranova

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform/config/module"
	"github.com/hashicorp/terraform/terraform"
)

// Platform is the platform to be managed by Terraform
type Platform struct {
	Path             string
	Code             string
	Providers        map[string]terraform.ResourceProvider
	Provisioners     map[string]terraform.ResourceProvisioner
	vars             map[string]interface{}
	state            *terraform.State
	plan             *terraform.Plan
	mod              *module.Tree
	context          *terraform.Context
	providerResolver terraform.ResourceProviderResolver
	provisioners     map[string]terraform.ResourceProvisionerFactory
}

// New return an instance of Platform
func New(path string, code string) (*Platform, error) {
	platform := &Platform{
		Path: path,
		Code: code,
	}
	platform.Providers = defaultProviders()
	platform.updateProviders()
	platform.Provisioners = defaultProvisioners()
	platform.updateProvisioners()

	if _, err := platform.setModule(); err != nil {
		return platform, err
	}

	return platform, nil
}

// Create is to create the platform
func (p *Platform) Create() error {
	return p.Apply(false)
}

// Destroy is to destroy/terminate an existing platform
func (p *Platform) Destroy() error {
	return p.Apply(true)
}

// Apply brings the platform to the desired state. It'll destroy the platform
// when destroy is true.
func (p *Platform) Apply(destroy bool) error {
	if p.context == nil {
		if _, err := p.Context(destroy); err != nil {
			return err
		}
	}

	if _, err := p.context.Plan(); err != nil {
		return err
	}

	if _, err := p.context.Refresh(); err != nil {
		return err
	}

	state, err := p.context.Apply()
	if err != nil {
		return err
	}
	p.state = state

	return nil
}

// Plan returns execution plan for an existing configuration to apply to the
// platform. It will create the plan if does not exists.
// It's required that a context/configuration exists
func (p *Platform) Plan() (*terraform.Plan, error) {
	if p.plan == nil {
		if p.context == nil {
			return nil, errors.New("Missing configuration to get the plan")
		}
		plan, err := p.context.Plan()
		if err != nil {
			return nil, err
		}
		p.plan = plan
	}

	return p.plan, nil
}

// Context creates the Terraform context or configuration
func (p *Platform) Context(destroy bool) (*terraform.Context, error) {

	// Create ContextOpts with the current state and variables to apply
	ctxOpts := &terraform.ContextOpts{
		Destroy:          destroy,
		State:            p.state,
		Variables:        p.vars,
		Module:           p.mod,
		ProviderResolver: p.providerResolver,
		Provisioners:     p.provisioners,
	}

	ctx, err := terraform.NewContext(ctxOpts)
	if err != nil {
		return nil, err
	}
	p.context = ctx

	return p.context, nil
}

func (p *Platform) setModule() (*module.Tree, error) {
	var cfgPath = p.Path
	// If path is not set, a temporal directory is created to store the
	// configuration files. It will be deleted when this function ends
	if len(cfgPath) == 0 {
		tmpDir, err := ioutil.TempDir("", "platform")
		if err != nil {
			return nil, err
		}
		cfgPath = tmpDir
		defer os.RemoveAll(cfgPath)
	}

	// Write the configuration file
	if len(p.Code) > 0 {
		cfgFileName := filepath.Join(cfgPath, "main.tf")
		// TODO: Verify there is no other file named `main.tf`. If it's there,
		// rename it to `main.org.tf` or similar, then defer to rename it as it was
		cfgFile, err := os.Create(cfgFileName)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(cfgFile, strings.NewReader(p.Code))
		if err != nil {
			return nil, err
		}
		cfgFile.Close()
		defer os.Remove(cfgFileName)
	}

	mod, err := module.NewTreeModule("", cfgPath)
	if err != nil {
		return nil, err
	}
	modStorage := &getter.FolderStorage{
		StorageDir: filepath.Join(cfgPath, ".tfmodules"),
	}
	if err = mod.Load(modStorage, module.GetModeNone); err != nil {
		return nil, err
	}
	p.mod = mod

	return p.mod, nil
}
