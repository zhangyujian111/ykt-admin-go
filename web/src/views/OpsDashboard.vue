<template>
  <div class="page">
    <el-card>
      <template #header><span>问题设备</span></template>
      <el-table :data="problemDevices" border size="small">
        <el-table-column prop="deviceId" label="设备 ID" />
        <el-table-column prop="deviceName" label="名称" />
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="reason" label="异常原因" />
        <el-table-column prop="lastSeen" label="最近活跃" width="180" />
      </el-table>
    </el-card>
    <el-card style="margin-top: 16px">
      <template #header><span>错误类型分布</span></template>
      <el-table :data="errorTypes" border size="small">
        <el-table-column prop="type" label="类型" />
        <el-table-column prop="count" label="次数" width="120" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const problemDevices = ref<any[]>([])
const errorTypes = ref<any[]>([])
onMounted(async () => {
  try {
    const [p, e] = await Promise.all([api.get('/ops/dashboard/problem-devices'), api.get('/ops/dashboard/error-type-stats')])
    problemDevices.value = p.data?.data || []
    errorTypes.value = e.data?.data || []
  } catch (err) { console.error(err) }
})
</script>