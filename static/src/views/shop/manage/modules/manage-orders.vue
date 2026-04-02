<!-- 订单管理面板（仅展示待发放订单） -->
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
      @pagination:size-change="handleSizeChange"
      @pagination:current-change="handleCurrentChange"
    />
  </ElCard>

  <!-- 发放备注对话框 -->
  <ElDialog
    v-model="reviewDialogVisible"
    :title="
      reviewAction === 'deliver'
        ? $t('shopAdmin.orders.dialogDeliver')
        : $t('shopAdmin.orders.dialogReject')
    "
    width="400px"
    destroy-on-close
  >
    <ElForm label-width="80px">
      <ElFormItem :label="$t('shopAdmin.orders.fields.orderNo')">
        <span class="font-medium">{{ reviewOrderNo }}</span>
      </ElFormItem>
      <ElFormItem :label="$t('shopAdmin.orders.fields.deliverRemark')">
        <ElInput
          v-model="reviewRemark"
          type="textarea"
          :rows="3"
          :placeholder="$t('shopAdmin.orders.placeholders.deliverRemark')"
        />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="reviewDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
      <ElButton
        :type="reviewAction === 'deliver' ? 'success' : 'danger'"
        :loading="reviewSubmitting"
        @click="submitReview"
      >
        {{
          reviewAction === 'deliver'
            ? $t('shopAdmin.orders.deliverConfirm')
            : $t('shopAdmin.orders.rejectConfirm')
        }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElButton, ElInput, ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { formatFuxiCoinWhole, formatTime } from '@utils/common'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtCopyButton from '@/components/core/forms/art-copy-button/index.vue'
  import { adminListOrders, adminDeliverOrder, adminRejectOrder } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'

  defineOptions({ name: 'ManageOrders' })
  const { t } = useI18n()

  type Order = Api.Shop.Order

  const formatContact = (row: Order) => {
    if (row.qq) return `QQ: ${row.qq}`
    if (row.discord_id) return `Discord: ${row.discord_id}`
    return '-'
  }

  // ─── 搜索过滤状态 ───
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
      apiFn: adminListOrders,
      apiParams: { current: 1, size: 20, statuses: ['requested'] },
      immediate: false,
      columnsFactory: () => [
        {
          prop: 'order_no',
          label: t('shopAdmin.orders.table.orderNo'),
          width: 140,
          formatter: (row: Order) =>
            h('div', { class: 'flex items-center gap-1 min-w-0' }, [
              h('span', { class: 'truncate' }, row.order_no || '-'),
              h(ArtCopyButton, { text: row.order_no })
            ])
        },
        {
          prop: 'main_character_name',
          label: t('shopAdmin.orders.table.mainCharacter'),
          width: 170,
          formatter: (row: Order) =>
            h('div', { class: 'flex items-center gap-1 min-w-0' }, [
              h('span', { class: 'truncate' }, row.main_character_name || '-'),
              h(ArtCopyButton, { text: row.main_character_name })
            ])
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
              `${formatFuxiCoinWhole(row.total_price)} ${t('shop.currency')}`
            )
        },
        {
          prop: 'remark',
          label: t('shopAdmin.orders.table.userRemark'),
          width: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'created_at',
          label: t('shopAdmin.orders.table.createdAt'),
          width: 170,
          formatter: (row: Order) => h('span', {}, formatTime(row.created_at))
        },
        {
          prop: 'actions',
          label: t('common.operation'),
          width: 140,
          fixed: 'right',
          formatter: (row: Order) =>
            h('div', { class: 'flex gap-1' }, [
              h(ArtButtonTable, {
                label: t('shopAdmin.orders.deliverButton'),
                elType: 'success',
                onClick: () => openDeliverDialog(row)
              }),
              h(ArtButtonTable, {
                label: t('shopAdmin.orders.rejectButton'),
                elType: 'danger',
                onClick: () => openRejectDialog(row)
              })
            ])
        }
      ]
    }
  })

  function handleSearch() {
    Object.assign(searchParams, {
      keyword: keywordFilter.value || undefined,
      statuses: ['requested'],
      current: 1
    })
    getData()
  }

  function handleReset() {
    keywordFilter.value = ''
    resetSearchParams()
    Object.assign(searchParams, { statuses: ['requested'] })
  }

  // ─── 发放/拒绝对话框状态 ───
  const reviewDialogVisible = ref(false)
  const reviewAction = ref<'deliver' | 'reject'>('deliver')
  const reviewOrderId = ref(0)
  const reviewOrderNo = ref('')
  const reviewRemark = ref('')
  const reviewSubmitting = ref(false)

  function openDeliverDialog(order: Order) {
    reviewAction.value = 'deliver'
    reviewOrderId.value = order.id
    reviewOrderNo.value = order.order_no
    reviewRemark.value = ''
    reviewDialogVisible.value = true
  }

  function openRejectDialog(order: Order) {
    reviewAction.value = 'reject'
    reviewOrderId.value = order.id
    reviewOrderNo.value = order.order_no
    reviewRemark.value = ''
    reviewDialogVisible.value = true
  }

  async function submitReview() {
    reviewSubmitting.value = true
    try {
      const params: Api.Shop.OrderReviewParams = {
        order_id: reviewOrderId.value,
        remark: reviewRemark.value
      }
      if (reviewAction.value === 'deliver') {
        await adminDeliverOrder(params)
        ElMessage.success(t('shopAdmin.orders.messages.deliverSuccess'))
      } else {
        await adminRejectOrder(params)
        ElMessage.success(t('shopAdmin.orders.messages.rejectSuccess'))
      }
      reviewDialogVisible.value = false
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('shopAdmin.orders.messages.actionFailed'))
    } finally {
      reviewSubmitting.value = false
    }
  }

  defineExpose({ load: getData, refresh: refreshData })
</script>
