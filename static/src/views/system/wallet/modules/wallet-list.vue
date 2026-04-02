<!-- 钱包列表子模块 -->
<template>
  <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
    <template #left>
      <ElButton type="success" @click="emit('adjust', 0, 'add')">
        {{ $t('walletAdmin.adjustBalance') }}
      </ElButton>
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
  import { useI18n } from 'vue-i18n'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { adminListWallets } from '@/api/sys-wallet'
  import { formatFuxiCoinAmount, formatTime } from '@utils/common'

  defineOptions({ name: 'WalletList' })
  const { t } = useI18n()

  type Wallet = Api.SysWallet.Wallet

  const emit = defineEmits<{
    (e: 'adjust', userId: number, action: 'add' | 'deduct' | 'set'): void
    (e: 'viewTransactions', userId: number): void
  }>()

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: adminListWallets,
      apiParams: { current: 1, size: 200 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        { prop: 'user_id', label: '用户 ID', width: 100 },
        {
          prop: 'character_name',
          label: '主人物',
          minWidth: 160,
          formatter: (row: Wallet) => h('span', {}, row.character_name || '-')
        },
        {
          prop: 'balance',
          label: t('walletAdmin.wallets.balance'),
          minWidth: 180,
          formatter: (row: Wallet) =>
            h(
              'span',
              { class: row.balance >= 0 ? 'text-green-600 font-bold' : 'text-red-500 font-bold' },
              formatFuxiCoinAmount(row.balance)
            )
        },
        {
          prop: 'updated_at',
          label: t('common.updatedAt'),
          minWidth: 200,
          formatter: (row: Wallet) => h('span', {}, formatTime(row.updated_at))
        },
        {
          prop: 'actions',
          label: t('common.operation'),
          width: 220,
          fixed: 'right',
          formatter: (row: Wallet) =>
            h('div', { class: 'flex gap-1' }, [
              h(ArtButtonTable, {
                label: t('walletAdmin.actions.add'),
                elType: 'success',
                onClick: () => emit('adjust', row.user_id, 'add')
              }),
              h(ArtButtonTable, {
                label: t('walletAdmin.actions.deduct'),
                elType: 'warning',
                onClick: () => emit('adjust', row.user_id, 'deduct')
              }),
              h(ArtButtonTable, {
                label: t('walletAdmin.actions.transactions'),
                elType: 'primary',
                onClick: () => emit('viewTransactions', row.user_id)
              })
            ])
        }
      ]
    }
  })

  defineExpose({ refreshData })
</script>
