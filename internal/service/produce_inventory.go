package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// ProduceInventoryService manages stock lots of harvested produce.
type ProduceInventoryService interface {
	Create(ctx context.Context, u *domain.ProduceInventoryUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.ProduceInventoryUpsert) error
	Get(ctx context.Context, id int64) (*domain.ProduceInventory, error)
	List(ctx context.Context, p domain.Pagination, status *int8, varietyID *int64) (domain.PageResult[*domain.ProduceInventory], error)
	Delete(ctx context.Context, id int64) error
}

type produceInventoryService struct {
	store *repository.Store
	audit AuditService
}

// NewProduceInventoryService returns the default ProduceInventoryService.
func NewProduceInventoryService(store *repository.Store, a AuditService) ProduceInventoryService {
	return &produceInventoryService{store: store, audit: a}
}

func (s *produceInventoryService) Create(ctx context.Context, u *domain.ProduceInventoryUpsert) (int64, error) {
	if err := validateProduce(u); err != nil {
		return 0, err
	}
	if _, err := s.store.VarietyRepo.GetByID(ctx, s.store.DB(), u.CropVarietyID); err != nil {
		return 0, err
	}
	harvest, err := s.store.HarvestRepo.GetByID(ctx, s.store.DB(), u.HarvestID)
	if err != nil {
		return 0, err
	}
	scope := domain.NewHarvestApprovalScope(harvest)
	approved, err := s.store.HarvestRepo.HasApprovedForPlan(ctx, s.store.DB(), scope.InventoryPlanID())
	if err != nil {
		return 0, err
	}
	if !approved {
		return 0, domain.Wrap(domain.CodeConflict, 409, "采收记录尚未审核，无法入库", nil)
	}
	p := &domain.ProduceInventory{
		CropVarietyID: u.CropVarietyID, HarvestID: u.HarvestID, Quantity: domain.Decimal(u.Quantity),
		Grade: u.Grade, Unit: u.Unit, StorageLocation: u.StorageLocation,
		Status: defaultIfZero(u.Status, domain.ProduceStatusInStock),
	}
	if p.Unit == "" {
		p.Unit = "kg"
	}
	id, err := s.store.ProduceRepo.Create(ctx, s.store.DB(), p)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "produce_inventory", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *produceInventoryService) Update(ctx context.Context, id int64, u *domain.ProduceInventoryUpsert) error {
	if err := validateProduce(u); err != nil {
		return err
	}
	if _, err := s.store.ProduceRepo.GetByID(ctx, s.store.DB(), id); err != nil {
		return err
	}
	if err := s.store.ProduceRepo.Update(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "produce_inventory", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *produceInventoryService) Get(ctx context.Context, id int64) (*domain.ProduceInventory, error) {
	return s.store.ProduceRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *produceInventoryService) List(ctx context.Context, p domain.Pagination, status *int8, varietyID *int64) (domain.PageResult[*domain.ProduceInventory], error) {
	p.Normalize()
	ps, total, err := s.store.ProduceRepo.List(ctx, s.store.DB(), p, status, varietyID)
	if err != nil {
		return domain.PageResult[*domain.ProduceInventory]{}, err
	}
	return domain.NewPageResult(ps, total, p), nil
}

func (s *produceInventoryService) Delete(ctx context.Context, id int64) error {
	if err := s.store.ProduceRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "produce_inventory", fmt.Sprintf("%d", id), nil)
	return nil
}

func validateProduce(u *domain.ProduceInventoryUpsert) error {
	if u.CropVarietyID == 0 || u.HarvestID == 0 {
		return domain.Wrap(domain.CodeValidation, 400, "作物品种与采收记录必填", nil)
	}
	if u.Quantity < 0 {
		return domain.Wrap(domain.CodeValidation, 400, "数量不能为负数", nil)
	}
	return nil
}
