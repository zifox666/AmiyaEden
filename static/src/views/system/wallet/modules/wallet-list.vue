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
    @pagination:size-change="handleSizeChange"
    @pagination:current-change="handleCurrentChange"
  />
</template>

<script setup lang="ts">
  import { ElButton } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useTable } from '@/hooks/core/useTable'
  import { adminListWallets } from '@/api/sys-wallet'

  defineOptions({ name: 'WalletList' })
  const { t } = useI18n()

  type Wallet = Api.SysWallet.Wallet

  const emit = defineEmits<{
    (e: 'adjust', userId: number, action: 'add' | 'deduct' | 'set'): void
    (e: 'viewTransactions', userId: number): void
  }>()

  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

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
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        { prop: 'user_id', label: '用户 ID', width: 100 },
        {
          prop: 'character_name',
          label: '主角色',
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
              formatISK(row.balance)
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
              h(
                ElButton,
                {
                  size: 'small',
                  type: 'success',
                  onClick: () => emit('adjust', row.user_id, 'add')
                },
                () => t('walletAdmin.actions.add')
              ),
              h(
                ElButton,
                {
                  size: 'small',
                  type: 'warning',
                  onClick: () => emit('adjust', row.user_id, 'deduct')
                },
                () => t('walletAdmin.actions.deduct')
              ),
              h(
                ElButton,
                {
                  size: 'small',
                  type: 'primary',
                  onClick: () => emit('viewTransactions', row.user_id)
                },
                () => t('walletAdmin.actions.transactions')
              )
            ])
        }
      ]
    }
  })

  defineExpose({ refreshData })
</script>
