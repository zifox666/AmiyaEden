<!-- SRP 补损审批管理页面 -->
<template>
  <div class="srp-manage-page art-full-height">
    <ElCard class="art-table-card srp-manage-card" shadow="never">
      <div class="srp-manage-content">
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
          <ElInput
            v-if="activeTab === 'history'"
            v-model="filter.keyword"
            :placeholder="$t('srp.manage.userKeywordFilter')"
            clearable
            style="width: 220px"
            @clear="handleSearch"
            @keyup="handleKeywordSearchKeyup"
          />
          <ElButton type="primary" @click="handleSearch">{{ $t('srp.manage.searchBtn') }}</ElButton>
          <ElButton @click="resetFilter">{{ $t('srp.manage.resetBtn') }}</ElButton>
          <div v-if="activeTab === 'pending' && canPayout" class="flex items-center gap-2">
            <span class="text-sm text-gray-500">{{ $t('srp.manage.payoutModeLabel') }}</span>
            <ElRadioGroup v-model="payoutMode" size="small">
              <ElRadioButton :value="'fuxi_coin'">
                {{ $t('srp.manage.payoutModes.fuxiCoin') }}
              </ElRadioButton>
              <ElRadioButton :value="'manual_transfer'">
                {{ $t('srp.manage.payoutModes.manualTransfer') }}
              </ElRadioButton>
            </ElRadioGroup>
          </div>
        </div>

        <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
          <template #left>
            <div class="flex items-center gap-2">
              <template v-if="activeTab === 'pending' && canPayout">
                <ElButton type="primary" @click="handleAutoApprove">
                  {{ $t('srp.manage.autoApproveBtn') }}
                </ElButton>
                <ElButton type="warning" @click="handleBatchPayoutClick">
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

        <div class="srp-manage-table-shell">
          <ArtTable
            :loading="loading"
            :data="data"
            :columns="columns"
            :pagination="pagination"
            visual-variant="ledger"
            @pagination:size-change="handleSizeChange"
            @pagination:current-change="handleCurrentChange"
          />
        </div>
      </div>
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
              :model-value="iskToMillionInput(reviewForm.final_amount)"
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
        <!-- 人物（可复制） -->
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
  import {
    ElCard,
    ElButton,
    ElSelect,
    ElOption,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElInput,
    ElLink,
    ElTable,
    ElTableColumn,
    ElTabs,
    ElTabPane,
    ElTooltip,
    ElRadioGroup,
    ElRadioButton
  } from 'element-plus'
  import ArtExcelExport from '@/components/core/forms/art-excel-export/index.vue'
  import KmPreviewDialog from '@/components/business/KmPreviewDialog.vue'
  import { CopyDocument } from '@element-plus/icons-vue'
  import { iskToMillionInput } from '@/utils/common'
  import { useSrpManage } from '@/hooks/srp/useSrpManage'
  import { useSrpWorkflow } from '@/hooks/srp/useSrpWorkflow'

  defineOptions({ name: 'SrpManage' })

  const {
    canPayout,
    fleets,
    fleetMap,
    formatFleetLabel,
    activeTab,
    payoutMode,
    filter,
    handleSearch,
    handleKeywordSearchKeyup,
    resetFilter,
    handleTabChange,
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    formatISK,
    manageExportHeaders,
    exportManageData,
    kmPreviewVisible,
    previewKillmailId,
    openKmPreview
  } = useSrpManage({
    openReviewDialog: (...args) => openReviewDialog(...args),
    handlePayoutAction: (...args) => handlePayoutAction(...args),
    openKmPreview: (row) => openKmPreview(row)
  })

  const {
    reviewDialogVisible,
    reviewAction,
    reviewTarget,
    reviewForm,
    actionLoading,
    reviewTargetFleetLabel,
    updateReviewFinalAmount,
    openReviewDialog,
    handleReview,
    payoutDialogVisible,
    payoutTarget,
    handlePayoutAction,
    handlePayout,
    batchPayoutDialogVisible,
    batchPayoutList,
    batchSummaryLoading,
    batchPayoutLoadingUserId,
    handleBatchPayoutClick,
    handleAutoApprove,
    handleBatchPayout,
    copyText,
    copyBatchPayoutListText
  } = useSrpWorkflow({
    fleetMap,
    formatFleetLabel,
    formatISK,
    filter,
    payoutMode,
    refreshData
  })
</script>

<style scoped>
  .srp-manage-page {
    gap: 12px;
  }

  .srp-manage-card {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .srp-manage-card :deep(.el-card__body) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .srp-manage-content {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .srp-manage-table-shell {
    flex: 1;
    min-height: 0;
  }

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
