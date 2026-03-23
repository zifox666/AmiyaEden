<!-- 福利审批页面 -->
<template>
  <div class="welfare-approval-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ElTabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 待发放 -->
        <ElTabPane :label="t('welfareApproval.pendingTab')" name="pending">
          <ArtTableHeader
            v-model:columns="pendingColumnChecks"
            :loading="pendingLoading"
            @refresh="loadPending"
          >
            <template #left>
              <span />
            </template>
          </ArtTableHeader>
          <ArtTable
            :loading="pendingLoading"
            :data="pendingData"
            :columns="pendingColumns"
            :pagination="pendingPagination"
            :pagination-options="{ pageSizes: [50, 100, 200] }"
            @pagination:size-change="pendingHandleSizeChange"
            @pagination:current-change="pendingHandleCurrentChange"
          />
          <ElEmpty
            v-if="!pendingLoading && pendingData.length === 0"
            :description="t('welfareApproval.noPending')"
          />
        </ElTabPane>

        <!-- 历史记录 -->
        <ElTabPane :label="t('welfareApproval.historyTab')" name="history">
          <ArtTableHeader
            v-model:columns="historyColumnChecks"
            :loading="historyLoading"
            @refresh="loadHistory"
          >
            <template #left>
              <span />
            </template>
          </ArtTableHeader>
          <ArtTable
            :loading="historyLoading"
            :data="historyData"
            :columns="historyColumns"
            :pagination="historyPagination"
            :pagination-options="{ pageSizes: [50, 100, 200] }"
            @pagination:size-change="historyHandleSizeChange"
            @pagination:current-change="historyHandleCurrentChange"
          />
          <ElEmpty
            v-if="!historyLoading && historyData.length === 0"
            :description="t('welfareApproval.noHistory')"
          />
        </ElTabPane>
      </ElTabs>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElMessage, ElMessageBox, ElEmpty } from 'element-plus'
  import { CopyDocument } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { adminListApplications, adminReviewApplication } from '@/api/welfare'

  defineOptions({ name: 'WelfareApproval' })
  const { t } = useI18n()

  // ─── Tab state ───
  const activeTab = ref('pending')

  type AppRow = Api.Welfare.AdminApplication

  const STATUS_CONFIG = computed(
    () =>
      ({
        requested: { label: t('welfareApproval.statusRequested'), type: 'warning' },
        delivered: { label: t('welfareApproval.statusDelivered'), type: 'success' },
        rejected: { label: t('welfareApproval.statusRejected'), type: 'danger' }
      }) as Record<string, { label: string; type: string }>
  )

  // ─── Copy helper ───
  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success(t('welfareApproval.copied'))
    } catch {
      ElMessage.warning(t('welfareApproval.copyFailed'))
    }
  }

  // ─── Shared column builders ───
  const buildBaseColumns = () => [
    {
      prop: 'applicant_nickname',
      label: t('welfareApproval.applicantNickname'),
      width: 130,
      showOverflowTooltip: true,
      formatter: (row: AppRow) =>
        h(
          'span',
          { class: row.applicant_nickname ? '' : 'text-gray-400' },
          row.applicant_nickname || '-'
        )
    },
    {
      prop: 'character_name',
      label: t('welfareApproval.characterName'),
      minWidth: 160,
      formatter: (row: AppRow) =>
        h('div', { class: 'flex items-center gap-1' }, [
          h('span', {}, row.character_name),
          h(
            ElButton,
            {
              size: 'small',
              icon: CopyDocument,
              type: '' as const,
              onClick: () => copyText(row.character_name)
            },
            () => ''
          )
        ])
    },
    {
      prop: 'contact',
      label: t('welfareApproval.contact'),
      width: 140,
      showOverflowTooltip: true,
      formatter: (row: AppRow) => {
        if (row.qq) return `${t('characters.profile.qq')}: ${row.qq}`
        if (row.discord_id) return `${t('characters.profile.discordId')}: ${row.discord_id}`
        return '-'
      }
    },
    {
      prop: 'welfare_name',
      label: t('welfareApproval.welfareName'),
      width: 160,
      showOverflowTooltip: true
    },
    {
      prop: 'welfare_description',
      label: t('welfareApproval.welfareDescription'),
      minWidth: 180,
      showOverflowTooltip: true,
      formatter: (row: AppRow) => row.welfare_description || '-'
    },
    {
      prop: 'reviewer_name',
      label: t('welfareApproval.reviewerName'),
      width: 130,
      showOverflowTooltip: true,
      formatter: (row: AppRow) =>
        h('span', { class: row.reviewer_name ? '' : 'text-gray-400' }, row.reviewer_name || '-')
    },
    {
      prop: 'created_at',
      label: t('welfareApproval.requestedAt'),
      width: 170,
      formatter: (row: AppRow) => formatTime(row.created_at)
    },
    {
      prop: 'reviewed_at',
      label: t('welfareApproval.processedAt'),
      width: 170,
      formatter: (row: AppRow) => (row.reviewed_at ? formatTime(row.reviewed_at) : '-')
    }
  ]

  // ─── Pending tab ───
  const {
    columns: pendingColumns,
    columnChecks: pendingColumnChecks,
    data: pendingData,
    loading: pendingLoading,
    pagination: pendingPagination,
    handleSizeChange: pendingHandleSizeChange,
    handleCurrentChange: pendingHandleCurrentChange,
    getData: loadPending
  } = useTable({
    core: {
      apiFn: adminListApplications,
      apiParams: { current: 1, size: 50, status: 'requested' },
      columnsFactory: () => [
        ...buildBaseColumns(),
        {
          prop: 'actions',
          label: '',
          width: 160,
          fixed: 'right' as const,
          formatter: (row: AppRow) =>
            h('div', { class: 'flex items-center gap-1' }, [
              h(ArtButtonTable, {
                label: t('welfareApproval.deliverBtn'),
                elType: 'success',
                onClick: () => handleDeliver(row)
              }),
              h(ArtButtonTable, {
                label: t('welfareApproval.rejectBtn'),
                elType: 'danger',
                onClick: () => handleReject(row)
              })
            ])
        }
      ]
    }
  })

  // ─── History tab ───
  const historyLoaded = ref(false)
  const {
    columns: historyColumns,
    columnChecks: historyColumnChecks,
    data: historyData,
    loading: historyLoading,
    pagination: historyPagination,
    handleSizeChange: historyHandleSizeChange,
    handleCurrentChange: historyHandleCurrentChange,
    getData: loadHistory
  } = useTable({
    core: {
      apiFn: adminListApplications,
      apiParams: { current: 1, size: 50, status: 'delivered,rejected' },
      immediate: false,
      columnsFactory: () => [
        ...buildBaseColumns(),
        {
          prop: 'status',
          label: t('welfareApproval.status'),
          width: 110,
          formatter: (row: AppRow) => {
            const cfg = STATUS_CONFIG.value[row.status] ?? { label: row.status, type: 'info' }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        }
      ]
    }
  })

  // ─── Actions ───
  const actionLoading = ref(false)

  async function handleDeliver(row: AppRow) {
    try {
      await ElMessageBox.confirm(
        t('welfareApproval.deliverConfirm', { name: row.character_name }),
        { type: 'info' }
      )
    } catch {
      return
    }

    actionLoading.value = true
    try {
      await adminReviewApplication({ id: row.id, action: 'deliver' })
      ElMessage.success(t('welfareApproval.deliverSuccess'))
      loadPending()
    } catch {
      /* handled by interceptor */
    } finally {
      actionLoading.value = false
    }
  }

  async function handleReject(row: AppRow) {
    try {
      await ElMessageBox.confirm(t('welfareApproval.rejectConfirm', { name: row.character_name }), {
        type: 'warning'
      })
    } catch {
      return
    }

    actionLoading.value = true
    try {
      await adminReviewApplication({ id: row.id, action: 'reject' })
      ElMessage.success(t('welfareApproval.rejectSuccess'))
      loadPending()
    } catch {
      /* handled by interceptor */
    } finally {
      actionLoading.value = false
    }
  }

  // ─── Tab switch ───
  function handleTabChange(tab: string | number) {
    if (tab === 'history' && !historyLoaded.value) {
      historyLoaded.value = true
      loadHistory()
    }
  }

  // ─── Helpers ───
  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')
</script>
