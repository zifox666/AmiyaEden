<template>
  <div class="info-contracts-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSelect
            v-model="filterType"
            :placeholder="$t('info.allContractTypes')"
            clearable
            style="width: 150px"
            @change="applyFilters"
          >
            <ElOption value="item_exchange" :label="$t('info.contractTypeItemExchange')" />
            <ElOption value="auction" :label="$t('info.contractTypeAuction')" />
            <ElOption value="courier" :label="$t('info.contractTypeCourier')" />
            <ElOption value="loan" :label="$t('info.contractTypeLoan')" />
          </ElSelect>
          <ElSelect
            v-model="filterStatus"
            :placeholder="$t('info.allContractStatuses')"
            clearable
            style="width: 150px; margin-left: 8px"
            @change="applyFilters"
          >
            <ElOption value="outstanding" :label="$t('info.contractStatusOutstanding')" />
            <ElOption value="in_progress" :label="$t('info.contractStatusInProgress')" />
            <ElOption value="finished" :label="$t('info.contractStatusFinished')" />
            <ElOption value="cancelled" :label="$t('info.contractStatusCancelled')" />
            <ElOption value="rejected" :label="$t('info.contractStatusRejected')" />
            <ElOption value="failed" :label="$t('info.contractStatusFailed')" />
            <ElOption value="deleted" :label="$t('info.contractStatusDeleted')" />
            <ElOption value="reversed" :label="$t('info.contractStatusReversed')" />
          </ElSelect>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ContractDetailDialog
      v-model:visible="detailVisible"
      :character-id="detailCharacterId"
      :contract-id="detailContractId"
      :contract-type="detailContractType"
      :contract-title="detailContractTitle"
    />
  </div>
</template>

<script setup lang="ts">
  import { ElTag, ElSelect, ElOption } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import { useI18n } from 'vue-i18n'
  import { useUserStore } from '@/store/modules/user'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { fetchInfoContracts } from '@/api/eve-info'
  import ContractDetailDialog from './modules/contract-detail-dialog.vue'

  defineOptions({ name: 'EveInfoContracts' })

  const { t } = useI18n()
  const userStore = useUserStore()

  type ContractItem = Api.EveInfo.ContractItem

  const filterType = ref('')
  const filterStatus = ref('')

  // 璇︽儏寮圭獥鐘舵€?
  const detailVisible = ref(false)
  const detailCharacterId = ref(0)
  const detailContractId = ref(0)
  const detailContractType = ref('')
  const detailContractTitle = ref('')

  const openDetail = (row: ContractItem) => {
    detailCharacterId.value = row.character_id
    detailContractId.value = row.contract_id
    detailContractType.value = row.type
    detailContractTitle.value = row.title ?? `#${row.contract_id}`
    detailVisible.value = true
  }

  const formatISK = (v: number | undefined) => {
    if (v == null) return '-'
    if (v >= 1_000_000_000) return `${(v / 1_000_000_000).toFixed(2)}B`
    if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(2)}M`
    if (v >= 1_000) return `${(v / 1_000).toFixed(2)}K`
    return v.toLocaleString()
  }

  const typeTagConfig: Record<
    string,
    { type: 'primary' | 'success' | 'warning' | 'info' | 'danger' }
  > = {
    item_exchange: { type: 'primary' },
    auction: { type: 'warning' },
    courier: { type: 'success' },
    loan: { type: 'info' },
    unknown: { type: 'info' }
  }

  const statusTagConfig: Record<
    string,
    { type: 'primary' | 'success' | 'warning' | 'info' | 'danger' }
  > = {
    outstanding: { type: 'primary' },
    in_progress: { type: 'warning' },
    finished: { type: 'success' },
    cancelled: { type: 'info' },
    rejected: { type: 'danger' },
    failed: { type: 'danger' },
    deleted: { type: 'info' },
    reversed: { type: 'info' }
  }

  const contractTypeLabels: Record<string, string> = {
    item_exchange: t('info.contractTypeItemExchange'),
    auction: t('info.contractTypeAuction'),
    courier: t('info.contractTypeCourier'),
    loan: t('info.contractTypeLoan'),
    unknown: t('info.contractTypeUnknown')
  }

  const contractStatusLabels: Record<string, string> = {
    outstanding: t('info.contractStatusOutstanding'),
    in_progress: t('info.contractStatusInProgress'),
    finished: t('info.contractStatusFinished'),
    cancelled: t('info.contractStatusCancelled'),
    rejected: t('info.contractStatusRejected'),
    failed: t('info.contractStatusFailed'),
    deleted: t('info.contractStatusDeleted'),
    reversed: t('info.contractStatusReversed')
  }

  const fetchContractsList = (params: Api.EveInfo.ContractsRequest) =>
    fetchInfoContracts({ ...params, language: userStore.language })

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
      apiFn: fetchContractsList,
      apiParams: { current: 1, size: 20, type: '', status: '' } as Api.EveInfo.ContractsRequest,
      columnsFactory: () => [
        { type: 'globalIndex', width: 60, label: '#', fixed: 'left' },
        {
          prop: 'type',
          label: t('info.contractType'),
          width: 130,
          formatter: (row: ContractItem) => {
            const typeConf = typeTagConfig[row.type] ?? { type: 'info' as const }
            const label = contractTypeLabels[row.type] ?? row.type
            return h(ElTag, { type: typeConf.type, size: 'small' }, () => label)
          }
        },
        {
          prop: 'status',
          label: t('info.contractStatus'),
          width: 120,
          formatter: (row: ContractItem) => {
            const statusConf = statusTagConfig[row.status] ?? { type: 'info' as const }
            const label = contractStatusLabels[row.status] ?? row.status
            return h(ElTag, { type: statusConf.type, size: 'small' }, () => label)
          }
        },
        {
          prop: 'title',
          label: t('info.contractTitle'),
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: ContractItem) => h('span', {}, row.title ?? `#${row.contract_id}`)
        },
        {
          prop: 'price',
          label: t('info.contractPrice'),
          width: 120,
          formatter: (row: ContractItem) => h('span', {}, formatISK(row.price))
        },
        {
          prop: 'reward',
          label: t('info.contractReward'),
          width: 120,
          formatter: (row: ContractItem) => h('span', {}, formatISK(row.reward))
        },
        {
          prop: 'character_name',
          label: t('info.owner'),
          width: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'date_expired',
          label: t('info.contractExpiry'),
          width: 180,
          formatter: (row: ContractItem) =>
            h('span', {}, row.date_expired ? new Date(row.date_expired).toLocaleString() : '-')
        },
        {
          prop: 'actions',
          label: t('common.operate'),
          width: 80,
          fixed: 'right',
          formatter: (row: ContractItem) =>
            h(ArtButtonTable, { type: 'view', onClick: () => openDetail(row) })
        }
      ]
    }
  })

  const applyFilters = () => {
    searchParams.type = filterType.value
    searchParams.status = filterStatus.value
    getData()
  }
</script>
