package verify

import (
	"context"
	"go-farm-production/internal/domain"
	"testing"
)

func TestBug005_BusinessRegression(t *testing.T) {
	env := newVerifyEnv(t, "BUG-005-farm-area-capacity", false)
	f, e := env.Services.Farm.Create(context.Background(), &domain.FarmUpsert{Name: "bug005-farm", TotalArea: 100, Status: domain.StatusActive})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = env.Services.Field.Create(context.Background(), &domain.FieldUpsert{FarmID: f, Code: "bug005-field", Name: "field", Area: 80, Status: domain.FieldStatusAvailable}); e != nil {
		t.Fatal(e)
	}
	if e = env.Services.Farm.Update(context.Background(), f, &domain.FarmUpsert{Name: "bug005-farm", TotalArea: 50, Status: domain.StatusActive}); e == nil {
		t.Fatal("farm area below field aggregate accepted")
	}
}
