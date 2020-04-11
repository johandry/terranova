package terranova

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/states"
	"github.com/zclconf/go-cty/cty/json"
)

// OutputValueAsString returns the value of the Terraform output parameter in the code
func (p *Platform) OutputValueAsString(name string) (string, error) {
	if p.State == nil || p.State.Empty() {
		return "", fmt.Errorf("no state found or empty state")
	}

	// p.State.RootModule() shouldn't be null, there's no need to check it's nil
	output := p.State.RootModule().OutputValues
	if output == nil {
		return "", fmt.Errorf("no output values in the state")
	}

	if _, ok := output[name]; !ok {
		return "", fmt.Errorf("value of %q not found", name)
	}

	return valueAsString(output[name])
}

// ValueAsString returns the given OutputValue as a string in JSON format.
// Examples: `15`, `Hello`, ``, `true`, `["hello", true]`
func valueAsString(v *states.OutputValue) (s string, err error) {
	if v == nil {
		return s, nil
	}

	var b []byte
	b, err = json.Marshal(v.Value, v.Value.Type())
	if err != nil {
		return s, err
	}
	s = string(b)

	s = strings.Trim(s, `"`)
	s = strings.Replace(s, `\n`, "\n", -1)

	return s, nil
}
