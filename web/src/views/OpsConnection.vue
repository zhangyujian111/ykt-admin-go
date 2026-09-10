<template><div class="page">
  <el-tabs v-model="tab">
    <el-tab-pane label="连接总览" name="summary">
      <el-row :gutter="16">
        <el-col :span="6"><el-card><div class="metric-label">今日连接</div><div class="metric-value">{{ summary.todayConnects ?? '—' }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">今日断开</div><div class="metric-value">{{ summary.todayDisconnects ?? '—' }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">当前在线</div><div class="metric-value">{{ summary.currentOnline ?? '—' }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">快速断连</div><div class="metric-value text-red">{{ summary.quickDisconnects ?? '—' }}</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="异常设备" name="abnormal">
      <el-table :data="abnormal" border size="small">
        <el-table-column prop="deviceId" label="设备 ID" />
        <el-table-column prop="deviceName" label="名称" />
        <el-table-column prop="reason" label="异常" />
        <el-table-column prop="count" label="次数" width="100" />
      </el-table>
    </el-tab-pane>
    <el-tab-pane label="事件流水" name="events">
      <el-table :data="events" border size="small">
        <el-table-column prop="eventId" label="事件 ID" width="120" />
        <el-table-column prop="deviceId" label="设备 ID" />
        <el-table-column prop="connectTime" label="连接时间" width="180" />
        <el-table-column prop="disconnectTime" label="断开时间" width="180" />
        <el-table-column prop="durationSec" label="时长(秒)" width="120" />
        <el-table-column prop="reason" label="原因" />
      </el-table>
    </el-tab-pane>
  </el-tabs>
</div></template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const tab = ref('summary')
const summary = ref<any>({})
const abnormal = ref<any[]>([])
const events = ref<any[]>([])
onMounted(async () => {
  try {
    const [s, a, e] = await Promise.all([api.get('/ops/connection/summary'), api.get('/ops/connection/abnormal-devices'), api.get('/ops/connection/events')])
    summary.value = s.data?.data || {}
    abnormal.value = a.data?.data || []
    events.value = (e.data?.data || []).slice(0, 50)
  } catch (err) { console.error(err) }
})
</script>
<style scoped>.metric-label{font-size:13px;color:#94a3b8;margin-bottom:8px}.metric-value{font-size:24px;font-weight:600;color:#1f3a5e}.text-red{color:#ef4444}.page{padding:4px}</style>