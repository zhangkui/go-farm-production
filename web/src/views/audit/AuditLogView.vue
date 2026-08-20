<template>
  <div>
    <div class="toolbar">
      <el-input v-model="q.user_id" placeholder="用户ID" clearable style="width:120px" @keyup.enter="load" />
      <el-input v-model="q.action" placeholder="操作 login/create/..." clearable style="width:160px" @keyup.enter="load" />
      <el-input v-model="q.resource_type" placeholder="资源类型" clearable style="width:160px" @keyup.enter="load" />
      <el-button type="primary" @click="load">查询</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="用户" width="120" />
      <el-table-column prop="action" label="操作" width="140" />
      <el-table-column prop="resource_type" label="资源类型" width="120" />
      <el-table-column prop="resource_id" label="资源ID" width="100" />
      <el-table-column prop="ip_address" label="IP" width="140" />
      <el-table-column prop="created_at" label="时间" width="170" />
      <el-table-column label="详情" show-overflow-tooltip>
        <template #default="{ row }"><pre class="detail">{{ formatDetails(row.details) }}</pre></template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { AuditAPI, type AuditLog } from '@/api/modules'

const rows = ref<AuditLog[]>([]); const total = ref(0); const loading = ref(false)
const q = reactive({ page: 1, page_size: 20, user_id: '', action: '', resource_type: '' })

async function load() {
  loading.value = true
  try { const res = await AuditAPI.list(q as any); rows.value = res.items; total.value = res.total } finally { loading.value = false }
}
function formatDetails(s: string): string {
  if (!s) return ''
  try { return JSON.stringify(JSON.parse(s), null, 2) } catch { return s }
}
load()
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }.detail { margin: 0; max-height: 80px; overflow: auto; font-size: 12px; }</style>
