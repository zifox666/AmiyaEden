<!-- 操作日志子模块 -->
<template>
  <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
    <template #left>
      <ElInput
        v-model="filterForm.target_uid"
        :placeholder="$t('walletAdmin.placeholders.targetUserId')"
        clearable
        style="width: 150px"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <ElInput
        v-model="filterForm.operator_id"
        :placeholder="$t('walletAdmin.placeholders.operatorId')"
        clearable
        style="width: 150px"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <ElSelect
        v-model="filterForm.action"
        :placeholder="$t('walletAdmin.placeholders.action')"
        clearable
        style="width: 130px"
        @change="handleSearch"
      >
        <ElOption :label="$t('walletAdmin.actions.add')" value="add" />
        <ElOption :label="$t('walletAdmin.actions.deduct')" value="deduct" />
        <ElOption :label="$t('walletAdmin.actions.set')" value="set" />
      </ElSelect>
      <ElButton type="primary" @click="handleSearch">{{ $t('common.search') }}</ElButton>
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
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElInput, ElSelect, ElOption } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { adminListWalletLogs } from '@/api/sys-wallet'

  defineOptions({ name: 'WalletLogs' })
  const { t } = useI18n()

  type WalletLog = Api.SysWallet.WalletLog

  const ACTION_MAP: Record<string, { label: string; tag: string }> = {
    add: { label: t('walletAdmin.actions.add'), tag: 'success' },
    deduct: { label: t('walletAdmin.actions.deduct'), tag: 'danger' },
    set: { label: t('walletAdmin.actions.set'), tag: 'warning' }
  }
  const getActionLabel = (a: string) => ACTION_MAP[a]?.label ?? a
  const getActionTag = (a: string): any => ACTION_MAP[a]?.tag ?? 'info'

  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  const filterForm = reactive({ target_uid: '', operator_id: '', action: '' })

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    getData,
    searchParams
  } = useTable({
    core: {
      apiFn: adminListWalletLogs,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'operator_id',
          label: '操作人',
          minWidth: 140,
          formatter: (row: WalletLog) =>
            h(
              'span',
              {},
              row.operator_character_name
                ? `${row.operator_character_name} (#${row.operator_id})`
                : `#${row.operator_id}`
            )
        },
        {
          prop: 'target_uid',
          label: '目标用户',
          minWidth: 140,
          formatter: (row: WalletLog) =>
            h(
              'span',
              {},
              row.target_character_name
                ? `${row.target_character_name} (#${row.target_uid})`
                : `#${row.target_uid}`
            )
        },
        {
          prop: 'action',
          label: t('common.operation'),
          width: 100,
          formatter: (row: WalletLog) =>
            h(ElTag, { size: 'small', type: getActionTag(row.action) }, () =>
              getActionLabel(row.action)
            )
        },
        {
          prop: 'amount',
          label: t('common.amount'),
          width: 140,
          formatter: (row: WalletLog) => h('span', {}, formatISK(row.amount))
        },
        {
          prop: 'before',
          label: t('walletAdmin.logs.before'),
          width: 140,
          formatter: (row: WalletLog) => h('span', {}, formatISK(row.before))
        },
        {
          prop: 'after',
          label: t('walletAdmin.logs.after'),
          width: 140,
          formatter: (row: WalletLog) => h('span', {}, formatISK(row.after))
        },
        {
          prop: 'reason',
          label: t('common.reason'),
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'created_at',
          label: t('common.time'),
          width: 200,
          formatter: (row: WalletLog) => h('span', {}, formatTime(row.created_at))
        }
      ]
    }
  })

  const handleSearch = () => {
    Object.assign(searchParams, {
      target_uid: filterForm.target_uid ? Number(filterForm.target_uid) : undefined,
      operator_id: filterForm.operator_id ? Number(filterForm.operator_id) : undefined,
      action: filterForm.action || undefined
    })
    getData()
  }
</script>
