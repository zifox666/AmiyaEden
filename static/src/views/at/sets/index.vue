<template>
  <div class="art-card p-5">
    <div class="flex-b mb-4">
      <div
        ><h2 class="text-xl">{{ t('at.sets.title') }}</h2
        ><p class="text-sm text-g-500 mt-1">{{ t('at.sets.workflowDescription') }}</p></div
      >
    </div>
    <ElTabs v-model="activeTab">
      <ElTabPane name="fits" :label="t('at.sets.fitLibrary')">
        <div class="flex-b mb-4">
          <ElInput
            v-model="fitKeyword"
            clearable
            :placeholder="t('at.sets.searchFits')"
            style="max-width: 320px"
          />
          <ElButton v-if="canManage" type="primary" @click="newFit">{{
            t('at.sets.createFit')
          }}</ElButton>
        </div>
        <div v-loading="loading" class="fit-grid">
          <button
            v-for="fit in filteredFits"
            :key="fit.id"
            class="fit-card"
            @click="viewFit(fit.id)"
          >
            <img :src="shipIcon(fit.ship_type_id, 64)" :alt="fit.ship_name" />
            <div
              ><strong>{{ fit.name }}</strong
              ><span>{{ fit.ship_name }}</span
              ><p
                ><ElTag
                  v-for="tag in [...(fit.base_ewar_tags || []), ...(fit.base_bonus_tags || [])]"
                  :key="tag"
                  size="small"
                  >{{ tag }}</ElTag
                ></p
              ></div
            >
          </button>
        </div>
        <ElEmpty v-if="!loading && !filteredFits.length" :description="t('at.sets.noFits')" />
      </ElTabPane>

      <ElTabPane name="sets" :label="t('at.sets.compositionSets')">
        <div class="flex-b mb-4"
          ><span class="text-sm text-g-500">{{ t('at.sets.compositionHelp') }}</span
          ><ElButton v-if="canManage" type="primary" @click="newSet">{{
            t('at.sets.create')
          }}</ElButton></div
        >
        <ElTable :data="sets" v-loading="loading">
          <ElTableColumn prop="name" :label="t('at.common.name')" min-width="220" />
          <ElTableColumn prop="format" :label="t('at.common.format')" width="110" />
          <ElTableColumn prop="status" :label="t('at.common.status')" width="130" />
          <ElTableColumn :label="t('at.common.actions')" width="180"
            ><template #default="{ row }"
              ><ElButton link type="primary" @click="viewSet(row.id)">{{
                t('at.common.view')
              }}</ElButton
              ><ElButton v-if="canManage" link @click="editSet(row.id)">{{
                t('at.common.edit')
              }}</ElButton></template
            ></ElTableColumn
          >
        </ElTable>
      </ElTabPane>
    </ElTabs>

    <ElDrawer
      v-model="fitDrawer"
      :title="editingFitId ? t('at.sets.editFit') : t('at.sets.createFit')"
      size="780px"
    >
      <ElForm label-position="top">
        <ElFormItem :label="t('at.sets.baseEft')"
          ><ElInput
            v-model="fitForm.base_eft"
            type="textarea"
            :rows="10"
            :placeholder="t('at.sets.eftPlaceholder')"
        /></ElFormItem>
        <ElButton type="primary" plain :loading="parsingFit" @click="parseFit">{{
          t('at.sets.parseEft')
        }}</ElButton>
        <div v-if="parsedFit" class="fit-preview mt-4"
          ><img :src="shipIcon(parsedFit.ship_type_id, 64)" /><div
            ><strong>{{ derivedFitName }}</strong
            ><div class="text-sm text-g-500"
              >{{ parsedFit.ship_name }} · {{ parsedFit.items.length }}
              {{ t('at.sets.items') }}</div
            ></div
          ></div
        >
        <ElRow :gutter="12" class="mt-4"
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.sets.baseEwarTags')"
              ><ElSelect
                v-model="fitForm.base_ewar_tags"
                multiple
                filterable
                allow-create
                default-first-option
                class="w-full" /></ElFormItem></ElCol
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.sets.baseBonusTags')"
              ><ElSelect
                v-model="fitForm.base_bonus_tags"
                multiple
                filterable
                allow-create
                default-first-option
                class="w-full" /></ElFormItem></ElCol
        ></ElRow>
        <div class="flex-b mb-2"
          ><strong>{{ t('at.sets.variantRules') }}</strong
          ><ElButton @click="addReplacement">{{ t('at.sets.addReplacement') }}</ElButton></div
        >
        <div v-for="(rule, index) in fitForm.replacements" :key="index" class="rule-row">
          <ElSelect v-model="rule.tag_type"
            ><ElOption :label="t('at.sets.ewar')" value="ewar" /><ElOption
              :label="t('at.sets.bonus')"
              value="bonus"
          /></ElSelect>
          <ElInput v-model="rule.tag" :placeholder="t('at.sets.variantTag')" />
          <ElSelect v-model="rule.operation"
            ><ElOption :label="t('at.sets.operationReplace')" value="replace" /><ElOption
              :label="t('at.sets.operationAdd')"
              value="add" /><ElOption :label="t('at.sets.operationRemove')" value="remove"
          /></ElSelect>
          <ElSelect v-if="rule.operation !== 'add'" v-model="rule.target_type_id" filterable
            ><ElOption
              v-for="item in parsedFit?.items || []"
              :key="item.type_id"
              :label="item.type_name"
              :value="item.type_id" /></ElSelect
          ><span v-else />
          <SdeSearchSelect
            v-if="rule.operation !== 'remove'"
            v-model="rule.replacement_type_id"
            :placeholder="t('at.sets.searchReplacement')"
          /><span v-else /> <ElInputNumber v-model="rule.quantity" :min="1" /><ElButton
            link
            type="danger"
            @click="fitForm.replacements.splice(index, 1)"
            >{{ t('at.common.remove') }}</ElButton
          >
        </div>
      </ElForm>
      <template #footer
        ><ElButton @click="fitDrawer = false">{{ t('at.common.cancel') }}</ElButton
        ><ElButton type="primary" :loading="saving" @click="saveFit">{{
          t('at.common.save')
        }}</ElButton></template
      >
    </ElDrawer>

    <ElDialog v-model="fitDetailVisible" :title="fitDetail?.fit.name || ''" width="720px">
      <div v-if="fitDetail"
        ><div class="fit-preview"
          ><img :src="shipIcon(fitDetail.fit.ship_type_id, 64)" /><div
            ><h3>{{ fitDetail.fit.name }}</h3
            ><span class="text-sm text-g-500">{{ fitDetail.fit.ship_name }}</span></div
          ></div
        >
        <div class="slot-groups mt-4"
          ><section v-for="group in groupedItems(fitDetail.parsed.items)" :key="group.flag"
            ><header>{{ group.flag }}</header
            ><div
              v-for="item in group.items"
              :key="`${item.type_id}-${item.flag}`"
              class="slot-item"
              ><img :src="shipIcon(item.type_id, 32)" /><span>{{ item.type_name }}</span
              ><small>x{{ item.quantity }}</small></div
            ></section
          ></div
        > </div
      ><template #footer
        ><ElButton v-if="canManage" @click="editFit">{{ t('at.common.edit') }}</ElButton
        ><ElButton @click="copyText(fitDetail?.fit.base_eft || '')"
          >{{ t('at.common.copy') }} EFT</ElButton
        ></template
      >
    </ElDialog>

    <ElDrawer
      v-model="setDrawer"
      :title="editingSetId ? t('at.sets.edit') : t('at.sets.create')"
      size="920px"
    >
      <ElForm label-position="top"
        ><ElRow :gutter="12"
          ><ElCol :span="8"
            ><ElFormItem :label="t('at.common.name')"
              ><ElInput v-model="setForm.name" /></ElFormItem></ElCol
          ><ElCol :span="8"
            ><ElFormItem :label="t('at.common.format')"
              ><ElSelect v-model="setForm.format" class="w-full"
                ><ElOption label="7v7" value="7v7" /><ElOption
                  label="10v10"
                  value="10v10" /></ElSelect></ElFormItem></ElCol
          ><ElCol :span="8"
            ><ElFormItem :label="t('at.sets.rule')"
              ><ElSelect v-model="setForm.rule_version_id" class="w-full"
                ><ElOption
                  v-for="rule in publishedRules"
                  :key="rule.id"
                  :label="rule.version"
                  :value="rule.id" /></ElSelect></ElFormItem></ElCol></ElRow
      ></ElForm>
      <ElAlert
        :type="scoreValid ? 'success' : 'warning'"
        :closable="false"
        class="mb-3"
        :title="t('at.sets.scoreSummary', { score: compositionScore, max: 200 })"
      />
      <ElTable :data="setSlots" row-key="key"
        ><ElTableColumn type="index" :label="t('at.sets.position')" width="80" /><ElTableColumn
          :label="t('at.sets.shipFit')"
          min-width="240"
          ><template #default="{ row }"
            ><ElSelect v-model="row.fit_id" filterable class="w-full" @change="selectFit(row)"
              ><ElOption
                v-for="fit in fits"
                :key="fit.id"
                :label="fit.name"
                :value="fit.id" /></ElSelect></template></ElTableColumn
        ><ElTableColumn :label="t('at.sets.ewarTags')" min-width="180"
          ><template #default="{ row }"
            ><ElSelect v-model="row.ewar_tags" multiple class="w-full"
              ><ElOption
                v-for="tag in fitOptions(row).ewar"
                :key="tag"
                :label="tag"
                :value="tag" /></ElSelect></template></ElTableColumn
        ><ElTableColumn :label="t('at.sets.bonusTags')" min-width="180"
          ><template #default="{ row }"
            ><ElSelect v-model="row.bonus_tags" multiple class="w-full"
              ><ElOption
                v-for="tag in fitOptions(row).bonus"
                :key="tag"
                :label="tag"
                :value="tag" /></ElSelect></template></ElTableColumn
        ><ElTableColumn :label="t('at.scoreRules.points')" width="90"
          ><template #default="{ row }">{{ slotPoints(row) }}</template></ElTableColumn
        ><ElTableColumn width="70"
          ><template #default="{ $index }"
            ><ElButton link type="danger" @click="setSlots.splice($index, 1)">×</ElButton></template
          ></ElTableColumn
        ></ElTable
      >
      <ElButton class="mt-3" :disabled="setSlots.length >= maxSlots" @click="addSlot">{{
        t('at.sets.addPosition')
      }}</ElButton>
      <template #footer
        ><ElButton @click="setDrawer = false">{{ t('at.common.cancel') }}</ElButton
        ><ElButton type="primary" :loading="saving" :disabled="!scoreValid" @click="saveSet">{{
          t('at.common.save')
        }}</ElButton></template
      >
    </ElDrawer>

    <ElDialog v-model="setDetailVisible" :title="setDetail?.set.name || ''" width="900px"
      ><div v-if="setDetail"
        ><ElAlert
          type="success"
          :closable="false"
          class="mb-3"
          :title="`${setDetail.score.total_points} / ${setDetail.score.max_points}`"
        /><div class="composition-grid"
          ><div
            v-for="bundle in setDetail.slots"
            :key="bundle.slot.slot_order"
            class="position-card"
            ><span class="position-number">{{ bundle.slot.slot_order }}</span
            ><img :src="shipIcon(bundle.slot.ship_type_id, 64)" /><div
              ><strong>{{ positionName(bundle) }}</strong
              ><div class="mt-1"
                ><ElTag
                  v-for="tag in [...bundle.slot.ewar_tags, ...bundle.slot.bonus_tags]"
                  :key="tag"
                  size="small"
                  >{{ tag }}</ElTag
                ></div
              ></div
            ></div
          ></div
        ></div
      ><template #footer
        ><ElButton @click="copySetEFT">{{ t('at.sets.wholeTeam') }} EFT</ElButton></template
      ></ElDialog
    >
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'
  import { atApi } from '@/api/at'
  import { useUserStore } from '@/store/modules/user'
  const { t } = useI18n(),
    user = useUserStore()
  const canManage = computed(() =>
    (user.info.roles || []).some((r) => ['fc', 'admin', 'super_admin'].includes(r))
  )
  const activeTab = ref('fits'),
    loading = ref(false),
    saving = ref(false),
    fitKeyword = ref(''),
    fits = ref<Api.AT.Fit[]>([]),
    sets = ref<Api.AT.Set[]>([]),
    rules = ref<Api.AT.ScoreRule[]>([]),
    ruleDetail = ref<Api.AT.ScoreRuleDetail>()
  const fitDrawer = ref(false),
    fitDetailVisible = ref(false),
    setDrawer = ref(false),
    setDetailVisible = ref(false),
    parsingFit = ref(false),
    editingFitId = ref<number | null>(null),
    editingSetId = ref<number | null>(null),
    parsedFit = ref<Api.AT.ParsedEFT>(),
    fitDetail = ref<Api.AT.FitDetail>(),
    setDetail = ref<Api.AT.SetDetail>()
  const fitForm = reactive<Api.AT.FitInput>({
    base_eft: '',
    base_ewar_tags: [],
    base_bonus_tags: [],
    replacements: []
  })
  type SlotForm = {
    key: number
    fit_id: number
    ship_type_id: number
    ewar_tags: string[]
    bonus_tags: string[]
  }
  let key = 0
  const setSlots = ref<SlotForm[]>([]),
    setForm = reactive({ name: '', format: '10v10' as '7v7' | '10v10', rule_version_id: 0 })
  const publishedRules = computed(() => rules.value.filter((r) => r.status === 'published')),
    filteredFits = computed(() => {
      const q = fitKeyword.value.toLowerCase()
      return fits.value.filter((f) => !q || `${f.name} ${f.ship_name}`.toLowerCase().includes(q))
    }),
    maxSlots = computed(() => (setForm.format === '10v10' ? 10 : 7))
  const penalties = computed(() =>
      Object.fromEntries((ruleDetail.value?.penalties || []).map((x) => [x.size_code, x.penalty]))
    ),
    shipRules = computed(() =>
      Object.fromEntries((ruleDetail.value?.ships || []).map((x) => [x.type_id, x]))
    ),
    counts = computed(() =>
      setSlots.value.reduce<Record<number, number>>(
        (m, x) => ((m[x.ship_type_id] = (m[x.ship_type_id] || 0) + 1), m),
        {}
      )
    )
  const slotPoints = (value: unknown) => {
    const row = value as SlotForm
    const r = shipRules.value[row.ship_type_id]
    return r
      ? r.base_points +
          ((counts.value[row.ship_type_id] || 1) - 1) * (penalties.value[r.size_code] || 0)
      : 0
  }
  const compositionScore = computed(() => setSlots.value.reduce((n, x) => n + slotPoints(x), 0)),
    scoreValid = computed(
      () =>
        setSlots.value.length > 0 &&
        setSlots.value.length <= maxSlots.value &&
        compositionScore.value <= 200 &&
        setSlots.value.every((x) => x.fit_id && shipRules.value[x.ship_type_id])
    )
  const derivedFitName = computed(() =>
    parsedFit.value
      ? `${parsedFit.value.ship_name}${[...fitForm.base_ewar_tags, ...fitForm.base_bonus_tags].length ? ' · ' + [...fitForm.base_ewar_tags, ...fitForm.base_bonus_tags].sort().join('+') : ''}`
      : ''
  )
  const shipIcon = (id: number, size: number) =>
    `https://images.evetech.net/types/${id}/icon?size=${size}`
  function groupedItems(items: Api.AT.ParsedEFTItem[]) {
    const m = new Map<string, Api.AT.ParsedEFTItem[]>()
    for (const x of items) {
      if (!m.has(x.flag)) m.set(x.flag, [])
      m.get(x.flag)!.push(x)
    }
    return [...m].map(([flag, rows]) => ({ flag, items: rows }))
  }
  function addReplacement() {
    fitForm.replacements.push({
      tag_type: 'ewar',
      tag: '',
      operation: 'replace',
      target_type_id: 0,
      replacement_type_id: 0,
      quantity: 1
    })
  }
  async function parseFit() {
    parsingFit.value = true
    try {
      parsedFit.value = await atApi.parseSetEFT(fitForm.base_eft)
    } finally {
      parsingFit.value = false
    }
  }
  function newFit() {
    editingFitId.value = null
    Object.assign(fitForm, {
      base_eft: '',
      base_ewar_tags: [],
      base_bonus_tags: [],
      replacements: []
    })
    parsedFit.value = undefined
    fitDrawer.value = true
  }
  async function viewFit(id: number) {
    fitDetail.value = await atApi.getFit(id)
    fitDetailVisible.value = true
  }
  function editFit() {
    if (!fitDetail.value) return
    editingFitId.value = fitDetail.value.fit.id
    Object.assign(fitForm, {
      base_eft: fitDetail.value.fit.base_eft,
      base_ewar_tags: [...(fitDetail.value.fit.base_ewar_tags || [])],
      base_bonus_tags: [...(fitDetail.value.fit.base_bonus_tags || [])],
      replacements: fitDetail.value.replacements.map((x) => ({ ...x }))
    })
    parsedFit.value = fitDetail.value.parsed
    fitDetailVisible.value = false
    fitDrawer.value = true
  }
  async function saveFit() {
    if (!parsedFit.value) await parseFit()
    saving.value = true
    try {
      await atApi.saveFit(editingFitId.value, fitForm)
      fitDrawer.value = false
      ElMessage.success(t('at.common.saved'))
      await load()
    } finally {
      saving.value = false
    }
  }
  const fitDetailCache = new Map<number, Api.AT.FitDetail>()
  function fitOptions(value: unknown) {
    const row = value as SlotForm
    const fit = fits.value.find((x) => x.id === row.fit_id),
      d = fitDetailCache.get(row.fit_id)
    return {
      ewar: d?.ewar_tags || fit?.base_ewar_tags || [],
      bonus: d?.bonus_tags || fit?.base_bonus_tags || []
    }
  }
  async function selectFit(value: unknown) {
    const row = value as SlotForm
    const d = await atApi.getFit(row.fit_id)
    fitDetailCache.set(row.fit_id, d)
    row.ship_type_id = d.fit.ship_type_id
    row.ewar_tags = [...(d.fit.base_ewar_tags || [])]
    row.bonus_tags = [...(d.fit.base_bonus_tags || [])]
  }
  function addSlot() {
    if (setSlots.value.length < maxSlots.value)
      setSlots.value.push({ key: ++key, fit_id: 0, ship_type_id: 0, ewar_tags: [], bonus_tags: [] })
  }
  async function newSet() {
    editingSetId.value = null
    Object.assign(setForm, {
      name: '',
      format: '10v10',
      rule_version_id: publishedRules.value[0]?.id || 0
    })
    setSlots.value = []
    addSlot()
    await loadRule()
    setDrawer.value = true
  }
  async function editSet(id: number) {
    const d = await atApi.getSet(id)
    editingSetId.value = id
    Object.assign(setForm, {
      name: d.set.name,
      format: d.set.format,
      rule_version_id: d.set.rule_version_id
    })
    setSlots.value = d.slots.map((x) => ({
      key: ++key,
      fit_id: x.slot.fit_id || 0,
      ship_type_id: x.slot.ship_type_id,
      ewar_tags: [...(x.slot.ewar_tags || [])],
      bonus_tags: [...(x.slot.bonus_tags || [])]
    }))
    await Promise.all(
      setSlots.value
        .filter((x) => x.fit_id)
        .map(async (x) => fitDetailCache.set(x.fit_id, await atApi.getFit(x.fit_id)))
    )
    await loadRule()
    setDrawer.value = true
  }
  async function loadRule() {
    if (setForm.rule_version_id)
      ruleDetail.value = await atApi.getScoreRule(setForm.rule_version_id)
  }
  watch(() => setForm.rule_version_id, loadRule)
  async function saveSet() {
    saving.value = true
    try {
      await atApi.saveSet(editingSetId.value, {
        ...setForm,
        status: 'draft',
        slots: setSlots.value.map((x, i) => ({
          slot: {
            slot_order: i + 1,
            fit_id: x.fit_id,
            ship_type_id: x.ship_type_id,
            base_eft: '',
            ewar_tags: x.ewar_tags,
            bonus_tags: x.bonus_tags
          },
          replacements: []
        }))
      })
      setDrawer.value = false
      ElMessage.success(t('at.common.saved'))
      await load()
    } finally {
      saving.value = false
    }
  }
  async function viewSet(id: number) {
    setDetail.value = await atApi.getSet(id)
    setDetailVisible.value = true
  }
  const positionName = (b: Api.AT.SetSlotBundle) =>
    `${b.fit?.ship_name || `TypeID ${b.slot.ship_type_id}`}${[...b.slot.ewar_tags, ...b.slot.bonus_tags].length ? ' · ' + [...b.slot.ewar_tags, ...b.slot.bonus_tags].sort().join('+') : ''}`
  async function copyText(text: string) {
    await navigator.clipboard.writeText(text)
    ElMessage.success(t('at.common.copied'))
  }
  async function copySetEFT() {
    if (!setDetail.value) return
    const out = await atApi.getSetEFT(setDetail.value.set.id)
    await copyText(out.combined)
  }
  async function load() {
    loading.value = true
    try {
      ;[fits.value, sets.value, rules.value] = await Promise.all([
        atApi.listFits(),
        atApi.listSets(),
        atApi.listScoreRules()
      ])
    } finally {
      loading.value = false
    }
  }
  onMounted(load)
</script>

<style scoped>
  .fit-grid,
  .composition-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
    gap: 12px;
  }
  .fit-card,
  .position-card {
    display: flex;
    align-items: center;
    gap: 12px;
    border: 1px solid var(--el-border-color-light);
    background: var(--el-bg-color);
    border-radius: 10px;
    padding: 12px;
    text-align: left;
    color: inherit;
  }
  .fit-card {
    cursor: pointer;
  }
  .fit-card:hover {
    border-color: var(--el-color-primary);
  }
  .fit-card img,
  .position-card img {
    width: 56px;
    height: 56px;
    border-radius: 6px;
  }
  .fit-card strong,
  .fit-card span {
    display: block;
  }
  .fit-card p {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
    margin: 5px 0 0;
  }
  .fit-preview {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 12px;
    background: var(--el-fill-color-light);
    border-radius: 8px;
  }
  .fit-preview img {
    width: 64px;
    height: 64px;
  }
  .rule-row {
    display: grid;
    grid-template-columns: 110px 120px 120px 1fr 1fr 110px auto;
    gap: 8px;
    align-items: center;
    margin-bottom: 8px;
  }
  .slot-groups section {
    margin-bottom: 12px;
  }
  .slot-groups header {
    font-weight: 600;
    padding: 6px 10px;
    background: var(--el-fill-color-light);
  }
  .slot-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 10px;
  }
  .slot-item img {
    width: 28px;
    height: 28px;
  }
  .slot-item span {
    flex: 1;
  }
  .position-card {
    position: relative;
  }
  .position-number {
    font-size: 18px;
    font-weight: 700;
    color: var(--el-color-primary);
  }
  @media (max-width: 1000px) {
    .rule-row {
      grid-template-columns: 1fr 1fr;
    }
  }
</style>
