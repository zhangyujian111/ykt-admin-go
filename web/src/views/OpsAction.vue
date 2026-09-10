<template><div class="page">
  <el-tabs v-model="tab">
    <el-tab-pane label="动作总览" name="summary">
      <el-row :gutter="16">
        <el-col :span="8"><el-card><div class="metric-label">今日动作</div><div class="metric-value">{{ summary.todayActions ?? '—' }}</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">成功率</div><div class="metric-value">{{ summary.successRate ?? '—' }}%</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">失败</div><div class="metric-value text-red">{{ summary.failures ?? '—' }}</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="失败流水" name="failures">
      <el-table :data="failures" border size="small">
        <el-table-column prop="deviceId" label="设备" /><el-table-column prop="actionName" label="动作" /><el-table-column prop="errorMsg" label="错误" /><el-table-column prop="createdAt" label="时间" width="180" />
      </el-table>
    </el-tab-pane>
  </el-tabs>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const tab = ref('summary')
const summary = ref<any>({})
const failures = ref<any[]>([])
onMounted(async () => {
  try {
    const [s, f] = await Promise.all([api.get('/ops/action/summary'), api.get('/ops/action/failures')])
    summary.value = s.data?.data || {}
    failures.value = (f.data?.data || []).slice(0, 50)
  } catch (err) { console.error(err) }
})
</script>
<style scoped>.metric-label{font-size:13px;color:#94a3b8;margin-bottom:8px}.metric-value{font-size:24px;font-weight:600;color:#1f3a5e}.text-red{color:#ef4444}.page{padding:4px}</style>