<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.status" placeholder="状态" clearable style="width:130px" @change="load">
        <el-option :value="1" label="在库" /><el-option :value="2" label="已售" /><el-option :value="3" label="已加工" /><el-option :value="0" label="已损耗" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button v-permission="'produce_inventory:manage'" type="success" @click="openCreate">新增库存</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="crop_variety_name" label="作物" />
      <el-table-column prop="quantity" label="数量" width="100" />
      <el-table-column prop="unit" label="单位" width="80" />
      <el-table-column prop="grade" label="等级" width="80" />
      <el-table-column prop="storage_location" label="存储位置" />
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="statusTag(row.status,'produce')">{{ statusText(row.status,'produce') }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button v-permission="'produce_inventory:manage'" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-permission="'produce_inventory:manage'" size="small" type="danger" @click="onDel(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑库存' : '新增库存'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="作物品种"><el-select v-model="form.crop_variety_id" filterable><el-option v-for="v in varieties" :key="v.id" :label="v.name" :value="v.id" /></el-select></el-form-item>
        <el-form-item label="采收记录"><el-select v-model="form.harvest_id" filterable placeholder="选择采收记录" style="width:100%"><el-option v-for="h in harvests" :key="h.id" :label="`#${h.id} ${h.harvest_date} (${h.total_weight}kg)`" :value="h.id" /></el-select></el-form-item>
        <el-form-item label="数量"><el-input-number v-model="form.quantity" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="等级"><el-input v-model="form.grade" /></el-form-item>
        <el-form-item label="单位"><el-input v-model="form.unit" /></el-form-item>
        <el-form-item label="存储位置"><el-input v-model="form.storage_location" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option :value="1" label="在库" /><el-option :value="2" label="已售" /><el-option :value="3" label="已加工" /><el-option :value="0" label="已损耗" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ProduceAPI, VarietyAPI, HarvestAPI, type Produce, type CropVariety, type Harvest } from '@/api/modules'
import { statusText, statusTag } from '@/utils/constants'

const rows = ref<Produce[]>([]); const total = ref(0); const loading = ref(false)
const varieties = ref<CropVariety[]>([]); const harvests = ref<Harvest[]>([])
const q = reactive({ page: 1, page_size: 20, status: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<Produce>>({ crop_variety_id: undefined, harvest_id: undefined, quantity: 0, grade: '', unit: 'kg', storage_location: '', status: 1 })

async function loadVarieties() { varieties.value = (await VarietyAPI.list({ page: 1, page_size: 200 })).items }
async function loadHarvests() { harvests.value = (await HarvestAPI.list({ page: 1, page_size: 200 })).items }
async function load() { loading.value = true; try { const res = await ProduceAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, crop_variety_id: varieties.value[0]?.id, harvest_id: harvests.value[0]?.id, quantity: 0, grade: '', unit: 'kg', storage_location: '', status: 1 }); dialog.value = true }
function openEdit(row: Produce) { Object.assign(form, row); dialog.value = true }
async function onSave() { saving.value = true; try { if (form.id) await ProduceAPI.update(form.id, form); else await ProduceAPI.create(form); ElMessage.success('保存成功'); dialog.value = false; await load() } finally { saving.value = false } }
async function onDel(row: Produce) { await ElMessageBox.confirm('确认删除该库存记录？', '提示', { type: 'warning' }); await ProduceAPI.remove(row.id); ElMessage.success('已删除'); await load() }
onMounted(() => { loadVarieties(); loadHarvests(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
