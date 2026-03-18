<!-- 订单管理面板 -->
<template>
  <ElCard class="art-table-card" shadow="never">
    <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
      <template #left>
        <div class="flex items-center gap-4">
          <ElInput
            v-model="userIdFilter"
            :placeholder="$t('shopAdmin.orders.userIdPlaceholder')"
            clearable
            style="width: 140px"
            @keyup.enter="handleSearch"
          />
          <ElSelect
            v-model="statusFilter"
            :placeholder="$t('shopAdmin.orders.statusPlaceholder')"
            clearable
            style="width: 140px"
            @change="handleSearch"
          >
            <ElOption :label="$t('shopAdmin.orders.status.pending')" value="pending" />
            <ElOption :label="$t('shopAdmin.orders.status.completed')" value="completed" />
            <ElOption :label="$t('shopAdmin.orders.status.rejected')" value="rejected" />
            <ElOption
              :label="$t('shopAdmin.orders.status.insufficient_funds')"
              value="insufficient_funds"
            />
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

  <!-- 审批备注对话框 -->
  <ElDialog
    v-model="reviewDialogVisible"
    :title="
      reviewAction === 'approve'
        ? $t('shopAdmin.orders.dialogApprove')
        : $t('shopAdmin.orders.dialogReject')
    "
    width="400px"
    destroy-on-close
  >
    <ElForm label-width="80px">
      <ElFormItem :label="$t('shopAdmin.orders.fields.orderNo')">
        <span class="font-medium">{{ reviewOrderNo }}</span>
      </ElFormItem>
      <ElFormItem :label="$t('shopAdmin.orders.fields.reviewRemark')">
        <ElInput
          v-model="reviewRemark"
          type="textarea"
          :rows="3"
          :placeholder="$t('shopAdmin.orders.placeholders.reviewRemark')"
        />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="reviewDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
      <ElButton
        :type="reviewAction === 'approve' ? 'success' : 'danger'"
        :loading="reviewSubmitting"
        @click="submitReview"
      >
        {{
          reviewAction === 'approve'
            ? $t('shopAdmin.orders.approveConfirm')
            : $t('shopAdmin.orders.rejectConfirm')
        }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElInput, ElSelect, ElOption, ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { adminListOrders, adminApproveOrder, adminRejectOrder } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'

  defineOptions({ name: 'ManageOrders' })
  const { t } = useI18n()

  type Order = Api.Shop.Order

  // ─── 订单状态映射 ───
  const ORDER_STATUS_CONFIG: Record<string, { label: string; type: string }> = {
    pending: { label: t('shopAdmin.orders.status.pending'), type: 'warning' },
    paid: { label: t('shopAdmin.orders.status.paid'), type: 'success' },
    approved: { label: t('shopAdmin.orders.status.approved'), type: 'success' },
    rejected: { label: t('shopAdmin.orders.status.rejected'), type: 'danger' },
    completed: { label: t('shopAdmin.orders.status.completed'), type: 'success' },
    cancelled: { label: t('shopAdmin.orders.status.cancelled'), type: 'info' },
    insufficient_funds: { label: t('shopAdmin.orders.status.insufficient_funds'), type: 'danger' }
  }

  const formatISK = (v: number) =>
    v.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

  const formatTime = (t: string) => new Date(t).toLocaleString()

  // ─── 搜索过滤状态 ───
  const userIdFilter = ref('')
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
      apiFn: adminListOrders,
      apiParams: { current: 1, size: 20 },
      immediate: false,
      columnsFactory: () => [
        {
          prop: 'order_no',
          label: t('shopAdmin.orders.table.orderNo'),
          width: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'user_id',
          label: t('shopAdmin.orders.table.userId'),
          width: 90
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
          width: 130,
          formatter: (row: Order) =>
            h('span', { class: 'font-medium text-orange-600' }, formatISK(row.total_price))
        },
        {
          prop: 'status',
          label: t('common.status'),
          width: 120,
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
          prop: 'remark',
          label: t('shopAdmin.orders.table.userRemark'),
          width: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'review_remark',
          label: t('shopAdmin.orders.table.reviewRemark'),
          width: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'created_at',
          label: t('shopAdmin.orders.table.createdAt'),
          width: 180,
          formatter: (row: Order) => h('span', {}, formatTime(row.created_at))
        },
        {
          prop: 'actions',
          label: t('common.operation'),
          width: 160,
          fixed: 'right',
          formatter: (row: Order) => {
            if (row.status !== 'pending') {
              return h('span', { class: 'text-gray-400 text-sm' }, '-')
            }
            return h('div', { class: 'flex gap-1' }, [
              h(
                ElButton,
                { size: 'small', type: 'success', onClick: () => openApproveDialog(row) },
                () => t('shopAdmin.orders.approveButton')
              ),
              h(
                ElButton,
                { size: 'small', type: 'danger', onClick: () => openRejectDialog(row) },
                () => t('shopAdmin.orders.rejectButton')
              )
            ])
          }
        }
      ]
    }
  })

  function handleSearch() {
    Object.assign(searchParams, {
      status: statusFilter.value || undefined,
      user_id: userIdFilter.value ? Number(userIdFilter.value) : undefined,
      current: 1
    })
    getData()
  }

  function handleReset() {
    userIdFilter.value = ''
    statusFilter.value = ''
    resetSearchParams()
  }

  // ─── 审批对话框状态 ───
  const reviewDialogVisible = ref(false)
  const reviewAction = ref<'approve' | 'reject'>('approve')
  const reviewOrderId = ref(0)
  const reviewOrderNo = ref('')
  const reviewRemark = ref('')
  const reviewSubmitting = ref(false)

  function openApproveDialog(order: Order) {
    reviewAction.value = 'approve'
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
      if (reviewAction.value === 'approve') {
        await adminApproveOrder(params)
        ElMessage.success(t('shopAdmin.orders.messages.approveSuccess'))
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
