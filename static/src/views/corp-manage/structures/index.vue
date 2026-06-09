<!-- 军团建筑管理页面 -->
<template>
  <div class="corp-structures art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="getData">
        <template #left>
          <ElButton
            v-if="canManageFuelSetting"
            type="primary"
            style="margin-right: 12px"
            @click="openFuelSettingDialog"
          >
            {{ $t('corpStructure.actions.openSettings') }}
          </ElButton>

          <ElInput
            v-model="keyword"
            :placeholder="$t('corpStructure.searchPlaceholder')"
            clearable
            style="width: 220px; margin-right: 12px"
            :prefix-icon="Search"
            @input="handleKeywordInput"
            @clear="handleKeywordInput"
          />

          <ElSelect
            v-if="corpIDs.length > 1"
            v-model="selectedCorpID"
            :placeholder="$t('corpStructure.selectCorp')"
            style="width: 220px; margin-right: 12px"
            @change="handleCorpChange"
          >
            <ElOption v-for="id in corpIDs" :key="id" :label="getCorpLabel(id)" :value="id" />
          </ElSelect>

          <ElSelect
            v-model="stateFilter"
            :placeholder="$t('corpStructure.stateFilter')"
            style="width: 180px; margin-right: 12px"
            clearable
            @change="handleFilterChange"
          >
            <ElOption
              v-for="opt in stateOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </ElSelect>

          <span style="display: inline-flex; align-items: center; gap: 6px">
            <ElSwitch v-model="fuelExpiresSoon" @change="handleFilterChange" />
            <span style="font-size: 13px; color: var(--el-text-color-regular)">
              {{ $t('corpStructure.fuelExpiresFilter') }}
            </span>
          </span>
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

    <ElDialog
      v-model="fuelSettingDialogVisible"
      :title="$t('corpStructure.fuelSetting.title')"
      width="760px"
      destroy-on-close
    >
      <div v-loading="fuelSettingLoading">
        <ElForm :model="fuelSettingForm" label-width="210px">
          <ElFormItem :label="$t('corpStructure.fuelSetting.enabled')">
            <ElSwitch v-model="fuelSettingForm.enabled" />
          </ElFormItem>

          <ElFormItem :label="$t('corpStructure.fuelSetting.claimMode')">
            <ElSelect v-model="fuelSettingForm.claim_mode" style="width: 240px">
              <ElOption value="all" :label="$t('corpStructure.fuelSetting.claimModes.all')" />
              <ElOption value="manual" :label="$t('corpStructure.fuelSetting.claimModes.manual')" />
              <ElOption
                value="condition"
                :label="$t('corpStructure.fuelSetting.claimModes.condition')"
              />
              <ElOption value="mixed" :label="$t('corpStructure.fuelSetting.claimModes.mixed')" />
            </ElSelect>
          </ElFormItem>

          <ElFormItem :label="$t('corpStructure.fuelSetting.manualStructures')">
            <ElSelect
              v-model="fuelSettingForm.manual_structure_ids"
              multiple
              filterable
              clearable
              style="width: 100%"
            >
              <ElOption
                v-for="opt in manualStructureOptions"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </ElSelect>
          </ElFormItem>

          <ElFormItem :label="$t('corpStructure.fuelSetting.conditionFuelHours')">
            <ElInputNumber
              v-model="fuelSettingForm.condition_fuel_hours_le"
              :min="0"
              :precision="1"
              :step="6"
            />
          </ElFormItem>

          <ElFormItem :label="$t('corpStructure.fuelSetting.conditionStates')">
            <ElSelect
              v-model="fuelSettingForm.condition_states"
              multiple
              clearable
              style="width: 100%"
            >
              <ElOption
                v-for="opt in stateOptions"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </ElSelect>
          </ElFormItem>

          <ElDivider>{{ $t('corpStructure.fuelSetting.walletSection') }}</ElDivider>
          <ElFormItem :label="$t('corpStructure.fuelSetting.walletEnabled')">
            <ElSwitch v-model="fuelSettingForm.wallet_enabled" />
          </ElFormItem>
          <ElFormItem :label="$t('corpStructure.fuelSetting.walletCalcMode')">
            <ElSelect v-model="fuelSettingForm.wallet_calc_mode" style="width: 240px">
              <ElOption value="per_hour" :label="$t('corpStructure.fuelSetting.calcModes.perHour')" />
              <ElOption value="fixed" :label="$t('corpStructure.fuelSetting.calcModes.fixed')" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem :label="$t('corpStructure.fuelSetting.walletValue')">
            <ElInputNumber
              v-model="fuelSettingForm.wallet_value"
              :min="0"
              :precision="2"
              :step="10"
            />
          </ElFormItem>

          <ElDivider>{{ $t('corpStructure.fuelSetting.iskSection') }}</ElDivider>
          <ElFormItem :label="$t('corpStructure.fuelSetting.iskEnabled')">
            <ElSwitch v-model="fuelSettingForm.isk_enabled" />
          </ElFormItem>
          <ElFormItem :label="$t('corpStructure.fuelSetting.iskCalcMode')">
            <ElSelect v-model="fuelSettingForm.isk_calc_mode" style="width: 240px">
              <ElOption value="per_hour" :label="$t('corpStructure.fuelSetting.calcModes.perHour')" />
              <ElOption value="fixed" :label="$t('corpStructure.fuelSetting.calcModes.fixed')" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem :label="$t('corpStructure.fuelSetting.iskValue')">
            <ElInputNumber
              v-model="fuelSettingForm.isk_value"
              :min="0"
              :precision="2"
              :step="1000000"
            />
          </ElFormItem>
        </ElForm>
      </div>

      <template #footer>
        <ElButton @click="fuelSettingDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="fuelSettingSaving" @click="saveFuelSetting">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, h, onMounted, computed, reactive } from 'vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchCorpStructureList,
    fetchCorpIDs,
    claimCorpStructureFuelTask,
    settleCorpStructureFuelTask,
    fetchCorpStructureFuelSetting,
    updateCorpStructureFuelSetting,
    markCorpStructureFuelTaskIskPaid
  } from '@/api/corp-structure'
  import { fetchNames } from '@/api/sde'
  import {
    ElTag,
    ElSelect,
    ElOption,
    ElSwitch,
    ElInput,
    ElButton,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElMessage,
    ElDivider
  } from 'element-plus'
  import { Search } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'CorpStructures' })
  const { t, locale } = useI18n()
  const userStore = useUserStore()
  const actionLoading = ref<Record<number, boolean>>({})

  const corpIDs = ref<number[]>([])
  const selectedCorpID = ref<number | undefined>(undefined)
  const corpNames = ref<Map<number, string>>(new Map())
  const solarSystemNames = ref<Map<number, string>>(new Map())
  const typeNames = ref<Map<number, string>>(new Map())
  const stateFilter = ref<string>('')
  const fuelExpiresSoon = ref<boolean>(false)
  const keyword = ref<string>('')

  const canManageFuelSetting = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((r) => ['super_admin', 'admin'].includes(r))
  })
  const canOperateFuel = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((r) => ['super_admin', 'admin', 'staff'].includes(r))
  })

  const fuelSettingDialogVisible = ref(false)
  const fuelSettingLoading = ref(false)
  const fuelSettingSaving = ref(false)
  const fuelSettingForm = reactive<Api.CorpStructure.FuelSettingUpdateRequest>({
    corporation_id: 0,
    enabled: false,
    claim_mode: 'all',
    manual_structure_ids: [],
    condition_fuel_hours_le: null,
    condition_states: [],
    contribution_unit: 'hour',
    wallet_enabled: true,
    wallet_calc_mode: 'per_hour',
    wallet_value: 0,
    isk_enabled: false,
    isk_calc_mode: 'per_hour',
    isk_value: 0
  })

  let keywordTimer: ReturnType<typeof setTimeout> | null = null

  function getCorpLabel(id: number): string {
    return corpNames.value.get(id) ?? t('corpStructure.corpPrefix', { id })
  }

  const stateOptions = computed(() => [
    { value: 'shield_vulnerable', label: t('corpStructure.stateLabels.shield_vulnerable') },
    { value: 'armor_reinforce', label: t('corpStructure.stateLabels.armor_reinforce') },
    { value: 'armor_vulnerable', label: t('corpStructure.stateLabels.armor_vulnerable') },
    { value: 'hull_reinforce', label: t('corpStructure.stateLabels.hull_reinforce') },
    { value: 'hull_vulnerable', label: t('corpStructure.stateLabels.hull_vulnerable') },
    { value: 'anchoring', label: t('corpStructure.stateLabels.anchoring') },
    { value: 'unanchored', label: t('corpStructure.stateLabels.unanchored') },
    { value: 'fitting_invulnerable', label: t('corpStructure.stateLabels.fitting_invulnerable') },
    { value: 'onlining_vulnerable', label: t('corpStructure.stateLabels.onlining_vulnerable') }
  ])

  const manualStructureOptions = computed(() => {
    const options = new Map<number, string>()
    ;(data.value as Api.CorpStructure.StructureItem[]).forEach((row) => {
      options.set(row.structure_id, row.name || String(row.structure_id))
    })
    fuelSettingForm.manual_structure_ids.forEach((id) => {
      if (!options.has(id)) options.set(id, `${t('corpStructure.structurePrefix')} ${id}`)
    })
    return Array.from(options.entries()).map(([value, label]) => ({ value, label }))
  })

  const STATE_MAP: Record<string, string> = {
    shield_vulnerable: 'success',
    armor_reinforce: 'warning',
    armor_vulnerable: 'warning',
    hull_reinforce: 'danger',
    hull_vulnerable: 'danger',
    anchoring: 'info',
    unanchored: 'info',
    fitting_invulnerable: 'success',
    onlining_vulnerable: 'info',
    online_deprecated: 'info',
    anchor_vulnerable: 'warning',
    deploy_vulnerable: 'warning',
    unknown: 'info'
  }

  const SERVICE_STATE_MAP: Record<string, string> = {
    online: 'success',
    offline: 'danger',
    cleanup: 'warning'
  }

  function formatReinforceHour(hour: number): string {
    return String(hour).padStart(2, '0') + ':00'
  }

  async function resolveNames(rows: Api.CorpStructure.StructureItem[]) {
    const systemIds = [...new Set(rows.map((r) => r.system_id).filter(Boolean))]
    const typeIds = [...new Set(rows.map((r) => r.type_id).filter(Boolean))]
    if (systemIds.length === 0 && typeIds.length === 0) return
    try {
      const lang = locale.value.startsWith('zh') ? 'zh' : 'en'
      const ids: Record<string, number[]> = {}
      if (systemIds.length > 0) ids['solar_system'] = systemIds
      if (typeIds.length > 0) ids['type'] = typeIds
      const nameMap = await fetchNames({ language: lang, ids })
      const raw = nameMap as Record<string, string>

      const sysMap = new Map<number, string>(solarSystemNames.value)
      systemIds.forEach((id) => {
        if (raw[String(id)]) sysMap.set(id, raw[String(id)])
      })
      solarSystemNames.value = sysMap

      const typeMap = new Map<number, string>(typeNames.value)
      typeIds.forEach((id) => {
        if (raw[String(id)]) typeMap.set(id, raw[String(id)])
      })
      typeNames.value = typeMap
    } catch {}
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    handleSizeChange,
    handleCurrentChange
  } = useTable({
    core: {
      apiFn: fetchCorpStructureList,
      apiParams: { current: 1, size: 20 },
      immediate: false,
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        { prop: 'name', label: t('corpStructure.fields.name'), minWidth: 200, showOverflowTooltip: true },
        {
          prop: 'state',
          label: t('corpStructure.fields.state'),
          width: 160,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            const label = t(`corpStructure.stateLabels.${row.state}`, row.state)
            return h(
              ElTag,
              { type: (STATE_MAP[row.state] ?? 'info') as any, effect: 'dark', size: 'small' },
              () => label
            )
          }
        },
        {
          prop: 'system_id',
          label: t('corpStructure.fields.systemId'),
          width: 160,
          formatter: (row: Api.CorpStructure.StructureItem) =>
            solarSystemNames.value.get(row.system_id) ?? String(row.system_id)
        },
        {
          prop: 'type_id',
          label: t('corpStructure.fields.typeId'),
          width: 220,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            const name = typeNames.value.get(row.type_id) ?? String(row.type_id)
            return h('span', { style: 'display:inline-flex;align-items:center;gap:6px' }, [
              h('img', {
                src: `https://images.newdoublex.space/types/${row.type_id}/icon`,
                style: 'width:24px;height:24px;border-radius:2px;flex-shrink:0',
                loading: 'lazy'
              }),
              h('span', {}, name)
            ])
          }
        },
        {
          prop: 'fuel_expires',
          label: t('corpStructure.fields.fuelExpires'),
          width: 180,
          formatter: (row: Api.CorpStructure.StructureItem) =>
            row.fuel_expires ? new Date(row.fuel_expires).toLocaleString() : '-'
        },
        {
          prop: 'reinforce_hour',
          label: t('corpStructure.fields.reinforceHour'),
          width: 110,
          formatter: (row: Api.CorpStructure.StructureItem) => formatReinforceHour(row.reinforce_hour)
        },
        {
          prop: 'services',
          label: t('corpStructure.fields.services'),
          minWidth: 200,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            if (!row.services || row.services.length === 0) return '-'
            return h(
              'div',
              { class: 'flex flex-wrap gap-1' },
              row.services.map((svc) =>
                h(
                  ElTag,
                  { type: (SERVICE_STATE_MAP[svc.state] ?? 'info') as any, size: 'small' },
                  () => svc.name
                )
              )
            )
          }
        },
        {
          prop: 'state_timer_end',
          label: t('corpStructure.fields.stateTimerEnd'),
          width: 180,
          formatter: (row: Api.CorpStructure.StructureItem) =>
            row.state_timer_end ? new Date(row.state_timer_end).toLocaleString() : '-'
        },
        {
          prop: 'unanchors_at',
          label: t('corpStructure.fields.unanchorsAt'),
          width: 180,
          formatter: (row: Api.CorpStructure.StructureItem) =>
            row.unanchors_at ? new Date(row.unanchors_at).toLocaleString() : '-'
        },
        {
          prop: 'actions',
          label: t('common.operation'),
          width: 280,
          fixed: 'right',
          formatter: (row: Api.CorpStructure.StructureItem) => {
            const ext = row as Api.CorpStructure.StructureItem & {
              fuel_task?: Api.CorpStructure.FuelTask
              can_claim?: boolean
              can_settle?: boolean
              claim_denied_reason?: string
            }

            const controls: any[] = []
            if (canOperateFuel.value && ext.can_claim) {
              controls.push(
                h(
                  ElButton,
                  {
                    type: 'primary',
                    size: 'small',
                    loading: actionLoading.value[row.structure_id] ?? false,
                    onClick: () => handleClaim(ext)
                  },
                  () => t('corpStructure.actions.claim')
                )
              )
            }
            if (canOperateFuel.value && ext.can_settle) {
              controls.push(
                h(
                  ElButton,
                  {
                    type: 'success',
                    size: 'small',
                    loading: actionLoading.value[row.structure_id] ?? false,
                    onClick: () => handleSettle(ext)
                  },
                  () => t('corpStructure.actions.settle')
                )
              )
            }
            if (ext.fuel_task?.status === 'claimed' && !ext.can_settle) {
              controls.push(h(ElTag, { type: 'warning', size: 'small' }, () => t('corpStructure.taskStatus.claimedByOther')))
            }
            if (ext.fuel_task?.status === 'completed') {
              controls.push(
                h(
                  ElTag,
                  { type: 'success', size: 'small' },
                  () =>
                    t('corpStructure.taskStatus.completedWithAmount', {
                      wallet: ext.fuel_task?.wallet_amount ?? 0,
                      isk: ext.fuel_task?.isk_amount ?? 0
                    })
                )
              )
            }
            if (
              canManageFuelSetting.value &&
              ext.fuel_task?.status === 'completed' &&
              (ext.fuel_task?.isk_amount ?? 0) > 0 &&
              ext.fuel_task?.isk_payout_status === 'pending'
            ) {
              controls.push(
                h(
                  ElButton,
                  {
                    type: 'warning',
                    size: 'small',
                    loading: actionLoading.value[row.structure_id] ?? false,
                    onClick: () => handleMarkIskPaid(ext)
                  },
                  () => t('corpStructure.actions.markIskPaid')
                )
              )
            }
            if (ext.fuel_task?.isk_payout_status === 'paid') {
              controls.push(h(ElTag, { type: 'info', size: 'small' }, () => t('corpStructure.taskStatus.iskPaid')))
            }
            if (controls.length === 0) {
              controls.push(
                h(
                  'span',
                  { style: 'color: var(--el-text-color-secondary); font-size: 12px;' },
                  ext.claim_denied_reason || t('corpStructure.taskStatus.notAvailable')
                )
              )
            }
            return h('div', { class: 'flex flex-wrap items-center gap-2' }, controls)
          }
        }
      ]
    },
    hooks: {
      onSuccess: (rows) => resolveNames(rows as Api.CorpStructure.StructureItem[])
    }
  })

  function handleCorpChange(corpID: number) {
    Object.assign(searchParams, { corp_id: corpID })
    getData()
  }

  function handleFilterChange() {
    Object.assign(searchParams, {
      state: stateFilter.value || undefined,
      fuel_expires_soon: fuelExpiresSoon.value || undefined
    })
    getData()
  }

  function handleKeywordInput() {
    if (keywordTimer) clearTimeout(keywordTimer)
    keywordTimer = setTimeout(() => {
      Object.assign(searchParams, { keyword: keyword.value || undefined })
      getData()
    }, 400)
  }

  async function handleClaim(row: Api.CorpStructure.StructureItem) {
    actionLoading.value[row.structure_id] = true
    try {
      await claimCorpStructureFuelTask(row.structure_id)
      ElMessage.success(t('corpStructure.messages.claimSuccess'))
      getData()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.error'))
    } finally {
      actionLoading.value[row.structure_id] = false
    }
  }

  async function handleSettle(row: Api.CorpStructure.StructureItem) {
    actionLoading.value[row.structure_id] = true
    try {
      const result = await settleCorpStructureFuelTask(row.structure_id)
      ElMessage.success(
        t('corpStructure.messages.settleSuccess', {
          wallet: result?.wallet_amount ?? 0,
          isk: result?.isk_amount ?? 0
        })
      )
      getData()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.error'))
    } finally {
      actionLoading.value[row.structure_id] = false
    }
  }

  async function handleMarkIskPaid(row: Api.CorpStructure.StructureItem) {
    const ext = row as Api.CorpStructure.StructureItem & { fuel_task?: Api.CorpStructure.FuelTask }
    const taskID = ext.fuel_task?.id
    if (!taskID) return
    actionLoading.value[row.structure_id] = true
    try {
      await markCorpStructureFuelTaskIskPaid(taskID)
      ElMessage.success(t('corpStructure.messages.markIskPaidSuccess'))
      getData()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.error'))
    } finally {
      actionLoading.value[row.structure_id] = false
    }
  }

  async function openFuelSettingDialog() {
    if (!selectedCorpID.value) {
      ElMessage.warning(t('corpStructure.messages.selectCorpFirst'))
      return
    }
    fuelSettingDialogVisible.value = true
    await loadFuelSetting(selectedCorpID.value)
  }

  async function loadFuelSetting(corpID: number) {
    fuelSettingLoading.value = true
    try {
      const setting = await fetchCorpStructureFuelSetting(corpID)
      Object.assign(fuelSettingForm, {
        corporation_id: corpID,
        enabled: setting?.enabled ?? false,
        claim_mode: setting?.claim_mode ?? 'all',
        manual_structure_ids: setting?.manual_structure_ids ?? [],
        condition_fuel_hours_le: setting?.condition_fuel_hours_le ?? null,
        condition_states: setting?.condition_states ?? [],
        contribution_unit: 'hour',
        wallet_enabled: setting?.wallet_enabled ?? true,
        wallet_calc_mode: setting?.wallet_calc_mode ?? 'per_hour',
        wallet_value: setting?.wallet_value ?? 0,
        isk_enabled: setting?.isk_enabled ?? false,
        isk_calc_mode: setting?.isk_calc_mode ?? 'per_hour',
        isk_value: setting?.isk_value ?? 0
      })
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.error'))
    } finally {
      fuelSettingLoading.value = false
    }
  }

  async function saveFuelSetting() {
    if (!fuelSettingForm.corporation_id) {
      ElMessage.warning(t('corpStructure.messages.selectCorpFirst'))
      return
    }
    fuelSettingSaving.value = true
    try {
      await updateCorpStructureFuelSetting({
        ...fuelSettingForm,
        condition_fuel_hours_le:
          fuelSettingForm.condition_fuel_hours_le === null ||
          fuelSettingForm.condition_fuel_hours_le === undefined
            ? null
            : Number(fuelSettingForm.condition_fuel_hours_le)
      })
      ElMessage.success(t('corpStructure.messages.settingSaved'))
      fuelSettingDialogVisible.value = false
      getData()
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.error'))
    } finally {
      fuelSettingSaving.value = false
    }
  }

  async function loadCorpIDs() {
    try {
      const res = await fetchCorpIDs()
      corpIDs.value = res ?? []
      if (corpIDs.value.length > 0) {
        try {
          const nameMap = await fetchNames({ esi: corpIDs.value })
          const map = new Map<number, string>()
          corpIDs.value.forEach((id) => {
            const name = (nameMap as Record<string, string>)[String(id)]
            if (name) map.set(id, name)
          })
          corpNames.value = map
        } catch {}
        selectedCorpID.value = corpIDs.value[0]
        Object.assign(searchParams, { corp_id: corpIDs.value[0] })
        getData()
      }
    } catch {
      corpIDs.value = []
    }
  }

  onMounted(() => {
    loadCorpIDs()
  })
</script>
