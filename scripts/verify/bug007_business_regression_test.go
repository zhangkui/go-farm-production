package verify

import (
	"context"
	"go-farm-production/internal/domain"
	"testing"
)

func TestBug007_BusinessRegression(t *testing.T) {
	env := newVerifyEnv(t, "BUG-007-variety-lifecycle", false)
	_, _, v := latePlanFixture(t, env, domain.PlanStatusGrowing)
	if e := env.Services.Variety.Update(context.Background(), v, &domain.CropVarietyUpsert{Code: "v", Name: "v", Status: domain.StatusInactive}); e == nil {
		t.Fatal("variety used by active plan was disabled")
	}
}
