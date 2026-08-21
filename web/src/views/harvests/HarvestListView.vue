<template>
  <div>
    <div class="toolbar">
      <el-input v-model="q.planting_plan_id" placeholder="种植计划ID" clearable style="width:140px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button v-permission="'harvest:manage'" type="success" @click="openCreate">新建采收</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="planting_plan_id" label="计划ID" width="90" />
      <el-table-column prop="harvest_date" label="日期" width="120" />
      <el-table-column prop="total_weight" label="总重量(kg)" width="120" />
      <el-table-column prop="grade" label="等级" width="80" />
      <el-table-column label="审核" width="90"><template #default="{ row }"><el-tag :type="row.approved?'success':'info'">{{ row.approved?'已审核':'待审核' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button v-permission="'harvest:approve'" v-if="!row.approved" size="small" type="warning" @click="onApprove(row)">审核</el-button>
          <el-button v-permission="'harvest:manage'" size="small" :disabled="row.approved" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="'harvest:manage'" size="small" type="danger" :disabled="row.approved" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑采收' : '新建采收'" width="640px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="种植计划"><el-select v-model="form.planting_plan_id" filterable placeholder="选择种植计划" style="width:100%"><el-option v-for="p in plans" :key="p.id" :label="`#${p.id} ${p.field_name||''} - ${p.crop_variety_name||''}`" :value="p.id" /></el-select></el-form-item>
        <el-form-item label="采收日期"><el-date-picker v-model="form.harvest_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="总重量(kg)"><el-input-number v-model="form.total_weight" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="等级"><el-input v-model="form.grade" placeholder="一级/二级" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { HarvestAPI, PlanAPI, type Harvest, type PlantingPlan } from '@/api/modules'

const rows = ref<Harvest[]>([]); const total = ref(0); const loading = ref(false)
const plans = ref<PlantingPlan[]>([])
const q = reactive({ page: 1, page_size: 20, planting_plan_id: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<Harvest>>({ planting_plan_id: undefined, harvest_date: '', total_weight: 0, grade: '', remark: '' })

async function loadPlans() { plans.value = (await PlanAPI.list({ page: 1, page_size: 200 })).items }
async function load() { loading.value = true; try { const res = await HarvestAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, planting_plan_id: plans.value[0]?.id, harvest_date: '', total_weight: 0, grade: '', remark: '' }); dialog.value = true }
function openEdit(row: Harvest) { Object.assign(form, row); dialog.value = true }
async function onSave() { saving.value = true; try { if (form.id) await HarvestAPI.update(form.id, form); else await HarvestAPI.create(form); ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false } }
async function onApprove(row: Harvest) { await HarvestAPI.approve(row.id); ElMessage.success('已审核'); await load() }
async function onDel(row: Harvest) { await ElMessageBox.confirm('确认删除该采收记录？', '提示', { type: 'warning' }); await HarvestAPI.remove(row.id); ElMessage.success('已删除'); await load() }
onMounted(() => { loadPlans(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
