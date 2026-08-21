<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.status" placeholder="状态" clearable style="width:130px" @change="load">
        <el-option :value="1" label="计划中" /><el-option :value="2" label="进行中" /><el-option :value="3" label="已完成" /><el-option :value="0" label="已取消" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button v-permission="'farm_task:manage'" type="success" @click="openCreate">新建任务</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" />
      <el-table-column label="类型" width="120"><template #default="{ row }">{{ taskTypeText[row.task_type] }}</template></el-table-column>
      <el-table-column prop="planned_date" label="计划日期" width="120" />
      <el-table-column prop="labour_hours" label="工时" width="80" />
      <el-table-column prop="equipment_cost" label="设备成本" width="100" />
      <el-table-column prop="assignee_name" label="负责人" width="100" />
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="statusTag(row.status,'task')">{{ statusText(row.status,'task') }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button v-permission="'farm_task:execute'" size="small" :disabled="row.status===3" @click="onTransition(row, row.status+1)">推进</el-button>
          <el-button v-permission="'farm_task:manage'" size="small" @click="openEdit(row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑任务' : '新建任务'" width="560px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="种植计划"><el-select v-model="form.planting_plan_id" filterable placeholder="选择种植计划" style="width:100%"><el-option v-for="p in plans" :key="p.id" :label="`#${p.id} ${p.field_name||''} - ${p.crop_variety_name||''}`" :value="p.id" /></el-select></el-form-item>
        <el-form-item label="类型"><el-select v-model="form.task_type"><el-option :value="1" label="施肥" /><el-option :value="2" label="灌溉" /><el-option :value="3" label="病虫害防治" /><el-option :value="4" label="采收" /><el-option :value="5" label="其他" /></el-select></el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="计划日期"><el-date-picker v-model="form.planned_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="工时(小时)"><el-input-number v-model="form.labour_hours" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="设备成本(元)"><el-input-number v-model="form.equipment_cost" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { TaskAPI, PlanAPI, type FarmTask, type PlantingPlan } from '@/api/modules'
import { statusText, statusTag, taskTypeText } from '@/utils/constants'

const rows = ref<FarmTask[]>([]); const total = ref(0); const loading = ref(false)
const plans = ref<PlantingPlan[]>([])
const q = reactive({ page: 1, page_size: 20, status: undefined as number | undefined, planting_plan_id: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<FarmTask>>({ planting_plan_id: undefined, task_type: 5, title: '', planned_date: '', labour_hours: 0, equipment_cost: 0, description: '' })

async function loadPlans() { plans.value = (await PlanAPI.list({ page: 1, page_size: 200 })).items }
async function load() { loading.value = true; try { const res = await TaskAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, planting_plan_id: plans.value[0]?.id, task_type: 5, title: '', planned_date: '', labour_hours: 0, equipment_cost: 0, description: '' }); dialog.value = true }
function openEdit(row: FarmTask) { Object.assign(form, row); dialog.value = true }
async function onSave() { saving.value = true; try { if (form.id) await TaskAPI.update(form.id, form); else { await TaskAPI.create(form) }; ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false } }
async function onTransition(row: FarmTask, to: number) { await TaskAPI.transition(row.id, to); ElMessage.success('状态已更新'); await load() }
onMounted(() => { loadPlans(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
