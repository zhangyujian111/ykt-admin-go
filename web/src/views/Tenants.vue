<template><div class="page">
  <el-card>
    <template #header><span>租户列表</span></template>
    <el-table :data="tenants" border size="small">
      <el-table-column prop="tenantId" label="ID" width="100" />
      <el-table-column prop="tenantName" label="名称" />
      <el-table-column prop="tenantCode" label="编码" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }"><el-tag :type="row.status === '1' ? 'success' : 'info'">{{ row.status === '1' ? '启用' : '停用' }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="180" />
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const tenants = ref<any[]>([])
onMounted(async () => {
  const r = await api.get('/tenant/list')
  tenants.value = r.data?.data || []
})
</script>