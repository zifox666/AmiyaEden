<!-- 兑换码管理面板 -->
<template>
  <ElCard class="art-table-card" shadow="never">
    <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
      <template #left>
        <div class="flex items-center gap-4">
          <ElInput
            v-model="productIdFilter"
            :placeholder="$t('shopAdmin.redeem.productIdPlaceholder')"
            clearable
            style="min-width: 140px"
            @keyup.enter="handleSearch"
          />
          <ElSelect
            v-model="statusFilter"
            :placeholder="$t('shopAdmin.redeem.statusPlaceholder')"
            clearable
            style="min-width: 120px"
            @change="handleSearch"
          >
            <ElOption :label="$t('shopAdmin.redeem.status.unused')" value="unused" />
            <ElOption :label="$t('shopAdmin.redeem.status.used')" value="used" />
            <ElOption :label="$t('shopAdmin.redeem.status.expired')" value="expired" />
          </ElSelect>
          <ElButton type="primary" @click="handleSearch">{{ $t('common.search') }}</ElButton>
          <ElButton @click="handleReset">{{ $t('common.reset') }}</ElButton>
        </div>
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
</template>

<script setup lang="ts">
  import { ElTag, ElInput, ElSelect, ElOption, ElButton } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { adminListRedeemCodes } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'

  defineOptions({ name: 'ManageRedeem' })
  const { t } = useI18n()

  type RedeemCode = Api.Shop.RedeemCode

  // ─── 兑换码状态映射 ───
  const REDEEM_STATUS_CONFIG: Record<string, { label: string; type: string }> = {
    unused: { label: t('shopAdmin.redeem.status.unused'), type: 'success' },
    used: { label: t('shopAdmin.redeem.status.used'), type: 'info' },
    expired: { label: t('shopAdmin.redeem.status.expired'), type: 'danger' }
  }

  const formatTime = (v: string | null) => (v ? new Date(v).toLocaleString() : '-')

  // ─── 搜索过滤状态 ───
  const productIdFilter = ref('')
  const statusFilter = ref('')

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    searchParams,
    getData,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: adminListRedeemCodes,
      apiParams: { current: 1, size: 20 },
      immediate: false,
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'product_id',
          label: t('shopAdmin.redeem.table.productId'),
          minWidth: 90
        },
        {
          prop: 'user_id',
          label: t('shopAdmin.redeem.table.userId'),
          minWidth: 90
        },
        {
          prop: 'order_id',
          label: t('shopAdmin.redeem.table.orderId'),
          minWidth: 90
        },
        {
          prop: 'code',
          label: t('shopAdmin.redeem.table.code'),
          minWidth: 220,
          showOverflowTooltip: true,
          formatter: (row: RedeemCode) => h('code', { class: 'text-sm font-mono' }, row.code)
        },
        {
          prop: 'status',
          label: t('common.status'),
          minWidth: 100,
          formatter: (row: RedeemCode) => {
            const cfg = REDEEM_STATUS_CONFIG[row.status] ?? { label: row.status, type: 'info' }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        },
        {
          prop: 'created_at',
          label: t('shopAdmin.redeem.table.createdAt'),
          minWidth: 180,
          formatter: (row: RedeemCode) => h('span', {}, formatTime(row.created_at))
        },
        {
          prop: 'expires_at',
          label: t('shopAdmin.redeem.table.expiresAt'),
          minWidth: 180,
          formatter: (row: RedeemCode) => h('span', {}, formatTime(row.expires_at))
        }
      ]
    }
  })

  function handleSearch() {
    Object.assign(searchParams, {
      status: statusFilter.value || undefined,
      product_id: productIdFilter.value ? Number(productIdFilter.value) : undefined,
      current: 1
    })
    getData()
  }

  function handleReset() {
    productIdFilter.value = ''
    statusFilter.value = ''
    resetSearchParams()
  }

  defineExpose({ load: getData, refresh: refreshData })
</script>
