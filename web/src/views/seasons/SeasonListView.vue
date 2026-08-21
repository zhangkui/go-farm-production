<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button v-permission="'season:manage'" type="success" @click="openCreate">新建季次</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="code" label="编码" width="120" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="start_date" label="开始日期" width="130" />
      <el-table-column prop="end_date" label="结束日期" width="130" />
      <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusTag(row.status,'plan')">{{ statusText(row.status,'plan') }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-permission="'season:manage'" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="'season:manage'" size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑季次' : '新建季次'" width="460px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="开始日期"><el-date-picker v-model="form.start_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="结束日期"><el-date-picker v-model="form.end_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option :value="1" label="计划中" /><el-option :value="2" label="进行中" /><el-option :value="3" label="已完成" /><el-option :value="0" label="已取消" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { SeasonAPI, type Season } from '@/api/modules'
import { statusText, statusTag } from '@/utils/constants'

const rows = ref<Season[]>([]); const total = ref(0); const loading = ref(false)
const q = reactive({ page: 1, page_size: 20 })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<Season>>({ code: '', name: '', start_date: '', end_date: '', status: 1 })

async function load() { loading.value = true; try { const res = await SeasonAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, code: '', name: '', start_date: '', end_date: '', status: 1 }); dialog.value = true }
function openEdit(row: Season) { Object.assign(form, row); dialog.value = true }
async function onSave() { saving.value = true; try { if (form.id) await SeasonAPI.update(form.id, form); else await SeasonAPI.create(form); ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false } }
async function onDel(row: Season) { await ElMessageBox.confirm(`确认删除季次「${row.name}」？`, '提示', { type: 'warning' }); await SeasonAPI.remove(row.id); ElMessage.success('已删除'); await load() }
load()
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
