import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/layouts/Layout.vue'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    { path: '/login', component: () => import('@/views/Login.vue') },
    {
      path: '/',
      component: Layout,
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', component: () => import('@/views/Dashboard.vue'), meta: { title: '总览', icon: 'Monitor' } },
        { path: 'ops/dashboard', component: () => import('@/views/OpsDashboard.vue'), meta: { title: '运维大盘', icon: 'DataAnalysis' } },
        { path: 'ops/connection', component: () => import('@/views/OpsConnection.vue'), meta: { title: '连接分析', icon: 'Link' } },
        { path: 'ops/dialog', component: () => import('@/views/OpsDialog.vue'), meta: { title: '对话分析', icon: 'ChatDotRound' } },
        { path: 'ops/action', component: () => import('@/views/OpsAction.vue'), meta: { title: '动作分析', icon: 'Aim' } },
        { path: 'ops/latency', component: () => import('@/views/OpsLatency.vue'), meta: { title: '延迟分析', icon: 'Timer' } },
        { path: 'ops/offline', component: () => import('@/views/OpsOffline.vue'), meta: { title: '离线设备', icon: 'CircleClose' } },
        { path: 'ops/ai-metrics', component: () => import('@/views/AIMetrics.vue'), meta: { title: 'AI 监控', icon: 'Cpu' } },
        { path: 'devices', component: () => import('@/views/Devices.vue'), meta: { title: '设备列表', icon: 'List' } },
        { path: 'devices/:id', component: () => import('@/views/DeviceDetail.vue'), meta: { title: '设备详情', hidden: true } },
        { path: 'alerts/rules', component: () => import('@/views/AlertRules.vue'), meta: { title: '告警规则', icon: 'AlarmClock' } },
        { path: 'alerts/history', component: () => import('@/views/AlertHistory.vue'), meta: { title: '告警历史', icon: 'Bell' } },
        { path: 'tenants', component: () => import('@/views/Tenants.vue'), meta: { title: '租户管理', icon: 'OfficeBuilding' } },
        { path: 'system/user', component: () => import('@/views/SysUser.vue'), meta: { title: '用户管理', icon: 'User' } },
        { path: 'system/role', component: () => import('@/views/SysRole.vue'), meta: { title: '角色管理', icon: 'UserFilled' } },
        { path: 'system/menu', component: () => import('@/views/SysMenu.vue'), meta: { title: '菜单管理', icon: 'Menu' } },
        { path: 'system/dept', component: () => import('@/views/SysDept.vue'), meta: { title: '部门管理', icon: 'Connection' } },
        { path: 'system/dict', component: () => import('@/views/SysDict.vue'), meta: { title: '字典管理', icon: 'Notebook' } },
        { path: 'system/log/oper', component: () => import('@/views/SysOperLog.vue'), meta: { title: '操作日志', icon: 'Document' } },
        { path: 'system/log/login', component: () => import('@/views/SysLoginLog.vue'), meta: { title: '登录日志', icon: 'Key' } }
      ]
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('admin_token')
  if (to.path === '/login') return next()
  if (!token) return next('/login')
  next()
})

export default router