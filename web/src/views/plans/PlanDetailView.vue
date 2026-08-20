<template>
  <div v-loading="loading">
    <el-page-header @back="$router.back()" :content="`种植计划 #${plan?.id}`" />
    <el-descriptions :column="3" border style="margin-top:12px">
      <el-descriptions-item label="地块">{{ plan?.field_name }}</el-descriptions-item>
      <el-descriptions-item label="作物品种">{{ plan?.crop_variety_name }}</el-descriptions-item>
      <el-descriptions-item label="种植季">{{ plan?.season_name }}</el-descriptions-item>
      <el-descriptions-item label="计划面积(亩)">{{ plan?.planned_area }}</el-descriptions-item>
      <el-descriptions-item label="计划播种">{{ plan?.planned_sow_date }}</el-descriptions-item>
      <el-descriptions-item label="计划采收">{{ plan?.planned_harvest_date || '-' }}</el-descriptions-item>
      <el-descriptions-item label="实际播种">{{ plan?.actual_sow_date || '-' }}</el-descriptions-item>
      <el-descriptions-item label="实际采收">{{ plan?.actual_harvest_date || '-' }}</el-descriptions-item>
      <el-descriptions-item label="状态"><el-tag :type="statusTag(plan?.status ?? 0,'plan')">{{ statusText(plan?.status ?? 0,'plan') }}</el-tag></el-descriptions-item>
    </el-descriptions>

    <el-card style="margin-top:12px" v-permission="'planting_plan:manage'">
      <template #header>状态流转</template>
      <el-space wrap>
        <el-button :disabled="plan?.status!==1" @click="transition(2)">播种</el-button>
        <el-button :disabled="plan?.status!==2" @click="transition(3)">生长中</el-button>
        <el-button :disabled="plan?.status!==3" @click="transition(4)">采收</el-button>
        <el-button :disabled="plan?.status!==4" @click="transition(5)">完成</el-button>
        <el-button type="danger" :disabled="!plan || plan.status===0 || plan.status===5" @click="transition(0)">取消</el-button>
      </el-space>
    </el-card>

    <el-tabs style="margin-top:12px">
      <el-tab-pane :label="`农事任务 (${tasks.length})`">
        <el-table :data="tasks" border size="small">
          <el-table-column prop="title" label="标题" />
          <el-table-column label="类型" width="120"><template #default="{ row }">{{ taskTypeText[row.task_type] }}</template></el-table-column>
          <el-table-column prop="planned_date" label="计划日期" width="120" />
          <el-table-column prop="labour_hours" label="工时" width="80" />
          <el-table-column prop="equipment_cost" label="设备成本" width="100" />
          <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="statusTag(row.status,'task')">{{ statusText(row.status,'task') }}</el-tag></template></el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane :label="`采收记录 (${harvests.length})`">
        <el-table :data="harvests" border size="small">
          <el-table-column prop="harvest_date" label="日期" width="120" />
          <el-table-column prop="total_weight" label="总重量(kg)" width="120" />
          <el-table-column prop="grade" label="等级" width="80" />
          <el-table-column label="审核"><template #default="{ row }"><el-tag :type="row.approved?'success':'info'">{{ row.approved?'已审核':'待审核' }}</el-tag></template></el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { PlanAPI, type PlantingPlan, type FarmTask, type Harvest } from '@/api/modules'
import { statusText, statusTag, taskTypeText } from '@/utils/constants'

const route = useRoute()
const id = Number(route.params.id)
const plan = ref<PlantingPlan | null>(null)
const tasks = ref<FarmTask[]>([])
const harvests = ref<Harvest[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try { const res = await PlanAPI.get(id); plan.value = res.plan; tasks.value = res.tasks; harvests.value = res.harvests } finally { loading.value = false }
}
async function transition(to: number) {
  await PlanAPI.transition(id, to); ElMessage.success('状态已更新'); await load()
}
onMounted(load)
</script>
