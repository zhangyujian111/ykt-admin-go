<template><div class="page">
  <el-card>
    <template #header><span>菜单列表</span></template>
    <el-table :data="rows" border size="small" row-key="menuId" :tree-props="{ children: 'children' }" default-expand-all>
      <el-table-column prop="menuName" label="菜单名" />
      <el-table-column prop="menuType" label="类型" width="80" />
      <el-table-column prop="path" label="路径" />
      <el-table-column prop="component" label="组件" />
      <el-table-column prop="perms" label="权限" />
      <el-table-column prop="orderNum" label="排序" width="80" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }"><el-tag :type="row.status === '1' ? 'success' : 'info'">{{ row.status === '1' ? '显示' : '隐藏' }}</el-tag></template>
      </el-table-column>
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'
const rows = ref<any[]>([])
function toTree(list: any[]) {
  const map: Record<number, any> = {}
  list.forEach(m => { m.children = []; map[m.menuId] = m })
  const roots: any[] = []
  list.forEach(m => { if (m.parentId && map[m.parentId]) map[m.parentId].children.push(m); else roots.push(m) })
  return roots
}
onMounted(async () => {
  const list = (await api.get('/system/menu/list')).data?.data || []
  rows.value = toTree(list)
})
</script>