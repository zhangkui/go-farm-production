package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// MaterialService manages the input-material dictionary.
type MaterialService interface {
	Create(ctx context.Context, u *domain.MaterialUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.MaterialUpsert) error
	Get(ctx context.Context, id int64) (*domain.Material, error)
	List(ctx context.Context, p domain.Pagination, category *int8) (domain.PageResult[*domain.Material], error)
	Delete(ctx context.Context, id int64) error
}

type materialService struct {
	store *repository.Store
	audit AuditService
}

// NewMaterialService returns the default MaterialService.
func NewMaterialService(store *repository.Store, a AuditService) MaterialService {
	return &materialService{store: store, audit: a}
}

func (s *materialService) Create(ctx context.Context, u *domain.MaterialUpsert) (int64, error) {
	if err := validateMaterial(u); err != nil {
		return 0, err
	}
	m := &domain.Material{Code: u.Code, Name: u.Name, Category: defaultIfZero(u.Category, domain.MaterialCatOther),
		Unit: u.Unit, UnitPrice: domain.Decimal(u.UnitPrice), Description: u.Description,
		Status: defaultIfZero(u.Status, domain.StatusActive)}
	if m.Unit == "" {
		m.Unit = "kg"
	}
	id, err := s.store.MaterialRepo.Create(ctx, s.store.DB(), m)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "material", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *materialService) Update(ctx context.Context, id int64, u *domain.MaterialUpsert) error {
	if err := validateMaterial(u); err != nil {
		return err
	}
	if _, err := s.store.MaterialRepo.GetByID(ctx, s.store.DB(), id); err != nil {
		return err
	}
	if err := s.store.MaterialRepo.Update(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "material", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *materialService) Get(ctx context.Context, id int64) (*domain.Material, error) {
	return s.store.MaterialRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *materialService) List(ctx context.Context, p domain.Pagination, category *int8) (domain.PageResult[*domain.Material], error) {
	p.Normalize()
	ms, total, err := s.store.MaterialRepo.List(ctx, s.store.DB(), p, category)
	if err != nil {
		return domain.PageResult[*domain.Material]{}, err
	}
	return domain.NewPageResult(ms, total, p), nil
}

func (s *materialService) Delete(ctx context.Context, id int64) error {
	if err := s.store.MaterialRepo.DeleteReportingReference(ctx, s.store.DB(), id).BusinessError(); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "material", fmt.Sprintf("%d", id), nil)
	return nil
}

func validateMaterial(u *domain.MaterialUpsert) error {
	if u.Code == "" || u.Name == "" {
		return domain.Wrap(domain.CodeValidation, 400, "物料编码和名称必填", nil)
	}
	if u.UnitPrice < 0 {
		return domain.Wrap(domain.CodeValidation, 400, "单价不能为负数", nil)
	}
	return nil
}
