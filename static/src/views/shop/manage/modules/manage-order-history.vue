<!-- 订单历史面板（已发放 / 已拒绝） -->
<template>
  <ElCard class="art-table-card" shadow="never">
    <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
      <template #left>
        <div class="flex items-center gap-4">
          <ElInput
            v-model="keywordFilter"
            :placeholder="$t('shopAdmin.orders.keywordPlaceholder')"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
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
      :pagination-options="{ pageSizes: [200, 500, 1000] }"
      @pagination:size-change="handleSizeChange"
      @pagination:current-change="handleCurrentChange"
    />
  </ElCard>
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElInput } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { formatTime } from '@utils/common'
  import { adminListOrderHistory } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'

  defineOptions({ name: 'ManageOrderHistory' })
  const { t } = useI18n()

  type Order = Api.Shop.Order

  const ORDER_STATUS_CONFIG: Record<string, { label: string; type: string }> = {
    delivered: { label: t('shopAdmin.orders.status.delivered'), type: 'success' },
    rejected: { label: t('shopAdmin.orders.status.rejected'), type: 'danger' }
  }

  const formatISK = (v: number) => Math.round(v).toLocaleString('en-US')
  const formatContact = (row: Order) => {
    if (row.qq) return `QQ: ${row.qq}`
    if (row.discord_id) return `Discord: ${row.discord_id}`
    return '-'
  }

  const keywordFilter = ref('')

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
      apiFn: adminListOrderHistory,
      apiParams: { current: 1, size: 200 },
      immediate: false,
      columnsFactory: () => [
        {
          prop: 'order_no',
          label: t('shopAdmin.orders.table.orderNo'),
          width: 120,
          showOverflowTooltip: true
        },
        {
          prop: 'main_character_name',
          label: t('shopAdmin.orders.table.mainCharacter'),
          width: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'nickname',
          label: t('shopAdmin.orders.table.nickname'),
          width: 120,
          showOverflowTooltip: true,
          formatter: (row: Order) => h('span', {}, row.nickname || '-')
        },
        {
          prop: 'contact',
          label: t('shopAdmin.orders.table.contact'),
          width: 160,
          showOverflowTooltip: true,
          formatter: (row: Order) => h('span', {}, formatContact(row))
        },
        {
          prop: 'product_name',
          label: t('shopAdmin.orders.table.product'),
          minWidth: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'quantity',
          label: t('shopAdmin.orders.table.quantity'),
          width: 70
        },
        {
          prop: 'total_price',
          label: t('shopAdmin.orders.table.totalPrice'),
          width: 120,
          formatter: (row: Order) =>
            h(
              'span',
              { class: 'font-medium text-orange-600' },
              `${formatISK(row.total_price)} ${t('shop.currency')}`
            )
        },
        {
          prop: 'status',
          label: t('common.status'),
          width: 100,
          formatter: (row: Order) => {
            const cfg = ORDER_STATUS_CONFIG[row.status] ?? { label: row.status, type: 'info' }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        },
        {
          prop: 'review_remark',
          label: t('shopAdmin.orders.fields.deliverRemark'),
          width: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'created_at',
          label: t('shopAdmin.orders.table.createdAt'),
          width: 170,
          formatter: (row: Order) => h('span', {}, formatTime(row.created_at))
        }
      ]
    }
  })

  function handleSearch() {
    Object.assign(searchParams, {
      keyword: keywordFilter.value || undefined,
      current: 1
    })
    getData()
  }

  function handleReset() {
    keywordFilter.value = ''
    resetSearchParams()
  }

  defineExpose({ load: getData, refresh: refreshData })
</script>
