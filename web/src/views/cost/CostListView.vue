<template>
  <div>
    <div class="toolbar">
      <el-input v-model="q.planting_plan_id" placeholder="种植计划ID" clearable style="width:140px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openCompute">计算成本</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="planting_plan_id" label="计划ID" width="90" />
      <el-table-column prop="analysis_date" label="分析日期" width="120" />
      <el-table-column prop="labour_cost" label="人工成本" width="110" />
      <el-table-column prop="input_cost" label="投入品成本" width="110" />
      <el-table-column prop="equipment_cost" label="设备成本" width="110" />
      <el-table-column prop="other_cost" label="其他成本" width="110" />
      <el-table-column prop="total_cost" label="总成本" width="110" />
      <el-table-column prop="total_yield" label="总产量(kg)" width="120" />
      <el-table-column label="单位成本(元/kg)" width="140"><template #default="{ row }">{{ row.unit_cost.toFixed(4) }}</template></el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" title="计算种植计划成本" width="460px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="种植计划"><el-select v-model="form.planting_plan_id" filterable placeholder="选择种植计划" style="width:100%"><el-option v-for="p in plans" :key="p.id" :label="`#${p.id} ${p.field_name||''} - ${p.crop_variety_name||''}`" :value="p.id" /></el-select></el-form-item>
        <el-form-item label="分析日期"><el-date-picker v-model="form.analysis_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="其他成本"><el-input-number v-model="form.other_cost" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">计算并保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { CostAPI, PlanAPI, type CostAnalysis, type PlantingPlan } from '@/api/modules'

const rows = ref<CostAnalysis[]>([]); const total = ref(0); const loading = ref(false)
const plans = ref<PlantingPlan[]>([])
const q = reactive({ page: 1, page_size: 20, planting_plan_id: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive({ planting_plan_id: undefined as number | undefined, analysis_date: '', other_cost: 0, remark: '' })

async function loadPlans() { plans.value = (await PlanAPI.list({ page: 1, page_size: 200 })).items }
async function load() { loading.value = true; try { const res = await CostAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCompute() { Object.assign(form, { planting_plan_id: plans.value[0]?.id, analysis_date: '', other_cost: 0, remark: '' }); dialog.value = true }
async function onSave() {
  saving.value = true
  try { await CostAPI.compute(form as any); ElMessage.success('成本已计算'); dialog.value = false; await load() } finally { saving.value = false }
}
onMounted(() => { loadPlans(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
