package verify

import (
	"context"
	"testing"
	"time"

	"go-farm-production/internal/domain"
)

func TestBug015_BusinessRegression(t *testing.T) {
	t.Run("natural dates persist exactly", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-015-natural", false)
		id := createBug015Season(t, env, "natural", "2026-03-01", "2026-06-30")
		assertBug015Dates(t, env, id, "2026-03-01", "2026-06-30")
	})
	t.Run("RFC3339 inputs preserve their calendar dates", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-015-rfc", false)
		id := createBug015Season(t, env, "rfc", "2026-03-01T08:00:00+08:00", "2026-06-30T08:00:00+08:00")
		assertBug015Dates(t, env, id, "2026-03-01", "2026-06-30")
	})
	t.Run("equal and reverse ranges are rejected without rows", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-015-invalid", false)
		for _, tc := range []struct{ code, start, end string }{{"equal", "2026-04-01", "2026-04-01"}, {"reverse", "2026-04-02", "2026-04-01"}} {
			_, err := env.Services.Season.Create(context.Background(), &domain.SeasonUpsert{Code: tc.code, Name: tc.code, StartDate: tc.start, EndDate: tc.end, Status: domain.StatusActive})
			if err == nil {
				t.Errorf("%s range accepted", tc.code)
			}
			if got := queryVerifyInt(t, env.DB, `SELECT COUNT(*) FROM seasons WHERE code=?`, tc.code); got != 0 {
				t.Fatalf("invalid season %s persisted", tc.code)
			}
		}
	})
	t.Run("empty dates and duplicate codes are rejected", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-015-boundary", false)
		_, err := env.Services.Season.Create(context.Background(), &domain.SeasonUpsert{Code: "empty", Name: "empty", StartDate: "", EndDate: "2026-05-01"})
		if err == nil {
			t.Fatalf("empty start accepted")
		}
		_ = createBug015Season(t, env, "duplicate", "2026-01-01", "2026-02-01")
		_, err = env.Services.Season.Create(context.Background(), &domain.SeasonUpsert{Code: "duplicate", Name: "dup", StartDate: "2026-03-01", EndDate: "2026-04-01"})
		if err == nil || domain.AsAppError(err).Code != domain.CodeDuplicate {
			t.Fatalf("duplicate error=%v", err)
		}
	})
}
func createBug015Season(t *testing.T, env *verifyEnv, code, start, end string) int64 {
	t.Helper()
	id, err := env.Services.Season.Create(context.Background(), &domain.SeasonUpsert{Code: code, Name: code, StartDate: start, EndDate: end, Status: domain.StatusActive})
	if err != nil {
		t.Fatalf("create season %s: %v", code, err)
	}
	return id
}
func assertBug015Dates(t *testing.T, env *verifyEnv, id int64, start, end string) {
	t.Helper()
	var gotStart, gotEnd time.Time
	if err := env.DB.QueryRow(`SELECT start_date,end_date FROM seasons WHERE id=?`, id).Scan(&gotStart, &gotEnd); err != nil {
		t.Fatalf("read season dates: %v", err)
	}
	if gotStart.Format("2006-01-02") != start || gotEnd.Format("2006-01-02") != end {
		t.Fatalf("dates=%s..%s, want %s..%s", gotStart.Format("2006-01-02"), gotEnd.Format("2006-01-02"), start, end)
	}
}
