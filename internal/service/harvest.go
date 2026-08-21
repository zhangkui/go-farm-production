package service

import (
	"context"
	"fmt"
	"time"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// HarvestService manages harvest events, their field-level details, and the
// post-harvest transition of the planting plan.
type HarvestService interface {
	Create(ctx context.Context, u *domain.HarvestUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.HarvestUpsert) error
	Approve(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*domain.Harvest, []*domain.HarvestDetail, error)
	List(ctx context.Context, p domain.Pagination, filter repository.HarvestFilter) (domain.PageResult[*domain.Harvest], error)
	Delete(ctx context.Context, id int64) error
}

type harvestService struct {
	store *repository.Store
	audit AuditService
}

// NewHarvestService returns the default HarvestService.
func NewHarvestService(store *repository.Store, a AuditService) HarvestService {
	return &harvestService{store: store, audit: a}
}

// Create records a harvest. The sum of detail weights must equal the total
// weight (design §4.6). On success, if the plan is still growing it is advanced
// to harvested, all within one transaction.
func (s *harvestService) Create(ctx context.Context, u *domain.HarvestUpsert) (int64, error) {
	if err := validateHarvest(u); err != nil {
		return 0, err
	}
	harvestDate, err := mustTimeValue(u.HarvestDate, "采收日期")
	if err != nil {
		return 0, err
	}
	if harvestDate.IsZero() {
		return 0, domain.Wrap(domain.CodeValidation, 400, "采收日期必填", nil)
	}
	plan, err := s.store.PlanRepo.GetByID(ctx, s.store.DB(), u.PlantingPlanID)
	if err != nil {
		return 0, err
	}

	h := &domain.Harvest{
		PlantingPlanID: u.PlantingPlanID, HarvestDate: harvestDate,
		TotalWeight: domain.Decimal(u.TotalWeight), Grade: u.Grade, Remark: u.Remark,
	}

	var harvestID int64
	err = s.store.WithTx(ctx, func(ctx context.Context, tx domain.DBTX) error {
		id, err := s.store.HarvestRepo.Create(ctx, tx, h)
		if err != nil {
			return err
		}
		harvestID = id
		for _, d := range u.Details {
			det := &domain.HarvestDetail{
				HarvestID: id, FieldID: d.FieldID, Weight: domain.Decimal(d.Weight),
				Grade: d.Grade, Remark: d.Remark,
			}
			if _, err := s.store.HDetailRepo.Create(ctx, tx, det); err != nil {
				return err
			}
		}
		// Advance plan to harvested if it is in a pre-harvest active state.
		if domain.IsPlanActive(plan.Status) && plan.Status != domain.PlanStatusHarvested {
			now := time.Now()
			if err := s.store.PlanRepo.UpdateStatus(ctx, tx, plan.ID, domain.PlanStatusHarvested); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE planting_plans SET actual_harvest_date=? WHERE id=?`, now, plan.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "harvest", fmt.Sprintf("%d", harvestID), u)
	return harvestID, nil
}

func (s *harvestService) Update(ctx context.Context, id int64, u *domain.HarvestUpsert) error {
	if err := validateHarvest(u); err != nil {
		return err
	}
	existing, err := s.store.HarvestRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if existing.Approved {
		return domain.Wrap(domain.CodeConflict, 409, "已审核的采收记录不可修改", nil)
	}
	harvestDate, err := mustTimeValue(u.HarvestDate, "采收日期")
	if err != nil {
		return err
	}
	h := &domain.Harvest{
		PlantingPlanID: u.PlantingPlanID, HarvestDate: harvestDate,
		TotalWeight: domain.Decimal(u.TotalWeight), Grade: u.Grade, Remark: u.Remark, Approved: existing.Approved,
	}
	err = s.store.WithTx(ctx, func(ctx context.Context, tx domain.DBTX) error {
		if err := s.store.HarvestRepo.Update(ctx, tx, id, h); err != nil {
			return err
		}
		if err := s.store.HDetailRepo.DeleteByHarvest(ctx, tx, id); err != nil {
			return err
		}
		for _, d := range u.Details {
			det := &domain.HarvestDetail{
				HarvestID: id, FieldID: d.FieldID, Weight: domain.Decimal(d.Weight),
				Grade: d.Grade, Remark: d.Remark,
			}
			if _, err := s.store.HDetailRepo.Create(ctx, tx, det); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "harvest", fmt.Sprintf("%d", id), u)
	return nil
}

// Approve marks a harvest as reviewed/approved. Approval binds to the
// individual harvest record, never to its planting plan: approving one harvest
// must not flip the approved flag on any sibling harvest sharing the plan.
func (s *harvestService) Approve(ctx context.Context, id int64) error {
	if _, err := s.store.HarvestRepo.GetByID(ctx, s.store.DB(), id); err != nil {
		return err
	}
	if err := s.store.HarvestRepo.SetApproved(ctx, s.store.DB(), id, true); err != nil {
		return err
	}
	audit(ctx, s.audit, "approve", "harvest", fmt.Sprintf("%d", id), nil)
	return nil
}

func (s *harvestService) Get(ctx context.Context, id int64) (*domain.Harvest, []*domain.HarvestDetail, error) {
	h, err := s.store.HarvestRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	details, err := s.store.HDetailRepo.ListByHarvest(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	return h, details, nil
}

func (s *harvestService) List(ctx context.Context, p domain.Pagination, filter repository.HarvestFilter) (domain.PageResult[*domain.Harvest], error) {
	p.Normalize()
	hs, total, err := s.store.HarvestRepo.List(ctx, s.store.DB(), p, filter)
	if err != nil {
		return domain.PageResult[*domain.Harvest]{}, err
	}
	return domain.NewPageResult(hs, total, p), nil
}

func (s *harvestService) Delete(ctx context.Context, id int64) error {
	existing, err := s.store.HarvestRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if existing.Approved {
		return domain.Wrap(domain.CodeConflict, 409, "已审核的采收记录不可删除", nil)
	}
	if err := s.store.HarvestRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "harvest", fmt.Sprintf("%d", id), nil)
	return nil
}

func validateHarvest(u *domain.HarvestUpsert) error {
	if u.PlantingPlanID == 0 {
		return domain.Wrap(domain.CodeValidation, 400, "种植计划必填", nil)
	}
	if u.TotalWeight < 0 {
		return domain.Wrap(domain.CodeValidation, 400, "总重量不能为负数", nil)
	}
	// Detail weights must sum to total weight when details are provided.
	if len(u.Details) > 0 {
		var sum float64
		for _, d := range u.Details {
			if d.FieldID == 0 {
				return domain.Wrap(domain.CodeValidation, 400, "采收明细地块必填", nil)
			}
			if d.Weight < 0 {
				return domain.Wrap(domain.CodeValidation, 400, "明细重量不能为负数", nil)
			}
			sum += d.Weight
		}
		// Allow a small float epsilon for rounding.
		if absFloat(sum-u.TotalWeight) > 0.01 {
			return domain.ErrWeightMismatch
		}
	}
	return nil
}

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
