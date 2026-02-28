<!-- 我的钱包页面 -->
<template>
  <div class="wallet-page art-full-height">
    <!-- 钱包余额卡片 -->
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm text-gray-500">{{ $t('fleet.wallet.balance') }}</p>
          <p class="text-3xl font-bold mt-1" :class="balanceColor">
            {{ wallet ? formatISK(wallet.balance) : '-' }}
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
      <template #header>
        <div class="flex items-center justify-between">
          <span class="card-title">{{ $t('fleet.wallet.transactions') }}</span>
          <ElButton :loading="txLoading" @click="loadTransactions">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
      </template>

      <ElTable v-loading="txLoading" :data="transactions" stripe border style="width: 100%">
        <ElTableColumn type="index" width="60" label="#" />
        <ElTableColumn prop="amount" :label="$t('fleet.wallet.amount')" width="140" align="right">
          <template #default="{ row }">
            <span :class="row.amount >= 0 ? 'text-green-600' : 'text-red-500'" class="font-medium">
              {{ row.amount >= 0 ? '+' : '' }}{{ formatISK(row.amount) }}
            </span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="balance_after" label="余额" width="140" align="right">
          <template #default="{ row }">
            {{ formatISK(row.balance_after) }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="reason" :label="$t('fleet.wallet.reason')" min-width="200" />
        <ElTableColumn prop="ref_type" label="类型" width="120" align="center">
          <template #default="{ row }">
            <ElTag size="small" effect="plain">{{ row.ref_type }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="created_at" :label="$t('fleet.wallet.time')" width="200">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </ElTableColumn>
      </ElTable>

      <ElEmpty v-if="!txLoading && transactions.length === 0" :description="$t('fleet.wallet.empty')" />

      <!-- 分页 -->
      <div v-if="txPagination.total > 0" class="pagination-wrapper">
        <ElPagination
          v-model:current-page="txPagination.current"
          v-model:page-size="txPagination.size"
          :total="txPagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElPagination,
    ElEmpty
  } from 'element-plus'
  import { fetchMyWallet, fetchMyWalletTransactions } from '@/api/sys-wallet'

  defineOptions({ name: 'Wallet' })

  // ---- 数据 ----
  const wallet = ref<Api.SysWallet.Wallet | null>(null)
  const transactions = ref<Api.SysWallet.WalletTransaction[]>([])
  const walletLoading = ref(false)
  const txLoading = ref(false)

  // ---- 分页 ----
  const txPagination = reactive({ current: 1, size: 20, total: 0 })

  // ---- 余额颜色 ----
  const balanceColor = computed(() => {
    if (!wallet.value) return ''
    return wallet.value.balance >= 0 ? 'text-green-600' : 'text-red-500'
  })

  // ---- 格式化 ----
  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  const formatISK = (v: number) => {
    return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)
  }

  // ---- 加载数据 ----
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

  const loadTransactions = async () => {
    txLoading.value = true
    try {
      const res = await fetchMyWalletTransactions({
        current: txPagination.current,
        size: txPagination.size
      })
      if (res) {
        transactions.value = res.list ?? []
        txPagination.total = res.total ?? 0
        txPagination.current = res.page ?? 1
        txPagination.size = res.pageSize ?? 20
      } else {
        transactions.value = []
        txPagination.total = 0
      }
    } catch {
      transactions.value = []
      txPagination.total = 0
    } finally {
      txLoading.value = false
    }
  }

  // ---- 分页 ----
  const handleSizeChange = () => {
    txPagination.current = 1
    loadTransactions()
  }
  const handleCurrentChange = () => {
    loadTransactions()
  }

  // ---- 初始化 ----
  onMounted(() => {
    loadWallet()
    loadTransactions()
  })
</script>

<style scoped>
  .card-title {
    font-size: 15px;
    font-weight: 500;
  }
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
