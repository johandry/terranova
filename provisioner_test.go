package terranova_test

import "testing"

func TestDefaultProvisioner(t *testing.T) {
	p, err := newPlatform()
	if err != nil {
		t.Errorf("Fail to initialize platform. %s", err)
	}
	if p.Provisioners == nil {
		t.Errorf("No default provisioner. %s", p.Provisioners)
	}
}
