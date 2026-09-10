<template><div class="page">
  <el-card>
    <template #header>
      <div class="toolbar">
        <el-input v-model="roleName" placeholder="角色名" style="width:200px" clearable />
        <el-button @click="load">刷新</el-button>
        <el-button type="primary" @click="openCreate">+ 新增</el-button>
      </div>
    </template>
    <el-table :data="rows" border stripe size="small">
      <el-table-column prop="roleId" label="ID" width="80" />
      <el-table-column prop="roleName" label="角色名" />
      <el-table-column prop="roleKey" label="权限字符" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }"><el-tag :type="row.status === '1' ? 'success' : 'info'">{{ row.status === '1' ? '正常' : '停用' }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="180" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
  <el-dialog v-model="dialogVisible" title="新增角色" width="480px">
    <el-form label-width="80px">
      <el-form-item label="角色名"><el-input v-model="form.roleName" /></el-form-item>
      <el-form-item label="权限字符"><el-input v-model="form.roleKey" placeholder="如 admin / ops / viewer" /></el-form-item>
      <el-form-item label="状态">
        <el-switch v-model="form.status" active-value="1" inactive-value="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible=false">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-dialog>
</div></template>
<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'
const rows = ref<any[]>([]); const roleName = ref('')
const dialogVisible = ref(false); const form = reactive({ roleName: '', roleKey: '', status: '1' })
async function load() {
  rows.value = (await api.get('/system/role/list', { params: { roleName: roleName.value } })).data?.data?.list || []
}
function openCreate() { Object.assign(form, { roleName: '', roleKey: '', status: '1' }); dialogVisible.value = true }
async function save() { await api.post('/system/role', form); ElMessage.success('已创建'); dialogVisible.value = false; load() }
async function remove(row: any) {
  try { await ElMessageBox.confirm(`确认删除 ${row.roleName}?`) } catch { return }
  await api.delete(`/system/role/${row.roleId}`); ElMessage.success('已删除'); load()
}
onMounted(load)
</script>
<style scoped>.toolbar{display:flex;gap:12px;align-items:center}</style>