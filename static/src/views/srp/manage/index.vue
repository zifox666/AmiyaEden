<!-- SRP 补损审批管理页面 -->
<template>
  <div class="srp-manage-page art-full-height">
    <ElCard class="art-search-card" shadow="never">
      <div class="flex items-center gap-3 flex-wrap">
        <ElSelect v-model="filter.review_status" :placeholder="$t('srp.apply.columns.reviewStatus')" clearable style="width: 130px" @change="handleSearch">
          <ElOption :label="$t('srp.status.pending')" value="pending" />
          <ElOption :label="$t('srp.status.approved')" value="approved" />
          <ElOption :label="$t('srp.status.rejected')" value="rejected" />
        </ElSelect>
        <ElSelect v-model="filter.payout_status" :placeholder="$t('srp.apply.columns.payoutStatus')" clearable style="width: 130px" @change="handleSearch">
          <ElOption :label="$t('srp.status.unpaid')" value="pending" />
          <ElOption :label="$t('srp.status.paid')" value="paid" />
        </ElSelect>
        <ElSelect v-model="filter.fleet_id" :placeholder="$t('srp.manage.selectFleet')" clearable filterable style="width: 220px" @change="handleSearch">
          <ElOption v-for="f in fleets" :key="f.id" :label="f.title" :value="f.id" />
        </ElSelect>
        <ElButton type="primary" @click="handleSearch">{{ $t('srp.manage.searchBtn') }}</ElButton>
        <ElButton @click="resetFilter">{{ $t('srp.manage.resetBtn') }}</ElButton>
      </div>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-base font-medium">{{ $t('srp.manage.title') }}</h2>
          <ElButton :loading="loading" @click="loadApplications">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('srp.manage.refresh') }}
          </ElButton>
        </div>
      </template>

      <ElTable v-loading="loading" :data="applications" stripe border style="width: 100%">
        <ElTableColumn prop="id" :label="$t('srp.manage.columns.id')" width="70" align="center" />
        <ElTableColumn prop="character_name" :label="$t('srp.manage.columns.character')" width="150" />
        <ElTableColumn prop="ship_name" :label="$t('srp.manage.columns.ship')" width="180">
          <template #default="{ row }">{{ row.ship_name || `TypeID: ${row.ship_type_id}` }}</template>
        </ElTableColumn>
        <ElTableColumn prop="solar_system_name" :label="$t('srp.manage.columns.system')" width="140">
          <template #default="{ row }">{{ row.solar_system_name || row.solar_system_id }}</template>
        </ElTableColumn>
        <ElTableColumn prop="killmail_id" :label="$t('srp.manage.columns.killId')" width="110" align="center">
          <template #default="{ row }">
            <ElLink :href="`https://zkillboard.com/kill/${row.killmail_id}/`" target="_blank" type="primary">{{ row.killmail_id }}</ElLink>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="killmail_time" :label="$t('srp.manage.columns.kmTime')" width="175">
          <template #default="{ row }">{{ formatTime(row.killmail_time) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="corporation_name" :label="$t('srp.manage.columns.corporation')" width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.corporation_name || `ID: ${row.corporation_id}` }}</template>
        </ElTableColumn>
        <ElTableColumn prop="alliance_name" :label="$t('srp.manage.columns.alliance')" width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.alliance_name || (row.alliance_id ? `ID: ${row.alliance_id}` : '-') }}</template>
        </ElTableColumn>
        <ElTableColumn prop="recommended_amount" :label="$t('srp.manage.columns.recommendedAmount')" width="140" align="right">
          <template #default="{ row }">{{ formatISK(row.recommended_amount) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="final_amount" :label="$t('srp.manage.columns.finalAmount')" width="140" align="right">
          <template #default="{ row }"><span class="font-semibold text-blue-600">{{ formatISK(row.final_amount) }}</span></template>
        </ElTableColumn>
        <ElTableColumn prop="review_status" :label="$t('srp.manage.columns.review')" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="reviewStatusType(row.review_status)" size="small">{{ reviewStatusLabel(row.review_status) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="payout_status" :label="$t('srp.manage.columns.payout')" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="payoutStatusType(row.payout_status)" size="small">
              {{ row.payout_status === 'paid' ? $t('srp.status.paid') : $t('srp.status.unpaid') }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('srp.manage.columns.action')" width="210" fixed="right" align="center">
          <template #default="{ row }">
            <template v-if="row.review_status === 'pending'">
              <ElButton size="small" type="success" @click="openReviewDialog(row, 'approve')">{{ $t('srp.manage.approveBtn') }}</ElButton>
              <ElButton size="small" type="danger" @click="openReviewDialog(row, 'reject')">{{ $t('srp.manage.rejectBtn') }}</ElButton>
            </template>
            <template v-else-if="row.review_status === 'approved' && row.payout_status === 'pending'">
              <ElButton size="small" type="primary" @click="openPayoutDialog(row)">{{ $t('srp.manage.payoutBtn') }}</ElButton>
            </template>
            <span v-else class="text-gray-400 text-xs">-</span>
          </template>
        </ElTableColumn>
      </ElTable>

      <div v-if="pagination.total > 0" class="pagination-wrapper">
        <ElPagination v-model:current-page="pagination.current" v-model:page-size="pagination.size"
          :total="pagination.total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper"
          background @size-change="() => { pagination.current = 1; loadApplications() }" @current-change="loadApplications" />
      </div>
    </ElCard>

    <ElDialog v-model="reviewDialogVisible" :title="reviewAction === 'approve' ? $t('srp.manage.approveDialog') : $t('srp.manage.rejectDialog')" width="460px">
      <ElForm label-width="90px">
        <ElFormItem :label="$t('srp.manage.finalAmount')" v-if="reviewAction === 'approve'">
          <ElInputNumber v-model="reviewForm.final_amount" :min="0" :precision="2" :step="1000000" style="width: 100%" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('srp.manage.finalAmountHint') }}</div>
        </ElFormItem>
        <ElFormItem :label="$t('srp.manage.reviewNote')" :required="reviewAction === 'reject'">
          <ElInput v-model="reviewForm.review_note" type="textarea" :rows="3"
            :placeholder="reviewAction === 'reject' ? $t('srp.manage.rejectNotePlaceholder') : $t('srp.manage.optionalPlaceholder')" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="reviewDialogVisible = false">{{ $t('srp.apply.cancelBtn') }}</ElButton>
        <ElButton :type="reviewAction === 'approve' ? 'success' : 'danger'" :loading="actionLoading" @click="handleReview">
          {{ reviewAction === 'approve' ? $t('srp.manage.confirmApprove') : $t('srp.manage.confirmReject') }}
        </ElButton>
      </template>
    </ElDialog>

    <ElDialog v-model="payoutDialogVisible" :title="$t('srp.manage.payoutDialog')" width="420px">
      <div class="mb-2">{{ $t('srp.manage.payoutCharacter') }}<strong>{{ payoutTarget?.character_name }}</strong></div>
      <div class="mb-4">{{ $t('srp.manage.payoutAmount') }}<strong class="text-blue-600">{{ formatISK(payoutTarget?.final_amount ?? 0) }} ISK</strong></div>
      <ElFormItem :label="$t('srp.manage.overrideAmount')">
        <ElInputNumber v-model="payoutOverrideAmount" :min="0" :precision="2" :step="1000000" style="width: 100%" />
        <div class="text-xs text-gray-400 mt-1">{{ $t('srp.manage.overrideAmountHint') }}</div>
      </ElFormItem>
      <template #footer>
        <ElButton @click="payoutDialogVisible = false">{{ $t('srp.apply.cancelBtn') }}</ElButton>
        <ElButton type="primary" :loading="actionLoading" @click="handlePayout">{{ $t('srp.manage.confirmPayout') }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { Refresh } from '@element-plus/icons-vue'
  import {
    ElCard, ElTable, ElTableColumn, ElTag, ElButton, ElPagination, ElSelect, ElOption,
    ElDialog, ElForm, ElFormItem, ElInputNumber, ElInput, ElLink, ElMessage
  } from 'element-plus'
  import { fetchFleetList } from '@/api/fleet'
  import { fetchApplicationList, reviewApplication, payoutApplication } from '@/api/srp'

  defineOptions({ name: 'SrpManage' })

  const { t } = useI18n()

  const applications = ref<Api.Srp.Application[]>([])
  const loading = ref(false)
  const pagination = reactive({ current: 1, size: 20, total: 0 })

  const fleets = ref<Api.Fleet.FleetItem[]>([])
  const loadFleets = async () => {
    try {
      const res = await fetchFleetList({ size: 200 } as any)
      fleets.value = res?.records ?? []
    } catch { fleets.value = [] }
  }

  const filter = reactive({ review_status: '', payout_status: '', fleet_id: '' })

  const loadApplications = async () => {
    loading.value = true
    try {
      const res = await fetchApplicationList({
        current: pagination.current, size: pagination.size,
        review_status: filter.review_status || undefined,
        payout_status: filter.payout_status || undefined,
        fleet_id: filter.fleet_id || undefined,
      })
      applications.value = res?.records ?? []
      pagination.total = res?.total ?? 0
    } catch { applications.value = [] }
    finally { loading.value = false }
  }

  const handleSearch = () => { pagination.current = 1; loadApplications() }
  const resetFilter = () => { filter.review_status = ''; filter.payout_status = ''; filter.fleet_id = ''; handleSearch() }

  const reviewDialogVisible = ref(false)
  const reviewAction = ref<'approve' | 'reject'>('approve')
  const reviewTarget = ref<Api.Srp.Application | null>(null)
  const reviewForm = reactive({ review_note: '', final_amount: 0 })
  const actionLoading = ref(false)

  const openReviewDialog = (row: Api.Srp.Application, action: 'approve' | 'reject') => {
    reviewTarget.value = row; reviewAction.value = action
    reviewForm.review_note = ''; reviewForm.final_amount = 0
    reviewDialogVisible.value = true
  }

  const handleReview = async () => {
    if (!reviewTarget.value) return
    if (reviewAction.value === 'reject' && !reviewForm.review_note) {
      ElMessage.warning(t('srp.manage.rejectRequired')); return
    }
    actionLoading.value = true
    try {
      await reviewApplication(reviewTarget.value.id, { action: reviewAction.value, review_note: reviewForm.review_note, final_amount: reviewForm.final_amount })
      ElMessage.success(reviewAction.value === 'approve' ? t('srp.manage.approveSuccess') : t('srp.manage.rejectSuccess'))
      reviewDialogVisible.value = false
      loadApplications()
    } catch { /* handled */ }
    finally { actionLoading.value = false }
  }

  const payoutDialogVisible = ref(false)
  const payoutTarget = ref<Api.Srp.Application | null>(null)
  const payoutOverrideAmount = ref(0)

  const openPayoutDialog = (row: Api.Srp.Application) => { payoutTarget.value = row; payoutOverrideAmount.value = 0; payoutDialogVisible.value = true }

  const handlePayout = async () => {
    if (!payoutTarget.value) return
    actionLoading.value = true
    try {
      await payoutApplication(payoutTarget.value.id, { final_amount: payoutOverrideAmount.value })
      ElMessage.success(t('srp.manage.payoutSuccess'))
      payoutDialogVisible.value = false
      loadApplications()
    } catch { /* handled */ }
    finally { actionLoading.value = false }
  }

  const formatTime = (v: string) => v ? new Date(v).toLocaleString() : '-'
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v ?? 0)

  type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'
  const reviewStatusType = (s: string): TagType =>
    (({ pending: 'info', approved: 'success', rejected: 'danger' } as Record<string, TagType>)[s] ?? 'info')
  const reviewStatusLabel = (s: string) =>
    ({ pending: t('srp.status.pending'), approved: t('srp.status.approved'), rejected: t('srp.status.rejected') })[s as 'pending' | 'approved' | 'rejected'] ?? s
  const payoutStatusType = (s: string): TagType => s === 'paid' ? 'success' : 'warning'

  onMounted(() => { loadFleets(); loadApplications() })
</script>

<style scoped>
  .pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>