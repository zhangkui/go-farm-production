<template>
  <div>
    <div class="toolbar">
      <el-input v-model="q.search" placeholder="搜索用户名/邮箱" clearable style="width:240px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
      <el-button type="success" @click="openCreate">新建用户</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="full_name" label="姓名" />
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="statusTag(row.status,'generic')">{{ statusText(row.status,'generic') }}</el-tag></template></el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="170" />
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="primary" @click="openAssign(row)">分配角色</el-button>
          <el-button size="small" type="warning" @click="openReset(row)">重置密码</el-button>
          <el-button size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑用户' : '新建用户'" width="480px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="用户名"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="姓名"><el-input v-model="form.full_name" /></el-form-item>
        <el-form-item v-if="!form.id" label="密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :label="1">活跃</el-radio><el-radio :label="0">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="assignDialog" title="分配角色" width="480px">
      <el-select v-model="selectedRoles" multiple style="width:100%">
        <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
      </el-select>
      <template #footer><el-button @click="assignDialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onAssign">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="resetDialog" title="重置密码" width="400px">
      <el-input v-model="newPwd" placeholder="新密码 (≥8位)" show-password />
      <template #footer><el-button @click="resetDialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onReset">重置</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MeAPI, FarmAPI, VarietyAPI, SeasonAPI, FieldAPI, type User, type Role } from '@/api/modules'
import { statusText, statusTag } from '@/utils/constants'

// User management uses the /users endpoints which require user:manage.
const rows = ref<User[]>([]); const total = ref(0); const loading = ref(false)
const roles = ref<Role[]>([])
const q = reactive({ page: 1, page_size: 20, search: '' })
const dialog = ref(false); const saving = ref(false)
const form = reactive<any>({ username: '', email: '', full_name: '', password: '', status: 1 })
const assignDialog = ref(false); const selectedRoles = ref<number[]>([]); const currentUser = ref<User | null>(null)
const resetDialog = ref(false); const newPwd = ref(''); const resetUser = ref<User | null>(null)

async function loadRoles() { roles.value = (await (await import('@/api/client')).api.get<Role[]>('/roles')) }
async function load() { loading.value = true; try { const res = await (await import('@/api/client')).api.get<{ items: User[]; total: number }>('/users', { params: q }); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, username: '', email: '', full_name: '', password: '', status: 1 }); dialog.value = true }
function openEdit(row: User) { Object.assign(form, row, { password: '' }); dialog.value = true }
async function onSave() {
  saving.value = true
  try {
    const { api } = await import('@/api/client')
    if (form.id) await api.put(`/users/${form.id}`, form); else await api.post('/users', form)
    ElMessage.success('保存成功'); dialog.value = false; await load()
  } finally { saving.value = false }
}
function openAssign(row: User) { currentUser.value = row; selectedRoles.value = []; assignDialog.value = true }
async function onAssign() {
  saving.value = true
  try { const { api } = await import('@/api/client'); await api.put(`/users/${currentUser.value!.id}/roles`, { role_ids: selectedRoles.value }); ElMessage.success('角色已分配'); assignDialog.value = false } finally { saving.value = false }
}
function openReset(row: User) { resetUser.value = row; newPwd.value = ''; resetDialog.value = true }
async function onReset() { saving.value = true; try { const { api } = await import('@/api/client'); await api.post(`/users/${resetUser.value!.id}/password`, { new_password: newPwd.value }); ElMessage.success('密码已重置'); resetDialog.value = false } finally { saving.value = false } }
async function onDel(row: User) { await ElMessageBox.confirm(`确认删除用户「${row.username}」？`, '提示', { type: 'warning' }); const { api } = await import('@/api/client'); await api.delete(`/users/${row.id}`); ElMessage.success('已删除'); await load() }
onMounted(() => { loadRoles(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
