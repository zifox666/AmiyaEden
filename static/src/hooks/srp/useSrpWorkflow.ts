import { ref, reactive, computed, type Ref, type ComputedRef } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/store/modules/user'
import { useNameResolver } from '@/hooks'
import {
  reviewApplication,
  payoutApplication,
  fetchBatchPayoutSummary,
  batchPayoutByUser,
  batchPayoutAsFuxiCoin,
  runFleetAutoApproval,
  openInfoWindow
} from '@/api/srp'
import { useClipboardCopy } from '@/hooks/core/useClipboardCopy'
import { formatIskPlain, millionInputToIsk } from '@/utils/common'

type SrpApp = Api.Srp.Application
type BatchPayoutSummary = Api.Srp.BatchPayoutSummary

export function useSrpWorkflow(deps: {
  fleetMap: ComputedRef<Map<string, Api.Fleet.FleetItem>>
  formatFleetLabel: (f: Api.Fleet.FleetItem) => string
  formatISK: (v: number) => string
  filter: { review_status: string; fleet_id: string; keyword: string }
  payoutMode: Ref<Api.Srp.PayoutMode>
  refreshData: () => void | Promise<void>
}) {
  const { t } = useI18n()
  const { getName } = useNameResolver()
  const userStore = useUserStore()
  const { copyText } = useClipboardCopy()

  // ─── Review Dialog ───
  const reviewDialogVisible = ref(false)
  const reviewAction = ref<'approve' | 'reject'>('approve')
  const reviewTarget = ref<SrpApp | null>(null)
  const reviewForm = reactive({ review_note: '', final_amount: 0 })
  const actionLoading = ref(false)

  const reviewTargetFleetLabel = computed(() => {
    const rt = reviewTarget.value
    if (!rt?.fleet_id) return ''
    const fleet = deps.fleetMap.value.get(rt.fleet_id as any)
    if (fleet) return deps.formatFleetLabel(fleet)
    return rt.fleet_fc_name
      ? `${rt.fleet_fc_name}: ${rt.fleet_title || rt.fleet_id}`
      : rt.fleet_title || rt.fleet_id
  })

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
    reviewForm.final_amount = millionInputToIsk(value)
  }

  const openReviewDialog = (row: SrpApp, action: 'approve' | 'reject') => {
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
      deps.refreshData()
    } catch {
      /* handled */
    } finally {
      actionLoading.value = false
    }
  }

  // ─── Payout Dialog ───
  const payoutDialogVisible = ref(false)
  const payoutTarget = ref<SrpApp | null>(null)

  const convertISKToFuxiCoin = (amount: number) =>
    Math.round((Number(amount ?? 0) / 1_000_000) * 100) / 100

  const formatFuxiCoin = (amount: number) => convertISKToFuxiCoin(amount).toFixed(2)

  const openPayoutDialog = (row: SrpApp) => {
    payoutTarget.value = row
    payoutDialogVisible.value = true
    handleOpenInfoWindow()
  }

  const handlePayoutAction = async (row: SrpApp) => {
    if (deps.payoutMode.value === 'manual_transfer') {
      openPayoutDialog(row)
      return
    }

    try {
      await ElMessageBox.confirm(
        t('srp.manage.fuxiPayoutConfirmText', {
          name: row.character_name,
          ship: getName(row.ship_type_id, `TypeID: ${row.ship_type_id}`, 'type'),
          amount: formatFuxiCoin(row.final_amount)
        }),
        t('srp.manage.fuxiPayoutConfirmTitle'),
        {
          type: 'warning',
          confirmButtonText: t('srp.manage.confirmPayout'),
          cancelButtonText: t('srp.apply.cancelBtn')
        }
      )
    } catch {
      return
    }

    actionLoading.value = true
    try {
      await payoutApplication(row.id, { mode: 'fuxi_coin' })
      ElMessage.success(
        t('srp.manage.fuxiPayoutSuccess', { amount: formatFuxiCoin(row.final_amount) })
      )
      await deps.refreshData()
    } catch {
      /* handled */
    } finally {
      actionLoading.value = false
    }
  }

  const handlePayout = async () => {
    if (!payoutTarget.value) return
    actionLoading.value = true
    try {
      await payoutApplication(payoutTarget.value.id, { mode: 'manual_transfer' })
      ElMessage.success(t('srp.manage.payoutSuccess'))
      payoutDialogVisible.value = false
      deps.refreshData()
    } catch {
      /* handled */
    } finally {
      actionLoading.value = false
    }
  }

  // ─── Batch Payout ───
  const autoApproveLoading = ref(false)
  const batchPayoutDialogVisible = ref(false)
  const batchPayoutList = ref<BatchPayoutSummary[]>([])
  const batchSummaryLoading = ref(false)
  const batchPayoutLoadingUserId = ref<number | null>(null)

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

  const handleBatchPayoutClick = async () => {
    if (deps.payoutMode.value === 'manual_transfer') {
      await openBatchPayoutDialog()
      return
    }

    try {
      await ElMessageBox.confirm(
        t('srp.manage.fuxiBatchPayoutConfirmText'),
        t('srp.manage.fuxiBatchPayoutConfirmTitle'),
        {
          type: 'warning',
          confirmButtonText: t('srp.manage.confirmPayout'),
          cancelButtonText: t('srp.apply.cancelBtn')
        }
      )
    } catch {
      return
    }

    batchSummaryLoading.value = true
    try {
      const result = await batchPayoutAsFuxiCoin()
      ElMessage.success(
        t('srp.manage.fuxiBatchPayoutSuccess', {
          users: result.user_count,
          applications: result.application_count,
          amount: result.total_fuxi_coin.toFixed(2)
        })
      )
      await deps.refreshData()
    } catch {
      /* handled */
    } finally {
      batchSummaryLoading.value = false
    }
  }

  const handleAutoApprove = async () => {
    if (!deps.filter.fleet_id) {
      ElMessage.warning(t('srp.manage.autoApproveFleetRequired'))
      return
    }

    autoApproveLoading.value = true
    try {
      const result = await runFleetAutoApproval({ fleet_id: deps.filter.fleet_id })
      ElMessage.success(
        t('srp.manage.autoApproveSuccess', {
          approved: result.approved_count,
          skipped: result.skipped_count
        })
      )
      deps.refreshData()
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
          amount: deps.formatISK(row.total_amount)
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
      await Promise.all([loadBatchPayoutSummary(), deps.refreshData()])
    } catch {
      /* handled */
    } finally {
      batchPayoutLoadingUserId.value = null
    }
  }

  const formatBatchPayoutLine = (characterId: number, characterName: string, totalAmount: number) =>
    `<a href="showinfo:1376//${characterId}">${characterName}</a>  ${formatIskPlain(totalAmount)}`

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

  // ─── Open Info Window (ESI) ───
  const handleOpenInfoWindow = async () => {
    if (!payoutTarget.value) return
    const userInfo = userStore.getUserInfo
    const primaryCharacterId = userInfo.primaryCharacterId
    if (!primaryCharacterId) {
      ElMessage.warning(t('srp.manage.noPrimaryCharacter'))
      return
    }
    try {
      await openInfoWindow({
        character_id: primaryCharacterId,
        target_id: payoutTarget.value.character_id
      })
      ElMessage.success(t('srp.manage.openInfoWindowSuccess'))
    } catch {
      /* handled */
    }
  }

  return {
    // review
    reviewDialogVisible,
    reviewAction,
    reviewTarget,
    reviewForm,
    actionLoading,
    reviewTargetFleetLabel,
    updateReviewFinalAmount,
    openReviewDialog,
    handleReview,
    // payout
    payoutDialogVisible,
    payoutTarget,
    handlePayoutAction,
    handlePayout,
    // batch payout
    autoApproveLoading,
    batchPayoutDialogVisible,
    batchPayoutList,
    batchSummaryLoading,
    batchPayoutLoadingUserId,
    handleBatchPayoutClick,
    handleAutoApprove,
    handleBatchPayout,
    // clipboard
    copyText,
    copyBatchPayoutListText,
    // formatters
    formatFuxiCoin
  }
}
