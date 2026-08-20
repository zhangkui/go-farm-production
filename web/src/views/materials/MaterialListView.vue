<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.category" placeholder="全部品类" clearable style="width:150px" @change="load">
        <el-option :value="1" label="肥料" /><el-option :value="2" label="农药" /><el-option :value="3" label="种子" /><el-option :value="4" label="燃料" /><el-option :value="5" label="其他" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button v-permission="'material:manage'" type="success" @click="openCreate">新建物料</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="code" label="编码" width="120" />
      <el-table-column prop="name" label="名称" />
      <el-table-column label="品类" width="90"><template #default="{ row }">{{ materialCategoryText[row.category] }}</template></el-table-column>
      <el-table-column prop="unit" label="单位" width="80" />
      <el-table-column prop="unit_price" label="单价" width="100" />
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="statusTag(row.status,'generic')">{{ statusText(row.status,'generic') }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-permission="'material:manage'" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="'material:manage'" size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑物料' : '新建物料'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="品类"><el-select v-model="form.category"><el-option :value="1" label="肥料" /><el-option :value="2" label="农药" /><el-option :value="3" label="种子" /><el-option :value="4" label="燃料" /><el-option :value="5" label="其他" /></el-select></el-form-item>
        <el-form-item label="单位"><el-input v-model="form.unit" /></el-form-item>
        <el-form-item label="单价"><el-input-number v-model="form.unit_price" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :label="1">活跃</el-radio><el-radio :label="0">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MaterialAPI, type Material } from '@/api/modules'
import { statusText, statusTag, materialCategoryText } from '@/utils/constants'

const rows = ref<Material[]>([]); const total = ref(0); const loading = ref(false)
const q = reactive({ page: 1, page_size: 20, category: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<Material>>({ code: '', name: '', category: 5, unit: 'kg', unit_price: 0, description: '', status: 1 })

async function load() { loading.value = true; try { const res = await MaterialAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, code: '', name: '', category: 5, unit: 'kg', unit_price: 0, description: '', status: 1 }); dialog.value = true }
function openEdit(row: Material) { Object.assign(form, row); dialog.value = true }
async function onSave() { saving.value = true; try { if (form.id) await MaterialAPI.update(form.id, form); else await MaterialAPI.create(form); ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false } }
async function onDel(row: Material) { await ElMessageBox.confirm(`确认删除物料「${row.name}」？`, '提示', { type: 'warning' }); await MaterialAPI.remove(row.id); ElMessage.success('已删除'); await load() }
load()
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
