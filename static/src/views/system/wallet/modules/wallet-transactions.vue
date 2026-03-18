<!-- 钱包流水查询子模块 -->
<template>
  <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
    <template #left>
      <ElInput
        v-model="filterForm.user_id"
        :placeholder="$t('walletAdmin.placeholders.userIdFilter')"
        clearable
        style="width: 160px"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <ElSelect
        v-model="filterForm.ref_type"
        :placeholder="$t('walletAdmin.placeholders.refType')"
        clearable
        style="width: 160px"
        @change="handleSearch"
      >
        <ElOption :label="$t('walletAdmin.refTypes.pap_reward')" value="pap_reward" />
        <ElOption :label="$t('walletAdmin.refTypes.admin_adjust')" value="admin_adjust" />
        <ElOption :label="$t('walletAdmin.refTypes.manual')" value="manual" />
        <ElOption :label="$t('walletAdmin.refTypes.redeem')" value="redeem" />
        <ElOption :label="$t('walletAdmin.refTypes.srp_payout')" value="srp_payout" />
        <ElOption :label="$t('walletAdmin.refTypes.shop_purchase')" value="shop_purchase" />
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
  import { adminListTransactions } from '@/api/sys-wallet'

  defineOptions({ name: 'WalletTransactions' })
  const { t } = useI18n()

  type WalletTransaction = Api.SysWallet.WalletTransaction

  const REF_TYPE_MAP: Record<string, { label: string; tag: string }> = {
    pap_reward: { label: t('walletAdmin.refTypes.pap_reward'), tag: 'success' },
    admin_adjust: { label: t('walletAdmin.refTypes.admin_adjust'), tag: 'warning' },
    manual: { label: t('walletAdmin.refTypes.manual'), tag: '' },
    redeem: { label: t('walletAdmin.refTypes.redeem'), tag: 'danger' },
    srp_payout: { label: t('walletAdmin.refTypes.srp_payout'), tag: 'primary' },
    shop_purchase: { label: t('walletAdmin.refTypes.shop_purchase'), tag: 'info' }
  }
  const getRefTypeLabel = (t: string) => REF_TYPE_MAP[t]?.label ?? t
  const getRefTypeTag = (t: string): any => REF_TYPE_MAP[t]?.tag ?? 'info'

  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  const filterForm = reactive({ user_id: '', ref_type: '' })

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
      apiFn: adminListTransactions,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        { prop: 'user_id', label: '用户 ID', width: 90 },
        {
          prop: 'character_name',
          label: '主角色',
          minWidth: 140,
          formatter: (row: WalletTransaction) => h('span', {}, row.character_name || '-')
        },
        {
          label: '金额',
          width: 140,
          formatter: (row: WalletTransaction) =>
            h(
              'span',
              {
                class: `font-medium ${row.amount >= 0 ? 'text-green-600' : 'text-red-500'}`
              },
              `${row.amount >= 0 ? '+' : ''}${formatISK(row.amount)}`
            )
        },
        {
          prop: 'balance_after',
          label: t('walletAdmin.transactions.balanceAfter'),
          width: 140,
          formatter: (row: WalletTransaction) => h('span', {}, formatISK(row.balance_after))
        },
        {
          prop: 'reason',
          label: t('common.reason'),
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'ref_type',
          label: t('common.type'),
          width: 120,
          formatter: (row: WalletTransaction) =>
            h(ElTag, { size: 'small', type: getRefTypeTag(row.ref_type) }, () =>
              getRefTypeLabel(row.ref_type)
            )
        },
        {
          prop: 'operator_id',
          label: t('walletAdmin.transactions.operator'),
          width: 100,
          formatter: (row: WalletTransaction) =>
            h(
              'span',
              {},
              row.operator_id === 0 ? t('walletAdmin.actions.system') : `#${row.operator_id}`
            )
        },
        {
          prop: 'created_at',
          label: t('common.time'),
          width: 200,
          formatter: (row: WalletTransaction) => h('span', {}, formatTime(row.created_at))
        }
      ]
    }
  })

  const handleSearch = () => {
    Object.assign(searchParams, {
      user_id: filterForm.user_id ? Number(filterForm.user_id) : undefined,
      ref_type: filterForm.ref_type || undefined
    })
    getData()
  }

  const filterByUser = (userId: number) => {
    filterForm.user_id = String(userId)
    handleSearch()
  }

  defineExpose({ filterByUser })
</script>
