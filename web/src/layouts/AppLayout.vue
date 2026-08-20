<template>
  <div class="app-layout">
    <!-- ===== Sidebar ===== -->
    <aside class="sidebar" :class="{ collapsed }">
      <div class="logo">
        <span class="logo-mark">
          <svg viewBox="0 0 48 48" width="30" height="30" aria-hidden="true">
            <rect width="48" height="48" rx="12" fill="rgba(255,255,255,0.16)" />
            <path d="M24 11c5 5 7.5 8.5 7.5 12.5A7.5 7.5 0 0 1 16.5 23.5C16.5 19.5 19 16 24 11Z" fill="#c8e6c9" />
            <path d="M15 33h18v2.5H15z" fill="#a5d6a7" />
          </svg>
        </span>
        <transition name="fade"><span v-if="!collapsed" class="logo-text">农场管理</span></transition>
      </div>

      <el-scrollbar class="menu-scroll">
        <el-menu
          :default-active="route.path"
          :default-openeds="openGroups"
          :collapse="collapsed"
          :collapse-transition="false"
          :router="true"
          class="side-menu"
        >
          <el-menu-item index="/">
            <el-icon><DataLine /></el-icon><template #title>仪表盘</template>
          </el-menu-item>

          <el-sub-menu index="g-resource">
            <template #title><el-icon><MapLocation /></el-icon><span>资源管理</span></template>
            <el-menu-item v-permission="'farm:view'" index="/farms"><el-icon><OfficeBuilding /></el-icon><template #title>农场管理</template></el-menu-item>
            <el-menu-item v-permission="'field:view'" index="/fields"><el-icon><Location /></el-icon><template #title>地块管理</template></el-menu-item>
            <el-menu-item v-permission="'crop_variety:view'" index="/crop-varieties"><el-icon><Cherry /></el-icon><template #title>作物品种</template></el-menu-item>
            <el-menu-item v-permission="'season:view'" index="/seasons"><el-icon><Calendar /></el-icon><template #title>种植季</template></el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="g-production">
            <template #title><el-icon><Document /></el-icon><span>生产管理</span></template>
            <el-menu-item v-permission="'planting_plan:view'" index="/planting-plans"><el-icon><Document /></el-icon><template #title>种植计划</template></el-menu-item>
            <el-menu-item v-permission="'farm_task:view'" index="/tasks"><el-icon><Tickets /></el-icon><template #title>农事任务</template></el-menu-item>
            <el-menu-item v-permission="'harvest:view'" index="/harvests"><el-icon><Orange /></el-icon><template #title>采收管理</template></el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="g-input">
            <template #title><el-icon><Box /></el-icon><span>投入品</span></template>
            <el-menu-item v-permission="'inventory:view'" index="/materials"><el-icon><Box /></el-icon><template #title>投入品物料</template></el-menu-item>
            <el-menu-item v-permission="'inventory:view'" index="/batches"><el-icon><Files /></el-icon><template #title>批次与库存</template></el-menu-item>
            <el-menu-item v-permission="'allocation:manage'" index="/allocations"><el-icon><Goods /></el-icon><template #title>领用记录</template></el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="g-stock">
            <template #title><el-icon><Coin /></el-icon><span>库存与成本</span></template>
            <el-menu-item v-permission="'produce_inventory:view'" index="/produce-inventory"><el-icon><Coin /></el-icon><template #title>农产品库存</template></el-menu-item>
            <el-menu-item v-permission="'cost_analysis:view'" index="/cost-analyses"><el-icon><TrendCharts /></el-icon><template #title>成本分析</template></el-menu-item>
            <el-menu-item v-permission="'cost_analysis:view'" index="/reports/cost-summary"><el-icon><PieChart /></el-icon><template #title>成本汇总报表</template></el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="g-system">
            <template #title><el-icon><Setting /></el-icon><span>系统管理</span></template>
            <el-menu-item v-permission="'user:manage'" index="/users"><el-icon><User /></el-icon><template #title>用户管理</template></el-menu-item>
            <el-menu-item v-permission="'role:manage'" index="/roles"><el-icon><UserFilled /></el-icon><template #title>角色权限</template></el-menu-item>
            <el-menu-item v-permission="'audit_log:view'" index="/audit-logs"><el-icon><List /></el-icon><template #title>审计日志</template></el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-scrollbar>
    </aside>

    <!-- ===== Main column ===== -->
    <div class="main-col">
      <header class="topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn" @click="collapsed = !collapsed">
            <Fold v-if="!collapsed" /><Expand v-else />
          </el-icon>
          <el-breadcrumb :separator-icon="ArrowRight" class="crumb">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="topbar-right">
          <el-tooltip content="刷新当前页" placement="bottom">
            <el-icon class="icon-btn" @click="reload"><Refresh /></el-icon>
          </el-tooltip>
          <el-dropdown>
            <span class="user-chip">
              <el-avatar :size="32" class="avatar">{{ initials }}</el-avatar>
              <span class="user-meta">
                <span class="user-name">{{ auth.fullName }}</span>
                <span class="user-role">{{ roleNames }}</span>
              </span>
              <el-icon class="caret"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/profile')"><el-icon><User /></el-icon>个人资料</el-dropdown-item>
                <el-dropdown-item divided @click="onLogout"><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <main class="content">
        <router-view v-slot="{ Component }">
          <transition name="page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowRight, ArrowDown, Refresh, SwitchButton } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const collapsed = ref(false)

// Expand all grouped sub-menus by default for discoverability.
const openGroups = ['g-resource', 'g-production', 'g-input', 'g-stock', 'g-system']

const initials = computed(() => (auth.fullName || '?').slice(0, 1).toUpperCase())
const roleNames = computed(() => auth.roles.map((r) => r.name).join('、') || '用户')

// Map route path → breadcrumb title.
const titleMap: Record<string, string> = {
  '/': '仪表盘',
  '/farms': '农场管理', '/fields': '地块管理', '/crop-varieties': '作物品种', '/seasons': '种植季',
  '/planting-plans': '种植计划', '/tasks': '农事任务', '/harvests': '采收管理',
  '/materials': '投入品物料', '/batches': '批次与库存', '/allocations': '领用记录',
  '/produce-inventory': '农产品库存', '/cost-analyses': '成本分析', '/reports/cost-summary': '成本汇总报表',
  '/users': '用户管理', '/roles': '角色权限', '/audit-logs': '审计日志', '/profile': '个人资料',
}
const currentTitle = computed(() => {
  if (route.path.startsWith('/planting-plans/')) return '种植计划详情'
  return titleMap[route.path] || ''
})

function reload() { router.replace({ path: route.path, query: { t: Date.now().toString() } }) }
async function onLogout() { await auth.logout(); router.push('/login') }
</script>

<style scoped>
.app-layout { display: flex; height: 100vh; overflow: hidden; }

/* ===== Sidebar ===== */
.sidebar {
  width: 232px; flex: none;
  display: flex; flex-direction: column;
  background: linear-gradient(180deg, #15301f 0%, #1b3a2b 100%);
  box-shadow: 2px 0 12px rgba(0,0,0,0.12);
  transition: width .22s ease;
  z-index: 10;
}
.sidebar.collapsed { width: 64px; }
.logo {
  height: 64px; flex: none;
  display: flex; align-items: center; gap: 12px;
  padding: 0 20px; color: #fff;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.sidebar.collapsed .logo { justify-content: center; padding: 0; }
.logo-mark { display: flex; }
.logo-text { font-size: 17px; font-weight: 700; letter-spacing: 1px; white-space: nowrap; }
.menu-scroll { flex: 1; }
.side-menu { border-right: none; background: transparent !important; }

/* Deep-style the el-menu within this sidebar (dark theme overrides). */
.side-menu :deep(.el-menu-item),
.side-menu :deep(.el-sub-menu__title) {
  color: #b9cfc1; height: 46px; line-height: 46px; border-radius: 8px; margin: 2px 10px;
}
.side-menu :deep(.el-menu-item:hover),
.side-menu :deep(.el-sub-menu__title:hover) { background: rgba(255,255,255,0.08); color: #fff; }
.side-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, rgba(67,160,71,0.45), rgba(67,160,71,0.12));
  color: #fff; font-weight: 600;
}
.side-menu :deep(.el-sub-menu .el-menu) { background: rgba(0,0,0,0.18) !important; margin: 0 10px; border-radius: 8px; padding: 4px 0; }
.side-menu :deep(.el-sub-menu .el-menu .el-menu-item) { height: 42px; line-height: 42px; margin: 0; border-radius: 6px; }
.side-menu :deep(.el-sub-menu__icon-arrow) { color: #b9cfc1; }

/* ===== Main column ===== */
.main-col { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.topbar {
  height: 60px; flex: none;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 20px; background: #fff;
  border-bottom: 1px solid #eef2ee;
  box-shadow: 0 1px 4px rgba(27,58,43,0.04);
}
.topbar-left { display: flex; align-items: center; gap: 16px; }
.collapse-btn { font-size: 20px; cursor: pointer; color: #5a6b60; }
.collapse-btn:hover { color: var(--el-color-primary); }
.crumb { font-size: 14px; }
.crumb :deep(.el-breadcrumb__inner) { color: #6b7b70; }
.crumb :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) { color: #1f2d27; font-weight: 600; }

.topbar-right { display: flex; align-items: center; gap: 18px; }
.icon-btn { font-size: 18px; cursor: pointer; color: #5a6b60; }
.icon-btn:hover { color: var(--el-color-primary); }
.user-chip {
  display: flex; align-items: center; gap: 10px; cursor: pointer; padding: 4px 6px;
  border-radius: 24px; transition: background .15s;
}
.user-chip:hover { background: #f4f7f5; }
.avatar { background: linear-gradient(135deg, #2e7d32, #43a047); color: #fff; font-weight: 600; }
.user-meta { display: flex; flex-direction: column; line-height: 1.25; }
.user-name { font-size: 14px; color: #1f2d27; font-weight: 600; }
.user-role { font-size: 12px; color: #8a9a90; }
.caret { font-size: 12px; color: #8a9a90; }

.content { flex: 1; overflow: auto; padding: 20px; background: #f4f7f5; }

/* ===== Transitions ===== */
.fade-enter-active, .fade-leave-active { transition: opacity .18s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.page-enter-active { transition: all .2s ease; }
.page-enter-from { opacity: 0; transform: translateY(8px); }
</style>
