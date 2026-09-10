<template>
  <div class="page">
    <el-row :gutter="16">
      <el-col :span="6"><el-card><div class="metric-label">今日连接事件</div><div class="metric-value">{{ summary.todayConnects ?? '—' }}</div></el-card></el-col>
      <el-col :span="6"><el-card><div class="metric-label">今日对话</div><div class="metric-value">{{ summary.todayDialogs ?? '—' }}</div></el-card></el-col>
      <el-col :span="6"><el-card><div class="metric-label">告警中</div><div class="metric-value text-red">{{ summary.activeAlerts ?? '—' }}</div></el-card></el-col>
      <el-col :span="6"><el-card><div class="metric-label">问题设备</div><div class="metric-value text-orange">{{ summary.problemDevices ?? '—' }}</div></el-card></el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="12"><el-card><template #header><span>告警趋势（近 7 天）</span></template><div ref="alertChart" style="height:240px"></div></el-card></el-col>
      <el-col :span="12"><el-card><template #header><span>设备活跃趋势</span></template><div ref="devChart" style="height:240px"></div></el-card></el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import * as echarts from 'echarts'
import { api } from '@/api'

const summary = reactive<any>({})
const alertChart = ref<HTMLDivElement>()
const devChart = ref<HTMLDivElement>()

onMounted(async () => {
  try {
    const r = await api.get('/dashboard/overview')
    Object.assign(summary, r.data?.data || {})
  } catch (e) { console.error(e) }
  try {
    const [at, dt] = await Promise.all([api.get('/dashboard/alert-trend'), api.get('/dashboard/device-trend')])
    const aTrend = at.data?.data || []
    const dTrend = dt.data?.data || []
    if (alertChart.value) {
      echarts.init(alertChart.value).setOption({
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: aTrend.map((d: any) => d.date) },
        yAxis: { type: 'value' },
        series: [{ name: '告警', data: aTrend.map((d: any) => d.count), type: 'line', smooth: true, areaStyle: {} }]
      })
    }
    if (devChart.value) {
      echarts.init(devChart.value).setOption({
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: dTrend.map((d: any) => d.date) },
        yAxis: { type: 'value' },
        series: [{ name: '活跃设备', data: dTrend.map((d: any) => d.count), type: 'bar' }]
      })
    }
  } catch (e) { console.error(e) }
})
</script>

<style scoped>
.page { padding: 4px; }
.metric-label { font-size: 13px; color: #94a3b8; margin-bottom: 8px; }
.metric-value { font-size: 26px; font-weight: 600; color: #1f3a5e; }
.text-red { color: #ef4444; }
.text-orange { color: #f97316; }
</style>