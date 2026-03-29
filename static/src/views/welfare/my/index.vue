<!-- 我的福利页面 -->
<template>
  <div class="welfare-my-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ElTabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 申请福利 -->
        <ElTabPane :label="t('welfareMy.applyTab')" name="apply">
          <ArtTable
            :loading="eligibleLoading"
            :data="eligibleRows"
            :columns="eligibleColumns"
            :empty-text="t('welfareMy.noEligibleWelfares')"
            :row-class-name="eligibleRowClassName"
          />
        </ElTabPane>

        <!-- 已领取福利 -->
        <ElTabPane :label="t('welfareMy.applicationsTab')" name="applications">
          <ArtTable
            :loading="applicationsLoading"
            :data="applications"
            :columns="applicationColumns"
            :empty-text="t('welfareMy.noApplications')"
            :pagination="applicationPagination"
            @pagination:size-change="handleApplicationSizeChange"
            @pagination:current-change="handleApplicationCurrentChange"
          />
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <!-- 证明图片上传对话框 -->
    <ElDialog
      v-model="evidenceDialogVisible"
      :title="t('welfareMy.evidenceDialogTitle')"
      width="480px"
      destroy-on-close
    >
      <div class="flex flex-col gap-3">
        <p class="text-sm text-gray-500">{{ t('welfareMy.evidenceDialogHint') }}</p>
        <div v-if="pendingApplyRow?.exampleEvidence" class="flex flex-col gap-1">
          <span class="text-xs text-gray-400">{{ t('welfareMy.exampleEvidenceLabel') }}</span>
          <img
            :src="pendingApplyRow.exampleEvidence"
            class="rounded border"
            style="max-height: 160px; max-width: 100%; object-fit: contain"
          />
        </div>
        <ElUpload
          :show-file-list="false"
          accept="image/*"
          :before-upload="handleEvidenceFileUpload"
        >
          <ElButton size="small" :loading="evidenceUploading">
            {{ t('welfareMy.uploadEvidenceBtn') }}
          </ElButton>
        </ElUpload>
        <img
          v-if="evidenceImageUrl"
          :src="evidenceImageUrl"
          class="rounded border"
          style="max-height: 160px; max-width: 100%; object-fit: contain"
        />
      </div>
      <template #footer>
        <ElButton @click="evidenceDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton
          type="primary"
          :disabled="!evidenceImageUrl"
          :loading="applyLoading"
          @click="handleEvidenceConfirm"
        >
          {{ t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElUpload, ElMessage, ElTooltip } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    getEligibleWelfares,
    applyForWelfare,
    getMyApplications,
    uploadWelfareEvidence
  } from '@/api/welfare'
  import { sortEligibleRows } from './eligibleRows'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'WelfareMy' })
  const { t } = useI18n()

  // ─── Tab state ───
  const activeTab = ref('apply')

  // ─── 申请福利 Tab ───
  const eligibleLoading = ref(false)
  const eligibleWelfares = ref<Api.Welfare.EligibleWelfare[]>([])

  // 将 eligible welfares 展平为表格行
  // per_user: 一条福利一行
  // per_character: 每个可申请人物一行
  interface EligibleRow {
    welfareId: number
    welfareName: string
    description: string
    distMode: string
    canApplyNow: boolean
    ineligibleReason?: 'pap' | 'skill' | 'pap_skill'
    characterId?: number
    characterName?: string
  }

  const eligibleRows = computed<EligibleRow[]>(() => {
    const rows: EligibleRow[] = []
    for (const w of eligibleWelfares.value) {
      if (w.dist_mode === 'per_user') {
        rows.push({
          welfareId: w.id,
          welfareName: w.name,
          description: w.description,
          distMode: w.dist_mode,
          canApplyNow: w.can_apply_now,
          ineligibleReason: w.ineligible_reason
        })
      } else {
        for (const char of w.eligible_characters) {
          rows.push({
            welfareId: w.id,
            welfareName: w.name,
            description: w.description,
            distMode: w.dist_mode,
            canApplyNow: char.can_apply_now,
            ineligibleReason: char.ineligible_reason,
            characterId: char.character_id,
            characterName: char.character_name
          })
        }
      }
    }
    return sortEligibleRows(rows)
  })

  const DIST_MODE_CONFIG = computed(
    () =>
      ({
        per_user: { label: t('welfareMy.distModePerUser'), type: 'primary' },
        per_character: { label: t('welfareMy.distModePerCharacter'), type: 'warning' }
      }) as Record<string, { label: string; type: string }>
  )

  const eligibleColumns = computed(() => [
    {
      prop: 'welfareName',
      label: t('welfareMy.welfareName'),
      minWidth: 160,
      showOverflowTooltip: true
    },
    {
      prop: 'description',
      label: t('welfareMy.description'),
      minWidth: 200,
      showOverflowTooltip: true,
      formatter: (row: EligibleRow) => row.description || '-'
    },
    {
      prop: 'distMode',
      label: t('welfareMy.deliveryMode'),
      width: 120,
      formatter: (row: EligibleRow) => {
        const cfg = DIST_MODE_CONFIG.value[row.distMode] ?? {
          label: row.distMode,
          type: 'info'
        }
        return h(ElTag, { type: cfg.type as any, size: 'small', effect: 'plain' }, () => cfg.label)
      }
    },
    {
      prop: 'eligibility',
      label: t('welfareMy.eligibility'),
      width: 140,
      formatter: (row: EligibleRow) => {
        const tag = h(
          ElTag,
          {
            type: row.canApplyNow ? 'success' : 'info',
            size: 'small',
            effect: row.canApplyNow ? 'light' : 'plain'
          },
          () => (row.canApplyNow ? t('welfareMy.eligibilityNow') : t('welfareMy.eligibilityFuture'))
        )
        if (row.canApplyNow || !row.ineligibleReason) return tag
        const reasonKey =
          row.ineligibleReason === 'pap_skill'
            ? 'welfareMy.ineligibleReasonPapSkill'
            : row.ineligibleReason === 'pap'
              ? 'welfareMy.ineligibleReasonPap'
              : 'welfareMy.ineligibleReasonSkill'
        return h(ElTooltip, { content: t(reasonKey), placement: 'top' }, () => tag)
      }
    },
    {
      prop: 'characterName',
      label: t('welfareMy.characterName'),
      width: 160,
      formatter: (row: EligibleRow) => row.characterName || '-'
    },
    {
      prop: 'actions',
      label: '',
      width: 100,
      fixed: 'right' as const,
      formatter: (row: EligibleRow) =>
        h(
          ElButton,
          {
            type: row.canApplyNow ? 'primary' : 'info',
            size: 'small',
            plain: !row.canApplyNow,
            disabled: !row.canApplyNow,
            onClick: () => handleApply(row)
          },
          () => (row.canApplyNow ? t('welfareMy.applyBtn') : t('welfareMy.futureApplyBtn'))
        )
    }
  ])

  function getEligibleRowClassName({ row }: { row: EligibleRow }) {
    return row.canApplyNow ? '' : 'welfare-future-row'
  }

  function eligibleRowClassName({ row }: { row: Record<string, any>; rowIndex: number }) {
    return getEligibleRowClassName({ row: row as EligibleRow })
  }

  async function loadEligibleWelfares() {
    eligibleLoading.value = true
    try {
      const res = await getEligibleWelfares()
      eligibleWelfares.value = res ?? []
    } catch {
      eligibleWelfares.value = []
    } finally {
      eligibleLoading.value = false
    }
  }

  // ─── 证明图片对话框 ───
  const evidenceDialogVisible = ref(false)
  const pendingApplyRow = ref<(EligibleRow & { exampleEvidence: string }) | null>(null)
  const evidenceImageUrl = ref('')
  const evidenceUploading = ref(false)
  const applyLoading = ref(false)

  function handleApply(row: EligibleRow) {
    const welfare = eligibleWelfares.value.find((w) => w.id === row.welfareId)
    if (welfare?.require_evidence) {
      pendingApplyRow.value = { ...row, exampleEvidence: welfare.example_evidence ?? '' }
      evidenceImageUrl.value = ''
      evidenceDialogVisible.value = true
    } else {
      submitApply(row, '')
    }
  }

  async function handleEvidenceFileUpload(file: File) {
    evidenceUploading.value = true
    try {
      const res = await uploadWelfareEvidence(file)
      evidenceImageUrl.value = res.url
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('welfareMy.applyFailed'))
    } finally {
      evidenceUploading.value = false
    }
    return false
  }

  async function handleEvidenceConfirm() {
    if (!pendingApplyRow.value || !evidenceImageUrl.value) return
    applyLoading.value = true
    try {
      await submitApply(pendingApplyRow.value, evidenceImageUrl.value)
      evidenceDialogVisible.value = false
    } finally {
      applyLoading.value = false
    }
  }

  async function submitApply(row: EligibleRow, evidenceImage: string) {
    try {
      await applyForWelfare({
        welfare_id: row.welfareId,
        character_id: row.characterId,
        evidence_image: evidenceImage || undefined
      })
      ElMessage.success(t('welfareMy.applySuccess'))
      loadEligibleWelfares()
      loadApplications()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('welfareMy.applyFailed'))
    }
  }

  // ─── 已领取福利 Tab ───
  const {
    data: applications,
    loading: applicationsLoading,
    pagination: applicationPagination,
    handleSizeChange: handleApplicationSizeChange,
    handleCurrentChange: handleApplicationCurrentChange,
    getData: loadApplications
  } = useTable({
    core: {
      apiFn: getMyApplications,
      apiParams: { current: 1, size: 10 },
      immediate: false
    }
  })

  const STATUS_CONFIG = computed(
    () =>
      ({
        requested: { label: t('welfareMy.statusRequested'), type: 'warning' },
        delivered: { label: t('welfareMy.statusDelivered'), type: 'success' },
        rejected: { label: t('welfareMy.statusRejected'), type: 'danger' }
      }) as Record<string, { label: string; type: string }>
  )

  const applicationColumns = computed(() => [
    {
      prop: 'welfare_name',
      label: t('welfareMy.welfareName'),
      minWidth: 160,
      showOverflowTooltip: true
    },
    {
      prop: 'character_name',
      label: t('welfareMy.characterName'),
      width: 160
    },
    {
      prop: 'status',
      label: t('welfareMy.status'),
      width: 100,
      formatter: (row: Api.Welfare.MyApplication) => {
        const cfg = STATUS_CONFIG.value[row.status] ?? {
          label: row.status,
          type: 'info'
        }
        return h(ElTag, { type: cfg.type as any, size: 'small', effect: 'plain' }, () => cfg.label)
      }
    },
    {
      prop: 'created_at',
      label: t('welfareMy.appliedAt'),
      width: 170
    },
    {
      prop: 'reviewed_at',
      label: t('welfareMy.reviewedAt'),
      width: 170,
      formatter: (row: Api.Welfare.MyApplication) => row.reviewed_at || '-'
    }
  ])

  // ─── Tab switch & init ───
  function handleTabChange(tab: string | number) {
    if (tab === 'applications') {
      loadApplications()
    } else {
      loadEligibleWelfares()
    }
  }

  onMounted(() => {
    loadEligibleWelfares()
    loadApplications()
  })
</script>

<style scoped>
  .welfare-my-page :deep(.welfare-future-row) {
    opacity: 0.58;
  }
</style>
