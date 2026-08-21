import { api, type PageResult } from './client'

// Aggregate typed API module for every resource. Each function returns the
// unwrapped `data` payload. Pagination params are spread from the caller.

export const AuthAPI = {
  login: (username: string, password: string) =>
    api.post<{ access_token: string; refresh_token: string; expires_at: string }>(
      '/auth/login',
      { username, password }
    ),
  refresh: (refresh_token: string) => api.post('/auth/refresh', { refresh_token }),
  logout: (refresh_token: string) => api.post('/auth/logout', { refresh_token }),
  register: (body: Record<string, unknown>) => api.post<{ id: number }>('/auth/register', body),
}

export const MeAPI = {
  me: () =>
    api.get<{ user: User; roles: Role[]; permissions: string[] }>('/me'),
  changePassword: (old_password: string, new_password: string) =>
    api.patch('/me/password', { old_password, new_password }),
}

export const FarmAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Farm>>('/farms', { params }),
  get: (id: number) => api.get<Farm>(`/farms/${id}`),
  create: (body: Partial<Farm>) => api.post<{ id: number }>('/farms', body),
  update: (id: number, body: Partial<Farm>) => api.put(`/farms/${id}`, body),
  remove: (id: number) => api.delete(`/farms/${id}`),
}

export const FieldAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Field>>('/fields', { params }),
  get: (id: number) => api.get<Field>(`/fields/${id}`),
  create: (body: Partial<Field>) => api.post<{ id: number }>('/fields', body),
  update: (id: number, body: Partial<Field>) => api.put(`/fields/${id}`, body),
  remove: (id: number) => api.delete(`/fields/${id}`),
}

export const VarietyAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<CropVariety>>('/crop-varieties', { params }),
  create: (body: Partial<CropVariety>) => api.post<{ id: number }>('/crop-varieties', body),
  update: (id: number, body: Partial<CropVariety>) => api.put(`/crop-varieties/${id}`, body),
  remove: (id: number) => api.delete(`/crop-varieties/${id}`),
}

export const SeasonAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Season>>('/seasons', { params }),
  create: (body: Partial<Season>) => api.post<{ id: number }>('/seasons', body),
  update: (id: number, body: Partial<Season>) => api.put(`/seasons/${id}`, body),
  remove: (id: number) => api.delete(`/seasons/${id}`),
}

export const PlanAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<PlantingPlan>>('/planting-plans', { params }),
  get: (id: number) =>
    api.get<{ plan: PlantingPlan; tasks: FarmTask[]; harvests: Harvest[] }>(`/planting-plans/${id}`),
  create: (body: Partial<PlantingPlan>) => api.post<{ id: number }>('/planting-plans', body),
  update: (id: number, body: Partial<PlantingPlan>) => api.put(`/planting-plans/${id}`, body),
  remove: (id: number) => api.delete(`/planting-plans/${id}`),
  transition: (id: number, to: number) =>
    api.post(`/planting-plans/${id}/transition`, { to }),
}

export const TaskAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<FarmTask>>('/tasks', { params }),
  get: (id: number) =>
    api.get<{ task: FarmTask; allocations: Allocation[] }>(`/tasks/${id}`),
  create: (body: Partial<FarmTask>) => api.post<{ id: number }>('/tasks', body),
  update: (id: number, body: Partial<FarmTask>) => api.put(`/tasks/${id}`, body),
  remove: (id: number) => api.delete(`/tasks/${id}`),
  transition: (id: number, to: number) => api.post(`/tasks/${id}/transition`, { to }),
}

export const MaterialAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Material>>('/materials', { params }),
  create: (body: Partial<Material>) => api.post<{ id: number }>('/materials', body),
  update: (id: number, body: Partial<Material>) => api.put(`/materials/${id}`, body),
  remove: (id: number) => api.delete(`/materials/${id}`),
}

export const BatchAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Batch>>('/batches', { params }),
  get: (id: number) => api.get<{ batch: Batch; allocations: Allocation[] }>(`/batches/${id}`),
  create: (body: Partial<Batch>) => api.post<{ id: number }>('/batches', body),
  update: (id: number, body: Partial<Batch>) => api.put(`/batches/${id}`, body),
  remove: (id: number) => api.delete(`/batches/${id}`),
  warnings: (within_days = 30) =>
    api.get<{ items: Batch[] }>('/inventory/warnings', { params: { within_days } }),
}

export const AllocationAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Allocation>>('/allocations', { params }),
  create: (body: Partial<Allocation>) => api.post<{ id: number }>('/allocations', body),
  get: (id: number) => api.get<Allocation>(`/allocations/${id}`),
}

export const HarvestAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Harvest>>('/harvests', { params }),
  get: (id: number) =>
    api.get<{ harvest: Harvest; details: HarvestDetail[] }>(`/harvests/${id}`),
  create: (body: Partial<Harvest>) => api.post<{ id: number }>('/harvests', body),
  update: (id: number, body: Partial<Harvest>) => api.put(`/harvests/${id}`, body),
  remove: (id: number) => api.delete(`/harvests/${id}`),
  approve: (id: number) => api.post(`/harvests/${id}/approve`),
}

export const ProduceAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<Produce>>('/produce-inventory', { params }),
  create: (body: Partial<Produce>) => api.post<{ id: number }>('/produce-inventory', body),
  update: (id: number, body: Partial<Produce>) => api.put(`/produce-inventory/${id}`, body),
  remove: (id: number) => api.delete(`/produce-inventory/${id}`),
}

export const CostAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<CostAnalysis>>('/cost-analyses', { params }),
  compute: (body: { planting_plan_id: number; analysis_date?: string; other_cost?: number; remark?: string }) =>
    api.post<{ id: number }>('/cost-analyses', body),
  get: (id: number) => api.get<CostAnalysis>(`/cost-analyses/${id}`),
  getByPlan: (plan_id: number) => api.get<CostAnalysis>(`/cost-analyses/plan/${plan_id}`),
  summary: (params: Record<string, unknown>) =>
    api.get<{ items: CostSummary[] }>('/reports/cost-summary', { params }),
}

export const AuditAPI = {
  list: (params: Record<string, unknown>) =>
    api.get<PageResult<AuditLog>>('/audit-logs', { params }),
}

// --- shared model types (mirrored from Go domain) ---
export interface User { id: number; username: string; email: string; full_name: string; status: number; created_at: string; updated_at: string }
export interface Role { id: number; code: string; name: string; description: string }
export interface Permission { id: number; code: string; name: string; resource: string; action: string }
export interface Farm { id: number; name: string; location: string; total_area: number; description: string; status: number; created_at: string }
export interface Field { id: number; farm_id: number; farm_name?: string; code: string; name: string; area: number; soil_type: string; irrigation_zone: string; status: number; remark: string }
export interface CropVariety { id: number; code: string; name: string; category: string; growth_cycle: number; description: string; status: number }
export interface Season { id: number; code: string; name: string; start_date: string; end_date: string; status: number }
export interface PlantingPlan { id: number; field_id: number; field_name?: string; crop_variety_id: number; crop_variety_name?: string; season_id: number; season_name?: string; planned_area: number; planned_sow_date: string; planned_harvest_date?: string; actual_sow_date?: string; actual_harvest_date?: string; status: number; remark: string }
export interface FarmTask { id: number; planting_plan_id: number; task_type: number; title: string; description: string; planned_date: string; completed_date?: string; status: number; labour_hours: number; equipment_cost: number; assignee_id?: number; assignee_name?: string; remark: string }
export interface Material { id: number; code: string; name: string; category: number; unit: string; unit_price: number; description: string; status: number }
export interface Batch { id: number; material_id: number; material_name?: string; batch_no: string; quantity: number; remaining_qty: number; purchase_date?: string; expiry_date?: string; purchase_price: number; supplier: string; status: number }
export interface Allocation { id: number; batch_id: number; batch_no?: string; material_id: number; material_name?: string; task_id?: number; quantity: number; type: number; remark: string; operator_id: number; operator_name?: string; created_at: string }
export interface Harvest { id: number; planting_plan_id: number; harvest_date: string; total_weight: number; grade: string; remark: string; approved: boolean; created_at: string }
export interface HarvestDetail { id: number; harvest_id: number; field_id: number; field_name?: string; weight: number; grade: string; remark: string }
export interface Produce { id: number; crop_variety_id: number; crop_variety_name?: string; harvest_id: number; quantity: number; grade: string; unit: string; storage_location: string; status: number }
export interface CostAnalysis { id: number; planting_plan_id: number; analysis_date: string; labour_cost: number; input_cost: number; equipment_cost: number; other_cost: number; total_cost: number; total_yield: number; unit_cost: number; remark: string }
export interface CostSummary { planting_plan_id: number; field_name: string; crop_variety_name: string; season_name: string; total_cost: number; total_yield: number; unit_cost: number }
export interface AuditLog { id: number; user_id?: number; username: string; action: string; resource_type: string; resource_id: string; details: string; ip_address: string; user_agent: string; created_at: string }
