<template>
  <div class="auth-page">
    <!-- Left branding panel -->
    <div class="brand-panel">
      <div class="brand-inner">
        <div class="brand-logo">
          <svg viewBox="0 0 48 48" width="56" height="56" aria-hidden="true">
            <rect width="48" height="48" rx="12" fill="rgba(255,255,255,0.14)" />
            <path d="M24 10c5.5 5.5 8.5 9.5 8.5 14A8.5 8.5 0 0 1 15.5 24C15.5 19.5 18.5 15.5 24 10Z" fill="#c8e6c9" />
            <path d="M14 34h20v3H14z" fill="#a5d6a7" />
          </svg>
        </div>
        <h1 class="brand-title">加入农场<br />数据化经营</h1>
        <p class="brand-sub">注册账号，开始记录从地块到采收的全流程，<br />用数据驱动农场的每一季生产。</p>
      </div>
      <div class="brand-decor decor-1"></div>
      <div class="brand-decor decor-2"></div>
    </div>

    <!-- Right form panel -->
    <div class="form-panel">
      <div class="form-card">
        <div class="form-head">
          <h2>创建账号</h2>
          <p>填写以下信息完成注册</p>
        </div>
        <el-form ref="formRef" :model="form" :rules="rules" label-width="0" size="large" @submit.prevent>
          <el-form-item prop="username"><el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" /></el-form-item>
          <el-form-item prop="email"><el-input v-model="form.email" placeholder="邮箱" :prefix-icon="Message" /></el-form-item>
          <el-form-item prop="full_name"><el-input v-model="form.full_name" placeholder="姓名" :prefix-icon="UserFilled" /></el-form-item>
          <el-form-item prop="password"><el-input v-model="form.password" type="password" placeholder="密码 (≥8位)" :prefix-icon="Lock" show-password /></el-form-item>
          <el-button type="primary" :loading="loading" class="submit" @click="onSubmit">注 册</el-button>
        </el-form>
        <div class="foot">已有账号？<router-link to="/login">去登录</router-link></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, Message, UserFilled } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { AuthAPI } from '@/api/modules'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', email: '', full_name: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ required: true, type: 'email' as const, message: '请输入有效邮箱', trigger: 'blur' }],
  password: [{ required: true, min: 8, message: '密码至少8位', trigger: 'blur' }],
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  loading.value = true
  try {
    await AuthAPI.register(form)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { display: flex; height: 100vh; overflow: hidden; }
.brand-panel {
  position: relative; flex: 1 1 44%;
  display: flex; align-items: center;
  background: linear-gradient(135deg, #14402a 0%, #2e7d32 60%, #43a047 100%);
  overflow: hidden;
}
.brand-inner { position: relative; z-index: 2; padding: 0 64px; max-width: 460px; }
.brand-logo { width: 56px; height: 56px; margin-bottom: 28px; }
.brand-title { color: #fff; font-size: 30px; line-height: 1.35; font-weight: 700; margin: 0 0 18px; }
.brand-sub { color: rgba(255,255,255,0.82); font-size: 15px; line-height: 1.8; margin: 0; }
.brand-decor { position: absolute; border-radius: 50%; filter: blur(8px); opacity: 0.5; z-index: 1; }
.decor-1 { width: 280px; height: 280px; background: rgba(174,215,176,0.35); top: -80px; right: -60px; }
.decor-2 { width: 200px; height: 200px; background: rgba(46,125,50,0.5); bottom: -60px; left: -40px; }

.form-panel { flex: 1 1 56%; display: flex; align-items: center; justify-content: center; background: #fff; padding: 24px; }
.form-card { width: 100%; max-width: 380px; }
.form-head h2 { margin: 0 0 8px; font-size: 26px; color: #14402a; font-weight: 700; }
.form-head p { margin: 0 0 28px; color: #8a9a90; font-size: 14px; }
.submit { width: 100%; height: 44px; font-size: 16px; letter-spacing: 2px; border-radius: 8px; }
.foot { text-align: center; margin-top: 18px; font-size: 14px; color: #8a9a90; }
.foot a { color: var(--el-color-primary); text-decoration: none; font-weight: 500; }

@media (max-width: 860px) {
  .auth-page { flex-direction: column; }
  .brand-panel { flex: none; min-height: 180px; padding: 32px; }
  .brand-inner { padding: 0; }
  .brand-title { font-size: 22px; }
  .brand-sub { display: none; }
  .form-panel { flex: 1; }
}
</style>
