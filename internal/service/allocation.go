package service

import (
	"context"
	"fmt"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// InputAllocationService handles allocate/return/waste movements. The stock
// mutation and the allocation record are written inside a single transaction
// (design §4.5.3, §15.2) so remaining_qty can never diverge from the ledger.
type InputAllocationService interface {
	Allocate(ctx context.Context, u *domain.InputAllocationUpsert) (int64, error)
	Get(ctx context.Context, id int64) (*domain.InputAllocation, error)
	List(ctx context.Context, p domain.Pagination, filter repository.AllocFilter) (domain.PageResult[*domain.InputAllocation], error)
	ListByPlan(ctx context.Context, planID int64) ([]*domain.InputAllocation, error)
}

type inputAllocationService struct {
	store *repository.Store
	audit AuditService
}

// NewInputAllocationService returns the default InputAllocationService.
func NewInputAllocationService(store *repository.Store, a AuditService) InputAllocationService {
	return &inputAllocationService{store: store, audit: a}
}

// Allocate creates an allocation record and atomically adjusts the batch's
// remaining quantity:
//   - allocate (type 1): decrement remaining_qty, refuse if insufficient;
//   - waste    (type 2): decrement remaining_qty, refuse if insufficient;
//   - return   (type 0): increment remaining_qty.
//
// All within one transaction.
func (s *inputAllocationService) Allocate(ctx context.Context, u *domain.InputAllocationUpsert) (int64, error) {
	if err := validateAllocation(u); err != nil {
		return 0, err
	}
	qty := domain.Decimal(u.Quantity)
	m, _ := MetaFrom(ctx)

	var allocID int64
	err := s.store.WithTx(ctx, func(ctx context.Context, tx domain.DBTX) error {
		batch, err := s.store.BatchRepo.GetByIDForUpdate(ctx, tx, u.BatchID)
		if err != nil {
			return err
		}
		if batch.Status == domain.BatchStatusVoid {
			return domain.Wrap(domain.CodeConflict, 409, "批次已失效，无法操作", nil)
		}
		switch u.Type {
		case domain.AllocationTypeAllocate, domain.AllocationTypeWaste:
			if _, err := s.store.BatchRepo.DecreaseStock(ctx, tx, u.BatchID, qty); err != nil {
				return err
			}
		case domain.AllocationTypeReturn:
			returnable, err := s.store.AllocRepo.ReturnableQuantity(ctx, tx, u.BatchID, u.TaskID)
			if err != nil {
				return err
			}
			if qty > returnable {
				return domain.ErrReturnExceedsUsed
			}
			if _, err := s.store.BatchRepo.IncreaseStock(ctx, tx, u.BatchID, qty); err != nil {
				return err
			}
		default:
			return domain.Wrap(domain.CodeValidation, 400, "未知的领用类型", nil)
		}
		a := &domain.InputAllocation{
			BatchID: u.BatchID, MaterialID: batch.MaterialID, TaskID: u.TaskID,
			Quantity: qty, Type: u.Type, Remark: u.Remark, OperatorID: m.UserID,
		}
		id, err := s.store.AllocRepo.Create(ctx, tx, a)
		if err != nil {
			return err
		}
		allocID = id
		return nil
	})
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, allocAction(u.Type), "allocation", fmt.Sprintf("%d", allocID), u)
	return allocID, nil
}

func (s *inputAllocationService) Get(ctx context.Context, id int64) (*domain.InputAllocation, error) {
	return s.store.AllocRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *inputAllocationService) List(ctx context.Context, p domain.Pagination, filter repository.AllocFilter) (domain.PageResult[*domain.InputAllocation], error) {
	p.Normalize()
	as, total, err := s.store.AllocRepo.List(ctx, s.store.DB(), p, filter)
	if err != nil {
		return domain.PageResult[*domain.InputAllocation]{}, err
	}
	return domain.NewPageResult(as, total, p), nil
}

func (s *inputAllocationService) ListByPlan(ctx context.Context, planID int64) ([]*domain.InputAllocation, error) {
	return s.store.AllocRepo.ListByPlan(ctx, s.store.DB(), planID)
}

func validateAllocation(u *domain.InputAllocationUpsert) error {
	if u.BatchID == 0 {
		return domain.Wrap(domain.CodeValidation, 400, "批次必填", nil)
	}
	if u.Quantity <= 0 {
		return domain.Wrap(domain.CodeValidation, 400, "数量必须大于0", nil)
	}
	if u.Type != domain.AllocationTypeAllocate && u.Type != domain.AllocationTypeReturn && u.Type != domain.AllocationTypeWaste {
		return domain.Wrap(domain.CodeValidation, 400, "类型必须为领用/退回/损耗", nil)
	}
	return nil
}

func allocAction(t int8) string {
	switch t {
	case domain.AllocationTypeAllocate:
		return "allocate"
	case domain.AllocationTypeReturn:
		return "return"
	case domain.AllocationTypeWaste:
		return "waste"
	}
	return "allocate"
}
