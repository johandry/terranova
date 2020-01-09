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

	"github.com/hashicorp/terraform/terraform"
)

func TestPlatform_AddFile_saveCode(t *testing.T) {
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
		{"no code", "", map[string]string{}, tmpDir, map[string]string{}, false},
		{"bad dir", "", map[string]string{}, "/fake", map[string]string{}, true},
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

			err := p.saveCode(thisTestDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Platform.saveCode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			got, err := getSavedFiles(thisTestDir)
			if err != nil {
				t.Errorf("Platform.saveCode() failed getting the saved files. %s", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Platform.saveCode() error = %v, want %v", got, tt.want)
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
	type fields struct {
		Code      string
		CodeFiles map[string]string
		Vars      map[string]interface{}
		Providers map[string]terraform.ResourceProvider
		State     *State
		Hooks     []terraform.Hook
	}

	tests := []struct {
		name    string
		fields  fields
		destroy bool
		wantErr bool
	}{
		{"null data source", fields{Code: nullDataSource}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlatform(tt.fields.Code, tt.fields.Hooks...).BindVars(tt.fields.Vars)
			if tt.fields.State != nil {
				p.State = tt.fields.State
			}
			for filename, code := range tt.fields.CodeFiles {
				p.AddFile(filename, code)
			}

			for pName, provider := range tt.fields.Providers {
				p.AddProvider(pName, provider)
			}

			if err := p.Apply(tt.destroy); (err != nil) != tt.wantErr {
				t.Errorf("Platform.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
