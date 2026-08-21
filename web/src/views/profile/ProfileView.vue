<template>
  <el-card style="max-width:640px">
    <template #header>个人资料</template>
    <el-descriptions :column="2" border>
      <el-descriptions-item label="用户名">{{ auth.user?.username }}</el-descriptions-item>
      <el-descriptions-item label="姓名">{{ auth.user?.full_name }}</el-descriptions-item>
      <el-descriptions-item label="邮箱">{{ auth.user?.email }}</el-descriptions-item>
      <el-descriptions-item label="状态"><el-tag :type="auth.user?.status===1?'success':'info'">{{ auth.user?.status===1?'活跃':'停用' }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="角色" :span="2">{{ auth.roles.map(r=>r.name).join('、') || '-' }}</el-descriptions-item>
    </el-descriptions>
    <el-divider />
    <h3>修改密码</h3>
    <el-form :model="pwd" label-width="100px" style="max-width:420px">
      <el-form-item label="当前密码"><el-input v-model="pwd.old_password" type="password" show-password /></el-form-item>
      <el-form-item label="新密码"><el-input v-model="pwd.new_password" type="password" show-password /></el-form-item>
      <el-form-item><el-button type="primary" :loading="saving" @click="onChange">提交</el-button></el-form-item>
    </el-form>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { MeAPI } from '@/api/modules'

const auth = useAuthStore()
const saving = ref(false)
const pwd = reactive({ old_password: '', new_password: '' })

async function onChange() {
  if (pwd.new_password.length < 8) { ElMessage.warning('新密码至少8位'); return }
  saving.value = true
  try { await MeAPI.changePassword(pwd.old_password, pwd.new_password); ElMessage.success('密码已修改，请重新登录'); await auth.logout(); location.href = '/login' } finally { saving.value = false }
}
</script>
