<template><div class="page">
  <el-card>
    <template #header><span>部门列表</span></template>
    <el-table :data="tree" border size="small" row-key="deptId" default-expand-all>
      <el-table-column prop="deptName" label="部门名" />
      <el-table-column prop="leader" label="负责人" width="120" />
      <el-table-column prop="phone" label="电话" width="160" />
      <el-table-column prop="orderNum" label="排序" width="80" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }"><el-tag :type="row.status === '1' ? 'success' : 'info'">{{ row.status === '1' ? '正常' : '停用' }}</el-tag></template>
      </el-table-column>
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const tree = ref<any[]>([])
function toTree(list: any[]) {
  const map: Record<number, any> = {}
  list.forEach(d => { d.children = []; map[d.deptId] = d })
  const roots: any[] = []
  list.forEach(d => { if (d.parentId && map[d.parentId]) map[d.parentId].children.push(d); else roots.push(d) })
  return roots
}
onMounted(async () => {
  const list = (await api.get('/system/dept/list')).data?.data || []
  tree.value = toTree(list)
})
</script>