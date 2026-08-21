package verify

import (
	"context"
	"testing"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/service"
)

func TestBug032_BusinessRegression(t *testing.T) {
	t.Run("disable revokes every device and old sessions never revive", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-032-disable")
		ctx := context.Background()
		userID := bug032User(t, env, "disable-user", "Before123!")
		first := bug032Login(t, env, "disable-user", "Before123!")
		second := bug032Login(t, env, "disable-user", "Before123!")

		if err := env.Services.User.UpdateStatus(ctx, userID, domain.StatusInactive); err != nil {
			t.Fatalf("disable user: %v", err)
		}
		assertBug032RevokedCount(t, env, userID, 2, "disable")
		beforeTokens := bug032TokenCount(t, env, userID)
		beforeAudits := bug032RefreshAuditCount(t, env, userID)
		illegal, err := env.Services.Auth.Refresh(ctx, first.RefreshToken)
		if err == nil {
			t.Errorf("disabled account refreshed an old device token: %+v", illegal)
		}
		if got := bug032TokenCount(t, env, userID); got != beforeTokens {
			t.Errorf("disabled refresh created tokens=%d, want %d", got, beforeTokens)
		}
		if got := bug032RefreshAuditCount(t, env, userID); got != beforeAudits {
			t.Errorf("disabled refresh wrote success audits=%d, want %d", got, beforeAudits)
		}

		if err := env.Services.User.UpdateStatus(ctx, userID, domain.StatusActive); err != nil {
			t.Fatalf("re-enable user: %v", err)
		}
		for name, token := range map[string]string{"first": first.RefreshToken, "second": second.RefreshToken} {
			if pair, err := env.Services.Auth.Refresh(ctx, token); err == nil {
				t.Errorf("%s pre-disable token revived after re-enable: %+v", name, pair)
			}
		}
		if illegal.RefreshToken != "" {
			if pair, err := env.Services.Auth.Refresh(ctx, illegal.RefreshToken); err == nil {
				t.Errorf("token issued while disabled survived re-enable: %+v", pair)
			}
		}
	})

	t.Run("password reset revokes every old device", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-032-reset")
		ctx := context.Background()
		userID := bug032User(t, env, "reset-user", "Before123!")
		first := bug032Login(t, env, "reset-user", "Before123!")
		second := bug032Login(t, env, "reset-user", "Before123!")

		if err := env.Services.User.ResetPassword(ctx, userID, "After456!"); err != nil {
			t.Fatalf("reset password: %v", err)
		}
		assertBug032RevokedCount(t, env, userID, 2, "password reset")
		beforeTokens := bug032TokenCount(t, env, userID)
		beforeAudits := bug032RefreshAuditCount(t, env, userID)
		for name, token := range map[string]string{"first": first.RefreshToken, "second": second.RefreshToken} {
			if pair, err := env.Services.Auth.Refresh(ctx, token); err == nil {
				t.Errorf("%s pre-reset token refreshed successfully: %+v", name, pair)
			}
		}
		if got := bug032TokenCount(t, env, userID); got != beforeTokens {
			t.Errorf("password-reset refresh attempts created tokens=%d, want %d", got, beforeTokens)
		}
		if got := bug032RefreshAuditCount(t, env, userID); got != beforeAudits {
			t.Errorf("password-reset refresh attempts wrote success audits=%d, want %d", got, beforeAudits)
		}

		fresh := bug032Login(t, env, "reset-user", "After456!")
		rotated, err := env.Services.Auth.Refresh(ctx, fresh.RefreshToken)
		if err != nil {
			t.Fatalf("new-password session failed normal refresh rotation: %v", err)
		}
		if rotated.RefreshToken == "" || rotated.RefreshToken == fresh.RefreshToken {
			t.Errorf("normal refresh did not rotate token: old=%q new=%q", fresh.RefreshToken, rotated.RefreshToken)
		}
	})

	t.Run("self password change revokes every old device", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-032-self-change")
		userID := bug032User(t, env, "self-user", "Before123!")
		first := bug032Login(t, env, "self-user", "Before123!")
		second := bug032Login(t, env, "self-user", "Before123!")
		ctx := service.WithMeta(context.Background(), service.RequestMeta{UserID: userID, Username: "self-user"})

		if err := env.Services.User.UpdatePassword(ctx, userID, "Before123!", "After456!"); err != nil {
			t.Fatalf("self password change: %v", err)
		}
		assertBug032RevokedCount(t, env, userID, 2, "self password change")
		beforeTokens := bug032TokenCount(t, env, userID)
		beforeAudits := bug032RefreshAuditCount(t, env, userID)
		for name, token := range map[string]string{"first": first.RefreshToken, "second": second.RefreshToken} {
			if pair, err := env.Services.Auth.Refresh(context.Background(), token); err == nil {
				t.Errorf("%s pre-change token refreshed successfully: %+v", name, pair)
			}
		}
		if got := bug032TokenCount(t, env, userID); got != beforeTokens {
			t.Errorf("self-change refresh attempts created tokens=%d, want %d", got, beforeTokens)
		}
		if got := bug032RefreshAuditCount(t, env, userID); got != beforeAudits {
			t.Errorf("self-change refresh attempts wrote success audits=%d, want %d", got, beforeAudits)
		}
	})
}

func bug032User(t *testing.T, env *verifyEnv, username, password string) int64 {
	t.Helper()
	id, err := env.Services.User.Create(context.Background(), &domain.UserUpsert{
		Username: username,
		Email:    username + "@example.com",
		FullName: username,
		Password: password,
		Status:   domain.StatusActive,
	})
	if err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return id
}

func bug032Login(t *testing.T, env *verifyEnv, username, password string) domain.TokenPair {
	t.Helper()
	pair, err := env.Services.Auth.Login(context.Background(), username, password)
	if err != nil {
		t.Fatalf("login %s: %v", username, err)
	}
	if pair.RefreshToken == "" {
		t.Fatalf("login %s returned empty refresh token", username)
	}
	return pair
}

func assertBug032RevokedCount(t *testing.T, env *verifyEnv, userID int64, want int, stage string) {
	t.Helper()
	if got := queryVerifyInt(t, env.DB, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id=? AND revoked=1`, userID); got != want {
		t.Errorf("%s revoked refresh tokens=%d, want %d", stage, got, want)
	}
}

func bug032TokenCount(t *testing.T, env *verifyEnv, userID int64) int {
	t.Helper()
	return queryVerifyInt(t, env.DB, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id=?`, userID)
}

func bug032RefreshAuditCount(t *testing.T, env *verifyEnv, userID int64) int {
	t.Helper()
	return queryVerifyInt(t, env.DB,
		`SELECT COUNT(*) FROM audit_logs WHERE action='refresh' AND resource_type='user' AND resource_id=?`, userID)
}
