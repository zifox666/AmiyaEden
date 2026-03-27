<!-- SRP 补损审批管理页面 -->
<template>
  <div class="srp-manage-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ElTabs v-model="activeTab" @tab-change="handleTabChange">
        <ElTabPane :label="$t('srp.manage.pendingTab')" name="pending" />
        <ElTabPane :label="$t('srp.manage.historyTab')" name="history" />
      </ElTabs>

        <div class="flex items-center gap-3 flex-wrap mb-3">
          <ElSelect
            v-model="filter.review_status"
            :placeholder="$t('srp.apply.columns.reviewStatus')"
            clearable
            style="width: 130px"
            @change="handleSearch"
          >
            <template v-if="activeTab === 'pending'">
              <ElOption :label="$t('srp.status.submitted')" value="submitted" />
              <ElOption :label="$t('srp.status.approved')" value="approved" />
            </template>
            <template v-else>
              <ElOption :label="$t('srp.status.approved')" value="approved" />
              <ElOption :label="$t('srp.status.rejected')" value="rejected" />
            </template>
          </ElSelect>
          <ElSelect
            v-model="filter.fleet_id"
            :placeholder="$t('srp.manage.selectFleet')"
            clearable
            filterable
            style="width: 220px"
            @change="handleSearch"
          >
            <ElOption v-for="f in fleets" :key="f.id" :label="formatFleetLabel(f)" :value="f.id" />
          </ElSelect>
          <ElButton type="primary" @click="handleSearch">{{ $t('srp.manage.searchBtn') }}</ElButton>
          <ElButton @click="resetFilter">{{ $t('srp.manage.resetBtn') }}</ElButton>
        </div>

        <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
          <template #left>
            <div class="flex items-center gap-2">
              <template v-if="activeTab === 'pending' && canPayout">
                <ElButton type="primary" @click="handleAutoApprove">
                  {{ $t('srp.manage.autoApproveBtn') }}
                </ElButton>
                <ElButton type="warning" @click="openBatchPayoutDialog">
                  {{ $t('srp.manage.batchPayoutBtn') }}
                </ElButton>
              </template>
              <ArtExcelExport
                v-if="activeTab === 'history'"
                :data="exportManageData"
                :headers="manageExportHeaders"
                :filename="`srp-manage_${new Date().toLocaleDateString()}`"
                sheet-name="补损申请"
                :button-text="$t('srp.manage.exportBtn')"
                type="success"
              />
            </div>
          </template>
          <template v-else>
            <ElOption :label="$t('srp.status.approved')" value="approved" />
            <ElOption :label="$t('srp.status.rejected')" value="rejected" />
          </template>
        </ElSelect>
        <ElSelect
          v-model="filter.fleet_id"
          :placeholder="$t('srp.manage.selectFleet')"
          clearable
          filterable
          style="width: 220px"
          @change="handleSearch"
        >
          <ElOption v-for="f in fleets" :key="f.id" :label="formatFleetLabel(f)" :value="f.id" />
        </ElSelect>
        <ElButton type="primary" @click="handleSearch">{{ $t('srp.manage.searchBtn') }}</ElButton>
        <ElButton @click="resetFilter">{{ $t('srp.manage.resetBtn') }}</ElButton>
      </div>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <div class="flex items-center gap-2">
            <template v-if="activeTab === 'pending'">
              <ElButton type="primary" @click="handleAutoApprove">
                {{ $t('srp.manage.autoApproveBtn') }}
              </ElButton>
              <ElButton type="warning" @click="openBatchPayoutDialog">
                {{ $t('srp.manage.batchPayoutBtn') }}
              </ElButton>
            </template>
            <ArtExcelExport
              v-if="activeTab === 'history'"
              :data="exportManageData"
              :headers="manageExportHeaders"
              :filename="`srp-manage_${new Date().toLocaleDateString()}`"
              sheet-name="补损申请"
              :button-text="$t('srp.manage.exportBtn')"
              type="success"
            />
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

    <!-- 审批弹窗 -->
    <ElDialog
      v-model="reviewDialogVisible"
      :title="
        reviewAction === 'approve' ? $t('srp.manage.approveDialog') : $t('srp.manage.rejectDialog')
      "
      width="460px"
    >
      <ElForm label-width="90px">
        <!-- 申请备注 & 舰队信息（审批前可见） -->
        <template v-if="reviewTarget">
          <ElFormItem :label="$t('srp.manage.columns.note')" v-if="reviewTarget.note">
            <span class="text-sm">{{ reviewTarget.note }}</span>
          </ElFormItem>
          <ElFormItem :label="$t('srp.manage.columns.fleet')" v-if="reviewTarget.fleet_id">
            <ElTooltip :content="reviewTargetFleetLabel" placement="top">
              <span class="font-medium cursor-default">{{
                reviewTarget.fleet_title || reviewTarget.fleet_id
              }}</span>
            </ElTooltip>
          </ElFormItem>
        </template>
        <ElFormItem :label="$t('srp.manage.finalAmount')" v-if="reviewAction === 'approve'">
          <div class="million-isk-input">
            <ElInputNumber
              :model-value="toMillionISKInput(reviewForm.final_amount)"
              :min="0"
              :precision="2"
              :step="1"
              class="million-isk-input__control"
              @update:model-value="updateReviewFinalAmount"
            />
            <span class="million-isk-input__suffix">{{ $t('common.millionIsk') }}</span>
          </div>
          <div class="text-xs text-gray-400 mt-1">{{ $t('srp.manage.finalAmountHint') }}</div>
        </ElFormItem>
        <ElFormItem :label="$t('srp.manage.reviewNote')" :required="reviewAction === 'reject'">
          <ElInput
            v-model="reviewForm.review_note"
            type="textarea"
            :rows="3"
            :placeholder="
              reviewAction === 'reject'
                ? $t('srp.manage.rejectNotePlaceholder')
                : $t('srp.manage.optionalPlaceholder')
            "
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="reviewDialogVisible = false">{{ $t('srp.apply.cancelBtn') }}</ElButton>
        <ElButton
          :type="reviewAction === 'approve' ? 'success' : 'danger'"
          :loading="actionLoading"
          @click="handleReview"
        >
          {{
            reviewAction === 'approve'
              ? $t('srp.manage.confirmApprove')
              : $t('srp.manage.confirmReject')
          }}
        </ElButton>
      </template>
    </ElDialog>

    <!-- 发放弹窗 -->
    <ElDialog v-model="payoutDialogVisible" :title="$t('srp.manage.payoutDialog')" width="480px">
      <div class="payout-info-list" v-if="payoutTarget">
        <!-- 角色（可复制） -->
        <div class="payout-info-row">
          <span class="payout-label">{{ $t('srp.manage.payoutCharacter') }}</span>
          <span class="payout-value">
            <strong>{{ payoutTarget.character_name }}</strong>
            <ElButton
              size="small"
              :icon="CopyDocument"
              type=""
              @click="copyText(payoutTarget!.character_name)"
              class="ml-2"
            >
            </ElButton>
          </span>
        </div>
        <!-- 金额（可复制） -->
        <div class="payout-info-row">
          <span class="payout-label">{{ $t('srp.manage.payoutAmount') }}</span>
          <span class="payout-value">
            <strong>{{ formatISK(payoutTarget.final_amount) }} ISK</strong>
            <ElButton
              size="small"
              :icon="CopyDocument"
              type=""
              @click="copyText(String(payoutTarget!.final_amount))"
              class="ml-2"
            >
            </ElButton>
          </span>
        </div>
        <!-- KillID（可复制） -->
        <div class="payout-info-row">
          <span class="payout-label">{{ $t('srp.manage.payoutKillId') }}</span>
          <span class="payout-value">
            <ElLink
              :href="`https://zkillboard.com/kill/${payoutTarget.killmail_id}/`"
              target="_blank"
              type="primary"
            >
              {{ payoutTarget.killmail_id }}
            </ElLink>
            <ElButton
              size="small"
              :icon="CopyDocument"
              type=""
              @click="copyText(String(payoutTarget!.killmail_id))"
              class="ml-2"
            >
            </ElButton>
          </span>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-between w-full">
          <div>
            <ElButton @click="payoutDialogVisible = false">{{
              $t('srp.apply.cancelBtn')
            }}</ElButton>
            <ElButton type="primary" :loading="actionLoading" @click="handlePayout">{{
              $t('srp.manage.confirmPayout')
            }}</ElButton>
          </div>
        </div>
      </template>
    </ElDialog>

    <ElDialog
      v-model="batchPayoutDialogVisible"
      :title="$t('srp.manage.batchPayoutDialog')"
      width="760px"
    >
      <div class="mb-3 flex items-center gap-2 text-sm text-gray-500">
        <span>{{ $t('srp.manage.batchPayoutHint') }}</span>
        <ElButton
          size="small"
          :icon="CopyDocument"
          :disabled="!batchPayoutList.length"
          @click="copyBatchPayoutListText"
        />
      </div>
      <ElTable
        v-loading="batchSummaryLoading"
        :data="batchPayoutList"
        :empty-text="$t('srp.manage.batchPayoutEmpty')"
        size="small"
      >
        <ElTableColumn
          prop="main_character_name"
          :label="$t('srp.manage.batchPayoutMainCharacter')"
          min-width="170"
        >
          <template #default="{ row }">
            {{ row.main_character_name || $t('srp.manage.unknownMainCharacter') }}
          </template>
        </ElTableColumn>
        <ElTableColumn
          prop="total_amount"
          :label="$t('srp.manage.batchPayoutAmount')"
          min-width="140"
        >
          <template #default="{ row }">
            <span class="font-semibold text-blue-600">{{ formatISK(row.total_amount) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn
          prop="application_count"
          :label="$t('srp.manage.batchPayoutCount')"
          width="110"
          align="center"
        />
        <ElTableColumn :label="$t('srp.manage.columns.action')" width="150" fixed="right">
          <template #default="{ row }">
            <ElButton
              type="primary"
              size="small"
              :loading="batchPayoutLoadingUserId === row.user_id"
              @click="handleBatchPayout(row)"
            >
              {{ $t('srp.manage.confirmPayout') }}
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
      <template #footer>
        <ElButton @click="batchPayoutDialogVisible = false">{{
          $t('srp.apply.cancelBtn')
        }}</ElButton>
      </template>
    </ElDialog>

    <!-- KM 预览弹窗 -->
    <KmPreviewDialog v-model="kmPreviewVisible" :killmail-id="previewKillmailId" />
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { formatTime } from '@utils/common'
  import {
    ElCard,
    ElTag,
    ElButton,
    ElSelect,
    ElOption,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElInput,
    ElLink,
    ElMessageBox,
    ElMessage,
    ElTable,
    ElTableColumn,
    ElTabs,
    ElTabPane,
    ElTooltip
  } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtExcelExport from '@/components/core/forms/art-excel-export/index.vue'
  import KmPreviewDialog from '@/components/business/KmPreviewDialog.vue'
  import { fetchFleetList } from '@/api/fleet'
  import {
    fetchApplicationList,
    runFleetAutoApproval,
    fetchBatchPayoutSummary,
    reviewApplication,
    batchPayoutByUser,
    payoutApplication,
    openInfoWindow
  } from '@/api/srp'
  import { useNameResolver } from '@/hooks'
  import { useUserStore } from '@/store/modules/user'
  import { CopyDocument } from '@element-plus/icons-vue'
  import { fromMillionISKInput, toMillionISKInput } from '@/utils/iskUnits'

  defineOptions({ name: 'SrpManage' })

  const { t } = useI18n()
  const { getName, resolve: resolveNames } = useNameResolver()
  const userStore = useUserStore()

  // fc can review, while srp/admin can also payout or trigger auto-approve
  const canPayout = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((r) => ['super_admin', 'admin', 'srp'].includes(r))
  })

  const fleets = ref<Api.Fleet.FleetItem[]>([])
  const fleetMap = computed(() => new Map(fleets.value.map((f) => [f.id, f])))
  const loadFleets = async () => {
    try {
      const res = await fetchFleetList({ size: 200 } as any)
      fleets.value = res?.list ?? []
    } catch {
      fleets.value = []
    }
  }

  const activeTab = ref('pending')
  const filter = reactive({ review_status: '', fleet_id: '' })

  type SrpApp = Api.Srp.Application
  type BatchPayoutSummary = Api.Srp.BatchPayoutSummary
  type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

  const reviewStatusType = (s: string): TagType =>
    (({ pending: 'info', approved: 'success', rejected: 'danger' }) as Record<string, TagType>)[
      s
    ] ?? 'info'
  const reviewStatusLabel = (s: string) =>
    ({
      submitted: t('srp.status.submitted'),
      approved: t('srp.status.approved'),
      rejected: t('srp.status.rejected')
    })[s as 'submitted' | 'approved' | 'rejected'] ?? s
  const payoutStatusType = (s: string): TagType => (s === 'paid' ? 'success' : 'warning')

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    getData,
    searchParams
  } = useTable({
    core: {
      apiFn: fetchApplicationList,
      apiParams: { current: 1, size: 200, tab: 'pending' },
      columnsFactory: () => [
        { type: 'index', width: 40, label: '#' },
        {
          prop: 'review_status',
          label: t('srp.manage.columns.review'),
          width: 60,
          formatter: (row: SrpApp) => {
            const tag = h(ElTag, { type: reviewStatusType(row.review_status), size: 'small' }, () =>
              reviewStatusLabel(row.review_status)
            )
            if (row.review_note) {
              return h(ElTooltip, { content: row.review_note, placement: 'top' }, () => tag)
            }
            return tag
          }
        },
        {
          prop: 'payout_status',
          label: t('srp.manage.columns.payout'),
          width: 60,
          formatter: (row: SrpApp) =>
            h(ElTag, { type: payoutStatusType(row.payout_status), size: 'small' }, () =>
              row.payout_status === 'paid' ? t('srp.status.paid') : t('srp.status.notpaid')
            )
        },
        {
          prop: 'nickname',
          label: t('srp.manage.columns.nickname'),
          width: 120,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', { class: row.nickname ? '' : 'text-gray-400' }, row.nickname || '-')
        },
        {
          prop: 'character_name',
          label: t('srp.manage.columns.character'),
          width: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'ship_type_id',
          label: t('srp.manage.columns.ship'),
          width: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', {}, getName(row.ship_type_id, `TypeID: ${row.ship_type_id}`, 'type'))
        },
        {
          prop: 'recommended_amount',
          label: t('srp.manage.columns.recommendedAmount'),
          width: 90,
          formatter: (row: SrpApp) => h('span', {}, formatISK(row.recommended_amount))
        },
        {
          prop: 'final_amount',
          label: t('srp.manage.columns.finalAmount'),
          width: 90,
          formatter: (row: SrpApp) =>
            h('span', { class: 'font-semibold text-blue-600' }, formatISK(row.final_amount))
        },
        {
          prop: 'fleet_title',
          label: t('srp.manage.columns.fleet'),
          width: 150,
          formatter: (row: SrpApp) => {
            if (!row.fleet_id) return h('span', { class: 'text-gray-400' }, '-')
            const fleet = fleetMap.value.get(row.fleet_id)
            const tooltipContent = fleet
              ? formatFleetLabel(fleet)
              : row.fleet_fc_name
                ? `${row.fleet_fc_name}: ${row.fleet_title || row.fleet_id}`
                : row.fleet_title || row.fleet_id
            return h(ElTooltip, { content: tooltipContent, placement: 'top' }, () =>
              h('span', { class: 'cursor-default' }, row.fleet_title || row.fleet_id || '')
            )
          }
        },
        {
          prop: 'solar_system_id',
          label: t('srp.manage.columns.system'),
          width: 128,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', {}, getName(row.solar_system_id, String(row.solar_system_id), 'solar_system'))
        },
        {
          prop: 'killmail_id',
          label: t('srp.manage.columns.killId'),
          width: 96,
          formatter: (row: SrpApp) =>
            h(
              ElLink,
              {
                href: `https://zkillboard.com/kill/${row.killmail_id}/`,
                target: '_blank',
                type: 'primary'
              },
              () => String(row.killmail_id)
            )
        },
        {
          prop: 'killmail_time',
          label: t('srp.manage.columns.kmTime'),
          width: 160,
          formatter: (row: SrpApp) => h('span', {}, formatTime(row.killmail_time))
        },
        {
          prop: 'corporation_id',
          label: t('srp.manage.columns.corporation'),
          width: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h(
              'span',
              {},
              getName(
                row.corporation_id,
                row.corporation_id ? `ID: ${row.corporation_id}` : '-',
                'esi'
              )
            )
        },
        {
          prop: 'alliance_id',
          label: t('srp.manage.columns.alliance'),
          width: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h(
              'span',
              {},
              getName(row.alliance_id, row.alliance_id ? `ID: ${row.alliance_id}` : '-', 'esi')
            )
        },
        {
          prop: 'note',
          label: t('srp.manage.columns.note'),
          minWidth: 150,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', { class: row.note ? '' : 'text-gray-400' }, row.note || '-')
        },
        {
          prop: 'review_note',
          label: t('srp.manage.columns.reviewNote'),
          minWidth: 170,
          showOverflowTooltip: true,
          formatter: (row: SrpApp) =>
            h('span', { class: row.review_note ? '' : 'text-gray-400' }, row.review_note || '-')
        },
        {
          prop: 'actions',
          label: t('srp.manage.columns.action'),
          width: 220,
          fixed: 'right',
          formatter: (row: SrpApp) => {
            const btns: ReturnType<typeof h>[] = [
              h(ArtButtonTable, { type: 'view', onClick: () => openKmPreview(row) })
            ]
            if (row.review_status === 'submitted') {
              // 待审批：批准 + 拒绝
              btns.push(
                h(ArtButtonTable, {
                  label: t('srp.manage.approveBtn'),
                  elType: 'success',
                  onClick: () => openReviewDialog(row, 'approve')
                }),
                h(ArtButtonTable, {
                  label: t('srp.manage.rejectBtn'),
                  elType: 'danger',
                  onClick: () => openReviewDialog(row, 'reject')
                })
              )
            } else if (row.review_status === 'approved' && row.payout_status === 'notpaid') {
              // 已批准 + 未发放：发放（仅 srp/admin）+ 编辑 + 重新拒绝
              if (canPayout.value) {
                btns.push(
                  h(ArtButtonTable, {
                    label: t('srp.manage.payoutBtn'),
                    elType: 'primary',
                    onClick: () => openPayoutDialog(row)
                  })
                )
              }
              btns.push(
                h(ArtButtonTable, {
                  label: t('srp.manage.editBtn'),
                  elType: 'warning',
                  onClick: () => openReviewDialog(row, 'approve')
                }),
                h(ArtButtonTable, {
                  label: t('srp.manage.reRejectBtn'),
                  elType: 'danger',
                  onClick: () => openReviewDialog(row, 'reject')
                })
              )
            } else if (row.review_status === 'rejected') {
              // 已拒绝：可重新批准
              btns.push(
                h(ArtButtonTable, {
                  label: t('srp.manage.reApproveBtn'),
                  elType: 'success',
                  onClick: () => openReviewDialog(row, 'approve')
                })
              )
            }
            return h('div', { class: 'flex items-center gap-1' }, btns)
          }
        }
      ]
    }
  })

  watch(data, async (list) => {
    if (list.length) await resolveManageNames(list)
  })

  /** 收集申请列表中所有需要解析的 ID，一次性查询 */
  const resolveManageNames = async (list: Api.Srp.Application[]) => {
    const typeIds = new Set<number>()
    const solarIds = new Set<number>()
    const esiIds = new Set<number>()
    for (const app of list) {
      if (app.ship_type_id) typeIds.add(app.ship_type_id)
      if (app.solar_system_id) solarIds.add(app.solar_system_id)
      if (app.corporation_id) esiIds.add(app.corporation_id)
      if (app.alliance_id) esiIds.add(app.alliance_id)
    }
    await resolveNames({
      ids: {
        ...(typeIds.size ? { type: [...typeIds] } : {}),
        ...(solarIds.size ? { solar_system: [...solarIds] } : {})
      },
      esi: esiIds.size ? [...esiIds] : undefined
    })
  }

  const handleSearch = () => {
    Object.assign(searchParams, {
      tab: activeTab.value,
      review_status: filter.review_status || undefined,
      fleet_id: filter.fleet_id || undefined
    })
    getData()
  }
  const resetFilter = () => {
    filter.review_status = ''
    filter.fleet_id = ''
    Object.assign(searchParams, {
      tab: activeTab.value,
      review_status: undefined,
      fleet_id: undefined
    })
    getData()
  }
  const handleTabChange = () => {
    filter.review_status = ''
    filter.fleet_id = ''
    Object.assign(searchParams, {
      tab: activeTab.value,
      review_status: undefined,
      fleet_id: undefined
    })
    getData()
  }

  const reviewDialogVisible = ref(false)
  const reviewAction = ref<'approve' | 'reject'>('approve')
  const reviewTarget = ref<Api.Srp.Application | null>(null)
  const reviewForm = reactive({ review_note: '', final_amount: 0 })
  const actionLoading = ref(false)

  const reviewTargetFleetLabel = computed(() => {
    const rt = reviewTarget.value
    if (!rt?.fleet_id) return ''
    const fleet = fleetMap.value.get(rt.fleet_id)
    if (fleet) return formatFleetLabel(fleet)
    return rt.fleet_fc_name
      ? `${rt.fleet_fc_name}: ${rt.fleet_title || rt.fleet_id}`
      : rt.fleet_title || rt.fleet_id
  })

  /** 当前操作人的主角色名（用于默认文案替换） */
  const primaryCharName = computed(() => {
    const info = userStore.getUserInfo
    if (!info.characters || !info.primaryCharacterId) return ''
    return (
      info.characters.find((c) => c.character_id === info.primaryCharacterId)?.character_name ?? ''
    )
  })

  const DEFAULT_APPROVE_NOTE = '补损由补损官{{mainChracterName}}手动批准，如有问题请Q群联系'
  const DEFAULT_REJECT_NOTE =
    '不符合现有补损条例。如有问题请Q群联系{{mainChracterName}}（或游戏内邮件{{mainChracterName}})'

  const fillTemplate = (tpl: string) =>
    tpl.replaceAll('{{mainChracterName}}', primaryCharName.value || t('srp.manage.unknownReviewer'))

  const updateReviewFinalAmount = (value: number | null | undefined) => {
    reviewForm.final_amount = fromMillionISKInput(value)
  }

  const openReviewDialog = (row: Api.Srp.Application, action: 'approve' | 'reject') => {
    reviewTarget.value = row
    reviewAction.value = action
    reviewForm.review_note =
      action === 'approve'
        ? row.review_status === 'submitted'
          ? fillTemplate(DEFAULT_APPROVE_NOTE)
          : row.review_note || ''
        : fillTemplate(DEFAULT_REJECT_NOTE)
    reviewForm.final_amount = action === 'approve' ? row.final_amount : 0
    reviewDialogVisible.value = true
  }

  const handleReview = async () => {
    if (!reviewTarget.value) return
    if (reviewAction.value === 'reject' && !reviewForm.review_note) {
      ElMessage.warning(t('srp.manage.rejectRequired'))
      return
    }
    actionLoading.value = true
    try {
      await reviewApplication(reviewTarget.value.id, {
        action: reviewAction.value,
        review_note: reviewForm.review_note,
        final_amount: reviewForm.final_amount
      })
      ElMessage.success(
        reviewAction.value === 'approve'
          ? t('srp.manage.approveSuccess')
          : t('srp.manage.rejectSuccess')
      )
      reviewDialogVisible.value = false
      refreshData()
    } catch {
      /* handled */
    } finally {
      actionLoading.value = false
    }
  }

  const payoutDialogVisible = ref(false)
  const payoutTarget = ref<Api.Srp.Application | null>(null)
  const autoApproveLoading = ref(false)
  const batchPayoutDialogVisible = ref(false)
  const batchPayoutList = ref<BatchPayoutSummary[]>([])
  const batchSummaryLoading = ref(false)
  const batchPayoutLoadingUserId = ref<number | null>(null)

  const openPayoutDialog = (row: Api.Srp.Application) => {
    payoutTarget.value = row
    payoutDialogVisible.value = true
    handleOpenInfoWindow()
  }

  const handlePayout = async () => {
    if (!payoutTarget.value) return
    actionLoading.value = true
    try {
      await payoutApplication(payoutTarget.value.id)
      ElMessage.success(t('srp.manage.payoutSuccess'))
      payoutDialogVisible.value = false
      refreshData()
    } catch {
      /* handled */
    } finally {
      actionLoading.value = false
    }
  }

  const loadBatchPayoutSummary = async () => {
    batchSummaryLoading.value = true
    try {
      batchPayoutList.value = await fetchBatchPayoutSummary()
    } catch {
      batchPayoutList.value = []
    } finally {
      batchSummaryLoading.value = false
    }
  }

  const openBatchPayoutDialog = async () => {
    batchPayoutDialogVisible.value = true
    await loadBatchPayoutSummary()
  }

  const handleAutoApprove = async () => {
    if (!filter.fleet_id) {
      ElMessage.warning(t('srp.manage.autoApproveFleetRequired'))
      return
    }

    autoApproveLoading.value = true
    try {
      const result = await runFleetAutoApproval({ fleet_id: filter.fleet_id })
      ElMessage.success(
        t('srp.manage.autoApproveSuccess', {
          approved: result.approved_count,
          skipped: result.skipped_count
        })
      )
      refreshData()
    } catch {
      /* handled */
    } finally {
      autoApproveLoading.value = false
    }
  }

  const handleBatchPayout = async (row: BatchPayoutSummary) => {
    try {
      await ElMessageBox.confirm(
        t('srp.manage.batchPayoutConfirmText', {
          name: row.main_character_name || t('srp.manage.unknownMainCharacter'),
          amount: formatISK(row.total_amount)
        }),
        t('srp.manage.batchPayoutConfirmTitle'),
        {
          type: 'warning',
          confirmButtonText: t('srp.manage.confirmPayout'),
          cancelButtonText: t('srp.apply.cancelBtn')
        }
      )
    } catch {
      return
    }

    batchPayoutLoadingUserId.value = row.user_id
    try {
      await batchPayoutByUser(row.user_id)
      ElMessage.success(t('srp.manage.batchPayoutSuccess'))
      await Promise.all([loadBatchPayoutSummary(), refreshData()])
    } catch {
      /* handled */
    } finally {
      batchPayoutLoadingUserId.value = null
    }
  }

  /* ── 复制到剪贴板 ── */
  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success(t('srp.manage.copied'))
    } catch {
      ElMessage.warning(t('srp.manage.copyFailed'))
    }
  }

  const formatBatchPayoutLine = (
    characterId: number,
    characterName: string,
    totalAmount: number
  ) => {
    const fullAmount = new Intl.NumberFormat('en-US', {
      maximumFractionDigits: 0,
      useGrouping: false
    }).format(totalAmount ?? 0)
    return `<a href="showinfo:1376//${characterId}">${characterName}</a>  ${fullAmount}`
  }

  const copyBatchPayoutListText = async () => {
    const lines = batchPayoutList.value
      .filter((row) => row.main_character_id && row.main_character_name)
      .map((row) =>
        formatBatchPayoutLine(row.main_character_id, row.main_character_name, row.total_amount)
      )

    if (!lines.length) {
      ElMessage.warning(t('srp.manage.batchPayoutEmpty'))
      return
    }

    await copyText(lines.join('\r\n'))
  }

  /* ── Open Info Window (ESI) ── */
  const openWindowLoading = ref(false)
  const handleOpenInfoWindow = async () => {
    if (!payoutTarget.value) return
    const userInfo = userStore.getUserInfo
    const primaryCharacterId = userInfo.primaryCharacterId
    if (!primaryCharacterId) {
      ElMessage.warning(t('srp.manage.noPrimaryCharacter'))
      return
    }
    openWindowLoading.value = true
    try {
      await openInfoWindow({
        character_id: primaryCharacterId,
        target_id: payoutTarget.value.character_id
      })
      ElMessage.success(t('srp.manage.openInfoWindowSuccess'))
    } catch {
      /* handled */
    } finally {
      openWindowLoading.value = false
    }
  }

  const formatShortTime = (v: string) => {
    if (!v) return '-'
    const d = new Date(v)
    return `${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
  const formatFleetLabel = (f: Api.Fleet.FleetItem) =>
    `${f.fc_character_name}: ${f.title} (${f.pap_count}PAP) @ ${formatShortTime(f.start_at)}~${formatShortTime(f.end_at)}`
  const formatISK = (v: number) => {
    const value = Number(v ?? 0)
    const abs = Math.abs(value)
    const units = [
      { threshold: 1_000_000_000_000, suffix: 'T' },
      { threshold: 1_000_000_000, suffix: 'B' },
      { threshold: 1_000_000, suffix: 'M' },
      { threshold: 1_000, suffix: 'K' }
    ]

    for (const unit of units) {
      if (abs >= unit.threshold) {
        return `${(value / unit.threshold).toFixed(1)}${unit.suffix}`
      }
    }

    return value.toFixed(1)
  }

  // ─── 导出 ───
  const manageExportHeaders = {
    character_name: '角色',
    ship_name: '舰船',
    solar_system: '星系',
    killmail_id: 'KillID',
    killmail_time: 'KM时间',
    corporation: '军团',
    alliance: '联盟',
    fleet_title: '关联舰队',
    fleet_fc_name: 'FC',
    note: '备注',
    recommended_amount: '推荐金额',
    final_amount: '最终金额',
    review_status: '审批状态',
    review_note: '审批备注',
    payout_status: '发放状态'
  }
  const exportManageData = computed(() =>
    data.value.map((app) => ({
      character_name: app.character_name,
      ship_name: getName(app.ship_type_id, `TypeID: ${app.ship_type_id}`, 'type'),
      solar_system: getName(app.solar_system_id, String(app.solar_system_id), 'solar_system'),
      killmail_id: app.killmail_id,
      killmail_time: formatTime(app.killmail_time),
      corporation: getName(
        app.corporation_id,
        app.corporation_id ? `ID: ${app.corporation_id}` : '-',
        'esi'
      ),
      alliance: getName(app.alliance_id, app.alliance_id ? `ID: ${app.alliance_id}` : '-', 'esi'),
      fleet_title: app.fleet_title || '-',
      fleet_fc_name: app.fleet_fc_name || '-',
      note: app.note || '-',
      recommended_amount: app.recommended_amount,
      final_amount: app.final_amount,
      review_status: reviewStatusLabel(app.review_status),
      review_note: app.review_note || '-',
      payout_status: app.payout_status === 'paid' ? t('srp.status.paid') : t('srp.status.notpaid')
    }))
  )

  /* ── KM 预览 ── */
  const kmPreviewVisible = ref(false)
  const previewKillmailId = ref(0)
  const openKmPreview = (row: Api.Srp.Application) => {
    previewKillmailId.value = row.killmail_id
    kmPreviewVisible.value = true
  }

  onMounted(() => {
    loadFleets()
  })
</script>

<style scoped>
  .million-isk-input {
    width: 100%;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
  }

  .million-isk-input__control {
    width: 100%;
  }

  .million-isk-input__suffix {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    white-space: nowrap;
  }

  .payout-info-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .payout-info-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .payout-label {
    min-width: 80px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
    flex-shrink: 0;
  }

  .payout-value {
    font-size: 14px;
    color: var(--el-text-color-primary);
    word-break: break-all;
  }
</style>
