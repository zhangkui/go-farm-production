package service

import (
	"context"
	"fmt"
	"time"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// CostAnalysisService computes the per-plan cost breakdown from tasks,
// allocations and manual entries (design §4.7).
type CostAnalysisService interface {
	Compute(ctx context.Context, u *domain.CostAnalysisUpsert) (int64, error)
	Get(ctx context.Context, id int64) (*domain.CostAnalysis, error)
	GetByPlan(ctx context.Context, planID int64) (*domain.CostAnalysis, error)
	List(ctx context.Context, p domain.Pagination, planID *int64) (domain.PageResult[*domain.CostAnalysis], error)
	Delete(ctx context.Context, id int64) error
	Summarise(ctx context.Context, seasonID, varietyID *int64) ([]*domain.CostSummary, error)
}

type costAnalysisService struct {
	store *repository.Store
}

// NewCostAnalysisService returns the default CostAnalysisService.
func NewCostAnalysisService(store *repository.Store) CostAnalysisService {
	return &costAnalysisService{store: store}
}

// Compute aggregates a plan's labour, input and equipment costs, adds the
// caller-supplied other cost, derives total + unit cost, and upserts a row.
// All aggregation reads use a single read snapshot; the upsert runs in a tx.
func (s *costAnalysisService) Compute(ctx context.Context, u *domain.CostAnalysisUpsert) (int64, error) {
	if u.PlantingPlanID == 0 {
		return 0, domain.Wrap(domain.CodeValidation, 400, "种植计划必填", nil)
	}
	plan, err := s.store.PlanRepo.GetByID(ctx, s.store.DB(), u.PlantingPlanID)
	if err != nil {
		return 0, err
	}

	tasks, err := s.store.TaskRepo.ListByPlan(ctx, s.store.DB(), u.PlantingPlanID)
	if err != nil {
		return 0, err
	}
	allocs, err := s.store.AllocRepo.ListByPlan(ctx, s.store.DB(), u.PlantingPlanID)
	if err != nil {
		return 0, err
	}
	harvests, err := s.store.HarvestRepo.ListByPlan(ctx, s.store.DB(), u.PlantingPlanID)
	if err != nil {
		return 0, err
	}

	// Labour cost = sum(task labour_hours) * labourRate.
	var labourHours, equipment domain.Decimal
	for _, t := range tasks {
		labourHours = labourHours.Add(t.LabourHours)
		equipment = equipment.Add(t.EquipmentCost)
	}
	labour := labourHours.Mul(domain.Decimal(labourRate))

	// Input cost = sum(allocation net quantity * material unit price). Returns
	// add back, allocate/waste deduct. We load each material once.
	materialPrice := map[int64]domain.Decimal{}
	var inputCost domain.Decimal
	for _, a := range allocs {
		price, ok := materialPrice[a.MaterialID]
		if !ok {
			mat, err := s.store.MaterialRepo.GetByID(ctx, s.store.DB(), a.MaterialID)
			if err != nil {
				return 0, err
			}
			price = mat.UnitPrice
			materialPrice[a.MaterialID] = price
		}
		line := a.Quantity.Mul(price)
		switch a.Type {
		case domain.AllocationTypeReturn:
			inputCost = inputCost.Sub(line)
		default: // allocate + waste both consume material
			inputCost = inputCost.Add(line)
		}
	}
	if inputCost < 0 {
		inputCost = 0 // returns exceeding allocations shouldn't yield negative cost
	}

	// Total yield = sum(harvest total_weight).
	var yield domain.Decimal
	for _, h := range harvests {
		yield = yield.Add(h.TotalWeight)
	}

	other := domain.Decimal(u.OtherCost)
	total := labour.Add(inputCost).Add(equipment).Add(other)
	unit := domain.Decimal(0)
	if !yield.IsZero() {
		unit = total.Div(yield)
	}

	analysisDate := time.Now()
	if u.AnalysisDate != "" {
		if t, err := mustTimeValue(u.AnalysisDate, "分析日期"); err == nil && !t.IsZero() {
			analysisDate = t
		}
	}
	c := &domain.CostAnalysis{
		PlantingPlanID: u.PlantingPlanID, AnalysisDate: analysisDate,
		LabourCost: labour, InputCost: inputCost, EquipmentCost: equipment, OtherCost: other,
		TotalCost: total, TotalYield: yield, UnitCost: unit, Remark: u.Remark,
	}

	var id int64
	err = s.store.WithTx(ctx, func(ctx context.Context, tx domain.DBTX) error {
		// Replace any existing analysis for this plan (one canonical row).
		if _, err := tx.ExecContext(ctx, `DELETE FROM cost_analyses WHERE planting_plan_id=?`, u.PlantingPlanID); err != nil {
			return err
		}
		rid, err := s.store.CostRepo.Create(ctx, tx, c)
		if err != nil {
			return err
		}
		id = rid
		return nil
	})
	if err != nil {
		return 0, err
	}
	_ = plan
	return id, nil
}

func (s *costAnalysisService) Get(ctx context.Context, id int64) (*domain.CostAnalysis, error) {
	return s.store.CostRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *costAnalysisService) GetByPlan(ctx context.Context, planID int64) (*domain.CostAnalysis, error) {
	return s.store.CostRepo.GetByPlan(ctx, s.store.DB(), planID)
}

func (s *costAnalysisService) List(ctx context.Context, p domain.Pagination, planID *int64) (domain.PageResult[*domain.CostAnalysis], error) {
	p.Normalize()
	cs, total, err := s.store.CostRepo.List(ctx, s.store.DB(), p, planID)
	if err != nil {
		return domain.PageResult[*domain.CostAnalysis]{}, err
	}
	return domain.NewPageResult(cs, total, p), nil
}

func (s *costAnalysisService) Delete(ctx context.Context, id int64) error {
	return s.store.CostRepo.Delete(ctx, s.store.DB(), id)
}

func (s *costAnalysisService) Summarise(ctx context.Context, seasonID, varietyID *int64) ([]*domain.CostSummary, error) {
	return s.store.CostRepo.Summarise(ctx, s.store.DB(), seasonID, varietyID)
}

// String formats a Decimal as yuan currency for reports.
func (s *costAnalysisService) String(d domain.Decimal) string {
	return fmt.Sprintf("%.2f", d.Float64())
}
