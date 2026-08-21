package service

import (
	"context"
	"fmt"
	"time"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// SeasonService manages planting seasons.
type SeasonService interface {
	Create(ctx context.Context, u *domain.SeasonUpsert) (int64, error)
	Update(ctx context.Context, id int64, u *domain.SeasonUpsert) error
	Get(ctx context.Context, id int64) (*domain.Season, error)
	List(ctx context.Context, p domain.Pagination, status *int8) (domain.PageResult[*domain.Season], error)
	Delete(ctx context.Context, id int64) error
}

type seasonService struct {
	store *repository.Store
	audit AuditService
}

// NewSeasonService returns the default SeasonService.
func NewSeasonService(store *repository.Store, a AuditService) SeasonService {
	return &seasonService{store: store, audit: a}
}

func (s *seasonService) Create(ctx context.Context, u *domain.SeasonUpsert) (int64, error) {
	start, err := mustSeasonCreateTimeValue(u.StartDate, "开始日期", true)
	if err != nil {
		return 0, err
	}
	end, err := mustSeasonCreateTimeValue(u.EndDate, "结束日期", false)
	if err != nil {
		return 0, err
	}
	window := domain.SeasonCreateWindow{Start: start, End: end}
	if u.Code == "" || u.Name == "" || u.StartDate == "" || u.EndDate == "" || !window.Valid() {
		return 0, domain.Wrap(domain.CodeValidation, 400, "季次日期或基础信息不合法", nil)
	}
	se := &domain.Season{Code: u.Code, Name: u.Name, StartDate: start, EndDate: end,
		Status: defaultIfZero(u.Status, domain.StatusActive)}
	id, err := s.store.SeasonRepo.CreateWindow(ctx, s.store.DB(), se, window)
	if err != nil {
		return 0, err
	}
	audit(ctx, s.audit, "create", "season", fmt.Sprintf("%d", id), u)
	return id, nil
}

func (s *seasonService) Update(ctx context.Context, id int64, u *domain.SeasonUpsert) error {
	start, end, err := parseSeasonRange(u)
	if err != nil {
		return err
	}
	if _, err := s.store.SeasonRepo.GetByID(ctx, s.store.DB(), id); err != nil {
		return err
	}
	// BUG-008: a season may only be deactivated when no plan in an effective
	// production phase still references it. Refuse the request and keep the
	// original status so an active plan never ends up under an unavailable
	// season. Seasons not used by any active plan can be deactivated normally.
	if u.Status == domain.StatusInactive {
		active, err := s.store.PlanRepo.CountActiveBySeason(ctx, s.store.DB(), id)
		if err != nil {
			return err
		}
		if active > 0 {
			return domain.Wrap(domain.CodeConflict, 409, "种植季仍被有效生产阶段的种植计划引用，无法停用", nil)
		}
	}
	projection := domain.SeasonUpdateProjection{ID: id, Code: u.Code, Name: u.Name, StartDate: start, EndDate: end, Status: u.Status}
	if err := s.store.SeasonRepo.UpdateProjected(ctx, s.store.DB(), projection); err != nil {
		return err
	}
	audit(ctx, s.audit, "update", "season", fmt.Sprintf("%d", id), u)
	return nil
}

func (s *seasonService) Get(ctx context.Context, id int64) (*domain.Season, error) {
	return s.store.SeasonRepo.GetByID(ctx, s.store.DB(), id)
}

func (s *seasonService) List(ctx context.Context, p domain.Pagination, status *int8) (domain.PageResult[*domain.Season], error) {
	p.Normalize()
	ss, total, err := s.store.SeasonRepo.List(ctx, s.store.DB(), p, status)
	if err != nil {
		return domain.PageResult[*domain.Season]{}, err
	}
	return domain.NewPageResult(ss, total, p), nil
}

func (s *seasonService) Delete(ctx context.Context, id int64) error {
	if err := s.store.SeasonRepo.Delete(ctx, s.store.DB(), id); err != nil {
		return err
	}
	audit(ctx, s.audit, "delete", "season", fmt.Sprintf("%d", id), nil)
	return nil
}

func parseSeasonRange(u *domain.SeasonUpsert) (time.Time, time.Time, error) {
	if u.Code == "" || u.Name == "" {
		return time.Time{}, time.Time{}, domain.Wrap(domain.CodeValidation, 400, "季次编码和名称必填", nil)
	}
	start, err := mustTimeValue(u.StartDate, "开始日期")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := mustTimeValue(u.EndDate, "结束日期")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if u.StartDate == "" || u.EndDate == "" {
		return time.Time{}, time.Time{}, domain.Wrap(domain.CodeValidation, 400, "开始日期和结束日期必填", nil)
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, domain.Wrap(domain.CodeValidation, 400, "结束日期必须晚于开始日期", nil)
	}
	return start, end, nil
}
