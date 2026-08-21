package service

import (
	"context"
	"fmt"
	"time"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// PlantingPlanService manages planting plans and enforces the no-overlap
// invariant and legal state transitions.
type PlantingPlanService interface {
	Create(ctx context.Context, u *domain.PlantingPlanUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.PlantingPlanUpsert) error
	Transition(ctx context.Context, id int64, to int8) error
	Get(ctx context.Context, id int64) (*domain.PlantingPlan, []*domain.FarmTask, []*domain.Harvest, error)
	List(ctx context.Context, p domain.Pagination, filter repository.PlanFilter) (domain.PageResult[*domain.PlantingPlan], error)
	Delete(ctx context.Context, id int64) error
}

type plantingPlanService struct {
	store *repository.Store
	audit AuditService
}

// NewPlantingPlanService returns the default PlantingPlanService.
func NewPlantingPlanService(store *repository.Store, a AuditService) PlantingPlanService {
	return &plantingPlanService{store: store, audit: a}
}

func (s *plantingPlanService) Create(ctx context.Context, u *domain.PlantingPlanUpsert) (int64, error) {
	pl, err := s.buildPlan(ctx, u)
	if err != nil {
		return 0, err
	}
	pl.Status = domain.PlanStatusPlanned

	// Overlap check: same field, overlapping [planned_sow, planned_harvest].
	harvestEnd := pl.PlannedSowDate
	if pl.PlannedHarvestDate != nil {
		harvestEnd = *pl.PlannedHarvestDate
	}
	if conflict, err := s.store.PlanRepo.FindOverlap(ctx, s.store.DB(), pl.FieldID, 0, pl.PlannedSowDate, harvestEnd); err != nil {
		return 0, err
	} else if conflict != nil {
		return 0, domain.ErrPlanOverlap
	}

	id, err := s.store.PlanRepo.Create(ctx, s.store.DB(), pl)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "planting_plan", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *plantingPlanService) Update(ctx context.Context, id int64, u *domain.PlantingPlanUpsert) error {
	existing, err := s.store.PlanRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	pl, err := s.buildPlan(ctx, u)
	if err != nil {
		return err
	}
	if domain.PlanExecutionIdentityChanged(existing, u) {
		u.FieldID = existing.FieldID
		u.CropVarietyID = existing.CropVarietyID
		u.SeasonID = existing.SeasonID
		pl.FieldID = existing.FieldID
		pl.CropVarietyID = existing.CropVarietyID
		pl.SeasonID = existing.SeasonID
	}
	if err := repository.ValidatePlanExecutionUpdate(ctx, s.store.DB(), id, u); err != nil {
		return err
	}
	pl.ID = id
	pl.Status = existing.Status

	harvestEnd := pl.PlannedSowDate
	if pl.PlannedHarvestDate != nil {
		harvestEnd = *pl.PlannedHarvestDate
	}
	if conflict, err := s.store.PlanRepo.FindOverlap(ctx, s.store.DB(), pl.FieldID, id, pl.PlannedSowDate, harvestEnd); err != nil {
		return err
	} else if conflict != nil {
		return domain.ErrPlanOverlap
	}
	if err := s.store.PlanRepo.Update(ctx, s.store.DB(), id, pl); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "planting_plan", fmt.Sprintf("%d", id), u)
	return nil
}

// Transition advances a plan's status along the legal state machine. When
// moving to planted, the actual_sow_date is stamped; when moving to harvested,
// the actual_harvest_date is stamped.
func (s *plantingPlanService) Transition(ctx context.Context, id int64, to int8) error {
	existing, err := s.store.PlanRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if !domain.AllowedPlanTransition(existing.Status, to) {
		return domain.Wrap(domain.CodeStateTransition, 409,
			fmt.Sprintf("状态 %s 无法流转至 %s", domain.PlanStatusName(existing.Status), domain.PlanStatusName(to)), nil)
	}
	err = s.store.WithTx(ctx, func(ctx context.Context, tx domain.DBTX) error {
		if err := s.store.PlanRepo.UpdateStatus(ctx, tx, id, to); err != nil {
			return err
		}
		// Stamp actual dates on the meaningful transitions.
		now := time.Now()
		switch to {
		case domain.PlanStatusPlanted:
			if _, err := tx.ExecContext(ctx, `UPDATE planting_plans SET actual_sow_date=? WHERE id=?`, now, id); err != nil {
				return err
			}
		case domain.PlanStatusHarvested:
			if _, err := tx.ExecContext(ctx, `UPDATE planting_plans SET actual_harvest_date=? WHERE id=?`, now, id); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	audit(ctx, s.audit, "status_change", "planting_plan", fmt.Sprintf("%d", id),
		map[string]any{"from": existing.Status, "to": to})
	return nil
}

func (s *plantingPlanService) Get(ctx context.Context, id int64) (*domain.PlantingPlan, []*domain.FarmTask, []*domain.Harvest, error) {
	pl, err := s.store.PlanRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, nil, err
	}
	tasks, err := s.store.TaskRepo.ListByPlan(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, nil, err
	}
	harvests, err := s.store.HarvestRepo.ListByPlan(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, nil, err
	}
	return pl, tasks, harvests, nil
}

func (s *plantingPlanService) List(ctx context.Context, p domain.Pagination, filter repository.PlanFilter) (domain.PageResult[*domain.PlantingPlan], error) {
	p.Normalize()
	plans, total, err := s.store.PlanRepo.List(ctx, s.store.DB(), p, filter)
	if err != nil {
		return domain.PageResult[*domain.PlantingPlan]{}, err
	}
	return domain.NewPageResult(plans, total, p), nil
}

func (s *plantingPlanService) Delete(ctx context.Context, id int64) error {
	existing, err := s.store.PlanRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if domain.IsPlanActive(existing.Status) {
		return domain.Wrap(domain.CodeConflict, 409, "活跃种植计划不可删除，请先取消", nil)
	}
	if err := s.store.PlanRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "planting_plan", fmt.Sprintf("%d", id), nil)
	return nil
}

// buildPlan validates and assembles a PlantingPlan domain object from input.
// It also enforces the planned-area <= field-area constraint.
func (s *plantingPlanService) buildPlan(ctx context.Context, u *domain.PlantingPlanUpsert) (*domain.PlantingPlan, error) {
	if u.FieldID == 0 || u.CropVarietyID == 0 || u.SeasonID == 0 {
		return nil, domain.Wrap(domain.CodeValidation, 400, "地块、作物品种、种植季必填", nil)
	}
	if u.PlannedArea <= 0 {
		return nil, domain.Wrap(domain.CodeValidation, 400, "计划面积必须大于0", nil)
	}
	field, err := s.store.FieldRepo.GetByID(ctx, s.store.DB(), u.FieldID)
	if err != nil {
		return nil, err
	}
	if domain.Decimal(u.PlannedArea) > field.Area {
		return nil, domain.ErrAreaExceeds
	}
	if _, err := s.store.VarietyRepo.GetByID(ctx, s.store.DB(), u.CropVarietyID); err != nil {
		return nil, err
	}
	if _, err := s.store.SeasonRepo.GetByID(ctx, s.store.DB(), u.SeasonID); err != nil {
		return nil, err
	}
	sow, err := mustTimeValue(u.PlannedSowDate, "计划播种日")
	if err != nil {
		return nil, err
	}
	if sow.IsZero() {
		return nil, domain.Wrap(domain.CodeValidation, 400, "计划播种日必填", nil)
	}
	harvest, err := mustTime(u.PlannedHarvestDate, "计划采收日")
	if err != nil {
		return nil, err
	}
	actualSow, err := mustTime(u.ActualSowDate, "实际播种日")
	if err != nil {
		return nil, err
	}
	actualHarvest, err := mustTime(u.ActualHarvestDate, "实际采收日")
	if err != nil {
		return nil, err
	}
	return &domain.PlantingPlan{
		FieldID: u.FieldID, CropVarietyID: u.CropVarietyID, SeasonID: u.SeasonID,
		PlannedArea: domain.Decimal(u.PlannedArea), PlannedSowDate: sow, PlannedHarvestDate: harvest,
		ActualSowDate: actualSow, ActualHarvestDate: actualHarvest, Remark: u.Remark,
	}, nil
}
