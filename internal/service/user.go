package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// UserService manages user accounts and their role assignments.
type UserService interface {
	Create(ctx context.Context, u *domain.UserUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.UserUpsert) error
	UpdatePassword(ctx context.Context, id int64, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, id int64, newPassword string) error
	UpdateStatus(ctx context.Context, id int64, status int8) error
	Get(ctx context.Context, id int64) (*domain.User, []*domain.Role, error)
	List(ctx context.Context, p domain.Pagination, search string) (domain.PageResult[*domain.User], error)
	Delete(ctx context.Context, id int64) error
	AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error
}

type userService struct {
	store  *repository.Store
	audit  AuditService
	hasher Hasher
}

// NewUserService returns the default UserService.
func NewUserService(store *repository.Store, a AuditService, h Hasher) UserService {
	return &userService{store: store, audit: a, hasher: h}
}

func (s *userService) Create(ctx context.Context, u *domain.UserUpsert) (int64, error) {
	if err := validateUser(u); err != nil {
		return 0, err
	}
	hash, err := s.hasher.Hash(u.Password)
	if err != nil {
		return 0, err
	}
	if u.Status == 0 {
		u.Status = domain.StatusActive
	}
	user := &domain.User{
		Username: u.Username, Email: u.Email, FullName: u.FullName,
		PasswordHash: hash, Status: u.Status,
	}
	id, err := s.store.UserRepo.Create(ctx, s.store.DB(), user)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "user", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *userService) Update(ctx context.Context, id int64, u *domain.UserUpsert) error {
	if err := validateUserUpdate(u); err != nil {
		return err
	}
	if err := s.store.UserRepo.Update(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "user", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *userService) UpdatePassword(ctx context.Context, id int64, oldPassword, newPassword string) error {
	user, err := s.store.UserRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	// Self-change requires the old password to match; admins use ResetPassword.
	m, _ := MetaFrom(ctx)
	if m.UserID == id {
		if err := s.hasher.Compare(user.PasswordHash, oldPassword); err != nil {
			return domain.ErrInvalidCredentials
		}
	}
	if err := s.store.UserRepo.UpdatePassword(ctx, s.store.DB(), id, hash); err != nil {
		return err
	}
	// A password change invalidates every prior session across all devices.
	_ = s.store.TokenRepo.RevokeAllForUser(ctx, s.store.DB(), id)
	audit(ctx, s.audit, "password_change", "user", fmt.Sprintf("%d", id), nil)
	return nil
}

func (s *userService) ResetPassword(ctx context.Context, id int64, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := s.store.UserRepo.UpdatePassword(ctx, s.store.DB(), id, hash); err != nil {
		return err
	}
	// A password reset invalidates every prior session across all devices.
	_ = s.store.TokenRepo.RevokeAllForUser(ctx, s.store.DB(), id)
	audit(ctx, s.audit, "password_reset", "user", fmt.Sprintf("%d", id), nil)
	return nil
}

func (s *userService) UpdateStatus(ctx context.Context, id int64, status int8) error {
	if status != domain.StatusActive && status != domain.StatusInactive {
		return domain.Wrap(domain.CodeValidation, 400, "状态值不合法", nil)
	}
	if err := s.store.UserRepo.UpdateStatus(ctx, s.store.DB(), id, status); err != nil {
		return err
	}
	if status == domain.StatusInactive {
		// Disabling an account invalidates every active session across all devices.
		_ = s.store.TokenRepo.RevokeAllForUser(ctx, s.store.DB(), id)
	}
	audit(ctx, s.audit, "status_change", "user", fmt.Sprintf("%d", id), map[string]any{"status": status})
	return nil
}

func (s *userService) Get(ctx context.Context, id int64) (*domain.User, []*domain.Role, error) {
	user, err := s.store.UserRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	user.PasswordHash = ""
	roles, err := s.store.UserRepo.ListRoles(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	return user, roles, nil
}

func (s *userService) List(ctx context.Context, p domain.Pagination, search string) (domain.PageResult[*domain.User], error) {
	p.Normalize()
	users, total, err := s.store.UserRepo.List(ctx, s.store.DB(), p, search)
	if err != nil {
		return domain.PageResult[*domain.User]{}, err
	}
	return domain.NewPageResult(users, total, p), nil
}

func (s *userService) Delete(ctx context.Context, id int64) error {
	if err := s.store.UserRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "user", fmt.Sprintf("%d", id), nil)
	return nil
}

func (s *userService) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	err := s.store.WithTx(ctx, func(ctx context.Context, tx domain.DBTX) error {
		if err := s.store.UserRepo.AssignRoles(ctx, tx, userID, roleIDs); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Permission cache is now stale; drop it.
	if s.store.RDB() != nil {
		_ = s.store.RDB().Del(ctx, repository.PermKey(userID)).Err()
	}
	audit(ctx, s.audit, "assign_roles", "user", fmt.Sprintf("%d", userID), map[string]any{"role_ids": roleIDs})
	return nil
}

// validateUser enforces create-time user input.
func validateUser(u *domain.UserUpsert) error {
	if u.Username == "" || len(u.Username) > 64 {
		return domain.Wrap(domain.CodeValidation, 400, "用户名必填且不超过64字符", nil)
	}
	if u.Email == "" || !contains(u.Email, "@") {
		return domain.Wrap(domain.CodeValidation, 400, "邮箱格式不正确", nil)
	}
	if err := validatePassword(u.Password); err != nil {
		return err
	}
	if u.FullName == "" {
		u.FullName = u.Username
	}
	return nil
}

// validateUserUpdate enforces update-time user input (password optional).
func validateUserUpdate(u *domain.UserUpsert) error {
	if u.Username == "" || len(u.Username) > 64 {
		return domain.Wrap(domain.CodeValidation, 400, "用户名必填且不超过64字符", nil)
	}
	if u.Email == "" || !contains(u.Email, "@") {
		return domain.Wrap(domain.CodeValidation, 400, "邮箱格式不正确", nil)
	}
	if u.FullName == "" {
		u.FullName = u.Username
	}
	return nil
}

// validatePassword enforces a minimum password policy.
func validatePassword(p string) error {
	if len(p) < 8 || len(p) > 128 {
		return domain.Wrap(domain.CodeValidation, 400, "密码长度需在8-128之间", nil)
	}
	return nil
}

// contains is a tiny strings.Contains shim to avoid importing strings here.
func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
