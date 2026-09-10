<template><div class="page">
  <el-card>
    <template #header>
      <div class="toolbar">
        <el-input v-model="username" placeholder="用户名" style="width:160px" clearable />
        <el-input v-model="nickname" placeholder="昵称" style="width:160px" clearable />
        <el-button @click="load">刷新</el-button>
        <el-button type="primary" @click="openCreate">+ 新增</el-button>
      </div>
    </template>
    <el-table :data="rows" border stripe size="small">
      <el-table-column prop="userId" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="phone" label="电话" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }"><el-tag :type="row.status === '1' ? 'success' : 'danger'">{{ row.status === '1' ? '正常' : '停用' }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="180" />
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link @click="resetPwd(row)">重置密码</el-button>
          <el-button link type="warning" @click="toggle(row)">{{ row.status === '1' ? '停用' : '启用' }}</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
  <el-dialog v-model="dialogVisible" :title="form.userId ? '编辑用户' : '新增用户'" width="540px">
    <el-form label-width="80px">
      <el-form-item label="用户名"><el-input v-model="form.username" :disabled="!!form.userId" /></el-form-item>
      <el-form-item v-if="!form.userId" label="密码"><el-input v-model="form.password" placeholder="默认 123456" /></el-form-item>
      <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
      <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
      <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
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
const rows = ref<any[]>([]); const username = ref(''); const nickname = ref('')
const dialogVisible = ref(false); const form = reactive<any>({ userId: 0, username: '', password: '', nickname: '', email: '', phone: '', status: '1' })
async function load() {
  const r = await api.get('/system/user/list', { params: { username: username.value, nickname: nickname.value } })
  rows.value = r.data?.data?.list || []
}
function openCreate() { Object.assign(form, { userId: 0, username: '', password: '', nickname: '', email: '', phone: '', status: '1' }); dialogVisible.value = true }
function openEdit(row: any) { Object.assign(form, row); dialogVisible.value = true }
async function save() {
  if (form.userId) {
    await api.put(`/system/user/${form.userId}`, form)
  } else {
    await api.post('/system/user', form)
  }
  ElMessage.success('已保存'); dialogVisible.value = false; load()
}
async function resetPwd(row: any) {
  try { await ElMessageBox.confirm(`重置 ${row.username} 的密码为 123456?`) } catch { return }
  await api.put(`/system/user/${row.userId}/resetPwd`, { password: '123456' })
  ElMessage.success('已重置')
}
async function toggle(row: any) {
  await api.put(`/system/user/${row.userId}/status`, { status: row.status === '1' ? '0' : '1' })
  load()
}
async function remove(row: any) {
  try { await ElMessageBox.confirm(`确认删除 ${row.username}?`) } catch { return }
  await api.delete(`/system/user/${row.userId}`); ElMessage.success('已删除'); load()
}
onMounted(load)
</script>
<style scoped>.toolbar{display:flex;gap:12px;align-items:center}</style>