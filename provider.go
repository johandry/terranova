package terranova

import (
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-null/null"
	"github.com/terraform-providers/terraform-provider-template/template"
)

func (p *Platform) updateProviders() terraform.ResourceProviderResolver {
	ctxProviders := make(map[string]terraform.ResourceProviderFactory)

	for name, provider := range p.Providers {
		ctxProviders[name] = terraform.ResourceProviderFactoryFixed(provider)
	}

	p.providerResolver = terraform.ResourceProviderResolverFixed(ctxProviders)

	// TODO: Reset the providers

	return p.providerResolver
}

// AddProvider adds a new provider to the providers list
func (p *Platform) AddProvider(name string, provider terraform.ResourceProvider) *Platform {
	if p.Providers == nil {
		p.Providers = defaultProviders()
	}
	p.Providers[name] = provider

	p.updateProviders()

	return p
}

func defaultProviders() map[string]terraform.ResourceProvider {
	return map[string]terraform.ResourceProvider{
		"template": template.Provider(),
		"null":     null.Provider(),
	}
}
