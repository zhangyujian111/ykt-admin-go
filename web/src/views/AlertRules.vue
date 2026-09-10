<template><div class="page">
  <el-card>
    <template #header><span>告警规则</span></template>
    <el-table :data="rules" border size="small">
      <el-table-column prop="ruleId" label="ID" width="80" />
      <el-table-column prop="ruleName" label="规则" />
      <el-table-column prop="metric" label="指标" />
      <el-table-column prop="expression" label="表达式" />
      <el-table-column prop="threshold" label="阈值" width="120" />
      <el-table-column prop="level" label="等级" width="100" />
      <el-table-column label="启用" width="100">
        <template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag></template>
      </el-table-column>
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const rules = ref<any[]>([])
onMounted(async () => {
  const r = await api.get('/alert/rule')
  rules.value = r.data?.data || []
})
</script>