<template><div class="page">
  <el-tabs v-model="tab">
    <el-tab-pane label="LLM" name="llm">
      <el-row :gutter="16">
        <el-col :span="6"><el-card><div class="metric-label">总请求</div><div class="metric-value">{{ llm.total ?? '—' }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">输入 tokens</div><div class="metric-value">{{ llm.tokensIn ?? '—' }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">输出 tokens</div><div class="metric-value">{{ llm.tokensOut ?? '—' }}</div></el-card></el-col>
        <el-col :span="6"><el-card><div class="metric-label">失败</div><div class="metric-value text-red">{{ llm.failures ?? '—' }}</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="TTS" name="tts">
      <el-row :gutter="16">
        <el-col :span="8"><el-card><div class="metric-label">合成次数</div><div class="metric-value">{{ tts.total ?? '—' }}</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">字符数</div><div class="metric-value">{{ tts.chars ?? '—' }}</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">失败</div><div class="metric-value text-red">{{ tts.failures ?? '—' }}</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="ASR" name="asr">
      <el-row :gutter="16">
        <el-col :span="8"><el-card><div class="metric-label">识别次数</div><div class="metric-value">{{ asr.total ?? '—' }}</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">音频秒数</div><div class="metric-value">{{ asr.audioSec ?? '—' }}</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">失败</div><div class="metric-value text-red">{{ asr.failures ?? '—' }}</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="RAG" name="rag">
      <el-row :gutter="16">
        <el-col :span="8"><el-card><div class="metric-label">检索次数</div><div class="metric-value">{{ rag.total ?? '—' }}</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">命中</div><div class="metric-value">{{ rag.hits ?? '—' }}</div></el-card></el-col>
        <el-col :span="8"><el-card><div class="metric-label">命中率</div><div class="metric-value">{{ rag.hitRate ?? '—' }}%</div></el-card></el-col>
      </el-row>
    </el-tab-pane>
    <el-tab-pane label="配额趋势" name="quota">
      <div ref="quotaChart" style="height:320px"></div>
    </el-tab-pane>
  </el-tabs>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { api } from '@/api'
const tab = ref('llm')
const llm = ref<any>({}); const tts = ref<any>({}); const asr = ref<any>({}); const rag = ref<any>({})
const quotaChart = ref<HTMLDivElement>()
onMounted(async () => {
  try {
    const [l, t, a, r, q] = await Promise.all([
      api.get('/ops/ai-metrics/llm'), api.get('/ops/ai-metrics/tts'), api.get('/ops/ai-metrics/asr'),
      api.get('/ops/ai-metrics/rag'), api.get('/ops/ai-metrics/quota-trend')
    ])
    llm.value = l.data?.data || {}
    tts.value = t.data?.data || {}
    asr.value = a.data?.data || {}
    rag.value = r.data?.data || {}
    const trend = (q.data?.data || []) as any[]
    if (quotaChart.value && trend.length) {
      echarts.init(quotaChart.value).setOption({
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: trend.map(d => d.date) },
        yAxis: { type: 'value' },
        series: [{ name: '用量', type: 'line', smooth: true, areaStyle: {}, data: trend.map(d => d.value) }]
      })
    }
  } catch (err) { console.error(err) }
})
</script>
<style scoped>.metric-label{font-size:13px;color:#94a3b8;margin-bottom:8px}.metric-value{font-size:24px;font-weight:600;color:#1f3a5e}.text-red{color:#ef4444}.page{padding:4px}</style>