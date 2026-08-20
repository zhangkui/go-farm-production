<template>
  <div>
    <div class="toolbar">
      <el-select v-model="q.season_id" placeholder="全部季次" clearable filterable style="width:180px" @change="load">
        <el-option v-for="s in seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-select v-model="q.status" placeholder="全部状态" clearable style="width:140px" @change="load">
        <el-option :value="1" label="计划中" /><el-option :value="2" label="已播种" /><el-option :value="3" label="生长中" /><el-option :value="4" label="已采收" /><el-option :value="5" label="已完成" /><el-option :value="0" label="已取消" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
      <el-button v-permission="'planting_plan:manage'" type="success" @click="openCreate">新建计划</el-button>
    </div>
    <el-table :data="rows" v-loading="loading" border stripe @row-click="(r:any)=>router.push(`/planting-plans/${r.id}`)">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="field_name" label="地块" />
      <el-table-column prop="crop_variety_name" label="作物" />
      <el-table-column prop="season_name" label="季次" />
      <el-table-column prop="planned_area" label="计划面积(亩)" width="120" />
      <el-table-column prop="planned_sow_date" label="计划播种" width="120" />
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="statusTag(row.status,'plan')">{{ statusText(row.status,'plan') }}</el-tag></template></el-table-column>
    </el-table>
    <el-pagination background layout="total, prev, pager, next" :total="total" v-model:current-page="q.page" :page-size="q.page_size" @current-change="load" style="margin-top:12px" />
    <el-dialog v-model="dialog" :title="form.id ? '编辑种植计划' : '新建种植计划'" width="560px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="地块"><el-select v-model="form.field_id" filterable><el-option v-for="f in fields" :key="f.id" :label="`${f.code} ${f.name}`" :value="f.id" /></el-select></el-form-item>
        <el-form-item label="作物品种"><el-select v-model="form.crop_variety_id" filterable><el-option v-for="v in varieties" :key="v.id" :label="v.name" :value="v.id" /></el-select></el-form-item>
        <el-form-item label="种植季"><el-select v-model="form.season_id" filterable><el-option v-for="s in seasons" :key="s.id" :label="s.name" :value="s.id" /></el-select></el-form-item>
        <el-form-item label="计划面积(亩)"><el-input-number v-model="form.planned_area" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="计划播种日"><el-date-picker v-model="form.planned_sow_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="计划采收日"><el-date-picker v-model="form.planned_harvest_date" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="onSave">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { PlanAPI, FieldAPI, VarietyAPI, SeasonAPI, type PlantingPlan, type Field, type CropVariety, type Season } from '@/api/modules'
import { statusText, statusTag } from '@/utils/constants'

const router = useRouter()
const rows = ref<PlantingPlan[]>([]); const total = ref(0); const loading = ref(false)
const fields = ref<Field[]>([]); const varieties = ref<CropVariety[]>([]); const seasons = ref<Season[]>([])
const q = reactive({ page: 1, page_size: 20, season_id: undefined as number | undefined, status: undefined as number | undefined })
const dialog = ref(false); const saving = ref(false)
const form = reactive<Partial<PlantingPlan>>({ field_id: undefined, crop_variety_id: undefined, season_id: undefined, planned_area: 0, planned_sow_date: '', planned_harvest_date: '', remark: '' })

async function loadRefs() {
  const [f, v, s] = await Promise.all([FieldAPI.list({ page: 1, page_size: 200 }), VarietyAPI.list({ page: 1, page_size: 200 }), SeasonAPI.list({ page: 1, page_size: 200 })])
  fields.value = f.items; varieties.value = v.items; seasons.value = s.items
}
async function load() { loading.value = true; try { const res = await PlanAPI.list(q); rows.value = res.items; total.value = res.total } finally { loading.value = false } }
function openCreate() { Object.assign(form, { id: undefined, field_id: fields.value[0]?.id, crop_variety_id: varieties.value[0]?.id, season_id: seasons.value[0]?.id, planned_area: 0, planned_sow_date: '', planned_harvest_date: '', remark: '' }); dialog.value = true }
async function onSave() {
  saving.value = true
  try {
    const headers: Record<string, string> = {}
    if (form.id) { await PlanAPI.update(form.id, form) } else {
      headers['Idempotency-Key'] = crypto.randomUUID()
      await PlanAPI.create(form)
    }
    ElMessage.success('保存成功'); dialog.value = false; await load()
  } finally { saving.value = false }
}
onMounted(() => { loadRefs(); load() })
</script>

<style scoped>.toolbar { display: flex; gap: 8px; margin-bottom: 12px; }</style>
