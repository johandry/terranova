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
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
)

func TestPlatform_WriteState(t *testing.T) {
	tests := []struct {
		name    string
		state   *State
		wantW   string
		wantErr bool
	}{
		{"empty state", nil, emptyStateStr, false},
		{"test state", testState(), testStateStr, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlatform("")
			if tt.state != nil {
				p.State = tt.state
			}
			w := &bytes.Buffer{}
			_, err := p.WriteState(w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Platform.WriteState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Platform.WriteState() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestPlatform_ReadState(t *testing.T) {
	tests := []struct {
		name     string
		stateStr string
		want     *State
		wantErr  bool
	}{
		{"no state", "", states.NewState(), true},
		{"empty state", emptyStateStr, states.NewState(), false},
		{"test state", testStateStr, testState(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlatform("")
			r := strings.NewReader(tt.stateStr)
			_, err := p.ReadState(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Platform.ReadState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := p.State
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Platform.ReadState() = %v, want %v", got, tt.want)
			}
		})
	}
}

const (
	emptyStateStr = `{
  "version": 4,
  "terraform_version": "0.12.17",
  "serial": 0,
  "lineage": "",
  "outputs": {},
  "resources": []
}
`
	testStateStr = `{
  "version": 4,
  "terraform_version": "0.12.17",
  "serial": 0,
  "lineage": "",
  "outputs": {},
  "resources": [
    {
      "module": "module.child",
      "mode": "managed",
      "type": "null_resource",
      "name": "somename",
      "each": "list",
      "provider": "provider.null",
      "instances": []
    }
  ]
}
`
)

func testState() *State {
	state := states.NewState()
	childMod := state.EnsureModule(addrs.RootModuleInstance.Child("child", addrs.NoKey))
	rAddr := addrs.Resource{
		Mode: addrs.ManagedResourceMode,
		Type: "null_resource",
		Name: "somename",
	}
	childMod.SetResourceMeta(rAddr, states.EachList, rAddr.DefaultProviderConfig().Absolute(addrs.RootModuleInstance))

	return state
}

func testStateFile() *statefile.File {
	state := testState()
	return &statefile.File{
		Lineage: "test-lineage",
		Serial:  0,
		// TerraformVersion: ver.Must(ver.NewVersion("0.12.0")),
		State: state,
	}
}
