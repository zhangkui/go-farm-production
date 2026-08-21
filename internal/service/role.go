package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// RoleService manages roles and their permission grants.
type RoleService interface {
	Create(ctx context.Context, u *domain.RoleUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.RoleUpsert) error
	Get(ctx context.Context, id int64) (*domain.Role, []*domain.Permission, error)
	List(ctx context.Context, p domain.Pagination, search string) (domain.PageResult[*domain.Role], error)
	Delete(ctx context.Context, id int64) error
	AssignPermissions(ctx context.Context, roleID int64, permIDs []int64) error
	ListPermissions(ctx context.Context) ([]*domain.Permission, error)
}

type roleService struct {
	store *repository.Store
	audit AuditService
}

// NewRoleService returns the default RoleService.
func NewRoleService(store *repository.Store, a AuditService) RoleService {
	return &roleService{store: store, audit: a}
}

func (s *roleService) Create(ctx context.Context, u *domain.RoleUpsert) (int64, error) {
	if u.Code == "" || u.Name == "" {
		return 0, domain.Wrap(domain.CodeValidation, 400, "角色编码和名称必填", nil)
	}
	id, err := s.store.RoleRepo.Create(ctx, s.store.DB(), &domain.Role{Code: u.Code, Name: u.Name, Description: u.Description})
	if err != nil {
		return 0, err
	}
	if len(u.Permissions) > 0 {
		if err := s.store.RoleRepo.AssignPermissions(ctx, s.store.DB(), id, u.Permissions); err != nil {
			return 0, err
		}
	}
	s.invalidateRoleCache(ctx, id)
	audit(ctx, s.audit, "create", "role", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *roleService) Update(ctx context.Context, id int64, u *domain.RoleUpsert) error {
	if u.Code == "" || u.Name == "" {
		return domain.Wrap(domain.CodeValidation, 400, "角色编码和名称必填", nil)
	}
	if _, err := s.store.RoleRepo.GetByID(ctx, s.store.DB(), id); err != nil {
		return err
	}
	if err := s.store.RoleRepo.Update(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	if u.Permissions != nil {
		if err := s.store.RoleRepo.AssignPermissions(ctx, s.store.DB(), id, u.Permissions); err != nil {
			return err
		}
	}
	s.invalidateRoleCache(ctx, id)
	audit(ctx, s.audit, "update", "role", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *roleService) Get(ctx context.Context, id int64) (*domain.Role, []*domain.Permission, error) {
	role, err := s.store.RoleRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	perms, err := s.store.RoleRepo.ListPermissions(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	return role, perms, nil
}

func (s *roleService) List(ctx context.Context, p domain.Pagination, search string) (domain.PageResult[*domain.Role], error) {
	p.Normalize()
	roles, total, err := s.store.RoleRepo.List(ctx, s.store.DB(), p, search)
	if err != nil {
		return domain.PageResult[*domain.Role]{}, err
	}
	return domain.NewPageResult(roles, total, p), nil
}

func (s *roleService) Delete(ctx context.Context, id int64) error {
	if err := s.store.RoleRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	s.invalidateRoleCache(ctx, id)
	audit(ctx, s.audit, "delete", "role", fmt.Sprintf("%d", id), nil)
	return nil
}

func (s *roleService) AssignPermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	if _, err := s.store.RoleRepo.GetByID(ctx, s.store.DB(), roleID); err != nil {
		return err
	}
	if err := s.store.RoleRepo.AssignPermissions(ctx, s.store.DB(), roleID, permIDs); err != nil {
		return err
	}
	s.invalidateRoleCache(ctx, roleID)
	audit(ctx, s.audit, "assign_permissions", "role", fmt.Sprintf("%d", roleID), map[string]any{"permission_ids": permIDs})
	return nil
}

func (s *roleService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	p := domain.Pagination{Page: 1, PageSize: domain.MaxPageSize}
	perms, _, err := s.store.PermRepo.List(ctx, s.store.DB(), p)
	return perms, err
}

// invalidateRoleCache clears cached permission sets for every user who holds
// roleID. We scan user_roles; on large deployments this could be replaced with
// a role-versioned cache key.
func (s *roleService) invalidateRoleCache(ctx context.Context, roleID int64) {
	if s.store.RDB() == nil {
		return
	}
	rows, err := s.store.DB().QueryContext(ctx, `SELECT user_id FROM user_roles WHERE role_id=?`, roleID)
	if err != nil {
		return
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err == nil {
			keys = append(keys, repository.PermKey(uid))
		}
	}
	if len(keys) > 0 {
		_ = s.store.RDB().Del(ctx, keys...).Err()
	}
}
