<template>
  <div>
    <div class="toolbar">
      <el-input v-model="q.search" placeholder="搜索农场名称/位置" clearable style="width:260px" @clear="load" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button v-permission="'farm:manage'" type="success" @click="openCreate">新建农场</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="location" label="位置" />
      <el-table-column prop="total_area" label="总面积(亩)" width="120" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }"><el-tag :type="statusTag(row.status,'generic')">{{ statusText(row.status,'generic') }}</el-tag></template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-permission="'farm:manage'" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="'farm:manage'" size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />

    <el-dialog v-model="dialog" :title="form.id ? '编辑农场' : '新建农场'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="位置"><el-input v-model="form.location" /></el-form-item>
        <el-form-item label="总面积(亩)"><el-input-number v-model="form.total_area" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status"><el-radio :label="1">活跃</el-radio><el-radio :label="0">停用</el-radio></el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FarmAPI, type Farm } from '@/api/modules'
import { statusText, statusTag } from '@/utils/constants'

const rows = ref<Farm[]>([])
const total = ref(0)
const loading = ref(false)
const q = reactive({ page: 1, page_size: 20, search: '' })
const dialog = ref(false)
const saving = ref(false)
const form = reactive<Partial<Farm>>({ name: '', location: '', total_area: 0, description: '', status: 1 })

async function load() {
  loading.value = true
  try {
    const res = await FarmAPI.list(q)
    rows.value = res.items
    total.value = res.total
  } finally { loading.value = false }
}
function openCreate() { Object.assign(form, { id: undefined, name: '', location: '', total_area: 0, description: '', status: 1 }); dialog.value = true }
function openEdit(row: Farm) { Object.assign(form, row); dialog.value = true }
async function onSave() {
  saving.value = true
  try {
    if (form.id) await FarmAPI.update(form.id, form); else await FarmAPI.create(form)
    ElMessage.success('保存成功'); dialog.value = false; await load()
  } finally { saving.value = false }
}
async function onDel(row: Farm) {
  await ElMessageBox.confirm(`确认删除农场「${row.name}」？`, '提示', { type: 'warning' })
  await FarmAPI.remove(row.id); ElMessage.success('已删除'); await load()
}
load()
</script>

<style scoped>
.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }
</style>
