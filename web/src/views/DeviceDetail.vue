<template><div class="page">
  <el-card v-if="detail">
    <template #header><span>设备详情：{{ detail.deviceId }}</span></template>
    <el-descriptions :column="2" border>
      <el-descriptions-item label="设备 ID">{{ detail.deviceId }}</el-descriptions-item>
      <el-descriptions-item label="名称">{{ detail.deviceName }}</el-descriptions-item>
      <el-descriptions-item label="类型">{{ detail.type }}</el-descriptions-item>
      <el-descriptions-item label="固件">{{ detail.version }}</el-descriptions-item>
      <el-descriptions-item label="状态"><el-tag :type="detail.state === 'online' ? 'success' : 'info'">{{ detail.state }}</el-tag></el-descriptions-item>
      <el-descriptions-item label="最近活跃">{{ detail.lastSeen }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api'
const route = useRoute()
const detail = ref<any>(null)
onMounted(async () => {
  const r = await api.get(`/device/detail/${route.params.id}`)
  detail.value = r.data?.data || null
})
</script>