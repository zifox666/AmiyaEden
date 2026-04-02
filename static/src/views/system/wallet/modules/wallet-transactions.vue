<!-- 钱包流水查询子模块 -->
<template>
  <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
    <template #left>
      <ElInput
        v-model="filterForm.user_keyword"
        :placeholder="$t('walletAdmin.placeholders.userKeywordFilter')"
        clearable
        style="width: 240px"
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
        <ElOption :label="$t('walletAdmin.refTypes.pap_fc_salary')" value="pap_fc_salary" />
        <ElOption :label="$t('walletAdmin.refTypes.admin_adjust')" value="admin_adjust" />
        <ElOption :label="$t('walletAdmin.refTypes.manual')" value="manual" />
        <ElOption :label="$t('walletAdmin.refTypes.redeem')" value="redeem" />
        <ElOption :label="$t('walletAdmin.refTypes.srp_payout')" value="srp_payout" />
        <ElOption :label="$t('walletAdmin.refTypes.welfare_payout')" value="welfare_payout" />
        <ElOption :label="$t('walletAdmin.refTypes.shop_purchase')" value="shop_purchase" />
        <ElOption :label="$t('walletAdmin.refTypes.shop_refund')" value="shop_refund" />
        <ElOption
          :label="$t('walletAdmin.refTypes.newbro_captain_reward')"
          value="newbro_captain_reward"
        />
        <ElOption :label="$t('walletAdmin.refTypes.mentor_reward')" value="mentor_reward" />
      </ElSelect>
      <ElButton type="primary" @click="handleSearch">{{ $t('common.search') }}</ElButton>
    </template>
  </ArtTableHeader>

  <ArtTable
    :loading="loading"
    :data="data"
    :columns="columns"
    :pagination="pagination"
    visual-variant="ledger"
    @pagination:size-change="handleSizeChange"
    @pagination:current-change="handleCurrentChange"
  />
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElInput, ElSelect, ElOption } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { formatFuxiCoinAmount, formatTime } from '@utils/common'
  import { useTable } from '@/hooks/core/useTable'
  import { adminListTransactions } from '@/api/sys-wallet'

  defineOptions({ name: 'WalletTransactions' })
  const { t } = useI18n()

  type WalletTransaction = Api.SysWallet.WalletTransaction

  const REF_TYPE_MAP: Record<string, { label: string; tag: string }> = {
    pap_reward: { label: t('walletAdmin.refTypes.pap_reward'), tag: 'success' },
    pap_fc_salary: { label: t('walletAdmin.refTypes.pap_fc_salary'), tag: 'success' },
    admin_adjust: { label: t('walletAdmin.refTypes.admin_adjust'), tag: 'warning' },
    manual: { label: t('walletAdmin.refTypes.manual'), tag: '' },
    redeem: { label: t('walletAdmin.refTypes.redeem'), tag: 'danger' },
    srp_payout: { label: t('walletAdmin.refTypes.srp_payout'), tag: 'primary' },
    welfare_payout: { label: t('walletAdmin.refTypes.welfare_payout'), tag: 'success' },
    shop_purchase: { label: t('walletAdmin.refTypes.shop_purchase'), tag: 'info' },
    shop_refund: { label: t('walletAdmin.refTypes.shop_refund'), tag: 'warning' },
    newbro_captain_reward: {
      label: t('walletAdmin.refTypes.newbro_captain_reward'),
      tag: 'success'
    },
    mentor_reward: {
      label: t('walletAdmin.refTypes.mentor_reward'),
      tag: 'success'
    }
  }
  const getRefTypeLabel = (t: string) => REF_TYPE_MAP[t]?.label ?? t
  const getRefTypeTag = (t: string): any => REF_TYPE_MAP[t]?.tag ?? 'info'

  const filterForm = reactive({
    user_id: undefined as number | undefined,
    user_keyword: '',
    ref_type: ''
  })

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
      apiParams: { current: 1, size: 200 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        { prop: 'user_id', label: t('walletAdmin.transactions.userId'), width: 90 },
        {
          prop: 'character_name',
          label: t('walletAdmin.transactions.characterName'),
          minWidth: 140,
          formatter: (row: WalletTransaction) => h('span', {}, row.character_name || '-')
        },
        {
          label: t('walletAdmin.transactions.amount'),
          width: 140,
          formatter: (row: WalletTransaction) =>
            h(
              'span',
              {
                class: `font-medium ${row.amount >= 0 ? 'text-green-600' : 'text-red-500'}`
              },
              `${row.amount >= 0 ? '+' : ''}${formatFuxiCoinAmount(row.amount)}`
            )
        },
        {
          prop: 'balance_after',
          label: t('walletAdmin.transactions.balanceAfter'),
          width: 140,
          formatter: (row: WalletTransaction) =>
            h('span', {}, formatFuxiCoinAmount(row.balance_after))
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
          minWidth: 120,
          formatter: (row: WalletTransaction) =>
            h(
              'span',
              {},
              row.operator_id === 0
                ? t('walletAdmin.actions.system')
                : row.operator_name || `#${row.operator_id}`
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
      user_id: filterForm.user_id,
      user_keyword: filterForm.user_keyword || undefined,
      ref_type: filterForm.ref_type || undefined
    })
    getData()
  }

  const filterByUser = (userId: number) => {
    filterForm.user_id = userId
    filterForm.user_keyword = ''
    handleSearch()
  }

  defineExpose({ filterByUser })
</script>
