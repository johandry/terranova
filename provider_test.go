package terranova_test

import "testing"

func TestDefaultProviders(t *testing.T) {
	p, err := newPlatform()
	if err != nil {
		t.Errorf("Fail to initialize platform. %s", err)
	}
	if p.Providers == nil {
		t.Errorf("No default providers. %s", p.Providers)
	}
}
