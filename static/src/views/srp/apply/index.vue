<!-- SRP 补损申请页面 -->
<template>
  <div class="srp-apply-page art-full-height">
    <!-- 我的申请历史 -->
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-base font-medium">{{ $t('srp.apply.title') }}</h2>
          <ElButton type="primary" @click="openApplyDialog">
            <el-icon class="mr-1"><Plus /></el-icon>
            {{ $t('srp.apply.submitBtn') }}
          </ElButton>
        </div>
      </template>

      <ElTable v-loading="loading" :data="applications" stripe border style="width: 100%">
        <ElTableColumn prop="character_name" :label="$t('srp.apply.columns.character')" width="150" />
        <ElTableColumn prop="ship_type_id" :label="$t('srp.apply.columns.ship')" width="180">
          <template #default="{ row }">
            <span>{{ getName(row.ship_type_id, `TypeID: ${row.ship_type_id}`) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="solar_system_id" :label="$t('srp.apply.columns.system')" width="140">
          <template #default="{ row }">
            {{ getName(row.solar_system_id, String(row.solar_system_id)) }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="killmail_id" :label="$t('srp.apply.columns.killId')" width="110" align="center">
          <template #default="{ row }">
            <ElLink :href="`https://zkillboard.com/kill/${row.killmail_id}/`" target="_blank" type="primary">
              {{ row.killmail_id }}
            </ElLink>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="killmail_time" :label="$t('srp.apply.columns.time')" width="180">
          <template #default="{ row }">{{ formatTime(row.killmail_time) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="recommended_amount" :label="$t('srp.apply.columns.recommendedAmount')" width="140" align="right">
          <template #default="{ row }">{{ formatISK(row.recommended_amount) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="final_amount" :label="$t('srp.apply.columns.finalAmount')" width="140" align="right">
          <template #default="{ row }">{{ formatISK(row.final_amount) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="review_status" :label="$t('srp.apply.columns.reviewStatus')" width="110" align="center">
          <template #default="{ row }">
            <ElTag :type="reviewStatusType(row.review_status)" size="small">
              {{ reviewStatusLabel(row.review_status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="payout_status" :label="$t('srp.apply.columns.payoutStatus')" width="110" align="center">
          <template #default="{ row }">
            <ElTag :type="payoutStatusType(row.payout_status)" size="small">
              {{ row.payout_status === 'paid' ? $t('srp.status.paid') : $t('srp.status.unpaid') }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="note" :label="$t('srp.apply.columns.note')" min-width="160" show-overflow-tooltip />
      </ElTable>

      <div v-if="pagination.total > 0" class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.current"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="() => { pagination.current = 1; loadApplications() }"
          @current-change="loadApplications"
        />
      </div>
    </ElCard>

    <ElDialog v-model="applyDialogVisible" :title="$t('srp.apply.dialogTitle')" width="560px" :close-on-click-modal="false">
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px" label-position="right">
        <ElFormItem :label="$t('srp.apply.selectCharacter')" prop="character_id">
          <ElSelect v-model="form.character_id" :placeholder="$t('srp.apply.selectCharacter')" style="width: 100%"
            @change="onCharacterChange">
            <ElOption v-for="c in characters" :key="c.character_id" :label="c.character_name" :value="c.character_id" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem :label="$t('srp.apply.associatedFleet')">
          <ElSelect v-model="form.fleet_id" :placeholder="$t('srp.apply.selectFleet')" clearable filterable style="width: 100%" @change="onFleetChange">
            <ElOption v-for="f in fleets" :key="f.id" :label="f.title" :value="f.id" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem :label="$t('srp.apply.killmail')" prop="killmail_id">
          <ElSelect
            v-model="form.killmail_id"
            :placeholder="form.character_id ? $t('srp.apply.selectKillmail') : $t('srp.apply.noKmHint')"
            :loading="kmLoading"
            :loading-text="$t('srp.apply.loadingKm')"
            style="width: 100%"
            filterable
            :disabled="!form.character_id"
            @change="onKillmailSelect"
          >
            <ElOption
              v-for="km in fleetKillmails"
              :key="km.killmail_id"
              :value="km.killmail_id"
              :label="`#${km.killmail_id}  ${km.victim_name} (${getName(km.ship_type_id, `TypeID: ${km.ship_type_id}`)})`"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem v-if="form.recommended_amount > 0" :label="$t('srp.apply.recommendedAmount')">
          <span class="text-green-600 font-medium">{{ formatISK(form.recommended_amount) }} ISK</span>
        </ElFormItem>

        <ElFormItem :label="$t('srp.apply.finalAmount')">
          <ElInputNumber v-model="form.final_amount" :min="0" :precision="2" :step="1000000" style="width: 100%" />
          <div class="text-xs text-gray-400 mt-1">{{ $t('srp.apply.finalAmountHint') }}</div>
        </ElFormItem>

        <ElFormItem :label="$t('srp.apply.note')" :prop="form.fleet_id ? '' : 'note'">
          <ElInput v-model="form.note" type="textarea" :rows="2"
            :placeholder="form.fleet_id ? $t('srp.apply.notePlaceholder') : $t('srp.apply.noteRequired')" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="applyDialogVisible = false">{{ $t('srp.apply.cancelBtn') }}</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">{{ $t('srp.apply.confirmSubmitBtn') }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useRoute } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import { Plus } from '@element-plus/icons-vue'
  import {
    ElCard, ElTable, ElTableColumn, ElTag, ElButton, ElPagination,
    ElDialog, ElForm, ElFormItem, ElSelect, ElOption, ElInput, ElInputNumber,
    ElLink, ElMessage, type FormInstance, type FormRules
  } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchFleetList } from '@/api/fleet'
  import { submitApplication, fetchMyApplications, fetchFleetKillmails, fetchMyKillmails } from '@/api/srp'
  import { useNameResolver } from '@/hooks'

  defineOptions({ name: 'SrpApply' })

  const route = useRoute()
  const { t } = useI18n()
  const { getName, resolve: resolveNames } = useNameResolver()

  const applications = ref<Api.Srp.Application[]>([])
  const loading = ref(false)
  const pagination = reactive({ current: 1, size: 20, total: 0 })

  /** 从列表数据中收集所有需要解析的 ID，一次性调 SDE /names */
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

  const loadApplications = async () => {
    loading.value = true
    try {
      const res = await fetchMyApplications({ current: pagination.current, size: pagination.size })
      applications.value = res?.records ?? []
      pagination.total = res?.total ?? 0
      if (applications.value.length) await resolveApplicationNames(applications.value)
    } catch { applications.value = [] }
    finally { loading.value = false }
  }

  const characters = ref<Api.Auth.EveCharacter[]>([])
  const loadCharacters = async () => {
    try {
      const list = await fetchMyCharacters()
      characters.value = list ?? []
    } catch { characters.value = [] }
  }

  const fleets = ref<Api.Fleet.FleetItem[]>([])
  const loadFleets = async () => {
    try {
      const res = await fetchFleetList({ size: 200 } as any)
      fleets.value = res?.records ?? []
    } catch { fleets.value = [] }
  }

  const applyDialogVisible = ref(false)
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
    recommended_amount: 0,
  })

  const rules: FormRules = {
    character_id: [{ required: true, message: t('srp.apply.selectCharacter'), trigger: 'change' }],
    killmail_id: [{ required: true, validator: (_r, v, cb) => v > 0 ? cb() : cb(new Error(t('srp.apply.selectKillmail'))), trigger: 'change' }],
    note: [{ validator: (_r: any, v: string, cb: (e?: Error) => void) => {
      if (!form.fleet_id && !v) return cb(new Error(t('srp.apply.noteRequired')))
      cb()
    }, trigger: 'blur' }]
  }

  const loadKillmails = async () => {
    if (!form.character_id) { fleetKillmails.value = []; return }
    kmLoading.value = true
    fleetKillmails.value = []
    form.killmail_id = 0
    try {
      if (form.fleet_id) {
        const list = await fetchFleetKillmails(form.fleet_id)
        fleetKillmails.value = list ?? []
        if (!list?.length) ElMessage.info(t('srp.apply.noKmFound'))
      } else {
        const list = await fetchMyKillmails(form.character_id)
        fleetKillmails.value = list ?? []
      }
      // 解析 KM 列表中的舰船名
      if (fleetKillmails.value.length) {
        const typeIds = [...new Set(fleetKillmails.value.map((km) => km.ship_type_id).filter(Boolean))]
        if (typeIds.length) await resolveNames({ ids: { type: typeIds } })
      }
    } catch { fleetKillmails.value = [] }
    finally { kmLoading.value = false }
  }

  const onCharacterChange = () => { form.killmail_id = 0; form.recommended_amount = 0; loadKillmails() }
  const onFleetChange = () => { form.killmail_id = 0; loadKillmails() }
  const onKillmailSelect = (_: number) => { form.recommended_amount = 0 }

  const openApplyDialog = () => {
    formRef.value?.resetFields()
    form.fleet_id = ''
    form.recommended_amount = 0
    fleetKillmails.value = []
    applyDialogVisible.value = true
  }

  const handleSubmit = async () => {
    await formRef.value?.validate()
    submitting.value = true
    try {
      await submitApplication({
        character_id: form.character_id,
        killmail_id: form.killmail_id,
        fleet_id: form.fleet_id || null,
        note: form.note,
        final_amount: form.final_amount,
      })
      ElMessage.success(t('srp.apply.submitSuccess'))
      applyDialogVisible.value = false
      formRef.value?.resetFields()
      form.fleet_id = ''
      form.recommended_amount = 0
      fleetKillmails.value = []
      loadApplications()
    } catch { /* handled */ }
    finally { submitting.value = false }
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

  onMounted(() => {
    const fid = route.query.fleet_id as string
    if (fid) { form.fleet_id = fid; applyDialogVisible.value = true }
    loadCharacters()
    loadFleets()
    loadApplications()
  })
</script>

<style scoped>
  .pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 16px; }
</style>