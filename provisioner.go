package terranova

import (
	"github.com/hashicorp/terraform/builtin/provisioners/file"
	localexec "github.com/hashicorp/terraform/builtin/provisioners/local-exec"
	remoteexec "github.com/hashicorp/terraform/builtin/provisioners/remote-exec"
	"github.com/hashicorp/terraform/terraform"
)

func (p *Platform) updateProvisioners() map[string]terraform.ResourceProvisionerFactory {
	ctxProvisioners := make(map[string]terraform.ResourceProvisionerFactory)

	for name, provisioner := range p.Provisioners {
		ctxProvisioners[name] = func() (terraform.ResourceProvisioner, error) {
			return provisioner, nil
		}
	}

	return ctxProvisioners
}

// AddProvisioner adds a new provisioner to the provisioner list
func (p *Platform) AddProvisioner(name string, provisioner terraform.ResourceProvisioner) *Platform {
	if p.Provisioners == nil {
		p.Provisioners = defaultProvisioners()
	}
	p.Provisioners[name] = provisioner

	p.updateProvisioners()

	return p
}

func defaultProvisioners() map[string]terraform.ResourceProvisioner {
	return map[string]terraform.ResourceProvisioner{
		"local-exec":  localexec.Provisioner(),
		"remote-exec": remoteexec.Provisioner(),
		"file":        file.Provisioner(),
	}
}
