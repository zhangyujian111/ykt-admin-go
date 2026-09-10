<template>
  <el-container class="layout-root">
    <el-aside width="220px" class="aside">
      <div class="logo">
        <el-icon class="logo-icon"><DataBoard /></el-icon>
        <span>Ops Console</span>
      </div>
      <el-menu :default-active="$route.path" router class="menu" :collapse="false">
        <template v-for="r in routes" :key="r.path">
          <el-menu-item v-if="!r.meta?.hidden" :index="'/' + r.path">
            <el-icon><component :is="iconFor(r.meta?.icon)" /></el-icon>
            <span>{{ r.meta?.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="title">{{ $route.meta?.title || '' }}</div>
        <div class="right">
          <el-tag v-if="auth.user" type="info" size="small">{{ auth.user }}</el-tag>
          <el-button link type="primary" @click="logout">退出</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  Monitor, DataAnalysis, Link, ChatDotRound, Aim, Timer, CircleClose, Cpu,
  List, AlarmClock, Bell, OfficeBuilding, User, UserFilled, Menu, Connection,
  Notebook, Document, Key, DataBoard
} from '@element-plus/icons-vue'

const auth = useAuthStore()
const router = useRouter()
const routes = ref<any[]>([])

function iconFor(name?: string) {
  const map: Record<string, any> = {
    Monitor, DataAnalysis, Link, ChatDotRound, Aim, Timer, CircleClose, Cpu,
    List, AlarmClock, Bell, OfficeBuilding, User, UserFilled, Menu, Connection,
    Notebook, Document, Key
  }
  return map[name || ''] || DataBoard
}

onMounted(async () => {
  auth.load()
  if (!auth.user) {
    try {
      const me = await fetch('/api/auth/oauth/me', { headers: { Authorization: 'Bearer ' + auth.token } }).then(r => r.json())
      if (me?.code === 0) auth.setUser(me.data?.username || me.data?.nickname || 'admin')
    } catch { auth.setUser('admin') }
  }
  routes.value = (router.options.routes[1].children || []).filter((c: any) => c.path)
})

function logout() {
  auth.clear()
  router.push('/login')
}
</script>

<style scoped>
.layout-root { height: 100vh; }
.aside { background: #001529; color: #fff; overflow: auto; }
.logo { padding: 18px 16px; display: flex; align-items: center; gap: 8px; font-weight: 600; font-size: 16px; border-bottom: 1px solid #1f3a5e; }
.logo-icon { font-size: 20px; color: #38bdf8; }
.menu { background: #001529; border-right: none; }
.header { background: #fff; display: flex; justify-content: space-between; align-items: center; padding: 0 20px; box-shadow: 0 1px 4px rgba(0,21,41,.08); }
.title { font-size: 16px; font-weight: 600; }
.right { display: flex; align-items: center; gap: 12px; }
.main { background: #f5f7fa; padding: 16px; }
:deep(.el-menu-item) { color: #a6adb4; }
:deep(.el-menu-item.is-active) { color: #fff; background: #1f3a5e !important; }
:deep(.el-menu-item:hover) { background: #1f3a5e !important; }
</style>