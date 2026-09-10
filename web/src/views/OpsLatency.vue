<template><div class="page">
  <el-tabs v-model="tab">
    <el-tab-pane label="延迟总览" name="summary">
      <el-row :gutter="16">
        <el-col :span="6"><el-card><div class="metric-label">P50</div><div class="metric-value">{{ summary.p50 ?? '—' }}ms</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">P90</div><div class="metric-value">{{ summary.p90 ?? '—' }}ms</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">P99</div><div class="metric-value text-orange">{{ summary.p99 ?? '—' }}ms</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">平均</div><div class="metric-value">{{ summary.avg ?? '—' }}ms</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="延迟趋势" name="trend">
      <div ref="trendChart" style="height:320px"></div>
    </el-tab-pane>
  </el-tabs>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { api } from '@/api'
const tab = ref('summary')
const summary = ref<any>({})
const trendChart = ref<HTMLDivElement>()
onMounted(async () => {
  try {
    const [s, t] = await Promise.all([api.get('/ops/latency/summary'), api.get('/ops/latency/trend')])
    summary.value = s.data?.data || {}
    const trend = (t.data?.data || []) as any[]
    if (trendChart.value && trend.length) {
      echarts.init(trendChart.value).setOption({
        tooltip: { trigger: 'axis' },
        legend: { data: ['p50', 'p90', 'p99'] },
        xAxis: { type: 'category', data: trend.map(d => d.time) },
        yAxis: { type: 'value', name: 'ms' },
        series: [
          { name: 'p50', type: 'line', smooth: true, data: trend.map(d => d.p50) },
          { name: 'p90', type: 'line', smooth: true, data: trend.map(d => d.p90) },
          { name: 'p99', type: 'line', smooth: true, data: trend.map(d => d.p99) }
        ]
      })
    }
  } catch (err) { console.error(err) }
})
</script>
<style scoped>.metric-label{font-size:13px;color:#94a3b8;margin-bottom:8px}.metric-value{font-size:24px;font-weight:600;color:#1f3a5e}.text-orange{color:#f97316}.page{padding:4px}</style>