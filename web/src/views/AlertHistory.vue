<template><div class="page">
  <el-card>
    <template #header><span>告警历史</span></template>
    <el-table :data="rows" border size="small">
      <el-table-column prop="alertId" label="ID" width="100" />
      <el-table-column prop="ruleName" label="规则" />
      <el-table-column prop="deviceId" label="设备" />
      <el-table-column prop="level" label="等级" width="100" />
      <el-table-column prop="message" label="消息" />
      <el-table-column prop="firedAt" label="触发时间" width="180" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'acked' ? 'success' : 'danger'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button v-if="row.status !== 'acked'" link type="primary" @click="ack(row)">确认</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/api'
const rows = ref<any[]>([])
async function load() { rows.value = (await api.get('/alert/history')).data?.data || [] }
async function ack(row: any) {
  await api.post(`/alert/history/${row.alertId}/ack`)
  ElMessage.success('已确认')
  load()
}
onMounted(load)
</script>