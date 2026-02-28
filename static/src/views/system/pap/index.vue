<!-- 系统管理 - 联盟 PAP 所有成员视图 -->
<template>
  <div class="alliance-pap-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between flex-wrap gap-3">
          <div class="flex items-center gap-3">
            <h2 class="text-lg font-medium">联盟 PAP - 全员视图</h2>
            <!-- 月份选择 -->
            <ElDatePicker
              v-model="selectedMonth"
              type="month"
              format="YYYY-MM"
              value-format="YYYY-MM"
              placeholder="选择月份"
              style="width: 145px"
              @change="loadData"
            />
          </div>
          <div class="flex gap-2">
            <ElButton :loading="loading" @click="loadData">
              <el-icon class="mr-1"><Refresh /></el-icon>
              刷新
            </ElButton>
            <ElButton type="primary" :loading="fetching" @click="triggerFetch">
              <el-icon class="mr-1"><Download /></el-icon>
              拉取最新数据
            </ElButton>
          </div>
        </div>
      </template>

      <!-- 月份汇总信息 -->
      <div v-if="list.length" class="mb-4 text-sm text-gray-500 px-1">
        共 <span class="font-medium text-primary">{{ list.length }}</span> 名成员有记录，
        总 PAP：<span class="font-medium text-green-600">{{ totalMonthlyPap.toFixed(1) }}</span>
      </div>

      <!-- 排行榜表格 -->
      <ElTable
        v-loading="loading"
        :data="list"
        stripe
        border
        style="width: 100%"
        :default-sort="{ prop: 'total_pap', order: 'descending' }"
      >
        <ElTableColumn type="index" width="55" label="排名" align="center" />
        <ElTableColumn prop="main_character" label="主角色" min-width="140" />
        <ElTableColumn prop="total_pap" label="月 PAP" width="100" align="center" sortable>
          <template #default="{ row }">
            <ElTag type="success" size="small">{{ row.total_pap }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="yearly_total_pap" label="年度 PAP" width="110" align="center" sortable>
          <template #default="{ row }">
            <ElTag type="info" size="small">{{ row.yearly_total_pap }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="monthly_rank" label="军团月排" width="100" align="center" sortable>
          <template #default="{ row }">
            <span class="text-green-600 font-medium">#{{ row.monthly_rank }}</span>
            <span class="text-xs text-gray-400"> / {{ row.total_in_corp }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="global_monthly_rank" label="联盟月排" width="110" align="center" sortable>
          <template #default="{ row }">
            <span class="text-yellow-500 font-medium">#{{ row.global_monthly_rank }}</span>
            <span class="text-xs text-gray-400"> / {{ row.total_global }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="yearly_rank" label="军团年排" width="100" align="center" sortable>
          <template #default="{ row }">
            <span class="text-purple-500 font-medium">#{{ row.yearly_rank }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="global_yearly_rank" label="联盟年排" width="100" align="center" sortable>
          <template #default="{ row }">
            <span class="text-blue-500 font-medium">#{{ row.global_yearly_rank }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="数据时间" width="170">
          <template #default="{ row }">
            {{ row.calculated_at ? formatTime(row.calculated_at) : '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="is_archived" label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="row.is_archived ? 'info' : 'success'" size="small">
              {{ row.is_archived ? '已归档' : '当月' }}
            </ElTag>
          </template>
        </ElTableColumn>
      </ElTable>

      <ElEmpty v-if="!loading && list.length === 0" description="暂无该月联盟 PAP 数据" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { Refresh, Download } from '@element-plus/icons-vue'
  import { ElCard, ElTable, ElTableColumn, ElTag, ElButton, ElEmpty, ElDatePicker, ElMessage } from 'element-plus'
  import {
    fetchAllAlliancePAP,
    triggerAlliancePAPFetch,
    type AlliancePAPSummary
  } from '@/api/alliance-pap'

  defineOptions({ name: 'AlliancePAP' })

  const now = new Date()
  const selectedMonth = ref<string>(
    `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  )

  const loading = ref(false)
  const fetching = ref(false)
  const list = ref<AlliancePAPSummary[]>([])

  const totalMonthlyPap = computed(() =>
    list.value.reduce((sum, s) => sum + Number(s.total_pap), 0)
  )

  const loadData = async () => {
    loading.value = true
    try {
      const [yearStr, monthStr] = selectedMonth.value.split('-')
      const result = await fetchAllAlliancePAP({
        year: Number(yearStr),
        month: Number(monthStr)
      })
      list.value = result?.list ?? []
    } catch {
      list.value = []
    } finally {
      loading.value = false
    }
  }

  const triggerFetch = async () => {
    fetching.value = true
    try {
      const [yearStr, monthStr] = selectedMonth.value.split('-')
      await triggerAlliancePAPFetch({
        year: Number(yearStr),
        month: Number(monthStr)
      })
      ElMessage.success('已触发后台拉取，请稍后刷新查看')
    } catch {
      ElMessage.error('触发拉取失败')
    } finally {
      fetching.value = false
    }
  }

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  onMounted(loadData)
</script>
