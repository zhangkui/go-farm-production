<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.type" placeholder="全部类型" clearable style="width:130px" @change="load">
        <el-option :value="1" label="领用" /><el-option :value="0" label="退回" /><el-option :value="2" label="损耗" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openCreate">新增领用</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="batch_no" label="批号" width="120" />
      <el-table-column prop="material_name" label="物料" />
      <el-table-column prop="quantity" label="数量" width="100" />
      <el-table-column label="类型" width="80"><template #default="{ row }"><el-tag :type="allocationTypeTag[row.type]">{{ allocationTypeText[row.type] }}</el-tag></template></el-table-column>
      <el-table-column prop="operator_name" label="操作人" width="100" />
      <el-table-column prop="created_at" label="时间" width="170" />
      <el-table-column prop="remark" label="备注" />
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" title="新增投入品领用" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="批次"><el-select v-model="form.batch_id" filterable><el-option v-for="b in batches" :key="b.id" :label="`${b.batch_no} (剩${b.remaining_qty})`" :value="b.id" /></el-select></el-form-item>
        <el-form-item label="关联任务"><el-select v-model="form.task_id" filterable clearable placeholder="可选" style="width:100%"><el-option v-for="t in tasks" :key="t.id" :label="`#${t.id} ${t.title}`" :value="t.id" /></el-select></el-form-item>
        <el-form-item label="数量"><el-input-number v-model="form.quantity" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="类型"><el-select v-model="form.type"><el-option :value="1" label="领用" /><el-option :value="0" label="退回" /><el-option :value="2" label="损耗" /></el-select></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { AllocationAPI, BatchAPI, TaskAPI, type Allocation, type Batch, type FarmTask } from '@/api/modules'
import { allocationTypeText, allocationTypeTag } from '@/utils/constants'

const rows = ref<Allocation[]>([]); const total = ref(0); const loading = ref(false)
const batches = ref<Batch[]>([]); const tasks = ref<FarmTask[]>([])
const q = reactive({ page: 1, page_size: 20, type: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<Allocation>>({ batch_id: undefined, task_id: undefined, quantity: 0, type: 1, remark: '' })

async function loadBatches() { batches.value = (await BatchAPI.list({ page: 1, page_size: 200, status: 1 })).items }
async function loadTasks() { tasks.value = (await TaskAPI.list({ page: 1, page_size: 200 })).items }
async function load() { loading.value = true; try { const res = await AllocationAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, batch_id: batches.value[0]?.id, task_id: undefined, quantity: 0, type: 1, remark: '' }); dialog.value = true }
async function onSave() {
  saving.value = true
  try { await AllocationAPI.create(form); ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false }
}
onMounted(() => { loadBatches(); loadTasks(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
