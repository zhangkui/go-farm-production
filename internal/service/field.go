package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// FieldService manages field (plot) records.
type FieldService interface {
	Create(ctx context.Context, u *domain.FieldUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.FieldUpsert) error
	Get(ctx context.Context, id int64) (*domain.Field, error)
	List(ctx context.Context, p domain.Pagination, farmID *int64, irrigationZone string) (domain.PageResult[*domain.Field], error)
	Delete(ctx context.Context, id int64) error
}

type fieldService struct {
	store *repository.Store
	audit AuditService
}

// NewFieldService returns the default FieldService.
func NewFieldService(store *repository.Store, a AuditService) FieldService {
	return &fieldService{store: store, audit: a}
}

func (s *fieldService) Create(ctx context.Context, u *domain.FieldUpsert) (int64, error) {
	if err := validateField(u); err != nil {
		return 0, err
	}
	if _, err := s.store.FarmRepo.GetByID(ctx, s.store.DB(), u.FarmID); err != nil {
		return 0, err
	}
	f := &domain.Field{
		FarmID: u.FarmID, Code: u.Code, Name: u.Name, Area: domain.Decimal(u.Area),
		SoilType: u.SoilType, IrrigationZone: u.IrrigationZone,
		Status: defaultIfZero(u.Status, domain.FieldStatusAvailable), Remark: u.Remark,
	}
	id, err := s.store.FieldRepo.Create(ctx, s.store.DB(), f)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "field", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *fieldService) Update(ctx context.Context, id int64, u *domain.FieldUpsert) error {
	if err := validateField(u); err != nil {
		return err
	}
	existing, err := s.store.FieldRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	// A field currently planting cannot be deleted; editing its farm is allowed.
	if _, err := s.store.FarmRepo.GetByID(ctx, s.store.DB(), u.FarmID); err != nil {
		return err
	}
	_ = existing
	if err := s.store.FieldRepo.Update(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "field", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *fieldService) Get(ctx context.Context, id int64) (*domain.Field, error) {
	return s.store.FieldRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *fieldService) List(ctx context.Context, p domain.Pagination, farmID *int64, irrigationZone string) (domain.PageResult[*domain.Field], error) {
	p.Normalize()
	fields, total, err := s.store.FieldRepo.List(ctx, s.store.DB(), p, farmID, irrigationZone)
	if err != nil {
		return domain.PageResult[*domain.Field]{}, err
	}
	return domain.NewPageResult(fields, total, p), nil
}

func (s *fieldService) Delete(ctx context.Context, id int64) error {
	existing, err := s.store.FieldRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if existing.Status == domain.FieldStatusPlanting {
		return domain.Wrap(domain.CodeConflict, 409, "地块正在种植中，无法删除", nil)
	}
	if err := s.store.FieldRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "field", fmt.Sprintf("%d", id), nil)
	return nil
}

func validateField(u *domain.FieldUpsert) error {
	if u.FarmID == 0 {
		return domain.Wrap(domain.CodeValidation, 400, "农场必填", nil)
	}
	if u.Code == "" || u.Name == "" {
		return domain.Wrap(domain.CodeValidation, 400, "地块编码和名称必填", nil)
	}
	if u.Area <= 0 {
		return domain.Wrap(domain.CodeValidation, 400, "地块面积必须大于0", nil)
	}
	return nil
}
