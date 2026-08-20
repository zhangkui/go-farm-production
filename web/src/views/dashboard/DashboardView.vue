<template>
  <div class="dashboard">
    <el-row :gutter="16">
      <el-col :span="6" v-for="card in cards" :key="card.label">
        <el-card shadow="hover">
          <div class="stat">
            <div class="num">{{ card.value }}</div>
            <div class="label">{{ card.label }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="14">
        <el-card>
          <template #header>单位产量成本排行 (前10)</template>
          <v-chart :option="costOption" class="chart" autoresize />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card>
          <template #header>批次库存预警</template>
          <el-table :data="warnings" size="small" max-height="320">
            <el-table-column prop="batch_no" label="批号" />
            <el-table-column prop="material_name" label="物料" />
            <el-table-column prop="remaining_qty" label="剩余" width="80" />
            <el-table-column label="有效期">
              <template #default="{ row }">{{ row.expiry_date || '-' }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import { TooltipComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { PlanAPI, BatchAPI, HarvestAPI, TaskAPI, CostAPI } from '@/api/modules'

use([CanvasRenderer, BarChart, TooltipComponent, GridComponent])

const planTotal = ref(0)
const taskTotal = ref(0)
const harvestTotal = ref(0)
const warnings = ref<Awaited<ReturnType<typeof BatchAPI['warnings']>>['items']>([])

const cards = computed(() => [
  { label: '种植计划数', value: planTotal.value },
  { label: '农事任务数', value: taskTotal.value },
  { label: '采收记录数', value: harvestTotal.value },
  { label: '库存预警批次', value: warnings.value.length },
])

const costOption = ref({})
const summary = ref<{ field_name: string; unit_cost: number }[]>([])

onMounted(async () => {
  const [plans, tasks, harvests, warn, summ] = await Promise.all([
    PlanAPI.list({ page: 1, page_size: 1 }),
    TaskAPI.list({ page: 1, page_size: 1 }),
    HarvestAPI.list({ page: 1, page_size: 1 }),
    BatchAPI.warnings(30).catch(() => ({ items: [] })),
    CostAPI.summary({}).catch(() => ({ items: [] })),
  ])
  planTotal.value = plans.total
  taskTotal.value = tasks.total
  harvestTotal.value = harvests.total
  warnings.value = warn.items
  summary.value = ((summ.items as any[])
    .sort((a, b) => a.unit_cost - b.unit_cost)
    .slice(0, 10))
  costOption.value = {
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: summary.value.map((s) => s.field_name) },
    yAxis: { type: 'value', name: '元/kg' },
    series: [{ type: 'bar', data: summary.value.map((s) => +s.unit_cost.toFixed(2)), itemStyle: { color: '#2e7d32' } }],
  }
})
</script>

<style scoped>
.stat .num { font-size: 28px; font-weight: 600; color: #2e7d32; }
.stat .label { color: #888; margin-top: 4px; }
.chart { height: 320px; }
</style>
