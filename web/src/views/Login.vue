<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('admin')
const password = ref('admin123')
const loading = ref(false)

async function login() {
  loading.value = true
  try {
    const r = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value })
    }).then(r => r.json())
    if (r.code === 0 && r.data?.token) {
      auth.setToken(r.data.token)
      auth.setUser(r.data.user?.username || r.data.username || 'admin')
      ElMessage.success('登录成功')
      router.push('/dashboard')
    } else {
      ElMessage.error(r.message || '登录失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="logo">
        <el-icon size="32"><DataBoard /></el-icon>
        <h2>Ops Console</h2>
        <p>平台运营控制台 · ykt-admin-go</p>
      </div>
      <el-form @submit.prevent="login" label-width="80px">
        <el-form-item label="账号">
          <el-input v-model="username" placeholder="admin" autofocus />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="password" type="password" show-password @keyup.enter="login" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="login" style="width:100%">登录</el-button>
        </el-form-item>
      </el-form>
      <p class="hint">默认账号 admin / admin123</p>
    </div>
  </div>
</template>

<style scoped>
.login-page { height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #001529 0%, #1f3a5e 100%); }
.login-card { background: #fff; padding: 32px 40px; border-radius: 8px; width: 360px; box-shadow: 0 8px 24px rgba(0,0,0,.15); }
.logo { text-align: center; margin-bottom: 24px; color: #1f3a5e; }
.logo h2 { margin: 8px 0 4px; }
.logo p { margin: 0; color: #94a3b8; font-size: 12px; }
.hint { text-align: center; color: #94a3b8; font-size: 12px; margin: 0; }
</style>