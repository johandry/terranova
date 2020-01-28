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
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/provisioners"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-null/null"
)

func TestNewPlatform(t *testing.T) {
	type args struct {
		code  string
		hooks []terraform.Hook
	}
	tests := []struct {
		name string
		args args
		want *Platform
	}{
		{"default values", args{"fake code", []terraform.Hook{}}, &Platform{
			Code: map[string]string{"main.tf": "fake code"},
			Providers: map[addrs.Provider]providers.Factory{
				addrs.NewLegacyProvider("null"): providers.FactoryFixed(NewProvider(null.Provider())),
			},
			Provisioners:  map[string]provisioners.Factory{},
			Vars:          nil,
			State:         states.NewState(),
			Hooks:         []terraform.Hook{},
			LogMiddleware: nil,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPlatform(tt.args.code, tt.args.hooks...); !reflect.DeepEqual(got, tt.want) {
				// t.Errorf("NewPlatform() = %+v, want %+v", got, tt.want)
				if !reflect.DeepEqual(got.Code, tt.want.Code) {
					t.Errorf("NewPlatform().Code = %+v, want %+v", got.Code, tt.want.Code)
				}
				if !reflect.DeepEqual(got.Providers, tt.want.Providers) {
					for name := range got.Providers {
						if _, ok := tt.want.Providers[name]; !ok {
							t.Errorf("NewPlatform().Providers[%s] not found in wanted", name)
						}
					}
					for name := range tt.want.Providers {
						if _, ok := got.Providers[name]; !ok {
							t.Errorf("NewPlatform().Providers[%s] not found", name)
						}
					}
					// t.Errorf("NewPlatform().Providers = \n'%#v'\n, want \n'%#v'", got.Providers, tt.want.Providers)
				}
				if !reflect.DeepEqual(got.Provisioners, tt.want.Provisioners) {
					for name := range got.Provisioners {
						if _, ok := tt.want.Provisioners[name]; !ok {
							t.Errorf("NewPlatform().Provisioners[%s] not found in wanted", name)
						}
					}
					for name := range tt.want.Provisioners {
						if _, ok := got.Provisioners[name]; !ok {
							t.Errorf("NewPlatform().Provisioners[%s] not found", name)
						}
					}
					// t.Errorf("NewPlatform().Provisioners = %+v, want %+v", got.Provisioners, tt.want.Provisioners)
				}
				if !reflect.DeepEqual(got.Vars, tt.want.Vars) {
					t.Errorf("NewPlatform().Vars = %+v, want %+v", got.Vars, tt.want.Vars)
				}
				if !reflect.DeepEqual(got.State, tt.want.State) {
					t.Errorf("NewPlatform().State = %+v, want %+v", got.State, tt.want.State)
				}
				if !reflect.DeepEqual(got.Hooks, tt.want.Hooks) {
					t.Errorf("NewPlatform().Hooks = %+v, want %+v", got.Hooks, tt.want.Hooks)
				}
				if !reflect.DeepEqual(got.LogMiddleware, tt.want.LogMiddleware) {
					t.Errorf("NewPlatform().LogMiddleware = %+v, want %+v", got.LogMiddleware, tt.want.LogMiddleware)
				}
			}
		})
	}
}

func TestPlatform_BindVars(t *testing.T) {
	tests := []struct {
		name string
		vars []map[string]interface{}
		want map[string]interface{}
	}{
		{"empty", []map[string]interface{}{}, nil},
		{"one value", []map[string]interface{}{map[string]interface{}{"foo": 1}}, map[string]interface{}{"foo": 1}},
		{"multiple simple values", []map[string]interface{}{map[string]interface{}{"foo": 1, "bar": "hey"}}, map[string]interface{}{"foo": 1, "bar": "hey"}},
		{"structs", []map[string]interface{}{map[string]interface{}{"foo": struct{ name string }{"bar"}}}, map[string]interface{}{"foo": struct{ name string }{"bar"}}},
		{"same value multiple times", []map[string]interface{}{map[string]interface{}{"foo": 1}, map[string]interface{}{"foo": 2}}, map[string]interface{}{"foo": 2}},
		{"diff value multiple times", []map[string]interface{}{map[string]interface{}{"foo": 1}, map[string]interface{}{"bar": 2}}, map[string]interface{}{"foo": 1, "bar": 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPlatform("")
			for _, v := range tt.vars {
				p.BindVars(v)
			}
			if !reflect.DeepEqual(p.Vars, tt.want) {
				t.Errorf("Platform.Vars() = %v, want %v", p.Vars, tt.want)
			}
		})
	}
}
