package service

import (
	"testing"

	"go-farm-production/internal/domain"
)

// TestAllowedPlanTransition exercises the planting-plan state machine.
func TestAllowedPlanTransition(t *testing.T) {
	cases := []struct {
		name string
		from int8
		to   int8
		want bool
	}{
		{"planned->planted", domain.PlanStatusPlanned, domain.PlanStatusPlanted, true},
		{"planted->growing", domain.PlanStatusPlanted, domain.PlanStatusGrowing, true},
		{"growing->harvested", domain.PlanStatusGrowing, domain.PlanStatusHarvested, true},
		{"harvested->completed", domain.PlanStatusHarvested, domain.PlanStatusCompleted, true},
		{"planned->completed (skip)", domain.PlanStatusPlanned, domain.PlanStatusCompleted, false},
		{"completed->planted (back)", domain.PlanStatusCompleted, domain.PlanStatusPlanted, false},
		{"planned->cancelled", domain.PlanStatusPlanned, domain.PlanStatusCancelled, true},
		{"growing->cancelled", domain.PlanStatusGrowing, domain.PlanStatusCancelled, true},
		{"cancelled->planned (from terminal)", domain.PlanStatusCancelled, domain.PlanStatusPlanned, false},
		{"completed->cancelled (from terminal)", domain.PlanStatusCompleted, domain.PlanStatusCancelled, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := domain.AllowedPlanTransition(c.from, c.to)
			if got != c.want {
				t.Fatalf("AllowedPlanTransition(%d,%d)=%v want %v", c.from, c.to, got, c.want)
			}
		})
	}
}

// TestIsPlanActive checks the overlap-relevant active-set predicate.
func TestIsPlanActive(t *testing.T) {
	active := []int8{domain.PlanStatusPlanned, domain.PlanStatusPlanted, domain.PlanStatusGrowing, domain.PlanStatusHarvested}
	for _, s := range active {
		if !domain.IsPlanActive(s) {
			t.Errorf("status %d should be active", s)
		}
	}
	inactive := []int8{domain.PlanStatusCancelled, domain.PlanStatusCompleted}
	for _, s := range inactive {
		if domain.IsPlanActive(s) {
			t.Errorf("status %d should NOT be active", s)
		}
	}
}

// TestDecimalArithmetic covers the Decimal helpers used by cost calc.
func TestDecimalArithmetic(t *testing.T) {
	a := domain.Decimal(10)
	b := domain.Decimal(4)
	if got := a.Add(b); got != 14 {
		t.Fatalf("Add=%v want 14", got)
	}
	if got := a.Sub(b); got != 6 {
		t.Fatalf("Sub=%v want 6", got)
	}
	if got := a.Mul(b); got != 40 {
		t.Fatalf("Mul=%v want 40", got)
	}
	if got := a.Div(b); got != 2.5 {
		t.Fatalf("Div=%v want 2.5", got)
	}
	if got := a.Div(domain.Decimal(0)); !got.IsZero() {
		t.Fatalf("Div by zero should be 0, got %v", got)
	}
}

// TestAllowedTaskTransition covers the farm-task state machine.
func TestAllowedTaskTransition(t *testing.T) {
	cases := []struct {
		from, to int8
		want     bool
	}{
		{domain.TaskStatusScheduled, domain.TaskStatusInProgress, true},
		{domain.TaskStatusInProgress, domain.TaskStatusCompleted, true},
		{domain.TaskStatusScheduled, domain.TaskStatusCompleted, false},
		{domain.TaskStatusCompleted, domain.TaskStatusInProgress, false},
		{domain.TaskStatusInProgress, domain.TaskStatusCancelled, true},
		{domain.TaskStatusCancelled, domain.TaskStatusCompleted, false},
	}
	for _, c := range cases {
		if got := allowedTaskTransition(c.from, c.to); got != c.want {
			t.Errorf("task %d->%d=%v want %v", c.from, c.to, got, c.want)
		}
	}
}

// TestDefaultIfZero covers the generic helper.
func TestDefaultIfZero(t *testing.T) {
	if got := defaultIfZero(0, 7); got != 7 {
		t.Fatalf("defaultIfZero(0,7)=%v want 7", got)
	}
	if got := defaultIfZero(5, 7); got != 5 {
		t.Fatalf("defaultIfZero(5,7)=%v want 5", got)
	}
	if got := defaultIfZero("", "x"); got != "x" {
		t.Fatalf(`defaultIfZero("","x")=%q want "x"`, got)
	}
}
