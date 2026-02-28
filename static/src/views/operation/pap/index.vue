<!-- 我的 PAP 记录页面（本系统 PAP + 联盟 PAP） -->
<template>
  <div class="pap-page art-full-height">
    <!-- ── 本系统 PAP ── -->
    <ElCard class="pap-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-medium">{{ $t('fleet.pap.myTitle') }}</h2>
          <ElButton :loading="loading" @click="loadPapLogs">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
      </template>

      <!-- 汇总统计 -->
      <div class="flex items-center gap-6 mb-4 px-2">
        <div class="text-center">
          <p class="text-2xl font-bold text-primary">{{ totalPap }}</p>
          <p class="text-xs text-gray-500 mt-1">总 PAP</p>
        </div>
        <div class="text-center">
          <p class="text-2xl font-bold text-green-600">{{ papLogs.length }}</p>
          <p class="text-xs text-gray-500 mt-1">参与次数</p>
        </div>
      </div>

      <!-- PAP 表格 -->
      <div class="pap-table-wrap">
        <ElTable v-loading="loading" :data="pagedPapLogs" stripe border style="width: 100%">
          <ElTableColumn
            type="index"
            :index="(i) => (papPage - 1) * papPageSize + i + 1"
            width="60"
            label="#"
          />
          <ElTableColumn prop="fleet_id" label="舰队 ID" min-width="260">
            <template #default="{ row }">
              <code class="text-xs">{{ row.fleet_id }}</code>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="character_id" label="角色 ID" width="120" align="center" />
          <ElTableColumn prop="pap_count" :label="$t('fleet.pap.count')" width="120" align="center">
            <template #default="{ row }">
              <ElTag type="success" size="small">+{{ row.pap_count }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn
            prop="issued_by"
            :label="$t('fleet.pap.issuedBy')"
            width="120"
            align="center"
          />
          <ElTableColumn prop="created_at" :label="$t('fleet.pap.issuedAt')" width="200">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </ElTableColumn>
        </ElTable>
      </div>
      <ElEmpty
        v-if="!loading && papLogs.length === 0"
        :description="$t('fleet.pap.empty')"
        class="my-4"
      />

      <!-- 本系统 PAP 分页 -->
      <div v-if="papLogs.length > 0" class="flex justify-end mt-4">
        <ElPagination
          v-model:current-page="papPage"
          v-model:page-size="papPageSize"
          :page-sizes="[5, 10, 20]"
          :total="papLogs.length"
          layout="total, sizes, prev, pager, next"
          background
          small
        />
      </div>
    </ElCard>

    <!-- ── 联盟 PAP ── -->
    <ElCard class="pap-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <h2 class="text-lg font-medium">联盟 PAP</h2>
            <ElDatePicker
              v-model="allianceMonth"
              type="month"
              format="YYYY-MM"
              value-format="YYYY-MM"
              placeholder="选择月份"
              style="width: 140px"
              @change="onAllianceMonthChange"
            />
          </div>
          <ElButton :loading="allianceLoading" @click="loadAlliancePAP">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
      </template>

      <!-- 联盟 PAP 统计卡片 -->
      <template v-if="allianceSummary">
        <div class="flex flex-wrap items-center gap-6 mb-4 px-2">
          <div class="text-center">
            <p class="text-2xl font-bold text-primary">{{ allianceSummary.total_pap }}</p>
            <p class="text-xs text-gray-500 mt-1">月 PAP</p>
          </div>
          <div class="text-center">
            <p class="text-2xl font-bold text-blue-500">{{ allianceSummary.yearly_total_pap }}</p>
            <p class="text-xs text-gray-500 mt-1">年度 PAP</p>
          </div>
          <div class="text-center">
            <p class="text-xl font-semibold text-green-600">#{{ allianceSummary.monthly_rank }}</p>
            <p class="text-xs text-gray-500 mt-1"
              >军团月排名 / {{ allianceSummary.total_in_corp }}</p
            >
          </div>
          <div class="text-center">
            <p class="text-xl font-semibold text-yellow-500"
              >#{{ allianceSummary.global_monthly_rank }}</p
            >
            <p class="text-xs text-gray-500 mt-1"
              >联盟月排名 / {{ allianceSummary.total_global }}</p
            >
          </div>
          <div class="text-center">
            <p class="text-xl font-semibold text-purple-500">#{{ allianceSummary.yearly_rank }}</p>
            <p class="text-xs text-gray-500 mt-1">军团年排名</p>
          </div>
          <div class="ml-auto text-xs text-gray-400 text-right">
            数据来源：联盟 PAP 系统<br />
            最后计算：{{
              allianceSummary.calculated_at ? formatTime(allianceSummary.calculated_at) : '-'
            }}
          </div>
        </div>
      </template>

      <!-- 联盟舰队明细 -->
      <div class="pap-table-wrap">
        <ElTable
          v-loading="allianceLoading"
          :data="pagedAllianceFleets"
          stripe
          border
          style="width: 100%"
        >
          <ElTableColumn
            type="index"
            :index="(i) => (alliancePage - 1) * alliancePageSize + i + 1"
            width="50"
            label="#"
          />
          <ElTableColumn prop="title" label="行动名称" min-width="100" />
          <ElTableColumn prop="character_name" label="角色" min-width="100" />
          <ElTableColumn prop="level" label="等级" width="110" align="center">
            <template #default="{ row }">
              <ElTag :type="levelTagType(row.level)" size="small">{{ row.level }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="pap" label="PAP" width="80" align="center">
            <template #default="{ row }">
              <ElTag type="success" size="small">{{ row.pap }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="舰船" min-width="160">
            <template #default="{ row }">
              {{ row.ship_type_name }}
              <span class="text-xs text-gray-400 ml-1">({{ row.ship_group_name }})</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="开始时间" width="170">
            <template #default="{ row }">{{ formatTime(row.start_at) }}</template>
          </ElTableColumn>
          <ElTableColumn label="结束时间" width="170">
            <template #default="{ row }">{{ row.end_at ? formatTime(row.end_at) : '-' }}</template>
          </ElTableColumn>
        </ElTable>
      </div>
      <ElEmpty
        v-if="!allianceLoading && allianceFleets.length === 0"
        description="暂无联盟 PAP 记录"
        class="my-4"
      />

      <!-- 联盟 PAP 分页 -->
      <div v-if="allianceFleets.length > 0" class="flex justify-end mt-4">
        <ElPagination
          v-model:current-page="alliancePage"
          v-model:page-size="alliancePageSize"
          :page-sizes="[5, 10, 20]"
          :total="allianceFleets.length"
          layout="total, sizes, prev, pager, next"
          background
          small
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElEmpty,
    ElDatePicker,
    ElPagination
  } from 'element-plus'
  import { fetchMyPapLogs } from '@/api/fleet'
  import {
    fetchMyAlliancePAP,
    type AlliancePAPSummary,
    type AlliancePAPFleet
  } from '@/api/alliance-pap'

  defineOptions({ name: 'MyPap' })

  // ── 本系统 PAP ──
  const papLogs = ref<Api.Fleet.PapLog[]>([])
  const loading = ref(false)

  const papPage = ref(1)
  const papPageSize = ref(5)

  const totalPap = computed(() => papLogs.value.reduce((sum, p) => sum + p.pap_count, 0))

  const pagedPapLogs = computed(() => {
    const start = (papPage.value - 1) * papPageSize.value
    return papLogs.value.slice(start, start + papPageSize.value)
  })

  const loadPapLogs = async () => {
    loading.value = true
    try {
      papLogs.value = (await fetchMyPapLogs()) ?? []
      papPage.value = 1
    } catch {
      papLogs.value = []
    } finally {
      loading.value = false
    }
  }

  // ── 联盟 PAP ──
  const now = new Date()
  const allianceMonth = ref<string>(
    `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  )
  const allianceLoading = ref(false)
  const allianceSummary = ref<AlliancePAPSummary | null>(null)
  const allianceFleets = ref<AlliancePAPFleet[]>([])

  const alliancePage = ref(1)
  const alliancePageSize = ref(5)

  const pagedAllianceFleets = computed(() => {
    const start = (alliancePage.value - 1) * alliancePageSize.value
    return allianceFleets.value.slice(start, start + alliancePageSize.value)
  })

  const loadAlliancePAP = async () => {
    allianceLoading.value = true
    try {
      const [yearStr, monthStr] = allianceMonth.value.split('-')
      const result = await fetchMyAlliancePAP({
        year: Number(yearStr),
        month: Number(monthStr)
      })
      allianceSummary.value = result?.summary ?? null
      allianceFleets.value = result?.fleets ?? []
      alliancePage.value = 1
    } catch {
      allianceSummary.value = null
      allianceFleets.value = []
    } finally {
      allianceLoading.value = false
    }
  }

  const onAllianceMonthChange = () => {
    alliancePage.value = 1
    loadAlliancePAP()
  }

  const levelTagType = (level: string): 'danger' | 'warning' | 'info' | 'success' => {
    if (level === 'CTA') return 'danger'
    if (level === 'Strat Op') return 'warning'
    return 'info'
  }

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  onMounted(() => {
    loadPapLogs()
    loadAlliancePAP()
  })
</script>

<style scoped lang="scss">
  .pap-page {
    gap: 12px;
  }

  .pap-card {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    border-radius: calc(var(--custom-radius) / 2 + 2px) !important;

    :deep(.el-card__body) {
      flex: 1;
      min-height: 0;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
  }

  .pap-table-wrap {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }
</style>
