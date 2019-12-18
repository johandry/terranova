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
)

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
