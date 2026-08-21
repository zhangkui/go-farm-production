<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.season_id" placeholder="全部季次" clearable filterable style="width:200px" @change="load">
        <el-option v-for="s in seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
    </div>
    <v-chart :option="option" class="chart" autoresize />
    <el-table :data="items" border stripe style="margin-top:12px">
      <el-table-column prop="field_name" label="地块" />
      <el-table-column prop="crop_variety_name" label="作物" />
      <el-table-column prop="season_name" label="季次" />
      <el-table-column prop="total_cost" label="总成本" width="120" />
      <el-table-column prop="total_yield" label="总产量(kg)" width="120" />
      <el-table-column label="单位成本(元/kg)" width="160"><template #default="{ row }">{{ row.unit_cost.toFixed(4) }}</template></el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { CostAPI, SeasonAPI, type CostSummary, type Season } from '@/api/modules'

use([CanvasRenderer, PieChart, TooltipComponent, LegendComponent, GridComponent])

const items = ref<CostSummary[]>([])
const seasons = ref<Season[]>([])
const q = reactive({ season_id: undefined as number | undefined })

const option = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} 元' },
  legend: { bottom: 0 },
  series: [{
    type: 'pie', radius: ['40%', '70%'],
    data: items.value.filter((i) => i.total_cost > 0).map((i) => ({ name: i.field_name || `计划#${i.planting_plan_id}`, value: +i.total_cost.toFixed(2) })),
    label: { formatter: '{b}\n{d}%' },
  }],
}))

async function loadSeasons() { seasons.value = (await SeasonAPI.list({ page: 1, page_size: 200 })).items }
async function load() { items.value = (await CostAPI.summary(q)).items }
onMounted(() => { loadSeasons(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }.chart { height: 360px; }</style>
