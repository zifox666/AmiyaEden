<!-- 我的钱包页面 -->
<template>
  <div class="wallet-page art-full-height">
    <!-- 钱包余额卡片 -->
    <ElCard shadow="never" class="art-card">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-500">{{ $t('shop.myBalance') }}</p>
          <p class="text-3xl font-bold mt-1" :class="balanceColor">
            {{ wallet ? `${formatFuxiCoinAmount(wallet.balance)} ${$t('shop.currency')}` : '-' }}
          </p>
          <p v-if="wallet" class="text-xs text-gray-400 mt-1">
            {{ $t('common.updatedAt') }}: {{ formatTime(wallet.updated_at) }}
          </p>
        </div>
        <ElButton :loading="walletLoading" @click="loadWallet">
          <el-icon class="mr-1"><Refresh /></el-icon>
          {{ $t('common.refresh') }}
        </ElButton>
      </div>
    </ElCard>

    <!-- 交易记录 -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        visual-variant="ledger"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { Refresh } from '@element-plus/icons-vue'
  import { ElCard, ElButton, ElTag } from 'element-plus'
  import { formatFuxiCoinAmount, formatTime } from '@utils/common'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchMyWallet, fetchMyWalletTransactions } from '@/api/sys-wallet'

  defineOptions({ name: 'Wallet' })

  const { t } = useI18n()

  type WalletTransaction = Api.SysWallet.WalletTransaction

  const getRefTypeLabel = (value: string) => {
    const key = `walletAdmin.refTypes.${value}`
    const translated = t(key)
    return translated === key ? value : translated
  }

  // ─── 钱包余额 ───
  const wallet = ref<Api.SysWallet.Wallet | null>(null)
  const walletLoading = ref(false)

  const balanceColor = computed(() => {
    if (!wallet.value) return ''
    return wallet.value.balance >= 0 ? 'text-green-600' : 'text-red-500'
  })

  const loadWallet = async () => {
    walletLoading.value = true
    try {
      wallet.value = await fetchMyWallet()
    } catch {
      wallet.value = null
    } finally {
      walletLoading.value = false
    }
  }

  // ─── 交易记录表格 ───
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
      apiFn: fetchMyWalletTransactions,
      apiParams: { current: 1, size: 200 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'amount',
          label: t('fleet.wallet.amount'),
          width: 140,
          formatter: (row: WalletTransaction) =>
            h(
              'span',
              { class: `font-medium ${row.amount >= 0 ? 'text-green-600' : 'text-red-500'}` },
              `${row.amount >= 0 ? '+' : ''}${formatFuxiCoinAmount(row.amount)} ${t('shop.currency')}`
            )
        },
        {
          prop: 'balance_after',
          label: t('fleet.wallet.balanceAfter'),
          width: 140,
          formatter: (row: WalletTransaction) =>
            h('span', {}, `${formatFuxiCoinAmount(row.balance_after)} ${t('shop.currency')}`)
        },
        {
          prop: 'reason',
          label: t('fleet.wallet.reason'),
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'ref_type',
          label: t('fleet.wallet.refType'),
          minWidth: 120,
          maxWidth: 200,
          formatter: (row: WalletTransaction) =>
            h(ElTag, { size: 'small', effect: 'plain' }, () => getRefTypeLabel(row.ref_type))
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
          label: t('fleet.wallet.time'),
          width: 200,
          formatter: (row: WalletTransaction) => h('span', {}, formatTime(row.created_at))
        }
      ]
    }
  })

  // ─── 初始化 ───
  onMounted(() => {
    loadWallet()
  })
</script>
