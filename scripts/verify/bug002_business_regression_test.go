package verify

import (
	"context"
	"testing"
	"time"

	"go-farm-production/internal/domain"
)

func TestBug002_BusinessRegression(t *testing.T) {
	t.Run("single criterion remains an exact filter", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-002-single", false)
		insertBug002Audit(t, env, 41, "alice", "delete", "field", "7", "2026-08-20 09:00:00")
		insertBug002Audit(t, env, 42, "bob", "update", "field", "7", "2026-08-20 09:30:00")
		result, err := env.Services.Audit.List(context.Background(), domain.Pagination{Page: 1, PageSize: 10}, domain.AuditLogQuery{Action: "delete"})
		if err != nil {
			t.Fatalf("single action filter: %v", err)
		}
		if result.Total != 1 || len(result.Items) != 1 || result.Items[0].Username != "alice" {
			t.Fatalf("single filter total=%d items=%d first=%v", result.Total, len(result.Items), auditUser(result.Items))
		}
	})

	t.Run("combined criteria intersect before pagination", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-002-combined", false)
		from := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
		to := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
		targetID := insertBug002Audit(t, env, 41, "alice", "update", "planting_plan", "900", "2026-08-20 11:00:00")
		insertBug002Audit(t, env, 42, "bob", "update", "planting_plan", "900", "2026-08-20 11:05:00")
		insertBug002Audit(t, env, 41, "alice", "delete", "planting_plan", "900", "2026-08-20 11:10:00")
		insertBug002Audit(t, env, 41, "alice", "update", "planting_plan", "901", "2026-08-20 11:15:00")
		insertBug002Audit(t, env, 41, "alice", "update", "planting_plan", "900", "2026-08-20 13:00:00")
		insertBug002Audit(t, env, 77, "other", "create", "material", "3", "2026-08-19 08:00:00")
		userID := int64(41)
		query := domain.AuditLogQuery{UserID: &userID, Action: "update", ResourceType: "planting_plan", ResourceID: "900", FromTime: &from, ToTime: &to}
		result, err := env.Services.Audit.List(context.Background(), domain.Pagination{Page: 1, PageSize: 1}, query)
		if err != nil {
			t.Fatalf("combined audit query: %v", err)
		}
		if result.Total != 1 || result.TotalPage != 1 {
			t.Fatalf("combined total=%d pages=%d, want 1/1", result.Total, result.TotalPage)
		}
		if len(result.Items) != 1 || result.Items[0].ID != targetID {
			t.Fatalf("combined item IDs=%v, want [%d]", auditIDs(result.Items), targetID)
		}
	})

	t.Run("time boundaries are inclusive and isolated", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-002-boundary", false)
		boundary := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
		targetID := insertBug002Audit(t, env, 51, "boundary", "status_change", "user", "51", "2026-08-20 10:00:00")
		insertBug002Audit(t, env, 52, "outside", "status_change", "user", "52", "2026-08-20 10:00:01")
		userID := int64(51)
		result, err := env.Services.Audit.List(context.Background(), domain.Pagination{Page: 1, PageSize: 10}, domain.AuditLogQuery{UserID: &userID, ResourceType: "user", ResourceID: "51", FromTime: &boundary, ToTime: &boundary})
		if err != nil {
			t.Fatalf("boundary query: %v", err)
		}
		if result.Total != 1 || len(result.Items) != 1 || result.Items[0].ID != targetID {
			t.Fatalf("boundary total=%d IDs=%v, want target %d", result.Total, auditIDs(result.Items), targetID)
		}
	})

	t.Run("invalid scopes fail before reading rows", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-002-invalid", false)
		from := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
		to := from.Add(-time.Hour)
		invalidUser := int64(0)
		for name, query := range map[string]domain.AuditLogQuery{
			"nonpositive user":      {UserID: &invalidUser},
			"resource without type": {ResourceID: "99"},
			"reversed time":         {FromTime: &from, ToTime: &to},
		} {
			t.Run(name, func(t *testing.T) {
				if _, err := env.Services.Audit.List(context.Background(), domain.Pagination{Page: 1, PageSize: 10}, query); err == nil || domain.AsAppError(err).Code != domain.CodeValidation {
					t.Fatalf("validation error=%v code=%v", err, domain.AsAppError(err))
				}
			})
		}
	})
}

func insertBug002Audit(t *testing.T, env *verifyEnv, userID int64, username, action, resourceType, resourceID, createdAt string) int64 {
	t.Helper()
	result := execVerifySQL(t, env.DB, `INSERT INTO audit_logs (user_id,username,action,resource_type,resource_id,details,ip_address,user_agent,created_at) VALUES (?,?,?,?,?,'{}','127.0.0.1','bug002',?)`, userID, username, action, resourceType, resourceID, createdAt)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("audit last insert id: %v", err)
	}
	return id
}

func auditIDs(rows []*domain.AuditLog) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

func auditUser(rows []*domain.AuditLog) string {
	if len(rows) == 0 {
		return ""
	}
	return rows[0].Username
}
