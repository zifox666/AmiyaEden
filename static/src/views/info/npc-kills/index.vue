<template>
  <div class="npc-kills-page">
    <!-- 人物切换 + 日期范围 -->
    <ElCard class="art-card" shadow="never">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-4 flex-wrap">
          <span class="text-sm text-gray-500">{{ $t('npcKill.selectCharacter') }}</span>
          <ElSelect
            v-model="selectedCharacterId"
            :placeholder="$t('npcKill.selectCharacterPlaceholder')"
            style="width: 240px"
            @change="onCharacterChange"
          >
            <!-- 全部人物 -->
            <ElOption :value="0" :label="$t('npcKill.allCharacters')">
              <span>{{ $t('npcKill.allCharacters') }}</span>
            </ElOption>
            <ElOption
              v-for="char in characters"
              :key="char.character_id"
              :value="char.character_id"
              :label="char.character_name"
            >
              <div class="flex items-center gap-2">
                <ElAvatar :src="char.portrait_url" :size="24" />
                <span>{{ char.character_name }}</span>
              </div>
            </ElOption>
          </ElSelect>

          <ElDatePicker
            v-model="dateRange"
            type="daterange"
            :start-placeholder="$t('npcKill.startDate')"
            :end-placeholder="$t('npcKill.endDate')"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 280px"
          />

          <ElButton type="primary" @click="handleSearch">{{ $t('npcKill.search') }}</ElButton>
          <ElButton @click="handleReset">{{ $t('npcKill.reset') }}</ElButton>
        </div>
      </div>
    </ElCard>

    <!-- 总览卡片 -->
    <div v-if="reportData" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4 my-4">
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.totalBounty') }}</p>
        <p class="text-xl font-bold text-green-600 mt-1">{{
          formatISK(reportData.summary.total_bounty)
        }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.totalTax') }}</p>
        <p class="text-xl font-bold text-red-500 mt-1">{{
          formatISK(reportData.summary.total_tax)
        }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.actualIncome') }}</p>
        <p class="text-xl font-bold text-green-600 mt-1">{{
          formatISK(reportData.summary.actual_income)
        }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.totalRecords') }}</p>
        <p class="text-xl font-bold mt-1">{{ reportData.summary.total_records }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.estimatedHours') }}</p>
        <p class="text-xl font-bold mt-1">{{ reportData.summary.estimated_hours }}</p>
      </ElCard>
    </div>

    <!-- 统计表格 -->
    <div v-if="reportData" class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">
      <!-- 按 NPC 分类 -->
      <ElCard shadow="never" class="art-table-card">
        <template #header>
          <span class="font-medium">{{ $t('npcKill.byNpc') }}</span>
        </template>
        <ElTable :data="reportData.by_npc" stripe border max-height="400">
          <ElTableColumn type="index" width="55" label="#" align="center" />
          <ElTableColumn
            prop="npc_name"
            :label="$t('npcKill.npcName')"
            min-width="200"
            show-overflow-tooltip
          />
          <ElTableColumn
            prop="count"
            :label="$t('npcKill.npcCount')"
            width="100"
            align="right"
            sortable
          />
        </ElTable>
      </ElCard>

      <!-- 按地点分类 -->
      <ElCard shadow="never" class="art-table-card">
        <template #header>
          <span class="font-medium">{{ $t('npcKill.bySystem') }}</span>
        </template>
        <ElTable :data="reportData.by_system" stripe border max-height="400">
          <ElTableColumn type="index" width="55" label="#" align="center" />
          <ElTableColumn
            prop="solar_system_name"
            :label="$t('npcKill.solarSystem')"
            min-width="160"
            show-overflow-tooltip
          />
          <ElTableColumn
            prop="count"
            :label="$t('npcKill.systemCount')"
            width="100"
            align="right"
            sortable
          />
          <ElTableColumn
            prop="amount"
            :label="$t('npcKill.systemAmount')"
            width="160"
            align="right"
            sortable
          >
            <template #default="{ row }">
              <span class="text-green-600 font-medium">{{ formatISK(row.amount) }}</span>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElCard>
    </div>

    <!-- 时间趋势 -->
    <ElCard
      v-if="reportData && reportData.trend?.length"
      shadow="never"
      class="art-table-card mb-4"
    >
      <template #header>
        <span class="font-medium">{{ $t('npcKill.trend') }}</span>
      </template>
      <ElTable :data="reportData.trend" stripe border max-height="400">
        <ElTableColumn prop="date" :label="$t('npcKill.trendDate')" width="140" />
        <ElTableColumn
          prop="amount"
          :label="$t('npcKill.trendAmount')"
          min-width="160"
          align="right"
          sortable
        >
          <template #default="{ row }">
            <span class="text-green-600 font-medium">{{ formatISK(row.amount) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn
          prop="count"
          :label="$t('npcKill.trendCount')"
          width="100"
          align="right"
          sortable
        />
      </ElTable>
    </ElCard>

    <!-- 流水明细 -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        :empty-text="$t('npcKill.noData')"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElAvatar, ElSelect, ElOption, ElDatePicker } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchNpcKills, fetchNpcKillsAll } from '@/api/npc-kill'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'NpcKillReport' })

  type JournalItem = Api.NpcKill.JournalItem

  const { t } = useI18n()

  // ─── 状态 ───
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterId = ref<number>(0) // 0 表示全部人物
  const dateRange = ref<[string, string] | null>(null)
  const reportData = ref<Api.NpcKill.NpcKillResponse | null>(null)

  // ─── ISK 格式化 ───
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  // ─── REF_TYPE 配置 ───
  const REF_TYPE_CONFIG: Record<string, { type: string; text: string }> = {
    bounty_prizes: { type: 'success', text: t('npcKill.refTypes.bounty_prizes') },
    ess_escrow_transfer: { type: 'warning', text: t('npcKill.refTypes.ess_escrow_transfer') }
  }

  // ─── API 适配器 ───
  const fetchNpcKillList = async (params: {
    character_id?: number
    start_date?: string
    end_date?: string
    current: number
    size: number
  }): Promise<Api.Common.PaginatedResponse<JournalItem>> => {
    const charId = params.character_id ?? 0
    let res: Api.NpcKill.NpcKillResponse | undefined
    if (charId === 0) {
      // 全部人物汇总
      res = await fetchNpcKillsAll({
        start_date: params.start_date,
        end_date: params.end_date,
        page: params.current,
        page_size: params.size
      })
    } else {
      res = await fetchNpcKills({
        character_id: charId,
        start_date: params.start_date,
        end_date: params.end_date,
        page: params.current,
        page_size: params.size
      })
    }
    reportData.value = res ?? null
    return {
      list: res?.journals ?? [],
      total: res?.total ?? 0,
      page: res?.page ?? 1,
      pageSize: res?.page_size ?? params.size
    }
  }

  // ─── 表格 ───
  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    searchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    getData
  } = useTable({
    core: {
      apiFn: fetchNpcKillList,
      apiParams: {
        character_id: 0 as number,
        start_date: undefined as string | undefined,
        end_date: undefined as string | undefined,
        current: 1,
        size: 20
      },
      immediate: true,
      columnsFactory: () => [
        {
          prop: 'date',
          label: t('npcKill.journalDate'),
          width: 180
        },
        {
          prop: 'ref_type',
          label: t('npcKill.journalRefType'),
          width: 120,
          formatter: (row: JournalItem) => {
            const cfg = REF_TYPE_CONFIG[row.ref_type]
            return h(
              ElTag,
              { type: (cfg?.type ?? 'info') as any, size: 'small', effect: 'plain' },
              () => cfg?.text || row.ref_type
            )
          }
        },
        {
          prop: 'amount',
          label: t('npcKill.journalAmount'),
          width: 160,
          formatter: (row: JournalItem) =>
            h(
              'span',
              { class: `font-medium ${row.amount >= 0 ? 'text-green-600' : 'text-red-500'}` },
              `${row.amount >= 0 ? '+' : ''}${formatISK(row.amount)}`
            )
        },
        {
          prop: 'tax',
          label: t('npcKill.journalTax'),
          width: 140,
          formatter: (row: JournalItem) =>
            h(
              'span',
              { class: 'font-medium text-red-500' },
              row.tax !== 0 ? formatISK(row.tax) : '-'
            )
        },
        {
          prop: 'solar_system_name',
          label: t('npcKill.journalSystem'),
          width: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'character_name',
          label: t('npcKill.characterName'),
          width: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'reason',
          label: t('npcKill.journalReason'),
          minWidth: 200,
          showOverflowTooltip: true
        }
      ]
    }
  })

  // ─── 事件 ───
  const onCharacterChange = () => {
    searchParams.character_id = selectedCharacterId.value
    getData()
  }

  const handleSearch = () => {
    searchParams.character_id = selectedCharacterId.value
    if (dateRange.value) {
      searchParams.start_date = dateRange.value[0]
      searchParams.end_date = dateRange.value[1]
    } else {
      searchParams.start_date = undefined
      searchParams.end_date = undefined
    }
    getData()
  }

  const handleReset = () => {
    dateRange.value = null
    searchParams.start_date = undefined
    searchParams.end_date = undefined
    getData()
  }

  // ─── 初始化 ───
  const loadCharacters = async () => {
    try {
      characters.value = (await fetchMyCharacters()) ?? []
      // 不自动选中第一个人物，保持默认"全部人物"模式
    } catch {
      characters.value = []
    }
  }

  onMounted(() => {
    loadCharacters()
  })
</script>
