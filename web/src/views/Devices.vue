<template><div class="page">
  <el-card>
    <template #header>
      <div class="toolbar">
        <el-input v-model="kw" placeholder="搜索 deviceId / deviceName" style="width:240px" clearable />
        <el-select v-model="status" placeholder="状态" clearable style="width:140px">
          <el-option label="在线" value="online" />
          <el-option label="离线" value="offline" />
        </el-select>
        <el-button @click="load">刷新</el-button>
      </div>
    </template>
    <el-table :data="rows" v-loading="loading" border stripe size="small">
      <el-table-column prop="deviceId" label="设备 ID" />
      <el-table-column prop="deviceName" label="名称" />
      <el-table-column prop="type" label="类型" width="100" />
      <el-table-column prop="version" label="固件" width="120" />
      <el-table-column prop="state" label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.state === 'online'" type="success" size="small">在线</el-tag>
          <el-tag v-else type="info" size="small">{{ row.state }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="lastSeen" label="最近活跃" width="180" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="primary" @click="$router.push(`/devices/${row.deviceId}`)">详情</el-button>
          <el-button link type="primary" @click="restart(row)">重启</el-button>
          <el-button link type="warning" @click="push(row)">推送</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'
const kw = ref(''); const status = ref(''); const rows = ref<any[]>([]); const loading = ref(false)
async function load() {
  loading.value = true
  try {
    const r = await api.get('/device', { params: { keyword: kw.value, state: status.value, pageSize: 200 } })
    rows.value = r.data?.data?.list || []
  } finally { loading.value = false }
}
async function restart(row: any) {
  try { await ElMessageBox.confirm(`确认重启设备 ${row.deviceId}?`, '提示') } catch { return }
  await api.post(`/device/restart/${row.deviceId}`)
  ElMessage.success('已下发重启指令')
}
async function push(row: any) {
  try { await ElMessageBox.confirm(`确认推送消息到 ${row.deviceId}?`, '提示') } catch { return }
  await api.post(`/device/push/${row.deviceId}`)
  ElMessage.success('已下发推送')
}
onMounted(load)
</script>
<style scoped>.toolbar{display:flex;gap:12px;align-items:center}</style>