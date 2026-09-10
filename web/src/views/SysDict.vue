<template><div class="page">
  <el-card>
    <template #header><span>字典类型</span></template>
    <el-table :data="types" border stripe size="small">
      <el-table-column prop="dictId" label="ID" width="80" />
      <el-table-column prop="dictName" label="字典名" />
      <el-table-column prop="dictType" label="字典类型" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }"><el-tag :type="row.status === '1' ? 'success' : 'info'">{{ row.status === '1' ? '正常' : '停用' }}</el-tag></template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" />
    </el-table>
  </el-card>
  <el-card style="margin-top:16px">
    <template #header>
      <div class="toolbar">
        <el-select v-model="dictType" placeholder="选择字典类型" style="width:240px" @change="loadData">
          <el-option v-for="t in types" :key="t.dictId" :label="`${t.dictName} (${t.dictType})`" :value="t.dictType" />
        </el-select>
      </div>
    </template>
    <el-table :data="data" border stripe size="small">
      <el-table-column prop="dictCode" label="ID" width="80" />
      <el-table-column prop="dictLabel" label="标签" />
      <el-table-column prop="dictValue" label="值" />
      <el-table-column prop="dictSort" label="排序" width="80" />
      <el-table-column prop="isDefault" label="默认" width="80" />
      <el-table-column prop="status" label="状态" width="80" />
    </el-table>
  </el-card>
</div></template>
<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '@/api'
const types = ref<any[]>([]); const data = ref<any[]>([]); const dictType = ref('')
async function loadTypes() { types.value = (await api.get('/system/dict/type/list')).data?.data || [] }
async function loadData() { if (!dictType.value) { data.value = []; return }; data.value = (await api.get('/system/dict/data/list', { params: { dictType: dictType.value } })).data?.data || [] }
onMounted(async () => { await loadTypes() })
</script>
<style scoped>.toolbar{display:flex;gap:12px}</style>