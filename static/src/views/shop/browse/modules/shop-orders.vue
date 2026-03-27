<!-- 我的订单面板 -->
<template>
  <ElCard class="art-table-card" shadow="never">
    <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
      <template #left>
        <ElSelect
          v-model="statusFilter"
          :placeholder="$t('shop.allStatuses')"
          clearable
          style="width: 160px"
          @change="handleStatusChange"
        >
          <ElOption :label="$t('shopAdmin.orders.status.requested')" value="requested" />
          <ElOption :label="$t('shopAdmin.orders.status.delivered')" value="delivered" />
          <ElOption :label="$t('shopAdmin.orders.status.rejected')" value="rejected" />
        </ElSelect>
      </template>
    </ArtTableHeader>

    <ArtTable
      :loading="loading"
      :data="data"
      :columns="columns"
      :pagination="pagination"
      :empty-text="$t('shop.noOrders')"
      @pagination:size-change="handleSizeChange"
      @pagination:current-change="handleCurrentChange"
    />
  </ElCard>
</template>

<script setup lang="ts">
  import { ElTag, ElSelect, ElOption } from 'element-plus'
  import { fetchMyOrders } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'
  import { useI18n } from 'vue-i18n'
  import { formatTime } from '@utils/common'

  defineOptions({ name: 'ShopOrders' })

  const { t } = useI18n()

  type Order = Api.Shop.Order

  // ─── 订单状态映射 ───
  const ORDER_STATUS_MAP: Record<string, { label: string; type: string }> = {
    requested: { label: t('shopAdmin.orders.status.requested'), type: 'warning' },
    delivered: { label: t('shopAdmin.orders.status.delivered'), type: 'success' },
    rejected: { label: t('shopAdmin.orders.status.rejected'), type: 'danger' }
  }

  const formatISK = (v: number) => Math.round(v).toLocaleString('en-US')

  const statusFilter = ref<string | undefined>(undefined)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    searchParams,
    getData,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchMyOrders,
      apiParams: { current: 1, size: 20, status: undefined as string | undefined },
      immediate: false,
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'order_no',
          label: t('shop.orderNo'),
          width: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'product_name',
          label: t('shop.productName'),
          minWidth: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'quantity',
          label: t('shop.quantity'),
          width: 80
        },
        {
          prop: 'unit_price',
          label: t('shop.unitPrice'),
          width: 140,
          formatter: (row: Order) =>
            h('span', {}, `${formatISK(row.unit_price)} ${t('shop.currency')}`)
        },
        {
          prop: 'total_price',
          label: t('shop.totalPrice'),
          width: 140,
          formatter: (row: Order) =>
            h(
              'span',
              { class: 'font-medium text-red-500' },
              `${formatISK(row.total_price)} ${t('shop.currency')}`
            )
        },
        {
          prop: 'status',
          label: t('shop.status'),
          width: 120,
          formatter: (row: Order) => {
            const cfg = ORDER_STATUS_MAP[row.status] ?? { label: row.status, type: 'info' }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        },
        {
          prop: 'created_at',
          label: t('shop.orderTime'),
          width: 180,
          formatter: (row: Order) => h('span', {}, formatTime(row.created_at))
        }
      ]
    }
  })

  function handleStatusChange() {
    searchParams.status = statusFilter.value || undefined
    getData()
  }

  // 供父页面按需触发首次加载
  defineExpose({ load: getData, refresh: refreshData })
</script>
