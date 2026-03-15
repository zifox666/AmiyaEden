<!-- SRP 补损申请页面 -->
<template>
  <div class="srp-apply-page art-full-height">
    <!-- 申请补损 -->
    <ElCard class="apply-card" shadow="never">
      <template #header>
        <h2 class="section-title">{{ $t('srp.apply.formTitle') }}</h2>
      </template>

      <ElAlert type="success" :closable="false" class="mb-4" show-icon>
        <p>{{ $t('srp.apply.infoText') }}</p>
        <ElLink type="primary" class="mt-1">{{ $t('srp.apply.faqLink') }}</ElLink>
      </ElAlert>

      <ElForm ref="formRef" :model="form" :rules="rules" label-position="top">
        <ElFormItem :label="$t('srp.apply.killmail')" prop="killmail_id">
          <div class="km-select-row">
            <ElSelect
              v-model="form.killmail_id"
              :placeholder="$t('srp.apply.selectKillmail')"
              :loading="kmLoading"
              :loading-text="$t('srp.apply.loadingKm')"
              class="flex-1"
              filterable
              @change="onKillmailSelect"
            >
              <ElOption
                v-for="km in fleetKillmails"
                :key="km.killmail_id"
                :value="km.killmail_id"
                :label="formatKmLabel(km)"
                :disabled="submittedKmIds.has(km.killmail_id)"
              />
            </ElSelect>
            <ElButton :disabled="!form.killmail_id" @click="openKmPreview">
              <el-icon class="mr-1"><View /></el-icon>
              {{ $t('srp.apply.previewKm') }}
            </ElButton>
          </div>
        </ElFormItem>

        <div class="fleet-info-section">
          <ElFormItem>
            <ElSelect
              v-model="form.fleet_id"
              :placeholder="$t('srp.apply.selectFleet')"
              clearable
              filterable
              style="width: 100%"
              @change="onFleetChange"
            >
              <ElOption key="__other__" :label="$t('srp.apply.otherAction')" value="__other__" />
              <ElOption v-for="f in fleets" :key="f.id" :label="formatFleetLabel(f)" :value="f.id" />
            </ElSelect>
          </ElFormItem>

          <!-- 选择指定舰队时：显示舰队详情（只读） -->
          <div v-if="selectedFleetDetail" class="fleet-detail-card">
            <h4 class="fleet-detail-title">{{ $t('srp.apply.fleetDetailTitle') }}</h4>
            <div class="fleet-detail-row">
              <span class="fleet-detail-label">{{ $t('srp.apply.fleetDetailFC') }}:</span>
              <span>{{ selectedFleetDetail.fc_character_name }}</span>
            </div>
            <div class="fleet-detail-row">
              <span class="fleet-detail-label">{{ $t('srp.apply.fleetDetailTime') }}:</span>
              <span
                >{{ formatTime(selectedFleetDetail.start_at) }} ~
                {{ formatTime(selectedFleetDetail.end_at) }}</span
              >
            </div>
            <div class="fleet-detail-row">
              <span class="fleet-detail-label">{{ $t('srp.apply.fleetDetailImportance') }}:</span>
              <ElTag size="small" :type="importanceTagType(selectedFleetDetail.importance)">{{
                selectedFleetDetail.importance
              }}</ElTag>
            </div>
            <div v-if="selectedFleetDetail.description" class="fleet-detail-row">
              <span class="fleet-detail-label">{{ $t('srp.apply.fleetDetailDesc') }}:</span>
              <span>{{ selectedFleetDetail.description }}</span>
            </div>
          </div>

          <!-- 选择"其他行动"时：显示可编辑备注 -->
          <ElFormItem v-if="showNoteArea" :prop="noteRequired ? 'note' : ''">
            <ElInput
              v-model="form.note"
              type="textarea"
              :rows="3"
              :placeholder="$t('srp.apply.fleetNotePlaceholder')"
            />
          </ElFormItem>
        </div>

        <div class="flex justify-end mt-2">
          <ElButton type="success" :loading="submitting" @click="handleSubmit">
            {{ $t('srp.apply.submitBtnText') }}
          </ElButton>
        </div>
      </ElForm>
    </ElCard>

    <!-- 我的补损申请 -->
    <ElCard class="art-table-card mt-4" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- KM 预览弹窗 -->
    <KmPreviewDialog v-model="kmPreviewVisible" :killmail-id="previewKillmailId" />
  </div>
</template>

<script setup lang="ts">
  import { useRoute } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import { View } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElTag,
    ElButton,
    ElForm,
    ElFormItem,
    ElSelect,
    ElOption,
    ElInput,
    ElLink,
    ElMessage,
    ElAlert,
    ElTooltip,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import KmPreviewDialog from '@/components/business/KmPreviewDialog.vue'
  import { fetchMyFleetList } from '@/api/fleet'
  import {
    submitApplication,
    fetchMyApplications,
    fetchFleetKillmails,
    fetchMyKillmails
  } from '@/api/srp'
  import { useNameResolver } from '@/hooks'

  defineOptions({ name: 'SrpApply' })

  const OTHER_ACTION = '__other__'
  const route = useRoute()
  const { t } = useI18n()
  const { getName, resolve: resolveNames } = useNameResolver()

  /* ── 申请列表 ── */
  const resolveApplicationNames = async (list: Api.Srp.Application[]) => {
    const typeIds = new Set<number>()
    const solarIds = new Set<number>()
    for (const app of list) {
      if (app.ship_type_id) typeIds.add(app.ship_type_id)
      if (app.solar_system_id) solarIds.add(app.solar_system_id)
    }
    await resolveNames({
      ids: {
        ...(typeIds.size ? { type: [...typeIds] } : {}),
        ...(solarIds.size ? { solar_system: [...solarIds] } : {})
      }
    })
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchMyApplications,
      apiParams: { current: 1, size: 5 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'killmail_id',
          label: t('srp.apply.columns.id'),
          width: 120,
          formatter: (row: Api.Srp.Application) =>
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
          prop: 'character_name',
          label: t('srp.apply.columns.character'),
          width: 130,
          showOverflowTooltip: true
        },
        {
          prop: 'ship_type_id',
          label: t('srp.apply.columns.ship'),
          minWidth: 140,
          showOverflowTooltip: true,
          formatter: (row: Api.Srp.Application) =>
            h('span', {}, getName(row.ship_type_id, `TypeID: ${row.ship_type_id}`))
        },
        {
          prop: 'recommended_amount',
          label: t('srp.apply.columns.estimatedValue'),
          width: 140,
          formatter: (row: Api.Srp.Application) =>
            h('span', {}, `${formatISK(row.recommended_amount)} ISK`)
        },
        {
          prop: 'review_status',
          label: t('srp.apply.columns.reviewStatus'),
          width: 110,
          formatter: (row: Api.Srp.Application) =>
            h(ElTag, { type: reviewStatusType(row.review_status), size: 'small' }, () =>
              reviewStatusLabel(row.review_status)
            )
        },
        {
          prop: 'final_amount',
          label: t('srp.apply.columns.actualAmount'),
          width: 130,
          formatter: (row: Api.Srp.Application) =>
            row.final_amount > 0
              ? h('span', {}, `${formatISK(row.final_amount)} ISK`)
              : h('span', {}, '-')
        },
        {
          prop: 'payout_status',
          label: t('srp.apply.columns.paid'),
          width: 100,
          formatter: (row: Api.Srp.Application) =>
            h(
              'span',
              {},
              row.payout_status === 'paid' ? t('srp.status.paid') : t('srp.status.unpaid')
            )
        },
        {
          prop: 'fleet_id',
          label: t('srp.apply.columns.fleetNote'),
          minWidth: 140,
          formatter: (row: Api.Srp.Application) => {
            if (row.fleet_id) {
              const fleet = fleetMap.value.get(row.fleet_id)
              const tooltipContent = fleet
                ? formatFleetLabel(fleet)
                : row.fleet_title || row.fleet_id
              return h(
                ElTooltip,
                { content: tooltipContent, placement: 'top' },
                () => h('span', { class: 'cursor-default' }, row.fleet_title || row.fleet_id || '')
              )
            }
            return h('span', { class: row.note ? '' : 'text-gray-400' }, row.note || '-')
          }
        },
        {
          prop: 'actions',
          label: t('srp.apply.columns.action'),
          width: 80,
          fixed: 'right',
          formatter: (row: Api.Srp.Application) =>
            h(ArtButtonTable, { type: 'view', onClick: () => openTableKmPreview(row) })
        }
      ]
    }
  })

  watch(data, async (list) => {
    if (list.length) await resolveApplicationNames(list)
  })

  /* ── 角色 & 舰队 ── */
  const fleets = ref<Api.Fleet.FleetItem[]>([])
  const fleetMap = computed(() => new Map(fleets.value.map((f) => [f.id, f])))
  const loadFleets = async () => {
    try {
      const list = await fetchMyFleetList()
      fleets.value = list ?? []
    } catch {
      fleets.value = []
    }
  }

  /* ── 表单 ── */
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const kmLoading = ref(false)
  const fleetKillmails = ref<Api.Srp.FleetKillmailItem[]>([])

  const form = reactive({
    character_id: 0,
    fleet_id: '',
    killmail_id: 0,
    note: '',
    final_amount: 0,
    recommended_amount: 0
  })

  const showNoteArea = computed(() => form.fleet_id === OTHER_ACTION)
  const noteRequired = computed(() => form.fleet_id === OTHER_ACTION || !form.fleet_id)
  const submittedKmIds = computed(() => new Set((data.value ?? []).map((a) => a.killmail_id)))

  // 选中的舰队详情（非"__other__"且非空时显示）
  const selectedFleetDetail = computed(() => {
    if (!form.fleet_id || form.fleet_id === OTHER_ACTION) return null
    return fleets.value.find((f) => f.id === form.fleet_id) ?? null
  })

  const importanceTagType = (v: string): TagType =>
    (({ strat_op: 'danger', cta: 'warning', other: 'info' }) as Record<string, TagType>)[v] ??
    'info'

  const rules: FormRules = {
    killmail_id: [
      {
        required: true,
        validator: (_r, v, cb) => (v > 0 ? cb() : cb(new Error(t('srp.apply.selectKillmail')))),
        trigger: 'change'
      }
    ],
    note: [
      {
        validator: (_r: any, v: string, cb: (e?: Error) => void) => {
          if (noteRequired.value && !v) return cb(new Error(t('srp.apply.noteRequired')))
          cb()
        },
        trigger: 'blur'
      }
    ]
  }

  const formatKmLabel = (km: Api.Srp.FleetKillmailItem) =>
    `${km.killmail_id}: ${getName(km.ship_type_id, `TypeID: ${km.ship_type_id}`)}` +
    `(${km.victim_name}) - ${formatTime(km.killmail_time)}` +
    ` @${getName(km.solar_system_id, String(km.solar_system_id))}`

  const loadKillmails = async () => {
    kmLoading.value = true
    fleetKillmails.value = []
    const prevKmId = form.killmail_id
    form.killmail_id = 0
    try {
      if (form.fleet_id && form.fleet_id !== OTHER_ACTION) {
        const list = await fetchFleetKillmails(form.fleet_id)
        fleetKillmails.value = list ?? []
        if (!list?.length) ElMessage.info(t('srp.apply.noKmFound'))
      } else {
        const list = await fetchMyKillmails()
        fleetKillmails.value = list ?? []
      }

      if (fleetKillmails.value.length) {
        const typeIds = [
          ...new Set(fleetKillmails.value.map((km) => km.ship_type_id).filter(Boolean))
        ]
        const solarIds = [
          ...new Set(fleetKillmails.value.map((km) => km.solar_system_id).filter(Boolean))
        ]
        const idsToResolve: Record<string, number[]> = {}
        if (typeIds.length) idsToResolve.type = typeIds
        if (solarIds.length) idsToResolve.solar_system = solarIds
        if (Object.keys(idsToResolve).length > 0) {
          await resolveNames({ ids: idsToResolve })
        }
      }

      // 如果之前选中的 KM 仍在新列表中，则保留选中状态
      if (prevKmId && fleetKillmails.value.some((k) => k.killmail_id === prevKmId)) {
        form.killmail_id = prevKmId
      }
    } catch {
      fleetKillmails.value = []
    } finally {
      kmLoading.value = false
    }
  }

  const onFleetChange = () => {
    if (form.fleet_id !== OTHER_ACTION) {
      form.note = ''
    }
    loadKillmails()
  }

  const onKillmailSelect = (kmId: number) => {
    // 从选中的 KM 自动推导 character_id
    const km = fleetKillmails.value.find((k) => k.killmail_id === kmId)
    form.character_id = km?.character_id ?? 0
    form.recommended_amount = 0
  }

  const handleSubmit = async () => {
    await formRef.value?.validate()
    submitting.value = true
    try {
      const fleetId = form.fleet_id === OTHER_ACTION ? null : form.fleet_id || null
      await submitApplication({
        character_id: form.character_id,
        killmail_id: form.killmail_id,
        fleet_id: fleetId,
        note: form.note,
        final_amount: form.final_amount
      })
      ElMessage.success(t('srp.apply.submitSuccess'))
      formRef.value?.resetFields()
      form.fleet_id = ''
      form.recommended_amount = 0
      form.character_id = 0
      loadKillmails()
      refreshData()
    } catch {
      /* handled */
    } finally {
      submitting.value = false
    }
  }

  /* ── KM 预览 ── */
  const kmPreviewVisible = ref(false)
  const previewKillmailId = ref(0)

  const openKmPreview = () => {
    if (!form.killmail_id) return
    previewKillmailId.value = form.killmail_id
    kmPreviewVisible.value = true
  }

  const openTableKmPreview = (row: Api.Srp.Application) => {
    previewKillmailId.value = row.killmail_id
    kmPreviewVisible.value = true
  }

  /* ── 工具函数 ── */
  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')
  const formatShortTime = (v: string) => {
    if (!v) return '-'
    const d = new Date(v)
    return `${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
  const formatFleetLabel = (f: Api.Fleet.FleetItem) =>
    `${f.fc_character_name}: ${f.title} (${f.pap_count}PAP) @ ${formatShortTime(f.start_at)}~${formatShortTime(f.end_at)}`
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(v ?? 0)

  type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'
  const reviewStatusType = (s: string): TagType =>
    (({ pending: 'info', approved: 'success', rejected: 'danger' }) as Record<string, TagType>)[
      s
    ] ?? 'info'
  const reviewStatusLabel = (s: string) =>
    ({
      pending: t('srp.status.pending'),
      approved: t('srp.status.approved'),
      rejected: t('srp.status.rejected')
    })[s as 'pending' | 'approved' | 'rejected'] ?? s

  /* ── 初始化 ── */
  onMounted(async () => {
    const fid = route.query.fleet_id as string
    await loadFleets()
    if (fid) {
      form.fleet_id = fid
    }
    loadKillmails()
  })
</script>

<style scoped>
  .apply-card :deep(.el-card__header) {
    padding: 12px 16px;
  }

  .km-select-row {
    display: flex;
    gap: 8px;
    width: 100%;
  }

  .fleet-info-section {
    padding-top: 4px;
    margin-top: 4px;
  }

  .fleet-info-label {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 12px;
    color: #606266;
  }

  .fleet-detail-card {
    border-radius: 6px;
    padding: 12px 16px;
    margin-bottom: 16px;
  }

  .fleet-detail-title {
    font-size: 13px;
    font-weight: 600;
    margin-bottom: 8px;
  }

  .fleet-detail-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    line-height: 24px;
  }

  .fleet-detail-label {
    font-weight: 500;
    min-width: 50px;
  }
</style>
