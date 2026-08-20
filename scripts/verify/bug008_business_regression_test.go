package verify

import (
	"context"
	"go-farm-production/internal/domain"
	"testing"
)

func TestBug008_BusinessRegression(t *testing.T) {
	env := newVerifyEnv(t, "BUG-008-season-lifecycle", false)
	p, _, _ := latePlanFixture(t, env, domain.PlanStatusGrowing)
	var s int64
	if e := env.DB.QueryRow(`SELECT season_id FROM planting_plans WHERE id=?`, p).Scan(&s); e != nil {
		t.Fatal(e)
	}
	if e := env.Services.Season.Update(context.Background(), s, &domain.SeasonUpsert{Code: "season", Name: "season", StartDate: "2026-01-01", EndDate: "2026-12-31", Status: domain.StatusInactive}); e == nil {
		t.Fatal("season used by active plan was disabled")
	}
}
