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
            <path d="M18 31l3-7M26 31l4-9" stroke="#a5d6a7" stroke-width="1.6" stroke-linecap="round" />
          </svg>
        </div>
        <h1 class="brand-title">农场种植与<br />投入品管理系统</h1>
        <p class="brand-sub">让每一寸土地、每一份投入、每一粒产出<br />都清晰可见、精准可控</p>
        <ul class="brand-features">
          <li><el-icon><CircleCheck /></el-icon> 种植计划全流程追溯</li>
          <li><el-icon><CircleCheck /></el-icon> 投入品批次与库存精细化管理</li>
          <li><el-icon><CircleCheck /></el-icon> 采收入库与成本精准核算</li>
        </ul>
      </div>
      <div class="brand-decor decor-1"></div>
      <div class="brand-decor decor-2"></div>
    </div>

    <!-- Right form panel -->
    <div class="form-panel">
      <div class="form-card">
        <div class="form-head">
          <h2>欢迎回来</h2>
          <p>登录以继续管理您的农场</p>
        </div>
        <el-form ref="formRef" :model="form" :rules="rules" label-width="0" size="large" @submit.prevent>
          <el-form-item prop="username">
            <el-input v-model="form.username" placeholder="用户名 / 邮箱" :prefix-icon="User" />
          </el-form-item>
          <el-form-item prop="password">
            <el-input v-model="form.password" type="password" placeholder="密码" :prefix-icon="Lock" show-password @keyup.enter="onSubmit" />
          </el-form-item>
          <el-button type="primary" :loading="loading" class="submit" @click="onSubmit">登 录</el-button>
        </el-form>
        <div class="foot">
          还没有账号？<router-link to="/register">立即注册</router-link>
        </div>
        <div class="hint">默认管理员 admin / Admin123!</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    ElMessage.success('登录成功')
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { display: flex; height: 100vh; overflow: hidden; }

/* ---- Brand panel ---- */
.brand-panel {
  position: relative; flex: 1 1 56%;
  display: flex; align-items: center;
  background: linear-gradient(135deg, #14402a 0%, #2e7d32 60%, #43a047 100%);
  overflow: hidden;
}
.brand-inner { position: relative; z-index: 2; padding: 0 64px; max-width: 520px; }
.brand-logo { width: 56px; height: 56px; margin-bottom: 28px; }
.brand-title { color: #fff; font-size: 34px; line-height: 1.3; font-weight: 700; margin: 0 0 18px; }
.brand-sub { color: rgba(255,255,255,0.82); font-size: 15px; line-height: 1.8; margin: 0 0 36px; }
.brand-features { list-style: none; margin: 0; padding: 0; }
.brand-features li {
  display: flex; align-items: center; gap: 10px;
  color: rgba(255,255,255,0.9); font-size: 15px; margin-bottom: 14px;
}
.brand-features .el-icon { color: #c8e6c9; font-size: 18px; }

/* Floating decorative blobs. */
.brand-decor { position: absolute; border-radius: 50%; filter: blur(8px); opacity: 0.5; z-index: 1; }
.decor-1 { width: 280px; height: 280px; background: rgba(174,215,176,0.35); top: -80px; right: -60px; }
.decor-2 { width: 200px; height: 200px; background: rgba(46,125,50,0.5); bottom: -60px; left: -40px; }

/* ---- Form panel ---- */
.form-panel { flex: 1 1 44%; display: flex; align-items: center; justify-content: center; background: #fff; padding: 24px; }
.form-card { width: 100%; max-width: 360px; }
.form-head h2 { margin: 0 0 8px; font-size: 26px; color: #14402a; font-weight: 700; }
.form-head p { margin: 0 0 32px; color: #8a9a90; font-size: 14px; }
.submit { width: 100%; height: 44px; font-size: 16px; letter-spacing: 2px; border-radius: 8px; }
.foot { text-align: center; margin-top: 20px; font-size: 14px; color: #8a9a90; }
.foot a { color: var(--el-color-primary); text-decoration: none; font-weight: 500; }
.hint { text-align: center; margin-top: 14px; font-size: 12px; color: #b0bbb3; }

/* ---- Responsive: stack on narrow screens ---- */
@media (max-width: 860px) {
  .auth-page { flex-direction: column; }
  .brand-panel { flex: none; min-height: 220px; padding: 32px; }
  .brand-inner { padding: 0; }
  .brand-title { font-size: 24px; }
  .brand-sub, .brand-features { display: none; }
  .form-panel { flex: 1; }
}
</style>
