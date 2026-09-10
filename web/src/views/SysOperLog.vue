<template><div class="page">
  <el-card>
    <template #header><span>操作日志</span></template>
    <el-table :data="rows" border stripe size="small">
      <el-table-column prop="operId" label="ID" width="100" />
      <el-table-column prop="title" label="标题" />
      <el-table-column prop="operName" label="操作人" width="100" />
      <el-table-column prop="operUrl" label="URL" />
      <el-table-column prop="operIp" label="IP" width="120" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }"><el-tag :type="row.status === '0' ? 'success' : 'danger'">{{ row.status === '0' ? '成功' : '失败' }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="operTime" label="时间" width="180" />
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const rows = ref<any[]>([])
onMounted(async () => { rows.value = (await api.get('/system/log/oper')).data?.data?.list || [] })
</script>