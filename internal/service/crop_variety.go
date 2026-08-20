package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// CropVarietyService manages the crop-variety dictionary.
type CropVarietyService interface {
	Create(ctx context.Context, u *domain.CropVarietyUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.CropVarietyUpsert) error
	Get(ctx context.Context, id int64) (*domain.CropVariety, error)
	List(ctx context.Context, p domain.Pagination, category string) (domain.PageResult[*domain.CropVariety], error)
	Delete(ctx context.Context, id int64) error
}

type cropVarietyService struct {
	store *repository.Store
	audit AuditService
}

// NewCropVarietyService returns the default CropVarietyService.
func NewCropVarietyService(store *repository.Store, a AuditService) CropVarietyService {
	return &cropVarietyService{store: store, audit: a}
}

func (s *cropVarietyService) Create(ctx context.Context, u *domain.CropVarietyUpsert) (int64, error) {
	if err := validateVariety(u); err != nil {
		return 0, err
	}
	v := &domain.CropVariety{Code: u.Code, Name: u.Name, Category: u.Category, GrowthCycle: u.GrowthCycle,
		Description: u.Description, Status: defaultIfZero(u.Status, domain.StatusActive)}
	id, err := s.store.VarietyRepo.Create(ctx, s.store.DB(), v)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "crop_variety", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *cropVarietyService) Update(ctx context.Context, id int64, u *domain.CropVarietyUpsert) error {
	if err := validateVariety(u); err != nil {
		return err
	}
	if _, err := s.store.VarietyRepo.GetByID(ctx, s.store.DB(), id); err != nil {
		return err
	}
	if err := s.store.VarietyRepo.Update(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "crop_variety", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *cropVarietyService) Get(ctx context.Context, id int64) (*domain.CropVariety, error) {
	return s.store.VarietyRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *cropVarietyService) List(ctx context.Context, p domain.Pagination, category string) (domain.PageResult[*domain.CropVariety], error) {
	p.Normalize()
	vs, total, err := s.store.VarietyRepo.List(ctx, s.store.DB(), p, category)
	if err != nil {
		return domain.PageResult[*domain.CropVariety]{}, err
	}
	return domain.NewPageResult(vs, total, p), nil
}

func (s *cropVarietyService) Delete(ctx context.Context, id int64) error {
	if err := s.store.VarietyRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "crop_variety", fmt.Sprintf("%d", id), nil)
	return nil
}

func validateVariety(u *domain.CropVarietyUpsert) error {
	if u.Code == "" || u.Name == "" {
		return domain.Wrap(domain.CodeValidation, 400, "品种编码和名称必填", nil)
	}
	if u.GrowthCycle < 0 {
		return domain.Wrap(domain.CodeValidation, 400, "生长周期不能为负数", nil)
	}
	return nil
}
