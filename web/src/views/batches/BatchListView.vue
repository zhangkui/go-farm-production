<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.material_id" placeholder="全部物料" clearable filterable style="width:200px" @change="load">
        <el-option v-for="m in materials" :key="m.id" :label="m.name" :value="m.id" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="warning" @click="loadWarnings">库存预警</el-button>
      <el-button v-permission="'batch:manage'" type="success" @click="openCreate">入库批次</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="batch_no" label="批号" width="120" />
      <el-table-column prop="material_name" label="物料" />
      <el-table-column prop="quantity" label="入库量" width="100" />
      <el-table-column prop="remaining_qty" label="剩余量" width="100" />
      <el-table-column prop="expiry_date" label="有效期" width="120" />
      <el-table-column prop="supplier" label="供应商" />
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="statusTag(row.status,'batch')">{{ statusText(row.status,'batch') }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-permission="'batch:manage'" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="'batch:manage'" size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑批次' : '入库批次'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="物料"><el-select v-model="form.material_id" filterable><el-option v-for="m in materials" :key="m.id" :label="m.name" :value="m.id" /></el-select></el-form-item>
        <el-form-item label="批号"><el-input v-model="form.batch_no" /></el-form-item>
        <el-form-item label="数量"><el-input-number v-model="form.quantity" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="采购日期"><el-date-picker v-model="form.purchase_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="有效期"><el-date-picker v-model="form.expiry_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="采购单价"><el-input-number v-model="form.purchase_price" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="供应商"><el-input v-model="form.supplier" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="warnDialog" title="库存预警 (30天内到期)" width="640px">
      <el-table :data="warnings" border size="small">
        <el-table-column prop="batch_no" label="批号" /><el-table-column prop="material_name" label="物料" /><el-table-column prop="remaining_qty" label="剩余" width="90" /><el-table-column prop="expiry_date" label="有效期" width="120" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { BatchAPI, MaterialAPI, type Batch, type Material } from '@/api/modules'
import { statusText, statusTag } from '@/utils/constants'

const rows = ref<Batch[]>([]); const total = ref(0); const loading = ref(false)
const materials = ref<Material[]>([])
const q = reactive({ page: 1, page_size: 20, material_id: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<Batch>>({ material_id: undefined, batch_no: '', quantity: 0, purchase_date: '', expiry_date: '', purchase_price: 0, supplier: '' })
const warnDialog = ref(false); const warnings = ref<Batch[]>([])

async function loadMaterials() { materials.value = (await MaterialAPI.list({ page: 1, page_size: 200 })).items }
async function load() { loading.value = true; try { const res = await BatchAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, material_id: materials.value[0]?.id, batch_no: '', quantity: 0, purchase_date: '', expiry_date: '', purchase_price: 0, supplier: '' }); dialog.value = true }
function openEdit(row: Batch) { Object.assign(form, row); dialog.value = true }
async function onSave() { saving.value = true; try { if (form.id) await BatchAPI.update(form.id, form); else await BatchAPI.create(form); ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false } }
async function onDel(row: Batch) { await ElMessageBox.confirm(`确认删除批次「${row.batch_no}」？`, '提示', { type: 'warning' }); await BatchAPI.remove(row.id); ElMessage.success('已删除'); await load() }
async function loadWarnings() { warnings.value = (await BatchAPI.warnings(30)).items; warnDialog.value = true }
onMounted(() => { loadMaterials(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
