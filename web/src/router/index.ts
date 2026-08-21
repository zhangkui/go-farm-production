import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api, registerRouter } from '@/api/client'

// Route table. Each authenticated route declares the permission(s) required;
// the guard below enforces them (backend is the source of truth).
const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/auth/LoginView.vue'), meta: { public: true } },
  { path: '/register', name: 'register', component: () => import('@/views/auth/RegisterView.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    children: [
      { path: '', name: 'dashboard', component: () => import('@/views/dashboard/DashboardView.vue') },
      { path: 'farms', name: 'farms', component: () => import('@/views/farms/FarmListView.vue'), meta: { perm: 'farm:view' } },
      { path: 'fields', name: 'fields', component: () => import('@/views/fields/FieldListView.vue'), meta: { perm: 'field:view' } },
      { path: 'crop-varieties', name: 'crop-varieties', component: () => import('@/views/varieties/VarietyListView.vue'), meta: { perm: 'crop_variety:view' } },
      { path: 'seasons', name: 'seasons', component: () => import('@/views/seasons/SeasonListView.vue'), meta: { perm: 'season:view' } },
      { path: 'planting-plans', name: 'planting-plans', component: () => import('@/views/plans/PlanListView.vue'), meta: { perm: 'planting_plan:view' } },
      { path: 'planting-plans/:id', name: 'plan-detail', component: () => import('@/views/plans/PlanDetailView.vue'), meta: { perm: 'planting_plan:view' } },
      { path: 'tasks', name: 'tasks', component: () => import('@/views/tasks/TaskListView.vue'), meta: { perm: 'farm_task:view' } },
      { path: 'materials', name: 'materials', component: () => import('@/views/materials/MaterialListView.vue'), meta: { perm: 'inventory:view' } },
      { path: 'batches', name: 'batches', component: () => import('@/views/batches/BatchListView.vue'), meta: { perm: 'inventory:view' } },
      { path: 'allocations', name: 'allocations', component: () => import('@/views/allocations/AllocationListView.vue'), meta: { perm: 'allocation:manage' } },
      { path: 'harvests', name: 'harvests', component: () => import('@/views/harvests/HarvestListView.vue'), meta: { perm: 'harvest:view' } },
      { path: 'produce-inventory', name: 'produce-inventory', component: () => import('@/views/produce/ProduceListView.vue'), meta: { perm: 'produce_inventory:view' } },
      { path: 'cost-analyses', name: 'cost-analyses', component: () => import('@/views/cost/CostListView.vue'), meta: { perm: 'cost_analysis:view' } },
      { path: 'reports/cost-summary', name: 'cost-summary', component: () => import('@/views/reports/CostSummaryView.vue'), meta: { perm: 'cost_analysis:view' } },
      { path: 'users', name: 'users', component: () => import('@/views/users/UserListView.vue'), meta: { perm: 'user:manage' } },
      { path: 'roles', name: 'roles', component: () => import('@/views/roles/RoleListView.vue'), meta: { perm: 'role:manage' } },
      { path: 'audit-logs', name: 'audit-logs', component: () => import('@/views/audit/AuditLogView.vue'), meta: { perm: 'audit_log:view' } },
      { path: 'profile', name: 'profile', component: () => import('@/views/profile/ProfileView.vue') },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

registerRouter(router)

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (auth.isLoggedIn && (to.name === 'login' || to.name === 'register')) {
      return { name: 'dashboard' }
    }
    return true
  }
  if (!auth.isLoggedIn) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (!auth.initialized) {
    try { await auth.fetchMe() } catch { await auth.logout(); return { name: 'login' } }
  }
  const perm = to.meta.perm as string | undefined
  if (perm && !auth.hasPermission(perm)) {
    return { name: 'dashboard' }
  }
  return true
})

// Expose for app bootstrap.
export default router
