package service

import (
	"context"
	"fmt"
	"time"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// FarmTaskService manages farm tasks and their status transitions.
type FarmTaskService interface {
	Create(ctx context.Context, u *domain.FarmTaskUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.FarmTaskUpsert) error
	Transition(ctx context.Context, id int64, to int8) error
	Get(ctx context.Context, id int64) (*domain.FarmTask, []*domain.InputAllocation, error)
	List(ctx context.Context, p domain.Pagination, filter repository.TaskFilter) (domain.PageResult[*domain.FarmTask], error)
	ListByPlan(ctx context.Context, planID int64) ([]*domain.FarmTask, error)
	Delete(ctx context.Context, id int64) error
}

type farmTaskService struct {
	store *repository.Store
	audit AuditService
}

// NewFarmTaskService returns the default FarmTaskService.
func NewFarmTaskService(store *repository.Store, a AuditService) FarmTaskService {
	return &farmTaskService{store: store, audit: a}
}

func (s *farmTaskService) Create(ctx context.Context, u *domain.FarmTaskUpsert) (int64, error) {
	if err := validateTask(u); err != nil {
		return 0, err
	}
	t, err := buildTask(u)
	if err != nil {
		return 0, err
	}
	if _, err := s.store.PlanRepo.GetByID(ctx, s.store.DB(), u.PlantingPlanID); err != nil {
		return 0, err
	}
	id, err := s.store.TaskRepo.Create(ctx, s.store.DB(), t)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "farm_task", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *farmTaskService) Update(ctx context.Context, id int64, u *domain.FarmTaskUpsert) error {
	existing, err := s.store.TaskRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if existing.Status == domain.TaskStatusCompleted {
		return domain.Wrap(domain.CodeConflict, 409, "已完成任务不可修改", nil)
	}
	if err := validateTask(u); err != nil {
		return err
	}
	t, err := buildTask(u)
	if err != nil {
		return err
	}
	t.ID = id
	if err := s.store.TaskRepo.Update(ctx, s.store.DB(), id, t); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "farm_task", fmt.Sprintf("%d", id), u)
	return nil
}

// Transition moves a task along scheduled -> in_progress -> completed, stamping
// the completed_date when finishing.
func (s *farmTaskService) Transition(ctx context.Context, id int64, to int8) error {
	existing, err := s.store.TaskRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return err
	}
	if !allowedTaskTransition(existing.Status, to) {
		return domain.Wrap(domain.CodeStateTransition, 409, "任务状态流转不合法", nil)
	}
	var completed *time.Time
	if to == domain.TaskStatusCompleted {
		now := time.Now()
		completed = &now
	}
	if err := s.store.TaskRepo.UpdateStatus(ctx, s.store.DB(), id, to, completed); err != nil {
		return err
	}
	audit(ctx, s.audit, "status_change", "farm_task", fmt.Sprintf("%d", id),
		map[string]any{"from": existing.Status, "to": to})
	return nil
}

func (s *farmTaskService) Get(ctx context.Context, id int64) (*domain.FarmTask, []*domain.InputAllocation, error) {
	t, err := s.store.TaskRepo.GetByID(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	allocs, err := s.store.AllocRepo.ListByTask(ctx, s.store.DB(), id)
	if err != nil {
		return nil, nil, err
	}
	return t, allocs, nil
}

func (s *farmTaskService) List(ctx context.Context, p domain.Pagination, filter repository.TaskFilter) (domain.PageResult[*domain.FarmTask], error) {
	p.Normalize()
	tasks, total, err := s.store.TaskRepo.List(ctx, s.store.DB(), p, filter)
	if err != nil {
		return domain.PageResult[*domain.FarmTask]{}, err
	}
	return domain.NewPageResult(tasks, total, p), nil
}

func (s *farmTaskService) ListByPlan(ctx context.Context, planID int64) ([]*domain.FarmTask, error) {
	return s.store.TaskRepo.ListByPlan(ctx, s.store.DB(), planID)
}

func (s *farmTaskService) Delete(ctx context.Context, id int64) error {
	if err := s.store.TaskRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "farm_task", fmt.Sprintf("%d", id), nil)
	return nil
}

func validateTask(u *domain.FarmTaskUpsert) error {
	if u.PlantingPlanID == 0 {
		return domain.Wrap(domain.CodeValidation, 400, "种植计划必填", nil)
	}
	if u.Title == "" {
		return domain.Wrap(domain.CodeValidation, 400, "任务标题必填", nil)
	}
	if u.LabourHours < 0 || u.EquipmentCost < 0 {
		return domain.Wrap(domain.CodeValidation, 400, "工时与设备成本不能为负数", nil)
	}
	return nil
}

func buildTask(u *domain.FarmTaskUpsert) (*domain.FarmTask, error) {
	planned, err := mustTimeValue(u.PlannedDate, "计划日期")
	if err != nil {
		return nil, err
	}
	if planned.IsZero() {
		return nil, domain.Wrap(domain.CodeValidation, 400, "计划日期必填", nil)
	}
	completed, err := mustTime(u.CompletedDate, "完成日期")
	if err != nil {
		return nil, err
	}
	status := defaultIfZero(u.Status, domain.TaskStatusScheduled)
	return &domain.FarmTask{
		PlantingPlanID: u.PlantingPlanID, TaskType: defaultIfZero(u.TaskType, domain.TaskTypeOther),
		Title: u.Title, Description: u.Description, PlannedDate: planned, CompletedDate: completed,
		Status: status, LabourHours: domain.Decimal(u.LabourHours), EquipmentCost: domain.Decimal(u.EquipmentCost),
		AssigneeID: u.AssigneeID, Remark: u.Remark,
	}, nil
}

// allowedTaskTransition enforces scheduled -> in_progress -> completed and
// cancellation from any non-terminal state.
func allowedTaskTransition(from, to int8) bool {
	if to == domain.TaskStatusCancelled {
		return from != domain.TaskStatusCancelled && from != domain.TaskStatusCompleted
	}
	switch from {
	case domain.TaskStatusScheduled:
		return to == domain.TaskStatusInProgress
	case domain.TaskStatusInProgress:
		return to == domain.TaskStatusCompleted
	}
	return false
}
