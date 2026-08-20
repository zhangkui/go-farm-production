<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.farm_id" placeholder="全部农场" clearable filterable style="width:200px" @change="load">
        <el-option v-for="f in farms" :key="f.id" :label="f.name" :value="f.id" />
      </el-select>
      <el-input v-model="q.irrigation_zone" placeholder="灌溉区" clearable style="width:160px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button v-permission="'field:manage'" type="success" @click="openCreate">新建地块</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="code" label="编码" width="110" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="farm_name" label="所属农场" />
      <el-table-column prop="area" label="面积(亩)" width="100" />
      <el-table-column prop="soil_type" label="土壤类型" width="100" />
      <el-table-column prop="irrigation_zone" label="灌溉区" width="100" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }"><el-tag :type="statusTag(row.status,'field')">{{ statusText(row.status,'field') }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-permission="'field:manage'" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="'field:manage'" size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />

    <el-dialog v-model="dialog" :title="form.id ? '编辑地块' : '新建地块'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="所属农场"><el-select v-model="form.farm_id" filterable><el-option v-for="f in farms" :key="f.id" :label="f.name" :value="f.id" /></el-select></el-form-item>
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="面积(亩)"><el-input-number v-model="form.area" :min="0.01" :precision="2" placeholder="请输入大于0的面积" /></el-form-item>
        <el-form-item label="土壤类型"><el-input v-model="form.soil_type" /></el-form-item>
        <el-form-item label="灌溉区"><el-input v-model="form.irrigation_zone" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option :value="1" label="可用" /><el-option :value="2" label="种植中" /><el-option :value="3" label="休耕" /><el-option :value="0" label="停用" /></el-select></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FieldAPI, FarmAPI, type Field, type Farm } from '@/api/modules'
import { statusText, statusTag } from '@/utils/constants'

const rows = ref<Field[]>([]); const total = ref(0); const loading = ref(false)
const farms = ref<Farm[]>([])
const q = reactive({ page: 1, page_size: 20, farm_id: undefined as number | undefined, irrigation_zone: '' })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<Field>>({ farm_id: undefined, code: '', name: '', area: 0, soil_type: '', irrigation_zone: '', status: 1, remark: '' })

async function loadFarms() { farms.value = (await FarmAPI.list({ page: 1, page_size: 200 })).items }
async function load() {
  loading.value = true
  try { const res = await FieldAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false }
}
function openCreate() { Object.assign(form, { id: undefined, farm_id: farms.value[0]?.id, code: '', name: '', area: undefined, soil_type: '', irrigation_zone: '', status: 1, remark: '' }); dialog.value = true }
function openEdit(row: Field) { Object.assign(form, row); dialog.value = true }
async function onSave() {
  saving.value = true
  try { if (form.id) await FieldAPI.update(form.id, form); else await FieldAPI.create(form); ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false }
}
async function onDel(row: Field) { await ElMessageBox.confirm(`确认删除地块「${row.name}」？`, '提示', { type: 'warning' }); await FieldAPI.remove(row.id); ElMessage.success('已删除'); await load() }
onMounted(() => { loadFarms(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
