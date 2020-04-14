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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func TestPlatform_AddFile_Export(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", ".terranova_test")
	if err != nil {
		t.Errorf("Failed to create temporal directory. %s", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		code      string
		codeFiles map[string]string
		cfgPath   string
		want      map[string]string
		wantErr   bool
	}{
		{"no code", "", map[string]string{}, tmpDir, map[string]string{}, true},
		{"bad dir", "some fake code", map[string]string{}, "/fake", map[string]string{"main.tf": "some fake code"}, true},
		{"bad dir and no code", "", map[string]string{}, "/fake", map[string]string{}, true},
		{"initial main code", "some fake code", map[string]string{}, tmpDir, map[string]string{"main.tf": "some fake code"}, false},
		{"added main code only", "", map[string]string{"": "some fake code"}, tmpDir, map[string]string{"main.tf": "some fake code"}, false},
		{"initial and added main code", "fake code you won't see", map[string]string{"": "some fake code"}, tmpDir, map[string]string{"main.tf": "some fake code"}, false},
		{"using filename", "fake code you won't see", map[string]string{"main.tf": "some fake code"}, tmpDir, map[string]string{"main.tf": "some fake code"}, false},
		{"2 files", "fake main code", map[string]string{"test.tf": "fake test code"}, tmpDir,
			map[string]string{
				"main.tf": "fake main code",
				"test.tf": "fake test code",
			}, false,
		},
		{"directories", "fake main code", map[string]string{"dir1/test.tf": "fake test code"}, tmpDir,
			map[string]string{
				"main.tf":      "fake main code",
				"dir1/test.tf": "fake test code",
			}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thisTestDir := filepath.Join(tt.cfgPath, strings.ReplaceAll(tt.name, " ", "_"))
			if _, err := os.Stat(tt.cfgPath); !os.IsNotExist(err) {
				os.MkdirAll(thisTestDir, 0700)
				defer os.RemoveAll(thisTestDir)
			}

			p := NewPlatform(tt.code)
			for file, content := range tt.codeFiles {
				p.AddFile(file, content)
			}
			if !reflect.DeepEqual(p.Code, tt.want) {
				t.Errorf("Platform.AddFile() error, Platform.Code = %v, want %v", p.Code, tt.want)
			}

			err := p.Export(thisTestDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Platform.Export() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			got, err := getSavedFiles(thisTestDir)
			if err != nil {
				t.Errorf("Platform.Export() failed getting the saved files. %s", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Platform.Export() error = %v, want %v", got, tt.want)
			}
		})
	}

	testMainCodeOrder(t)
}

// Special test case, to test the order of adding a main code with empty file name and 'main.tf' file name
func testMainCodeOrder(t *testing.T) {
	// "" -> "main.tf"
	p := NewPlatform("").AddFile("", "fake code 1").AddFile("main.tf", "fake code 2")
	want, ok := p.Code["main.tf"]
	if !ok {
		t.Errorf("Platform.AddFile() error, the 'main.tf' was not found in the platform codes")
	}
	if want != "fake code 2" {
		t.Errorf("Platform.AddFile() error = %v, want %v", "fake code 2", want)
	}

	// "main.tf" -> ""
	p1 := NewPlatform("").AddFile("main.tf", "fake code 2").AddFile("", "fake code 1")
	want, ok = p1.Code["main.tf"]
	if !ok {
		t.Errorf("Platform.AddFile() error, the 'main.tf' was not found in the platform codes")
	}
	if want != "fake code 1" {
		t.Errorf("Platform.AddFile() error = %v, want %v", "fake code 1", want)
	}
}

func getSavedFiles(cfgPath string) (map[string]string, error) {
	got := map[string]string{}

	err := filepath.Walk(cfgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %q. %s", path, err)
		}

		if info.IsDir() {
			return nil
		}

		fileContent, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("fail to read file %q", path)
		}
		fname := strings.TrimPrefix(path, cfgPath+string(filepath.Separator))
		got[fname] = string(fileContent)

		return nil
	})

	return got, err
}

func TestPlatform_Apply(t *testing.T) {
	tests := []struct {
		name           string
		platformFields platformFields
		destroy        bool
		wantErr        bool
	}{
		{"null data source", testsPlatformsFields["null data source"], false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPlatformForTest(tt.platformFields)

			if err := p.Apply(tt.destroy); (err != nil) != tt.wantErr {
				t.Errorf("Platform.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlatform_Plan(t *testing.T) {
	tests := []struct {
		name           string
		platformFields platformFields
		destroy        bool
		wantErr        bool
	}{
		{"null data source", testsPlatformsFields["null data source"], false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPlatformForTest(tt.platformFields)

			_, err := p.Plan(tt.destroy)
			if (err != nil) != tt.wantErr {
				t.Errorf("Platform.Plan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func newPlatformForTest(tt platformFields) *Platform {
	p := NewPlatform(tt.Code, tt.Hooks...).BindVars(tt.Vars)
	if tt.State != nil {
		p.State = tt.State
	}
	for filename, code := range tt.CodeFiles {
		p.AddFile(filename, code)
	}

	for pName, provider := range tt.Providers {
		p.Providers[addrs.NewLegacyProvider(pName)] = provider
	}

	return p
}

// NewMockProvider creates a MockProvider from the given schema.
func NewMockProvider(t *testing.T, name string, schema *terraform.ProviderSchema) *terraform.MockProvider {
	p := new(terraform.MockProvider)

	if schema == nil {
		schema = &terraform.ProviderSchema{} // default schema is empty
	}
	p.GetSchemaReturn = schema

	p.ApplyResourceChangeFn = func(req providers.ApplyResourceChangeRequest) providers.ApplyResourceChangeResponse {
		return providers.ApplyResourceChangeResponse{
			NewState: cty.UnknownAsNull(req.PlannedState),
		}
	}
	p.PlanResourceChangeFn = func(req providers.PlanResourceChangeRequest) providers.PlanResourceChangeResponse {
		// return providers.PlanResourceChangeResponse{
		// 	PlannedState: req.ProposedNewState,
		// }
		rSchema, _ := schema.SchemaForResourceType(addrs.ManagedResourceMode, req.TypeName)
		if rSchema == nil {
			rSchema = &configschema.Block{} // default schema is empty
		}
		plannedVals := map[string]cty.Value{}
		for name, attrS := range rSchema.Attributes {
			val := req.ProposedNewState.GetAttr(name)
			if attrS.Computed && val.IsNull() {
				val = cty.UnknownVal(attrS.Type)
			}
			plannedVals[name] = val
		}
		for name := range rSchema.BlockTypes {
			plannedVals[name] = req.ProposedNewState.GetAttr(name)
		}

		return providers.PlanResourceChangeResponse{
			PlannedState:   cty.ObjectVal(plannedVals),
			PlannedPrivate: req.PriorPrivate,
		}
	}
	p.ReadResourceFn = func(req providers.ReadResourceRequest) providers.ReadResourceResponse {
		return providers.ReadResourceResponse{NewState: req.PriorState}
	}
	p.ReadDataSourceFn = func(req providers.ReadDataSourceRequest) providers.ReadDataSourceResponse {
		return providers.ReadDataSourceResponse{State: req.Config}
	}

	return p
}

type platformFields struct {
	Code      string
	CodeFiles map[string]string
	Vars      map[string]interface{}
	Providers map[string]providers.Factory
	State     *State
	Hooks     []terraform.Hook
}

var testsPlatformsFields = map[string]platformFields{
	"null data source": platformFields{Code: nullDataSource},
	"test instance": platformFields{
		Code: testSimpleInstance,
		Providers: map[string]providers.Factory{
			"test": providers.FactoryFixed(NewMockProvider(nil, "test", testSimpleSchema())),
		},
	},
	"test instance with data source": platformFields{
		Code: testInstanceWithDataSource,
		Providers: map[string]providers.Factory{
			"test": providers.FactoryFixed(NewMockProvider(nil, "test", testSchemaWithDataSource())),
		},
	},
}

const nullDataSource = `data "null_data_source" "values" {
  inputs = {
    all_server_ids = "foo"
    all_server_ips = "bar"
  }
}

output "all_server_ids" {
  value = "${data.null_data_source.values.outputs["all_server_ids"]}"
}

output "all_server_ips" {
  value = "${data.null_data_source.values.outputs["all_server_ips"]}"
}
`

const testSimpleInstance = `
resource "test_instance" "foo" {
	ami = "bar"
}`

const testInstanceWithDataSource = `
resource "test_instance" "foo" {
	ami = "bar"

	network_interface {
		device_index = 0
		description = "Main network interface"
	}
}
data "test_ds" "bar" {
  filter = "foo"
}
`

func testSimpleSchema() *terraform.ProviderSchema {
	return &terraform.ProviderSchema{
		ResourceTypes: map[string]*configschema.Block{
			"test_instance": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"ami": {Type: cty.String, Optional: true},
				},
			},
		},
	}
}

func testSchemaWithDataSource() *terraform.ProviderSchema {
	return &terraform.ProviderSchema{
		Provider: &configschema.Block{
			Attributes: map[string]*configschema.Attribute{
				"region": {Type: cty.String, Required: true},
			},
		},
		ResourceTypes: map[string]*configschema.Block{
			"test_instance": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"ami": {Type: cty.String, Optional: true},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"network_interface": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"device_index": {Type: cty.String, Optional: true},
								"description":  {Type: cty.String, Optional: true},
							},
						},
					},
				},
			},
		},
		DataSources: map[string]*configschema.Block{
			"test_data_source": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"ami": {Type: cty.String, Optional: true},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"network_interface": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"device_index": {Type: cty.String, Optional: true},
								"description":  {Type: cty.String, Optional: true},
							},
						},
					},
				},
			},
		},

		ResourceTypeSchemaVersions: map[string]uint64{
			"test_instance":    42,
			"test_data_source": 3,
		},
	}
}
