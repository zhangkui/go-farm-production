package service

import (
	"context"
	"fmt"
	"time"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// InputBatchService manages input batches and expiry warnings.
type InputBatchService interface {
	Create(ctx context.Context, u *domain.InputBatchUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.InputBatchUpsert) error
	Get(ctx context.Context, id int64) (*domain.InputBatch, []*domain.InputAllocation, error)
	List(ctx context.Context, p domain.Pagination, materialID *int64, status *int8) (domain.PageResult[*domain.InputBatch], error)
	Delete(ctx context.Context, id int64) error
	ExpiryWarnings(ctx context.Context, withinDays int) ([]*domain.InputBatch, error)
}

type inputBatchService struct {
	store *repository.Store
	audit AuditService
}

// NewInputBatchService returns the default InputBatchService.
func NewInputBatchService(store *repository.Store, a AuditService) InputBatchService {
	return &inputBatchService{store: store, audit: a}
}

func (s *inputBatchService) Create(ctx context.Context, u *domain.InputBatchUpsert) (int64, error) {
	if err := validateBatch(u); err != nil {
		return 0, err
	}
	b, err := buildBatch(u)
	if err != nil {
		return 0, err
	}
	if _, err := s.store.MaterialRepo.GetByID(ctx, s.store.DB(), u.MaterialID); err != nil {
		return 0, err
	}
	// On create, remaining equals the received quantity.
	b.RemainingQty = (domain.BatchCreateState{Quantity: b.Quantity}).Remaining()
	id, err := s.store.BatchRepo.Create(ctx, s.store.DB(), b)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "batch", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *inputBatchService) Update(ctx context.Context, id int64, u *domain.InputBatchUpsert) error {
	if err := validateBatch(u); err != nil {
		return err
	}
	existing, err := s.store.BatchRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	b, err := buildBatch(u)
	if err != nil {
		return err
	}
	// Preserve remaining_qty (driven by allocations) unless quantity dropped below it.
	b.RemainingQty = (domain.BatchUpdateState{NewQuantity: b.Quantity, ExistingRemaining: existing.RemainingQty}).Remaining()
	b.Status = defaultIfZero(u.Status, existing.Status)
	if err := s.store.BatchRepo.UpdateWithRemaining(ctx, s.store.DB(), id, u, b.RemainingQty); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "batch", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *inputBatchService) Get(ctx context.Context, id int64) (*domain.InputBatch, []*domain.InputAllocation, error) {
	b, err := s.store.BatchRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	allocs, _, err := s.store.AllocRepo.List(ctx, s.store.DB(), domain.Pagination{Page: 1, PageSize: 100}, repository.AllocFilter{BatchID: &id})
	if err != nil {
		return nil, nil, err
	}
	return b, allocs, nil
}

func (s *inputBatchService) List(ctx context.Context, p domain.Pagination, materialID *int64, status *int8) (domain.PageResult[*domain.InputBatch], error) {
	p.Normalize()
	bs, total, err := s.store.BatchRepo.List(ctx, s.store.DB(), p, materialID, status)
	if err != nil {
		return domain.PageResult[*domain.InputBatch]{}, err
	}
	return domain.NewPageResult(bs, total, p), nil
}

func (s *inputBatchService) Delete(ctx context.Context, id int64) error {
	if err := s.store.BatchRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "batch", fmt.Sprintf("%d", id), nil)
	return nil
}

// ExpiryWarnings returns active batches expiring on or before withinDays from
// today (used by the inventory-warning page).
func (s *inputBatchService) ExpiryWarnings(ctx context.Context, withinDays int) ([]*domain.InputBatch, error) {
	if withinDays <= 0 {
		withinDays = 30
	}
	cutoff := time.Now().AddDate(0, 0, withinDays).Format("2006-01-02")
	candidates, err := s.store.BatchRepo.ListExpiring(ctx, s.store.DB(), cutoff)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.InputBatch, 0, len(candidates))
	for _, batch := range candidates {
		if (domain.ExpiryWarningCandidate{Status: batch.Status, Remaining: batch.RemainingQty}).Eligible() {
			out = append(out, batch)
		}
	}
	return out, nil
}

func validateBatch(u *domain.InputBatchUpsert) error {
	if u.MaterialID == 0 || u.BatchNo == "" {
		return domain.Wrap(domain.CodeValidation, 400, "物料和批号必填", nil)
	}
	if u.Quantity <= 0 {
		return domain.Wrap(domain.CodeValidation, 400, "入库数量必须大于0", nil)
	}
	if u.PurchasePrice < 0 {
		return domain.Wrap(domain.CodeValidation, 400, "采购单价不能为负数", nil)
	}
	return nil
}

func buildBatch(u *domain.InputBatchUpsert) (*domain.InputBatch, error) {
	purchase, err := mustTime(u.PurchaseDate, "采购日期")
	if err != nil {
		return nil, err
	}
	expiry, err := mustTime(u.ExpiryDate, "有效期")
	if err != nil {
		return nil, err
	}
	return &domain.InputBatch{
		MaterialID: u.MaterialID, BatchNo: u.BatchNo, Quantity: domain.Decimal(u.Quantity),
		PurchaseDate: purchase, ExpiryDate: expiry, PurchasePrice: domain.Decimal(u.PurchasePrice),
		Supplier: u.Supplier, Status: defaultIfZero(u.Status, domain.BatchStatusActive),
	}, nil
}
