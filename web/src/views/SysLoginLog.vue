<template><div class="page">
  <el-card>
    <template #header><span>登录日志</span></template>
    <el-table :data="rows" border stripe size="small">
      <el-table-column prop="infoId" label="ID" width="100" />
      <el-table-column prop="loginName" label="用户" />
      <el-table-column prop="ipaddr" label="IP" width="120" />
      <el-table-column prop="browser" label="浏览器" />
      <el-table-column prop="os" label="OS" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }"><el-tag :type="row.status === '0' ? 'success' : 'danger'">{{ row.status === '0' ? '成功' : '失败' }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="msg" label="消息" />
      <el-table-column prop="loginTime" label="时间" width="180" />
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const rows = ref<any[]>([])
onMounted(async () => { rows.value = (await api.get('/system/log/login')).data?.data?.list || [] })
</script>