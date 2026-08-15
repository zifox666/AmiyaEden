<template>
  <div class="combat-log-page">
    <div class="art-card p-5 mb-4">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 class="text-xl">{{ t('combatLog.title') }}</h2>
          <p class="text-sm text-g-500 mt-1">{{ t('combatLog.description') }}</p>
          <p class="text-xs text-g-500 mt-2">{{ t('combatLog.privacyNote') }}</p>
        </div>
        <ElUpload
          accept=".txt,.log,text/plain"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleUpload"
        >
          <ElButton type="primary" :loading="uploading">
            <ArtSvgIcon icon="ri:upload-cloud-2-line" class="mr-1" />
            {{ t('combatLog.upload') }}
          </ElButton>
        </ElUpload>
      </div>
    </div>

    <div class="art-card p-5">
      <div class="flex flex-wrap items-center justify-between gap-3 mb-4">
        <div>
          <h3 class="text-lg">{{ t('combatLog.reportList') }}</h3>
          <p class="text-sm text-g-500 mt-1">{{ t('combatLog.reportListHint') }}</p>
        </div>
        <div class="flex items-center gap-2">
          <ElRadioGroup v-if="isAdmin" v-model="scope" @change="changeScope">
            <ElRadioButton value="mine">{{ t('combatLog.scopeMine') }}</ElRadioButton>
            <ElRadioButton value="all">{{ t('combatLog.scopeAll') }}</ElRadioButton>
          </ElRadioGroup>
          <ElButton @click="loadReports">{{ t('combatLog.refresh') }}</ElButton>
        </div>
      </div>

      <ElTable :data="reports" v-loading="loading" @row-dblclick="openDetail">
        <ElTableColumn
          v-if="isAdmin && scope === 'all'"
          prop="user_nickname"
          :label="t('combatLog.member')"
          min-width="130"
        />
        <ElTableColumn prop="character_name" :label="t('combatLog.character')" min-width="150" />
        <ElTableColumn
          prop="original_filename"
          :label="t('combatLog.filename')"
          min-width="210"
          show-overflow-tooltip
        />
        <ElTableColumn :label="t('combatLog.startedAt')" min-width="180">
          <template #default="{ row }">{{ formatDate(row.started_at) }}</template>
        </ElTableColumn>
        <ElTableColumn :label="t('combatLog.duration')" width="120">
          <template #default="{ row }">{{ formatDuration(row.duration_seconds) }}</template>
        </ElTableColumn>
        <ElTableColumn :label="t('combatLog.damageOut')" min-width="120" align="right">
          <template #default="{ row }">{{ formatNumber(row.damage_out) }}</template>
        </ElTableColumn>
        <ElTableColumn :label="t('combatLog.damageIn')" min-width="120" align="right">
          <template #default="{ row }">{{ formatNumber(row.damage_in) }}</template>
        </ElTableColumn>
        <ElTableColumn :label="t('combatLog.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <ElButton link type="primary" @click="openDetail(row)">{{
              t('combatLog.view')
            }}</ElButton>
          </template>
        </ElTableColumn>
        <template #empty>{{ t('combatLog.empty') }}</template>
      </ElTable>

      <div class="flex justify-end mt-4">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          background
          layout="total, sizes, prev, pager, next"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          @current-change="loadReports"
          @size-change="changePageSize"
        />
      </div>
    </div>

    <ElDrawer v-model="detailVisible" size="82%" :title="t('combatLog.detailTitle')">
      <div v-if="detail" v-loading="detailLoading" class="pb-6">
        <ElDescriptions :column="isNarrow ? 1 : 3" border class="mb-4">
          <ElDescriptionsItem :label="t('combatLog.character')">
            {{ detail.report.character_name }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="isAdmin" :label="t('combatLog.member')">
            {{ detail.report.user_nickname || `#${detail.report.user_id}` }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="t('combatLog.language')">
            {{ t(`combatLog.languages.${detail.report.language}`) }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="t('combatLog.timeRange')">
            {{ formatDate(detail.report.started_at) }} — {{ formatDate(detail.report.ended_at) }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="t('combatLog.duration')">
            {{ formatDuration(detail.report.duration_seconds) }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="t('combatLog.events')">
            {{ formatNumber(detail.report.event_count) }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <div class="metric-grid mb-4">
          <div v-for="metric in metrics" :key="metric.key" class="metric-card">
            <span class="text-xs text-g-500">{{ metric.label }}</span>
            <strong>{{ formatNumber(metric.value) }}</strong>
          </div>
        </div>

        <div class="quality-grid mb-4">
          <div v-for="chart in hitQualityCharts" :key="chart.direction" class="art-card border p-4">
            <h4 class="mb-2">{{ chart.title }}</h4>
            <ArtRingChart
              v-if="chart.data.length"
              height="20rem"
              :data="chart.data"
              :center-text="chart.centerText"
              :colors="hitQualityColors"
              :show-label="true"
              :show-legend="true"
              legend-position="bottom"
            />
            <ElEmpty v-else :description="t('combatLog.noHitQualityData')" />
          </div>
        </div>

        <div class="art-card border p-4 mb-4">
          <h4 class="mb-3">{{ t('combatLog.timeline') }}</h4>
          <ArtLineChart
            height="21rem"
            :data="timelineSeries"
            :x-axis-data="timelineLabels"
            :show-legend="true"
            legend-position="bottom"
            :show-axis-line="false"
          />
        </div>

        <div class="art-card border p-4">
          <div class="flex flex-wrap items-center justify-between gap-3 mb-3">
            <h4>{{ t('combatLog.breakdown') }}</h4>
            <ElSelect v-model="breakdownCategory" class="w-56">
              <ElOption
                v-for="category in availableCategories"
                :key="category"
                :label="categoryLabel(category)"
                :value="category"
              />
            </ElSelect>
          </div>
          <ElTable :data="filteredBreakdowns" max-height="430">
            <ElTableColumn prop="pilot_name" :label="t('combatLog.pilot')" min-width="170">
              <template #default="{ row }">{{ row.pilot_name || t('combatLog.unknown') }}</template>
            </ElTableColumn>
            <ElTableColumn prop="ship_type" :label="t('combatLog.ship')" min-width="150">
              <template #default="{ row }">{{ row.ship_type || t('combatLog.unknown') }}</template>
            </ElTableColumn>
            <ElTableColumn prop="weapon_type" :label="t('combatLog.weapon')" min-width="190">
              <template #default="{ row }">{{
                row.weapon_type || t('combatLog.unknown')
              }}</template>
            </ElTableColumn>
            <ElTableColumn :label="t('combatLog.amount')" width="140" align="right">
              <template #default="{ row }">{{ formatNumber(row.amount) }}</template>
            </ElTableColumn>
            <ElTableColumn
              v-if="breakdownCategory === 'mining'"
              :label="t('combatLog.volumeM3')"
              width="140"
              align="right"
            >
              <template #default="{ row }">{{ formatNumber(row.volume_m3) }}</template>
            </ElTableColumn>
            <ElTableColumn :label="t('combatLog.events')" width="110" align="right">
              <template #default="{ row }">{{ formatNumber(row.event_count) }}</template>
            </ElTableColumn>
          </ElTable>
        </div>
      </div>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import type { UploadFile } from 'element-plus'
  import type { LineDataItem, PieDataItem } from '@/types/component/chart'
  import { useI18n } from 'vue-i18n'
  import { useWindowSize } from '@vueuse/core'
  import { useUserStore } from '@/store/modules/user'
  import { getCombatLogDetail, getCombatLogReports, uploadCombatLog } from '@/api/combat-log'

  defineOptions({ name: 'CombatLogReports' })

  const { t, locale } = useI18n()
  const userStore = useUserStore()
  const { width } = useWindowSize()
  const isNarrow = computed(() => width.value < 900)
  const isAdmin = computed(() =>
    (userStore.info.roles ?? []).some((role) => ['admin', 'super_admin'].includes(role))
  )
  const scope = ref<'mine' | 'all'>('mine')
  const reports = ref<Api.CombatLog.Report[]>([])
  const detail = ref<Api.CombatLog.Detail>()
  const loading = ref(false)
  const uploading = ref(false)
  const detailLoading = ref(false)
  const detailVisible = ref(false)
  const breakdownCategory = ref<Api.CombatLog.Category>('damage_out')
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

  const categories: Api.CombatLog.Category[] = [
    'damage_out',
    'damage_in',
    'logistics_out',
    'logistics_in',
    'cap_transfer_out',
    'cap_transfer_in',
    'cap_damage_out',
    'cap_damage_in',
    'mining'
  ]
  const hitQualityOrder: Api.CombatLog.HitQuality[] = [
    'barely_scratches',
    'hits',
    'smashes',
    'penetrates',
    'wrecks',
    'glances_off'
  ]
  const hitQualityColors = ['#94a3b8', '#3b82f6', '#f59e0b', '#8b5cf6', '#ef4444', '#14b8a6']

  const hitQualityCharts = computed(() =>
    (['damage_out', 'damage_in'] as Api.CombatLog.HitDirection[]).map((direction) => ({
      direction,
      title: t(direction === 'damage_out' ? 'combatLog.hitQualityOut' : 'combatLog.hitQualityIn'),
      centerText: t(direction === 'damage_out' ? 'combatLog.output' : 'combatLog.incoming'),
      data: hitQualityChartData(direction)
    }))
  )

  function hitQualityChartData(direction: Api.CombatLog.HitDirection): PieDataItem[] {
    const rows = new Map(
      (detail.value?.hit_qualities ?? [])
        .filter((row) => row.direction === direction)
        .map((row) => [row.quality, row])
    )
    return hitQualityOrder.flatMap((quality) => {
      const row = rows.get(quality)
      if (!row || row.hit_count === 0) return []
      return [
        { name: t(`combatLog.hitQualities.${quality}`), value: Number(row.percentage.toFixed(2)) }
      ]
    })
  }

  const availableCategories = computed(() => {
    const present = new Set(detail.value?.breakdowns.map((item) => item.category) ?? [])
    return categories.filter((category) => present.has(category))
  })

  const filteredBreakdowns = computed(() =>
    (detail.value?.breakdowns ?? []).filter((item) => item.category === breakdownCategory.value)
  )

  const metrics = computed(() => {
    if (!detail.value) return []
    const report = detail.value.report
    return [
      { key: 'damageOut', label: t('combatLog.damageOut'), value: report.damage_out },
      { key: 'damageIn', label: t('combatLog.damageIn'), value: report.damage_in },
      { key: 'dpsOut', label: t('combatLog.averageDpsOut'), value: detail.value.average_dps_out },
      { key: 'dpsIn', label: t('combatLog.averageDpsIn'), value: detail.value.average_dps_in },
      { key: 'logisticsOut', label: t('combatLog.logisticsOut'), value: report.logistics_out },
      { key: 'logisticsIn', label: t('combatLog.logisticsIn'), value: report.logistics_in },
      { key: 'capOut', label: t('combatLog.capTransferOut'), value: report.cap_transfer_out },
      { key: 'capIn', label: t('combatLog.capTransferIn'), value: report.cap_transfer_in },
      { key: 'neutOut', label: t('combatLog.capDamageOut'), value: report.cap_damage_out },
      { key: 'neutIn', label: t('combatLog.capDamageIn'), value: report.cap_damage_in },
      { key: 'mining', label: t('combatLog.mining'), value: report.mining_units },
      { key: 'miningM3', label: t('combatLog.miningM3'), value: report.mining_m3 }
    ]
  })

  const timelineLabels = computed(() => timelineShape.value.labels)
  const timelineSeries = computed<LineDataItem[]>(() => timelineShape.value.series)
  const timelineShape = computed(() => {
    const points = detail.value?.timeline ?? []
    const timestamps = [...new Set(points.map((point) => point.occurred_at))].sort()
    const index = new Map(timestamps.map((timestamp, position) => [timestamp, position]))
    const shownCategories: Api.CombatLog.Category[] = [
      'damage_out',
      'damage_in',
      'logistics_out',
      'logistics_in'
    ]
    const series = shownCategories.map((category) => {
      const data = Array(timestamps.length).fill(0) as number[]
      points
        .filter((point) => point.category === category)
        .forEach((point) => {
          const position = index.get(point.occurred_at)
          if (position !== undefined) data[position] += point.amount
        })
      return { name: categoryLabel(category), data, symbol: 'none' as const }
    })
    return {
      labels: timestamps.map((timestamp) =>
        new Date(timestamp).toLocaleTimeString(locale.value, { hour12: false })
      ),
      series
    }
  })

  async function loadReports() {
    loading.value = true
    try {
      const result = await getCombatLogReports({
        page: pagination.page,
        page_size: pagination.pageSize,
        scope: isAdmin.value ? scope.value : 'mine'
      })
      reports.value = result.list
      pagination.total = result.total
    } finally {
      loading.value = false
    }
  }

  async function handleUpload(file: UploadFile) {
    if (!file.raw || uploading.value) return
    uploading.value = true
    try {
      detail.value = await uploadCombatLog(file.raw)
      setDefaultBreakdownCategory()
      detailVisible.value = true
      ElMessage.success(t('combatLog.uploaded'))
      await loadReports()
    } finally {
      uploading.value = false
    }
  }

  async function openDetail(report: unknown) {
    const reportID = (report as Api.CombatLog.Report).id
    detailVisible.value = true
    detailLoading.value = true
    try {
      detail.value = await getCombatLogDetail(reportID)
      setDefaultBreakdownCategory()
    } finally {
      detailLoading.value = false
    }
  }

  function setDefaultBreakdownCategory() {
    const available = availableCategories.value
    breakdownCategory.value = available.includes('damage_out')
      ? 'damage_out'
      : (available[0] ?? 'damage_out')
  }

  function changeScope() {
    pagination.page = 1
    loadReports()
  }

  function changePageSize() {
    pagination.page = 1
    loadReports()
  }

  function categoryLabel(category: Api.CombatLog.Category) {
    return t(`combatLog.categories.${category}`)
  }

  function formatNumber(value: number) {
    return new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(value)
  }

  function formatDate(value: string) {
    return new Date(value).toLocaleString(locale.value, { hour12: false })
  }

  function formatDuration(seconds: number) {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const remaining = seconds % 60
    return t('combatLog.durationFormat', { hours, minutes, seconds: remaining })
  }

  onMounted(loadReports)
</script>

<style scoped>
  .combat-log-page {
    padding-bottom: 1rem;
  }

  .metric-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(145px, 1fr));
    gap: 0.75rem;
  }

  .metric-card {
    display: flex;
    min-height: 84px;
    padding: 1rem;
    background: var(--art-main-bg-color, var(--el-fill-color-light));
    border: 1px solid var(--art-card-border);
    border-radius: 10px;
    flex-direction: column;
    justify-content: center;
  }

  .metric-card strong {
    margin-top: 0.35rem;
    font-size: 1.35rem;
    line-height: 1.2;
  }

  .quality-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1rem;
  }

  @media (max-width: 900px) {
    .quality-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
