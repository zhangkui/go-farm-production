package domain

// Status enums are stored as TINYINT in MySQL. Centralising the numeric
// constants here keeps the Service, Repository and Transport layers in sync.

// Generic active/inactive status.
const (
	StatusInactive int8 = 0 // 停用
	StatusActive   int8 = 1 // 活跃
)

// Field status (地块状态).
const (
	FieldStatusInactive  int8 = 0 // 停用
	FieldStatusAvailable int8 = 1 // 可用
	FieldStatusPlanting  int8 = 2 // 种植中
	FieldStatusFallow    int8 = 3 // 休耕
)

// Planting plan status (种植计划状态).
const (
	PlanStatusCancelled int8 = 0 // 已取消（不可逆）
	PlanStatusPlanned   int8 = 1 // 计划中
	PlanStatusPlanted   int8 = 2 // 已播种
	PlanStatusGrowing   int8 = 3 // 生长中
	PlanStatusHarvested int8 = 4 // 已采收
	PlanStatusCompleted int8 = 5 // 已完成
)

// Task status (农事任务状态).
const (
	TaskStatusCancelled  int8 = 0 // 已取消
	TaskStatusScheduled  int8 = 1 // 计划中
	TaskStatusInProgress int8 = 2 // 进行中
	TaskStatusCompleted  int8 = 3 // 已完成
)

// Batch status (投入品批次状态).
const (
	BatchStatusVoid     int8 = 0 // 无效
	BatchStatusActive   int8 = 1 // 有效
	BatchStatusDepleted int8 = 2 // 已耗尽
	BatchStatusExpired  int8 = 3 // 已过期
)

// Allocation type (投入品领用类型).
const (
	AllocationTypeReturn   int8 = 0 // 退回（增加库存）
	AllocationTypeAllocate int8 = 1 // 领用（扣减库存）
	AllocationTypeWaste    int8 = 2 // 损耗（扣减库存，不可恢复）
)

// Produce inventory status (农产品库存状态).
const (
	ProduceStatusWasted    int8 = 0 // 已损耗
	ProduceStatusInStock   int8 = 1 // 在库
	ProduceStatusSold      int8 = 2 // 已售
	ProduceStatusProcessed int8 = 3 // 已加工
)

// Task type (农事任务类型).
const (
	TaskTypeFertilizing int8 = 1 // 施肥
	TaskTypeIrrigation  int8 = 2 // 灌溉
	TaskTypePestControl int8 = 3 // 病虫害防治
	TaskTypeHarvesting  int8 = 4 // 采收
	TaskTypeOther       int8 = 5 // 其他
)

// Material category (投入品品类).
const (
	MaterialCatFertilizer int8 = 1 // 肥料
	MaterialCatPesticide  int8 = 2 // 农药
	MaterialCatSeed       int8 = 3 // 种子
	MaterialCatFuel       int8 = 4 // 燃料
	MaterialCatOther      int8 = 5 // 其他
)

// PlanStatusName returns a human-readable name for a plan status.
func PlanStatusName(s int8) string {
	switch s {
	case PlanStatusCancelled:
		return "cancelled"
	case PlanStatusPlanned:
		return "planned"
	case PlanStatusPlanted:
		return "planted"
	case PlanStatusGrowing:
		return "growing"
	case PlanStatusHarvested:
		return "harvested"
	case PlanStatusCompleted:
		return "completed"
	}
	return "unknown"
}

// IsPlanActive reports whether the status represents an active (non-terminal,
// non-cancelled) plan that occupies land and must obey the overlap constraint.
func IsPlanActive(s int8) bool {
	switch s {
	case PlanStatusPlanned, PlanStatusPlanted, PlanStatusGrowing, PlanStatusHarvested:
		return true
	}
	return false
}

// AllowedPlanTransition reports whether transitioning from -> to is legal.
func AllowedPlanTransition(from, to int8) bool {
	if to == PlanStatusCancelled {
		return from != PlanStatusCancelled && from != PlanStatusCompleted
	}
	switch from {
	case PlanStatusPlanned:
		return to == PlanStatusPlanted
	case PlanStatusPlanted:
		return to == PlanStatusGrowing
	case PlanStatusGrowing:
		return to == PlanStatusHarvested
	case PlanStatusHarvested:
		return to == PlanStatusCompleted
	}
	return false
}
