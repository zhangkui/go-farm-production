package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// Container wires services together with their shared dependencies. The
// transport layer builds one Container at startup and resolves services by name.
type Container struct {
	Store  *repository.Store
	Cache  *repository.Cache
	Hasher Hasher
	Tokens TokenSigner

	Audit    AuditService
	Auth     AuthService
	User     UserService
	Role     RoleService
	Farm     FarmService
	Field    FieldService
	Variety  CropVarietyService
	Season   SeasonService
	Plan     PlantingPlanService
	Task     FarmTaskService
	Material MaterialService
	Batch    InputBatchService
	Alloc    InputAllocationService
	Harvest  HarvestService
	Produce  ProduceInventoryService
	Cost     CostAnalysisService
}

// New constructs the full service container.
func New(store *repository.Store, cache *repository.Cache, h Hasher, t TokenSigner) *Container {
	c := &Container{Store: store, Cache: cache, Hasher: h, Tokens: t}
	c.Audit = NewAuditService(store, cache)
	c.Auth = NewAuthService(store, cache, h, t, c.Audit)
	c.User = NewUserService(store, c.Audit, h)
	c.Role = NewRoleService(store, c.Audit)
	c.Farm = NewFarmService(store, c.Audit)
	c.Field = NewFieldService(store, c.Audit)
	c.Variety = NewCropVarietyService(store, c.Audit)
	c.Season = NewSeasonService(store, c.Audit)
	c.Plan = NewPlantingPlanService(store, c.Audit)
	c.Task = NewFarmTaskService(store, c.Audit)
	c.Material = NewMaterialService(store, c.Audit)
	c.Batch = NewInputBatchService(store, c.Audit)
	c.Alloc = NewInputAllocationService(store, c.Audit)
	c.Harvest = NewHarvestService(store, c.Audit)
	c.Produce = NewProduceInventoryService(store, c.Audit)
	c.Cost = NewCostAnalysisService(store)
	return c
}

// labourRate is the configurable cost per labour hour (yuan). Centralised so
// the cost service can override it per deployment if needed.
const labourRate = 50.0

// nowUTC returns the current UTC time. Centralised so tests can stub it.
func nowUTC() time.Time { return time.Now().UTC() }

// jsonMarshal serialises v to a JSON string, returning "{}" on error.
func jsonMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// audit emits an audit entry through the store, swallowing errors so a failed
// audit write never breaks the primary business operation. Details is JSON-
// encoded before persistence.
func audit(ctx context.Context, svc AuditService, action, resourceType, resourceID string, details any) {
	m, _ := MetaFrom(ctx)
	var detailJSON string
	if details != nil {
		b, err := json.Marshal(details)
		if err == nil {
			detailJSON = string(b)
		} else {
			detailJSON = fmt.Sprintf(`{"error":"%s"}`, err.Error())
		}
	}
	uid := m.UserID
	e := domain.AuditEntry{
		UserID:       uid,
		Username:     m.Username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      detailJSON,
		IPAddress:    m.IPAddress,
		UserAgent:    m.UserAgent,
	}
	_ = svc.Log(ctx, e)
}

// mustTime parses an RFC3339 date string, returning a pointer. Empty string
// returns nil. Invalid strings raise a validation error.
func mustTime(s string, field string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, domain.Wrap(domain.CodeValidation, 400, field+" 格式无效", err)
		}
	}
	return &t, nil
}

// mustTimeValue is like mustTime but returns a non-pointer value (zero on empty).
func mustTimeValue(s string, field string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := mustTime(s, field)
	if err != nil {
		return time.Time{}, err
	}
	if t == nil {
		return time.Time{}, nil
	}
	return *t, nil
}
