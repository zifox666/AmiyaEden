<template>
  <div class="newbro-manage-page art-full-height">
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <div class="text-lg font-semibold">{{ t('newbro.manage.title') }}</div>
          <div class="text-sm text-gray-500 mt-1">{{ t('newbro.manage.subtitle') }}</div>
        </div>
        <div class="flex items-center gap-3 flex-wrap">
          <ElButton type="primary" :disabled="syncing" @click="runSync">
            {{ t('newbro.manage.runSync') }}
          </ElButton>
          <ElButton type="success" :disabled="processingRewards" @click="runRewardProcessing">
            {{ t('newbro.manage.runRewardProcessing') }}
          </ElButton>
        </div>
      </div>
    </ElCard>

    <ElTabs v-model="activeTab">
      <ElTabPane :label="t('newbro.manage.performanceTab')" name="performance">
        <ElCard shadow="never" class="mb-4">
          <template #header>
            <div class="flex items-center justify-between gap-4 flex-wrap">
              <span>{{ t('newbro.manage.captainPerformance') }}</span>
              <div class="flex items-center gap-3 flex-wrap">
                <ElInput
                  v-model="keyword"
                  clearable
                  style="width: 220px"
                  :placeholder="t('newbro.manage.keyword')"
                  @clear="handleCaptainSearch"
                  @keyup="handleCaptainSearchKeyup"
                />
                <ElButton @click="handleCaptainSearch">{{ $t('common.search') }}</ElButton>
              </div>
            </div>
          </template>

          <ArtTable
            :loading="loadingCaptains"
            :data="captains"
            :columns="captainColumns"
            :pagination="page"
            @pagination:size-change="handleCaptainSizeChange"
            @pagination:current-change="handleCaptainCurrentChange"
          />
        </ElCard>

        <ElCard v-if="detail" shadow="never">
          <template #header>
            <div>
              <div class="font-medium">
                {{
                  t('newbro.manage.detailTitle', { name: detail.overview.captain_character_name })
                }}
              </div>
              <div class="text-xs text-gray-500">
                {{ t('newbro.manage.nicknameLabel') }}:
                {{ detail.overview.captain_nickname || '-' }}
              </div>
            </div>
          </template>

          <div class="grid grid-cols-2 xl:grid-cols-4 gap-4 mb-4">
            <ElCard shadow="never" class="summary-card">
              <div class="text-sm text-gray-500">{{ t('newbro.manage.activeNewbroCount') }}</div>
              <div class="text-2xl font-semibold summary-card__value">
                {{ detail.overview.active_player_count }}
              </div>
            </ElCard>
            <ElCard shadow="never" class="summary-card">
              <div class="text-sm text-gray-500">
                {{ t('newbro.manage.historicalNewbroCount') }}
              </div>
              <div class="text-2xl font-semibold summary-card__value">{{
                detail.overview.historical_player_count
              }}</div>
            </ElCard>
            <ElCard shadow="never" class="summary-card">
              <div class="text-sm text-gray-500">{{ t('newbro.captain.totalBounty') }}</div>
              <div class="text-2xl font-semibold summary-card__value">{{
                formatIsk(detail.overview.attributed_bounty_total)
              }}</div>
            </ElCard>
            <ElCard shadow="never" class="summary-card">
              <div class="text-sm text-gray-500">{{ t('newbro.captain.recordCount') }}</div>
              <div class="text-2xl font-semibold summary-card__value">{{
                detail.overview.attribution_record_count
              }}</div>
            </ElCard>
          </div>

          <div class="grid grid-cols-1 xl:grid-cols-2 gap-4">
            <ElCard shadow="never">
              <template #header>
                <span>{{ t('newbro.manage.detailPlayers') }}</span>
              </template>
              <ElTable :data="detail.players" stripe border>
                <ElTableColumn
                  prop="player_character_name"
                  :label="t('newbro.common.player')"
                  min-width="160"
                />
                <ElTableColumn prop="started_at" :label="t('newbro.common.startedAt')" width="180">
                  <template #default="{ row }">{{ formatDateTime(row.started_at) }}</template>
                </ElTableColumn>
                <ElTableColumn prop="ended_at" :label="t('newbro.common.endedAt')" width="180">
                  <template #default="{ row }">{{
                    row.ended_at ? formatDateTime(row.ended_at) : '-'
                  }}</template>
                </ElTableColumn>
              </ElTable>
            </ElCard>

            <ElCard shadow="never">
              <template #header>
                <span>{{ t('newbro.manage.detailAttributions') }}</span>
              </template>
              <ElTable :data="detail.attributions" stripe border>
                <ElTableColumn
                  prop="player_character_name"
                  :label="t('newbro.common.player')"
                  min-width="160"
                />
                <ElTableColumn prop="ref_type" :label="t('newbro.common.refType')" width="160" />
                <ElTableColumn prop="amount" :label="t('newbro.common.amount')" width="160">
                  <template #default="{ row }">{{ formatIsk(row.amount) }}</template>
                </ElTableColumn>
                <ElTableColumn prop="journal_at" :label="t('newbro.common.journalAt')" width="180">
                  <template #default="{ row }">{{ formatDateTime(row.journal_at) }}</template>
                </ElTableColumn>
              </ElTable>
            </ElCard>
          </div>
        </ElCard>
      </ElTabPane>

      <ElTabPane :label="t('newbro.manage.rewardHistoryTab')" name="rewards">
        <div class="grid grid-cols-1 xl:grid-cols-3 gap-4 mb-4">
          <ElCard shadow="never">
            <div class="text-sm text-gray-500">{{ t('newbro.manage.rewardSettlementCount') }}</div>
            <div class="text-2xl font-semibold mt-2">{{ rewardSummary.settlement_count }}</div>
          </ElCard>
          <ElCard shadow="never">
            <div class="text-sm text-gray-500">{{ t('newbro.manage.rewardTotalCredited') }}</div>
            <div class="text-2xl font-semibold mt-2">
              {{ formatCredit(rewardSummary.total_credited_value) }}
            </div>
          </ElCard>
          <ElCard shadow="never">
            <div class="text-sm text-gray-500">{{ t('newbro.manage.rewardLastProcessedAt') }}</div>
            <div class="text-lg font-semibold mt-2">
              {{ formatDateTime(rewardSummary.last_processed_at) }}
            </div>
          </ElCard>
        </div>

        <ElCard shadow="never">
          <template #header>
            <div class="flex items-center justify-between gap-4 flex-wrap">
              <span>{{ t('newbro.manage.rewardHistoryTab') }}</span>
              <div class="flex items-center gap-3 flex-wrap">
                <ElInput
                  v-model="rewardKeyword"
                  clearable
                  style="width: 240px"
                  :placeholder="t('newbro.manage.keyword')"
                  @clear="handleRewardSearch"
                  @keyup="handleRewardSearchKeyup"
                />
                <ElButton type="primary" @click="handleRewardSearch">
                  {{ $t('common.search') }}
                </ElButton>
                <ElButton @click="handleRewardReset">{{ $t('common.reset') }}</ElButton>
              </div>
            </div>
          </template>
          <ArtTable
            :loading="loadingRewards"
            :data="rewardRows"
            :columns="rewardColumns"
            :pagination="rewardPage"
            visual-variant="ledger"
            :show-table-header="false"
            @pagination:size-change="handleRewardSizeChange"
            @pagination:current-change="handleRewardCurrentChange"
          />
        </ElCard>
      </ElTabPane>

      <ElTabPane :label="t('newbro.manage.affiliationHistoryTab')" name="history">
        <ElCard shadow="never" class="mb-4">
          <template #header>
            <span>{{ t('newbro.manage.affiliationHistoryTitle') }}</span>
          </template>

          <div class="flex items-center gap-3 flex-wrap">
            <ElDatePicker
              v-model="historyFilters.dateRange"
              type="daterange"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              :start-placeholder="t('newbro.manage.historyStartDate')"
              :end-placeholder="t('newbro.manage.historyEndDate')"
              style="width: 280px"
            />
            <ElInput
              v-model="historyFilters.captainSearch"
              clearable
              style="width: 240px"
              :placeholder="t('newbro.manage.historyCaptainFilter')"
            />
            <ElInput
              v-model="historyFilters.playerSearch"
              clearable
              style="width: 260px"
              :placeholder="t('newbro.manage.historyPlayerFilter')"
            />
            <ElButton type="primary" @click="handleHistorySearch">
              {{ $t('common.search') }}
            </ElButton>
            <ElButton @click="handleHistoryReset">{{ $t('common.reset') }}</ElButton>
          </div>

          <div class="text-xs text-gray-500 mt-3">
            {{ t('newbro.manage.historyFilterHint') }}
          </div>
        </ElCard>

        <ElCard shadow="never">
          <ArtTable
            :loading="loadingHistory"
            :data="historyRows"
            :columns="historyColumns"
            :pagination="historyPage"
            visual-variant="ledger"
            :show-table-header="false"
            @pagination:size-change="handleHistorySizeChange"
            @pagination:current-change="handleHistoryCurrentChange"
          />
        </ElCard>
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script setup lang="ts">
  import type { ColumnOption } from '@/types/component'
  import { useI18n } from 'vue-i18n'
  import { useEnterSearch } from '@/hooks/core/useEnterSearch'
  import {
    fetchAdminAffiliationHistory,
    fetchAdminCaptainDetail,
    fetchAdminCaptainList,
    fetchAdminRewardSettlements,
    fetchRunCaptainAttributionSync,
    fetchRunCaptainRewardProcessing
  } from '@/api/newbro'
  import { useNewbroFormatters } from '@/hooks/newbro/useNewbroFormatters'

  defineOptions({ name: 'NewbroManage' })

  const { t } = useI18n()
  const { formatDateTime, formatIsk, formatCredit, formatPercentage } = useNewbroFormatters()
  const { createEnterSearchHandler } = useEnterSearch()

  const activeTab = ref('performance')
  const loadingCaptains = ref(false)
  const loadingHistory = ref(false)
  const loadingRewards = ref(false)
  const syncing = ref(false)
  const processingRewards = ref(false)
  const keyword = ref('')
  const rewardKeyword = ref('')
  const captains = ref<Api.Newbro.CaptainOverview[]>([])
  const detail = ref<Api.Newbro.AdminCaptainDetail | null>(null)
  const historyRows = ref<Api.Newbro.AdminAffiliationHistoryItem[]>([])
  const rewardRows = ref<Api.Newbro.CaptainRewardSettlementItem[]>([])
  const historyLoaded = ref(false)
  const rewardLoaded = ref(false)
  const page = reactive({ current: 1, size: 20, total: 0 })
  const rewardPage = reactive({ current: 1, size: 200, total: 0 })
  const historyPage = reactive({ current: 1, size: 200, total: 0 })
  const rewardSummary = ref<Api.Newbro.CaptainRewardSummary>({
    settlement_count: 0,
    total_credited_value: 0,
    last_processed_at: null
  })
  const historyFilters = reactive({
    captainSearch: '',
    playerSearch: '',
    dateRange: [] as string[]
  })

  const captainColumns = computed<ColumnOption<Api.Newbro.CaptainOverview>[]>(() => [
    {
      prop: 'captain_character_name',
      label: t('newbro.common.captain'),
      minWidth: 240,
      formatter: (row) => h('span', { class: 'font-medium' }, row.captain_character_name)
    },
    {
      prop: 'captain_nickname',
      label: t('newbro.manage.nicknameColumn'),
      minWidth: 140
    },
    {
      prop: 'active_player_count',
      label: t('newbro.manage.activeNewbroCount'),
      width: 140
    },
    {
      prop: 'historical_player_count',
      label: t('newbro.manage.historicalNewbroCount'),
      width: 140
    },
    {
      prop: 'attributed_bounty_total',
      label: t('newbro.captain.totalBounty'),
      width: 160,
      formatter: (row) => formatIsk(row.attributed_bounty_total)
    },
    {
      prop: 'attribution_record_count',
      label: t('newbro.captain.recordCount'),
      width: 120
    },
    {
      label: t('common.operation'),
      width: 120,
      fixed: 'right',
      formatter: (row) =>
        h(
          ElButton,
          {
            link: true,
            type: 'primary',
            onClick: () => showDetail(row.captain_user_id)
          },
          () => t('newbro.manage.viewDetail')
        )
    }
  ])

  const rewardColumns = computed<ColumnOption<Api.Newbro.CaptainRewardSettlementItem>[]>(() => [
    { prop: 'captain_character_name', label: t('newbro.common.captain'), minWidth: 180 },
    {
      prop: 'captain_nickname',
      label: t('newbro.manage.captainNicknameColumn'),
      minWidth: 140,
      formatter: (row) =>
        h(
          'span',
          { class: row.captain_nickname ? '' : 'text-gray-400' },
          row.captain_nickname || '-'
        )
    },
    {
      prop: 'processed_at',
      label: t('newbro.common.processedAt'),
      width: 180,
      formatter: (row) => formatDateTime(row.processed_at)
    },
    {
      prop: 'attribution_count',
      label: t('newbro.manage.rewardAttributionCount'),
      width: 160
    },
    {
      prop: 'attributed_isk_total',
      label: t('newbro.manage.rewardAttributedTotal'),
      width: 180,
      formatter: (row) => formatIsk(row.attributed_isk_total)
    },
    {
      prop: 'bonus_rate',
      label: t('newbro.manage.rewardBonusRate'),
      width: 140,
      formatter: (row) => formatPercentage(row.bonus_rate)
    },
    {
      prop: 'credited_value',
      label: t('newbro.manage.rewardCreditedValue'),
      width: 160,
      formatter: (row) => formatCredit(row.credited_value)
    }
  ])

  const historyColumns = computed<ColumnOption<Api.Newbro.AdminAffiliationHistoryItem>[]>(() => [
    { prop: 'captain_character_name', label: t('newbro.common.captain'), minWidth: 180 },
    {
      prop: 'captain_user_id',
      label: t('newbro.manage.historyCaptainUserId'),
      width: 140
    },
    {
      prop: 'captain_nickname',
      label: t('newbro.manage.captainNicknameColumn'),
      minWidth: 140,
      formatter: (row) =>
        h(
          'span',
          { class: row.captain_nickname ? '' : 'text-gray-400' },
          row.captain_nickname || '-'
        )
    },
    { prop: 'player_character_name', label: t('newbro.manage.newbroColumn'), minWidth: 180 },
    {
      prop: 'player_character_id',
      label: t('newbro.manage.historyPlayerCharacterId'),
      width: 150
    },
    {
      prop: 'player_nickname',
      label: t('newbro.manage.newbroNicknameColumn'),
      minWidth: 140,
      formatter: (row) =>
        h('span', { class: row.player_nickname ? '' : 'text-gray-400' }, row.player_nickname || '-')
    },
    {
      prop: 'changed_by_character_name',
      label: t('newbro.manage.changedByCharacterColumn'),
      minWidth: 180,
      formatter: (row) =>
        h(
          'span',
          { class: row.changed_by_character_name ? '' : 'text-gray-400' },
          row.changed_by_character_name || '-'
        )
    },
    {
      prop: 'started_at',
      label: t('newbro.common.startedAt'),
      width: 180,
      formatter: (row) => formatDateTime(row.started_at)
    },
    {
      prop: 'ended_at',
      label: t('newbro.common.endedAt'),
      width: 180,
      formatter: (row) => (row.ended_at ? formatDateTime(row.ended_at) : '-')
    }
  ])

  const loadCaptains = async () => {
    loadingCaptains.value = true
    try {
      const data = await fetchAdminCaptainList({
        current: page.current,
        size: page.size,
        keyword: keyword.value.trim() || undefined
      })
      captains.value = data.list
      page.total = data.total
    } finally {
      loadingCaptains.value = false
    }
  }

  const loadHistory = async () => {
    loadingHistory.value = true
    try {
      const [changeStartDate, changeEndDate] = historyFilters.dateRange
      const data = await fetchAdminAffiliationHistory({
        current: historyPage.current,
        size: historyPage.size,
        captain_search: historyFilters.captainSearch || undefined,
        player_search: historyFilters.playerSearch || undefined,
        change_start_date: changeStartDate || undefined,
        change_end_date: changeEndDate || undefined
      })
      historyRows.value = data.list
      historyPage.total = data.total
      historyLoaded.value = true
    } finally {
      loadingHistory.value = false
    }
  }

  const loadRewards = async () => {
    loadingRewards.value = true
    try {
      const data = await fetchAdminRewardSettlements({
        current: rewardPage.current,
        size: rewardPage.size,
        keyword: rewardKeyword.value.trim() || undefined
      })
      rewardRows.value = data.list
      rewardSummary.value = data.summary
      rewardPage.total = data.total
      rewardLoaded.value = true
    } finally {
      loadingRewards.value = false
    }
  }

  const ensureHistoryLoaded = async () => {
    if (!historyLoaded.value) {
      await loadHistory()
    }
  }

  const ensureRewardsLoaded = async () => {
    if (!rewardLoaded.value) {
      await loadRewards()
    }
  }

  const handleCaptainSearch = async () => {
    page.current = 1
    await loadCaptains()
  }
  const handleCaptainSearchKeyup = createEnterSearchHandler(handleCaptainSearch)

  const handleCaptainCurrentChange = async (value: number) => {
    page.current = value
    await loadCaptains()
  }

  const handleCaptainSizeChange = async (value: number) => {
    page.size = value
    page.current = 1
    await loadCaptains()
  }

  const handleHistorySearch = async () => {
    historyPage.current = 1
    await loadHistory()
  }

  const handleRewardCurrentChange = async (value: number) => {
    rewardPage.current = value
    await loadRewards()
  }

  const handleRewardSizeChange = async (value: number) => {
    rewardPage.size = value
    rewardPage.current = 1
    await loadRewards()
  }

  const handleRewardSearch = async () => {
    rewardPage.current = 1
    await loadRewards()
  }
  const handleRewardSearchKeyup = createEnterSearchHandler(handleRewardSearch)

  const handleRewardReset = async () => {
    rewardKeyword.value = ''
    rewardPage.current = 1
    await loadRewards()
  }

  const handleHistoryReset = async () => {
    historyFilters.captainSearch = ''
    historyFilters.playerSearch = ''
    historyFilters.dateRange = []
    historyPage.current = 1
    await loadHistory()
  }

  const handleHistoryCurrentChange = async (value: number) => {
    historyPage.current = value
    await loadHistory()
  }

  const handleHistorySizeChange = async (value: number) => {
    historyPage.size = value
    historyPage.current = 1
    await loadHistory()
  }

  const showDetail = async (captainUserId: number) => {
    detail.value = await fetchAdminCaptainDetail(captainUserId)
  }

  const runSync = async () => {
    syncing.value = true
    try {
      const result = await fetchRunCaptainAttributionSync()
      ElMessage.success(
        t('newbro.manage.syncSuccess', {
          inserted: result.inserted_count,
          processed: result.processed_count
        })
      )
      await loadCaptains()
      if (rewardLoaded.value) {
        await loadRewards()
      }
      if (historyLoaded.value) {
        await loadHistory()
      }
      if (detail.value) {
        await showDetail(detail.value.overview.captain_user_id)
      }
    } finally {
      syncing.value = false
    }
  }

  const runRewardProcessing = async () => {
    processingRewards.value = true
    try {
      const result = await fetchRunCaptainRewardProcessing()
      ElMessage.success(
        t('newbro.manage.rewardProcessSuccess', {
          captains: result.processed_captain_count,
          attributions: result.processed_attribution_count,
          credited: formatCredit(result.total_credited_value)
        })
      )
      await loadCaptains()
      await loadRewards()
      if (detail.value) {
        await showDetail(detail.value.overview.captain_user_id)
      }
    } finally {
      processingRewards.value = false
    }
  }

  watch(activeTab, (value) => {
    if (value === 'rewards') {
      void ensureRewardsLoaded()
    }
    if (value === 'history') {
      void ensureHistoryLoaded()
    }
  })

  onMounted(() => {
    loadCaptains()
  })
</script>

<style scoped>
  .summary-card {
    min-height: 0;
  }

  .summary-card :deep(.el-card__body) {
    padding: 0.75rem 1rem;
  }

  .summary-card__value {
    margin-top: 0.25rem;
  }
</style>
