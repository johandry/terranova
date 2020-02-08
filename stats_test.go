package terranova

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/backend/local"
)

func TestStats_FromPlan(t *testing.T) {
	type statsFields struct {
		Add      int
		Change   int
		Destroy  int
		fromPlan bool
	}
	tests := []struct {
		name           string
		platformFields platformFields
		destroy        bool
		fields         statsFields
		want           *Stats
		wantStr        string
	}{
		{"null data source", testsPlatformsFields["null data source"], false, statsFields{-1, -1, -1, false}, &Stats{0, 0, 0, true}, "resources: 0 to add, 0 to change, 0 to destroy"},
		{"test instance", testsPlatformsFields["test instance"], false, statsFields{-1, -1, -1, false}, &Stats{1, 0, 0, true}, "resources: 1 to add, 0 to change, 0 to destroy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{
				Add:      tt.fields.Add,
				Change:   tt.fields.Change,
				Destroy:  tt.fields.Destroy,
				fromPlan: tt.fields.fromPlan,
			}

			p := newPlatformForTest(tt.platformFields)

			plan, err := p.Plan(tt.destroy)
			if err != nil {
				t.Errorf("Failed to create the platform")
				return
			}

			if got := s.FromPlan(plan); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Stats.FromPlan() = %+v, want %+v", got, tt.want)
			}

			if gotStr := s.String(); gotStr != tt.wantStr {
				t.Errorf("Stats.String() = %v, want %v", gotStr, tt.wantStr)
			}

			if err := p.Apply(tt.destroy); err != nil {
				t.Errorf("Failed to apply the platform")
				return
			}

			if got := p.ExpectedStats; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Platform.ExpectedStats = %+v, want %+v", got, tt.want)
			}

		})
	}
}

func TestPlatform_Stats(t *testing.T) {
	tests := []struct {
		name           string
		platformFields platformFields
		destroy        bool
		want           *Stats
		wantStr        string
	}{
		{"null data source", testsPlatformsFields["null data source"], false, &Stats{0, 0, 0, false}, "resources: 0 added, 0 changed, 0 destroyed"},
		{"test instance", testsPlatformsFields["test instance"], false, &Stats{1, 0, 0, false}, "resources: 1 added, 0 changed, 0 destroyed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPlatformForTest(tt.platformFields)
			if err := p.Apply(tt.destroy); err != nil {
				t.Errorf("Failed to apply the changes. Error: %s", err)
				return
			}
			got := p.Stats()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Platform.Stats() = %v, want %v", got, tt.want)
			}

			if gotStr := got.String(); gotStr != tt.wantStr {
				t.Errorf("Stats.String() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestStats_FromCountHook(t *testing.T) {
	type statsFields struct {
		Add      int
		Change   int
		Destroy  int
		fromPlan bool
	}
	tests := []struct {
		name      string
		fields    statsFields
		countHook *local.CountHook
		want      *Stats
		wantStr   string
	}{
		{"simple", statsFields{0, 0, 0, false}, &local.CountHook{
			Added:          0,
			Changed:        0,
			Removed:        0,
			ToAdd:          0,
			ToChange:       0,
			ToRemove:       0,
			ToRemoveAndAdd: 0,
		}, &Stats{0, 0, 0, false}, "resources: 0 added, 0 changed, 0 destroyed"},
		{"added-n-removed", statsFields{0, 0, 0, false}, &local.CountHook{
			Added:          5,
			Changed:        0,
			Removed:        4,
			ToAdd:          0,
			ToChange:       0,
			ToRemove:       0,
			ToRemoveAndAdd: 0,
		}, &Stats{5, 0, 4, false}, "resources: 5 added, 0 changed, 4 destroyed"},
		{"to-add-n-remove", statsFields{0, 0, 0, false}, &local.CountHook{
			Added:          0,
			Changed:        0,
			Removed:        0,
			ToAdd:          5,
			ToChange:       0,
			ToRemove:       4,
			ToRemoveAndAdd: 0,
		}, &Stats{0, 0, 0, false}, "resources: 0 added, 0 changed, 0 destroyed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{
				Add:      tt.fields.Add,
				Change:   tt.fields.Change,
				Destroy:  tt.fields.Destroy,
				fromPlan: tt.fields.fromPlan,
			}
			got := s.FromCountHook(tt.countHook)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Stats.FromCountHook() = %v, want %v", got, tt.want)
			}
			if gotStr := got.String(); gotStr != tt.wantStr {
				t.Errorf("Stats.FromCountHook().String() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
