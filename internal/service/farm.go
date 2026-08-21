package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// FarmService manages farm records.
type FarmService interface {
	Create(ctx context.Context, u *domain.FarmUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.FarmUpsert) error
	Get(ctx context.Context, id int64) (*domain.Farm, error)
	List(ctx context.Context, p domain.Pagination, search string) (domain.PageResult[*domain.Farm], error)
	Delete(ctx context.Context, id int64) error
}

type farmService struct {
	store *repository.Store
	audit AuditService
}

// NewFarmService returns the default FarmService.
func NewFarmService(store *repository.Store, a AuditService) FarmService {
	return &farmService{store: store, audit: a}
}

func (s *farmService) Create(ctx context.Context, u *domain.FarmUpsert) (int64, error) {
	if err := validateFarm(u); err != nil {
		return 0, err
	}
	f := &domain.Farm{Name: u.Name, Location: u.Location, TotalArea: domain.Decimal(u.TotalArea),
		Description: u.Description, Status: defaultIfZero(u.Status, domain.StatusActive)}
	id, err := s.store.FarmRepo.Create(ctx, s.store.DB(), f)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "farm", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *farmService) Update(ctx context.Context, id int64, u *domain.FarmUpsert) error {
	if err := validateFarm(u); err != nil {
		return err
	}
	if _, err := s.store.FarmRepo.GetByID(ctx, s.store.DB(), id); err != nil {
		return err
	}
	if err := s.store.FarmRepo.Update(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "farm", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *farmService) Get(ctx context.Context, id int64) (*domain.Farm, error) {
	return s.store.FarmRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *farmService) List(ctx context.Context, p domain.Pagination, search string) (domain.PageResult[*domain.Farm], error) {
	p.Normalize()
	farms, total, err := s.store.FarmRepo.List(ctx, s.store.DB(), p, search)
	if err != nil {
		return domain.PageResult[*domain.Farm]{}, err
	}
	return domain.NewPageResult(farms, total, p), nil
}

func (s *farmService) Delete(ctx context.Context, id int64) error {
	count, err := s.store.FarmRepo.CountFields(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.Wrap(domain.CodeConflict, 409, fmt.Sprintf("农场存在 %d 个地块，无法删除", count), nil)
	}
	if err := s.store.FarmRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "farm", fmt.Sprintf("%d", id), nil)
	return nil
}

func validateFarm(u *domain.FarmUpsert) error {
	if u.Name == "" {
		return domain.Wrap(domain.CodeValidation, 400, "农场名称必填", nil)
	}
	if u.TotalArea < 0 {
		return domain.Wrap(domain.CodeValidation, 400, "总面积不能为负数", nil)
	}
	return nil
}
