<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button type="success" @click="openCreate">新建角色</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="code" label="编码" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="description" label="描述" />
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="primary" @click="openAssign(row)">分配权限</el-button>
          <el-button size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="dialog" :title="form.id ? '编辑角色' : '新建角色'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="assignDialog" title="分配权限" width="520px">
      <el-tree ref="treeRef" :data="permTree" :props="{ label: 'name', children: 'children' }" node-key="id" show-checkbox default-expand-all />
      <template #footer><el-button @click="assignDialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onAssign">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/client'
import type { Role, Permission } from '@/api/modules'

const rows = ref<Role[]>([]); const loading = ref(false)
const allPerms = ref<Permission[]>([])
const dialog = ref(false); const saving = ref(false)
const form = reactive<any>({ code: '', name: '', description: '' })
const assignDialog = ref(false); const currentRole = ref<Role | null>(null); const treeRef = ref()

async function load() { loading.value = true; try { const res = await api.get<{ items: Role[] }>('/roles', { params: { page: 1, page_size: 100 } }); rows.value = res.items } finally { loading.value = false } }
async function loadPerms() { allPerms.value = (await api.get<{ items: Permission[] }>('/permissions')).items }

const permTree = computed(() => {
  const groups: Record<string, any> = {}
  for (const p of allPerms.value) {
    const g = groups[p.resource] || (groups[p.resource] = { id: `r-${p.resource}`, name: p.resource, children: [] })
    g.children.push({ id: p.id, name: `${p.action} (${p.code})` })
  }
  return Object.values(groups)
})

function openCreate() { Object.assign(form, { id: undefined, code: '', name: '', description: '' }); dialog.value = true }
function openEdit(row: Role) { Object.assign(form, row); dialog.value = true }
async function onSave() {
  saving.value = true
  try {
    if (form.id) await api.put(`/roles/${form.id}`, form); else await api.post('/roles', form)
    ElMessage.success('保存成功'); dialog.value = false; await load()
  } finally { saving.value = false }
}
async function openAssign(row: Role) {
  currentRole.value = row; assignDialog.value = true
  const detail = await api.get<{ permissions: Permission[] }>(`/roles/${row.id}`)
  setTimeout(() => treeRef.value?.setCheckedKeys(detail.permissions.map((p) => p.id)), 50)
}
async function onAssign() {
  saving.value = true
  try { const ids = treeRef.value.getCheckedKeys(); await api.put(`/roles/${currentRole.value!.id}/permissions`, { permission_ids: ids }); ElMessage.success('权限已分配'); assignDialog.value = false } finally { saving.value = false }
}
async function onDel(row: Role) { await ElMessageBox.confirm(`确认删除角色「${row.name}」？`, '提示', { type: 'warning' }); await api.delete(`/roles/${row.id}`); ElMessage.success('已删除'); await load() }
onMounted(() => { load(); loadPerms() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
