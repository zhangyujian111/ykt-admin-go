<template><div class="page">
  <el-tabs v-model="tab">
    <el-tab-pane label="离线设备" name="devices">
      <el-table :data="devices" border size="small">
        <el-table-column prop="deviceId" label="设备 ID" />
        <el-table-column prop="deviceName" label="名称" />
        <el-table-column prop="lastSeen" label="最后活跃" width="180" />
        <el-table-column prop="offlineMinutes" label="离线时长(分)" width="120" />
      </el-table>
    </el-tab-pane>
    <el-tab-pane label="离线历史" name="history">
      <el-table :data="history" border size="small">
        <el-table-column prop="deviceId" label="设备" />
        <el-table-column prop="offlineAt" label="离线时间" width="180" />
        <el-table-column prop="recoveryAt" label="恢复时间" width="180" />
        <el-table-column prop="durationMin" label="离线(分)" width="120" />
      </el-table>
    </el-tab-pane>
  </el-tabs>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const tab = ref('devices')
const devices = ref<any[]>([])
const history = ref<any[]>([])
onMounted(async () => {
  try {
    const [d, h] = await Promise.all([api.get('/ops/offline/devices'), api.get('/ops/offline/history')])
    devices.value = d.data?.data || []
    history.value = (h.data?.data || []).slice(0, 50)
  } catch (err) { console.error(err) }
})
</script>