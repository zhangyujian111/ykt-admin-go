<template><div class="page">
  <el-tabs v-model="tab">
    <el-tab-pane label="对话总览" name="summary">
      <el-row :gutter="16">
        <el-col :span="6"><el-card><div class="metric-label">今日对话</div><div class="metric-value">{{ summary.todayDialogs ?? '—' }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">成功率</div><div class="metric-value">{{ summary.successRate ?? '—' }}%</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">平均延迟</div><div class="metric-value">{{ summary.avgLatency ?? '—' }}ms</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">失败对话</div><div class="metric-value text-red">{{ summary.failures ?? '—' }}</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="异常设备" name="abnormal">
      <el-table :data="abnormal" border size="small">
        <el-table-column prop="deviceId" label="设备 ID" /><el-table-column prop="reason" label="原因" /><el-table-column prop="count" label="次数" width="100" />
      </el-table>
    </el-tab-pane>
    <el-tab-pane label="失败流水" name="failures">
      <el-table :data="failures" border size="small">
        <el-table-column prop="deviceId" label="设备" /><el-table-column prop="failureType" label="类型" /><el-table-column prop="errorMsg" label="错误" /><el-table-column prop="createdAt" label="时间" width="180" />
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
const failures = ref<any[]>([])
onMounted(async () => {
  try {
    const [s, a, f] = await Promise.all([api.get('/ops/dialog/summary'), api.get('/ops/dialog/abnormal-devices'), api.get('/ops/dialog/failures')])
    summary.value = s.data?.data || {}
    abnormal.value = a.data?.data || []
    failures.value = (f.data?.data || []).slice(0, 50)
  } catch (err) { console.error(err) }
})
</script>
<style scoped>.metric-label{font-size:13px;color:#94a3b8;margin-bottom:8px}.metric-value{font-size:24px;font-weight:600;color:#1f3a5e}.text-red{color:#ef4444}.page{padding:4px}</style>