package terranova

import (
	"fmt"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/backend/local"
	"github.com/hashicorp/terraform/plans"
)

// Stats encapsulate the statistics of changes to apply or applied
type Stats struct {
	Add, Change, Destroy int
	fromPlan             bool
}

// NewStats creates an empty stats
func NewStats() *Stats {
	return &Stats{}
}

// Reset resets, set everything to zeros, the current stats
func (s *Stats) Reset() {
	s.Add, s.Change, s.Destroy = 0, 0, 0
}

// FromPlan return stats from a given plan
func (s *Stats) FromPlan(plan *plans.Plan) *Stats {
	s.Reset()
	s.fromPlan = true

	if plan == nil || plan.Changes == nil || plan.Changes.Empty() {
		return s
	}

	for _, r := range plan.Changes.Resources {
		// Do not count data resources
		if r.Addr.Resource.Resource.Mode == addrs.DataResourceMode {
			continue
		}

		switch r.Action {
		case plans.Create:
			s.Add++
		case plans.Update:
			s.Change++
		case plans.DeleteThenCreate, plans.CreateThenDelete:
			s.Add++
			s.Destroy++
		case plans.Delete:
			s.Destroy++
		}
	}

	return s
}

// FromCountHook return stats from a given count hook
func (s *Stats) FromCountHook(countHook *local.CountHook) *Stats {
	if countHook == nil {
		return s
	}

	s.Add, s.Change, s.Destroy = countHook.Added, countHook.Changed, countHook.Removed
	s.fromPlan = false

	return s
}

func (s *Stats) String() string {
	if s.fromPlan {
		return fmt.Sprintf("resources: %d to add, %d to change, %d to destroy", s.Add, s.Change, s.Destroy)
	}
	return fmt.Sprintf("resources: %d added, %d changed, %d destroyed", s.Add, s.Change, s.Destroy)
}

// Stats return the current status from the count hook
func (p *Platform) Stats() *Stats {
	return NewStats().FromCountHook(p.countHook)
}
