<template>
  <div class="info-wallet-page art-full-height">
    <!-- 人物切换器 + 余额展示（特殊上下文选择器，非标准搜索栏） -->
    <ElCard class="art-card" shadow="never">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-4 flex-wrap">
          <span class="text-sm text-gray-500">{{ $t('info.selectCharacter') }}</span>
          <ElSelect
            v-model="selectedCharacterId"
            :placeholder="$t('info.selectCharacterPlaceholder')"
            style="width: 240px"
            @update:model-value="onCharacterChange"
          >
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
          <span class="text-sm text-gray-500">{{ $t('info.journalTypeFilter') }}</span>
          <ElSelect
            v-model="selectedRefTypes"
            multiple
            filterable
            clearable
            collapse-tags
            collapse-tags-tooltip
            :placeholder="$t('info.journalTypeFilterPlaceholder')"
            style="width: 320px"
            @update:model-value="onJournalTypeChange"
          >
            <ElOption
              v-for="type in journalTypeOptions"
              :key="type.value"
              :value="type.value"
              :label="type.label"
            />
          </ElSelect>
        </div>
        <div v-if="walletBalance !== null">
          <p class="text-sm text-gray-500">{{ $t('info.walletBalance') }}</p>
          <p
            class="text-2xl font-bold mt-1"
            :class="walletBalance >= 0 ? 'text-green-600' : 'text-red-500'"
          >
            {{ formatISK(walletBalance) }} ISK
          </p>
        </div>
      </div>
    </ElCard>

    <!-- 流水表格 -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        visual-variant="ledger"
        :empty-text="$t('info.noJournalData')"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElAvatar, ElSelect, ElOption } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchInfoWallet } from '@/api/eve-info'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'EveInfoWallet' })

  type WalletJournal = Api.EveInfo.WalletJournal

  const { t } = useI18n()

  // ─── 余额（从 API 响应中捕获） ───
  const walletBalance = ref<number | null>(null)

  // ─── ISK 格式化 ───
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  // ─── API 适配器：标准化非标准响应并捕获余额 ───
  const fetchWalletJournalList = async (params: {
    character_id?: number
    ref_types?: string[]
    current: number
    size: number
  }): Promise<Api.Common.PaginatedResponse<WalletJournal>> => {
    if (!params.character_id) {
      return { list: [], total: 0, page: 1, pageSize: params.size }
    }
    const res = await fetchInfoWallet({
      character_id: params.character_id,
      ref_types: params.ref_types?.length ? params.ref_types : undefined,
      page: params.current,
      page_size: params.size
    })
    walletBalance.value = res?.balance ?? null
    walletJournalTypes.value = res?.ref_types ?? []
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
      apiFn: fetchWalletJournalList,
      apiParams: {
        character_id: undefined as number | undefined,
        ref_types: [] as string[],
        current: 1,
        size: 200
      },
      immediate: false,
      columnsFactory: () => [
        {
          prop: 'date',
          label: t('info.journalDate'),
          width: 180
        },
        {
          prop: 'ref_type',
          label: t('info.journalType'),
          width: 180,
          formatter: (row: WalletJournal) =>
            h(ElTag, { size: 'small', effect: 'plain' }, () => formatJournalTypeLabel(row.ref_type))
        },
        {
          prop: 'amount',
          label: t('info.journalAmount'),
          width: 160,
          formatter: (row: WalletJournal) =>
            h(
              'span',
              { class: `font-medium ${row.amount >= 0 ? 'text-green-600' : 'text-red-500'}` },
              `${row.amount >= 0 ? '+' : ''}${formatISK(row.amount)}`
            )
        },
        {
          prop: 'balance',
          label: t('info.journalBalance'),
          width: 160,
          formatter: (row: WalletJournal) =>
            h('span', { class: 'font-medium text-green-600' }, formatISK(row.balance))
        },
        {
          prop: 'description',
          label: t('info.journalDescription'),
          minWidth: 240,
          showOverflowTooltip: true
        },
        {
          prop: 'reason',
          label: t('info.journalReason'),
          minWidth: 160,
          showOverflowTooltip: true
        }
      ]
    }
  })

  // ─── 人物列表 ───
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterId = ref<number>()
  const selectedRefTypes = ref<string[]>([])
  const walletJournalTypes = ref<string[]>([])

  const formatJournalTypeLabel = (value: string) => {
    const npcKey = `npcKill.refTypes.${value}`
    const npcTranslated = t(npcKey)
    if (npcTranslated !== npcKey) return npcTranslated

    const walletAdminKey = `walletAdmin.refTypes.${value}`
    const walletAdminTranslated = t(walletAdminKey)
    if (walletAdminTranslated !== walletAdminKey) return walletAdminTranslated

    const key = `info.wallet.refTypes.${value}`
    const translated = t(key)
    if (translated !== key) return translated

    return value
      .split('_')
      .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
      .join(' ')
  }

  const journalTypeOptions = computed(() =>
    walletJournalTypes.value
      .slice()
      .sort((a, b) => a.localeCompare(b))
      .map((value) => ({
        value,
        label: formatJournalTypeLabel(value)
      }))
  )

  const applyFilters = () => {
    searchParams.character_id = selectedCharacterId.value
    searchParams.ref_types = selectedRefTypes.value.length ? [...selectedRefTypes.value] : undefined
    searchParams.current = 1
    getData()
  }

  const onCharacterChange = () => {
    applyFilters()
  }

  const onJournalTypeChange = () => {
    applyFilters()
  }

  const loadCharacters = async () => {
    try {
      characters.value = (await fetchMyCharacters()) ?? []
      if (characters.value.length > 0) {
        selectedCharacterId.value = characters.value[0].character_id
        applyFilters()
      }
    } catch {
      characters.value = []
    }
  }

  // ─── 初始化 ───
  onMounted(() => {
    loadCharacters()
  })
</script>
