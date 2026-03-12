<template>
  <div class="corporation-pap-page art-full-height">
    <ElCard class="art-search-card" shadow="never">
      <div class="filter-toolbar">
        <div class="filter-toolbar__main">
          <ElSelect
            v-model="filters.period"
            :placeholder="t('fleet.corporationPap.filters.period')"
            style="width: 180px"
          >
            <ElOption
              v-for="option in periodOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </ElSelect>
          <ElInputNumber
            v-if="filters.period === 'at_year'"
            v-model="filters.year"
            :min="2003"
            :max="2100"
            :step="1"
            style="width: 180px"
          />
          <span v-if="filters.period === 'at_year'" class="text-sm text-gray-500">
            {{ t('fleet.corporationPap.filters.year') }}
          </span>
          <ElSelect
            v-model="filters.corpTickers"
            class="ticker-filter"
            multiple
            filterable
            allow-create
            default-first-option
            collapse-tags
            collapse-tags-tooltip
            :reserve-keyword="false"
            :placeholder="t('fleet.corporationPap.filters.corpTickers')"
          >
            <ElOption
              v-for="ticker in corpTickerOptions"
              :key="ticker"
              :label="ticker"
              :value="ticker"
            />
          </ElSelect>
        </div>
        <div class="filter-toolbar__actions">
          <ElButton type="primary" :loading="loading" @click="handleSearch">
            {{ t('fleet.corporationPap.search') }}
          </ElButton>
          <ElButton @click="handleReset">{{ t('fleet.corporationPap.reset') }}</ElButton>
        </div>
      </div>
    </ElCard>

    <div class="stats-grid">
      <ElCard v-for="card in statsCards" :key="card.label" shadow="never" class="stat-card">
        <p class="stat-label">{{ card.label }}</p>
        <p class="stat-value">{{ card.value }}</p>
      </ElCard>
    </div>

    <ElCard class="art-table-card table-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-lg font-medium">{{ t('fleet.corporationPap.summaryTitle') }}</h2>
          <ElButton :loading="loading" @click="loadData">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
      </template>

      <div class="table-wrap">
        <ElTable v-loading="loading" :data="records" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" :index="tableIndex" />
          <ElTableColumn
            prop="user_id"
            :label="t('fleet.corporationPap.columns.userId')"
            min-width="120"
          />
          <ElTableColumn
            prop="corp_ticker"
            :label="t('fleet.corporationPap.columns.corpTicker')"
            min-width="140"
            align="center"
          />
          <ElTableColumn
            prop="main_character_name"
            :label="t('fleet.corporationPap.columns.mainCharacter')"
            min-width="220"
            show-overflow-tooltip
          />
          <ElTableColumn
            prop="character_count"
            :label="t('fleet.corporationPap.columns.characterCount')"
            min-width="140"
            align="center"
          />
          <ElTableColumn
            prop="strat_op_paps"
            :label="t('fleet.corporationPap.columns.stratOpPaps')"
            min-width="160"
            align="center"
          >
            <template #default="{ row }">
              <ElTag type="warning" size="small">{{ formatPap(row.strat_op_paps) }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn
            prop="skirmish_paps"
            :label="t('fleet.corporationPap.columns.skirmishPaps')"
            min-width="160"
            align="center"
          >
            <template #default="{ row }">
              <ElTag type="success" size="small">{{ formatPap(row.skirmish_paps) }}</ElTag>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>

      <ElEmpty
        v-if="!loading && records.length === 0"
        :description="t('fleet.corporationPap.empty')"
        class="my-4"
      />

      <div v-if="pagination.total > 0" class="flex justify-end mt-4">
        <ElPagination
          v-model:current-page="pagination.current"
          v-model:page-size="pagination.size"
          :page-sizes="[200, 500, 1000]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next"
          background
          @current-change="loadData"
          @size-change="handleSizeChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import {
    ElButton,
    ElCard,
    ElEmpty,
    ElInputNumber,
    ElOption,
    ElPagination,
    ElSelect,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { fetchCorporationPapSummary } from '@/api/fleet'

  defineOptions({ name: 'CorporationPap' })

  const { t } = useI18n()

  const loading = ref(false)
  const records = ref<Api.Fleet.CorporationPapSummaryItem[]>([])
  const currentYear = new Date().getFullYear()
  const filters = reactive<
    Required<Pick<Api.Fleet.CorporationPapSummaryParams, 'period' | 'year'>> & {
      corpTickers: string[]
    }
  >({
    period: 'last_month',
    year: currentYear,
    corpTickers: ['FUXI', 'FMA.1']
  })
  const pagination = reactive({
    current: 1,
    size: 200,
    total: 0
  })
  const overview = ref<Api.Fleet.CorporationPapOverview>({
    filtered_pap_total: 0,
    all_pap_total: 0,
    last_month_pap_total: 0,
    filtered_user_count: 0,
    period: 'last_month'
  })

  const periodOptions = computed(() => [
    { label: t('fleet.corporationPap.periods.currentMonth'), value: 'current_month' as const },
    { label: t('fleet.corporationPap.periods.lastMonth'), value: 'last_month' as const },
    { label: t('fleet.corporationPap.periods.atYear'), value: 'at_year' as const },
    { label: t('fleet.corporationPap.periods.all'), value: 'all' as const }
  ])
  const corpTickerOptions = computed(() => {
    const defaults = ['FUXI', 'FMA.1']
    return Array.from(
      new Set([...defaults, ...filters.corpTickers.map((ticker) => ticker.trim()).filter(Boolean)])
    )
  })

  const statsCards = computed(() => [
    {
      label: t('fleet.corporationPap.stats.filteredPap'),
      value: formatPap(overview.value.filtered_pap_total)
    },
    {
      label: t('fleet.corporationPap.stats.lastMonthPap'),
      value: formatPap(overview.value.last_month_pap_total)
    },
    {
      label: t('fleet.corporationPap.stats.allPap'),
      value: formatPap(overview.value.all_pap_total)
    },
    {
      label: t('fleet.corporationPap.stats.users'),
      value: String(overview.value.filtered_user_count)
    }
  ])

  const formatPap = (value: number) =>
    new Intl.NumberFormat(undefined, {
      minimumFractionDigits: Number.isInteger(value) ? 0 : 1,
      maximumFractionDigits: 1
    }).format(value ?? 0)

  const tableIndex = (index: number) => (pagination.current - 1) * pagination.size + index + 1

  async function loadData() {
    loading.value = true
    try {
      const params: Api.Fleet.CorporationPapSummaryParams = {
        current: pagination.current,
        size: pagination.size,
        period: filters.period,
        corp_tickers: filters.corpTickers
          .map((ticker) => ticker.trim())
          .filter(Boolean)
          .join(',')
      }
      if (filters.period === 'at_year') {
        params.year = filters.year
      }

      const result = await fetchCorporationPapSummary(params)
      records.value = result?.list ?? []
      pagination.total = result?.total ?? 0
      pagination.current = result?.page ?? pagination.current
      pagination.size = result?.pageSize ?? pagination.size
      overview.value = result?.overview ?? {
        filtered_pap_total: 0,
        all_pap_total: 0,
        last_month_pap_total: 0,
        filtered_user_count: 0,
        period: filters.period
      }
    } catch {
      records.value = []
      pagination.total = 0
      overview.value = {
        filtered_pap_total: 0,
        all_pap_total: 0,
        last_month_pap_total: 0,
        filtered_user_count: 0,
        period: filters.period
      }
    } finally {
      loading.value = false
    }
  }

  function handleSearch() {
    pagination.current = 1
    loadData()
  }

  function handleReset() {
    filters.period = 'last_month'
    filters.year = currentYear
    filters.corpTickers = ['FUXI', 'FMA.1']
    pagination.current = 1
    pagination.size = 200
    loadData()
  }

  function handleSizeChange() {
    pagination.current = 1
    loadData()
  }

  watch(
    () => filters.period,
    (period) => {
      if (period === 'at_year' && (!filters.year || filters.year < 2003)) {
        filters.year = currentYear
      }
    }
  )

  onMounted(() => {
    loadData()
  })
</script>

<style scoped lang="scss">
  .corporation-pap-page {
    gap: 12px;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 12px;
  }

  .filter-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }

  .filter-toolbar__main {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    flex: 1 1 560px;
  }

  .filter-toolbar__actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .ticker-filter {
    width: 240px;
  }

  .stat-card {
    border-radius: calc(var(--custom-radius) / 2 + 2px) !important;
  }

  .stat-label {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-bottom: 8px;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 700;
    line-height: 1.2;
  }

  .table-card {
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

  .table-wrap {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }
</style>
