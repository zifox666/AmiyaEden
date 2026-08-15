<template>
  <div class="art-card p-5">
    <div class="flex-b mb-4">
      <div>
        <h2 class="text-xl">{{ t('at.contributions.title') }}</h2>
        <p class="text-sm text-g-500 mt-1">{{ t('at.contributions.description') }}</p>
      </div>
      <div class="flex gap-2">
        <ElSelect
          v-model="seasonId"
          clearable
          :placeholder="t('at.contributions.career')"
          @change="loadBoard"
        >
          <ElOption v-for="item in seasons" :key="item.id" :label="item.name" :value="item.id" />
        </ElSelect>
        <ElButton v-if="isAdmin" type="primary" @click="adminVisible = true">{{
          t('at.contributions.manage')
        }}</ElButton>
      </div>
    </div>

    <ElTabs v-model="activeTab">
      <ElTabPane name="leaderboard" :label="t('at.contributions.leaderboard')">
        <ArtTable
          :data="pagedRows"
          :loading="loading"
          :columns="leaderboardColumns"
          :pagination="boardPagination"
          @pagination:size-change="changeBoardSize"
          @pagination:current-change="boardPagination.current = $event"
        />
      </ElTabPane>
      <ElTabPane name="mine" :label="t('at.contributions.myLedger')">
        <ArtTable
          :data="pagedLedger"
          :loading="ledgerLoading"
          :columns="ledgerColumns"
          :pagination="ledgerPagination"
          @pagination:size-change="changeLedgerSize"
          @pagination:current-change="ledgerPagination.current = $event"
        >
          <template #source_type="{ row }">
            <ElTag
              v-if="row.source_type"
              :type="
                row.source_type === 'reversal'
                  ? 'danger'
                  : row.source_type === 'manual'
                    ? 'warning'
                    : 'success'
              "
            >
              {{ t(`at.contributions.sources.${row.source_type}`) }} </ElTag
            ><span v-else>-</span>
          </template>
          <template #amount="{ row }">
            <span
              v-if="typeof row.amount === 'number'"
              :class="row.amount < 0 ? 'text-danger' : 'text-success'"
              >{{ row.amount > 0 ? '+' : '' }}{{ row.amount }}</span
            >
            <span v-else>-</span>
          </template>
        </ArtTable>
      </ElTabPane>
    </ElTabs>

    <ElDialog v-model="adminVisible" :title="t('at.contributions.manage')" width="900px">
      <ElTabs>
        <ElTabPane :label="t('at.contributions.seasons')">
          <div class="manage-toolbar">
            <strong>{{
              seasonForm.id ? t('at.contributions.editSeason') : t('at.contributions.newSeason')
            }}</strong>
            <ElButton v-if="seasonForm.id" link type="primary" @click="resetSeasonForm">{{
              t('at.contributions.newSeason')
            }}</ElButton>
          </div>
          <ElForm label-position="top">
            <ElRow :gutter="16">
              <ElCol :span="9"
                ><ElFormItem :label="t('at.common.name')"
                  ><ElInput v-model="seasonForm.name" /></ElFormItem
              ></ElCol>
              <ElCol :span="11"
                ><ElFormItem :label="t('at.contributions.dateRange')">
                  <ElDatePicker
                    v-model="seasonRange"
                    type="datetimerange"
                    value-format="YYYY-MM-DDTHH:mm:ssZ"
                    class="w-full"
                  /> </ElFormItem
              ></ElCol>
              <ElCol :span="4"
                ><ElFormItem :label="t('at.common.status')">
                  <ElCheckbox v-model="seasonForm.is_current">{{
                    t('at.contributions.current')
                  }}</ElCheckbox>
                </ElFormItem></ElCol
              >
            </ElRow>
            <ElButton type="primary" @click="saveSeason">{{ t('at.common.save') }}</ElButton>
          </ElForm>
          <ElTable :data="seasons" class="mt-5" max-height="340">
            <ElTableColumn prop="name" :label="t('at.common.name')" min-width="160" />
            <ElTableColumn :label="t('at.contributions.dateRange')" min-width="260"
              ><template #default="{ row }">{{ formatRange(row) }}</template></ElTableColumn
            >
            <ElTableColumn :label="t('at.contributions.current')" width="100"
              ><template #default="{ row }"
                ><ElTag v-if="row.is_current" type="success">{{ t('at.common.yes') }}</ElTag
                ><span v-else>-</span></template
              ></ElTableColumn
            >
            <ElTableColumn :label="t('at.common.actions')" width="190" fixed="right"
              ><template #default="{ row }">
                <ElButton link type="primary" @click="editSeason(row)">{{
                  t('at.common.edit')
                }}</ElButton>
                <ElButton link type="primary" @click="openCleanup(row)">{{
                  t('at.contributions.cleanupShips')
                }}</ElButton>
              </template></ElTableColumn
            >
          </ElTable>
        </ElTabPane>

        <ElTabPane :label="t('at.contributions.matchTypes')">
          <div class="manage-toolbar">
            <strong>{{
              matchTypeForm.id
                ? t('at.contributions.editMatchType')
                : t('at.contributions.newMatchType')
            }}</strong>
            <ElButton v-if="matchTypeForm.id" link type="primary" @click="resetMatchTypeForm">{{
              t('at.contributions.newMatchType')
            }}</ElButton>
          </div>
          <ElForm label-position="top">
            <ElRow :gutter="16">
              <ElCol :span="10"
                ><ElFormItem :label="t('at.common.name')"
                  ><ElInput v-model="matchTypeForm.name" /></ElFormItem
              ></ElCol>
              <ElCol :span="8"
                ><ElFormItem :label="t('at.contributions.value')"
                  ><ElInputNumber
                    v-model="matchTypeForm.contribution_value"
                    :min="0"
                    class="!w-full" /></ElFormItem
              ></ElCol>
              <ElCol :span="6"
                ><ElFormItem :label="t('at.common.status')"
                  ><ElCheckbox v-model="matchTypeForm.active">{{
                    t('at.common.active')
                  }}</ElCheckbox></ElFormItem
                ></ElCol
              >
            </ElRow>
            <ElButton type="primary" @click="saveMatchType">{{ t('at.common.save') }}</ElButton>
          </ElForm>
          <ElTable :data="matchTypes" class="mt-5" max-height="340">
            <ElTableColumn prop="name" :label="t('at.common.name')" />
            <ElTableColumn prop="contribution_value" :label="t('at.contributions.value')" />
            <ElTableColumn :label="t('at.common.status')"
              ><template #default="{ row }"
                ><ElTag :type="row.active ? 'success' : 'info'">{{
                  row.active ? t('at.common.active') : t('at.common.inactive')
                }}</ElTag></template
              ></ElTableColumn
            >
            <ElTableColumn :label="t('at.common.actions')" width="100"
              ><template #default="{ row }"
                ><ElButton link type="primary" @click="editMatchType(row)">{{
                  t('at.common.edit')
                }}</ElButton></template
              ></ElTableColumn
            >
          </ElTable>
        </ElTabPane>

        <ElTabPane :label="t('at.contributions.manual')">
          <ElAlert
            class="mb-4"
            type="info"
            :closable="false"
            :title="t('at.contributions.manualHint')"
          />
          <ElForm label-position="top">
            <ElFormItem :label="t('at.matches.season')"
              ><ElSelect v-model="manualForm.season_id" class="w-full"
                ><ElOption
                  v-for="item in seasons"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id" /></ElSelect
            ></ElFormItem>
            <ElFormItem :label="t('at.contributions.user')">
              <ElSelect
                v-model="manualForm.user_id"
                filterable
                remote
                clearable
                :remote-method="searchUsers"
                :loading="userSearchLoading"
                :placeholder="t('at.contributions.userSearchPlaceholder')"
                class="w-full"
              >
                <ElOption
                  v-for="item in userOptions"
                  :key="item.id"
                  :label="userOptionLabel(item)"
                  :value="item.id"
                />
              </ElSelect>
            </ElFormItem>
            <ElFormItem :label="t('at.contributions.amount')"
              ><ElInputNumber v-model="manualForm.amount" class="!w-full"
            /></ElFormItem>
            <ElFormItem :label="t('at.common.reason')"
              ><ElInput v-model="manualForm.reason" type="textarea"
            /></ElFormItem>
            <ElButton type="primary" @click="saveManual">{{ t('at.common.save') }}</ElButton>
          </ElForm>
        </ElTabPane>
      </ElTabs>
    </ElDialog>

    <ElDialog
      v-model="cleanupVisible"
      :title="t('at.contributions.cleanupForSeason', { name: cleanupSeason?.name ?? '' })"
      width="620px"
      append-to-body
    >
      <ElAlert
        type="info"
        :closable="false"
        :title="t('at.contributions.cleanupHint')"
        class="mb-4"
      />
      <SdeSearchSelect
        v-model="cleanupCandidate"
        :category-ids="[6]"
        :placeholder="t('at.contributions.cleanupSearchPlaceholder')"
        @select="addCleanupShip"
      />
      <div v-if="cleanupShips.length" class="cleanup-list mt-4">
        <div v-for="ship in cleanupShips" :key="ship.type_id" class="cleanup-item">
          <img :src="`https://images.evetech.net/types/${ship.type_id}/icon?size=32`" />
          <span class="flex-1">{{ ship.type_name }}</span
          ><span class="text-xs text-g-500">{{ ship.type_id }}</span>
          <ElButton link type="danger" @click="removeCleanupShip(ship.type_id)">{{
            t('at.common.delete')
          }}</ElButton>
        </div>
      </div>
      <ElEmpty v-else :description="t('at.contributions.noCleanupShips')" :image-size="70" />
      <template #footer
        ><ElButton @click="cleanupVisible = false">{{ t('at.common.cancel') }}</ElButton
        ><ElButton type="primary" :loading="cleanupSaving" @click="saveCleanup">{{
          t('at.common.save')
        }}</ElButton></template
      >
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import type { ColumnOption } from '@/types'
  import { atApi } from '@/api/at'
  import { useUserStore } from '@/store/modules/user'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'

  const { t } = useI18n()
  const userStore = useUserStore()
  const isAdmin = computed(() =>
    (userStore.info.roles ?? []).some((role) => ['admin', 'super_admin'].includes(role))
  )
  const seasons = ref<Api.AT.Season[]>([])
  const matchTypes = ref<Api.AT.MatchType[]>([])
  const rows = ref<Api.AT.ContributionRow[]>([])
  const ledger = ref<Api.AT.ContributionLedger[]>([])
  const activeTab = ref('leaderboard')
  const seasonId = ref<number>()
  const loading = ref(false)
  const ledgerLoading = ref(false)
  const boardPagination = reactive({ current: 1, size: 10, total: 0 })
  const ledgerPagination = reactive({ current: 1, size: 10, total: 0 })
  const adminVisible = ref(false)
  const seasonRange = ref<[string, string]>()
  const seasonForm = reactive<Partial<Api.AT.Season>>({ name: '', is_current: false })
  const matchTypeForm = reactive<Partial<Api.AT.MatchType>>({
    name: '',
    contribution_value: 1,
    active: true
  })
  const manualForm = reactive({
    season_id: 0,
    user_id: null as number | null,
    amount: 1,
    reason: ''
  })
  const userOptions = ref<Api.AT.ContributionUserOption[]>([])
  const userSearchLoading = ref(false)
  const cleanupVisible = ref(false)
  const cleanupSaving = ref(false)
  const cleanupSeason = ref<Api.AT.Season>()
  const cleanupShips = ref<Api.AT.CleanupShip[]>([])
  const cleanupCandidate = ref<number | null>(null)

  const pagedRows = computed(() =>
    rows.value.slice(
      (boardPagination.current - 1) * boardPagination.size,
      boardPagination.current * boardPagination.size
    )
  )
  const pagedLedger = computed(() =>
    ledger.value.slice(
      (ledgerPagination.current - 1) * ledgerPagination.size,
      ledgerPagination.current * ledgerPagination.size
    )
  )
  watch(
    () => rows.value.length,
    (total) => (boardPagination.total = total),
    { immediate: true }
  )
  watch(
    () => ledger.value.length,
    (total) => (ledgerPagination.total = total),
    { immediate: true }
  )

  const leaderboardColumns = computed<ColumnOption[]>(() => [
    { type: 'globalIndex', width: 90, label: t('at.contributions.rank') },
    { prop: 'nickname', label: t('at.contributions.nickname'), minWidth: 180 },
    {
      prop: 'tournament_character_name',
      label: t('at.contributions.tournamentCharacter'),
      minWidth: 220
    },
    { prop: 'total', label: t('at.contributions.total'), width: 140, sortable: true }
  ])
  const ledgerColumns = computed<ColumnOption[]>(() => [
    { prop: 'created_at', label: t('at.contributions.time'), minWidth: 190 },
    { prop: 'source_type', label: t('at.contributions.source'), width: 130, useSlot: true },
    { prop: 'match_id', label: t('at.contributions.matchId'), width: 120 },
    { prop: 'reason', label: t('at.common.reason'), minWidth: 240 },
    { prop: 'amount', label: t('at.contributions.amountShort'), width: 120, useSlot: true }
  ])

  async function loadReference() {
    ;[seasons.value, matchTypes.value] = await Promise.all([
      atApi.listSeasons(),
      atApi.listMatchTypes()
    ])
  }
  async function loadBoard() {
    loading.value = true
    try {
      rows.value = await atApi.leaderboard(seasonId.value)
    } finally {
      loading.value = false
    }
  }
  async function loadLedger() {
    ledgerLoading.value = true
    try {
      ledger.value = await atApi.myContributionLedger(seasonId.value)
    } finally {
      ledgerLoading.value = false
    }
  }
  function changeBoardSize(size: number) {
    boardPagination.size = size
    boardPagination.current = 1
  }
  function changeLedgerSize(size: number) {
    ledgerPagination.size = size
    ledgerPagination.current = 1
  }
  function formatRange(value: unknown) {
    const row = value as Api.AT.Season
    return `${new Date(row.starts_at).toLocaleString()} — ${new Date(row.ends_at).toLocaleString()}`
  }
  function resetSeasonForm() {
    Object.assign(seasonForm, { id: undefined, name: '', is_current: false })
    seasonRange.value = undefined
  }
  function editSeason(value: unknown) {
    const row = value as Api.AT.Season
    Object.assign(seasonForm, row)
    seasonRange.value = [row.starts_at, row.ends_at]
  }
  async function saveSeason() {
    if (!seasonRange.value) return ElMessage.warning(t('at.contributions.dateRequired'))
    await atApi.saveSeason({
      id: seasonForm.id,
      name: seasonForm.name?.trim() ?? '',
      starts_at: seasonRange.value[0],
      ends_at: seasonRange.value[1],
      is_current: Boolean(seasonForm.is_current)
    })
    ElMessage.success(t('at.common.saved'))
    resetSeasonForm()
    await loadReference()
  }
  function resetMatchTypeForm() {
    Object.assign(matchTypeForm, { id: undefined, name: '', contribution_value: 1, active: true })
  }
  function editMatchType(value: unknown) {
    const row = value as Api.AT.MatchType
    Object.assign(matchTypeForm, row)
  }
  async function saveMatchType() {
    await atApi.saveMatchType({
      id: matchTypeForm.id,
      name: matchTypeForm.name?.trim() ?? '',
      contribution_value: matchTypeForm.contribution_value ?? 0,
      active: Boolean(matchTypeForm.active)
    })
    ElMessage.success(t('at.common.saved'))
    resetMatchTypeForm()
    await loadReference()
  }
  async function openCleanup(value: unknown) {
    const row = value as Api.AT.Season
    cleanupSeason.value = row
    cleanupCandidate.value = null
    cleanupShips.value = await atApi.listCleanupShips(row.id, userStore.gameLanguage || 'en')
    cleanupVisible.value = true
  }
  function addCleanupShip(item: Api.Sde.FuzzySearchItem | null) {
    if (!item || item.category !== 'type') return
    if (!cleanupShips.value.some((ship) => ship.type_id === item.id))
      cleanupShips.value.push({ type_id: item.id, type_name: item.name })
    nextTick(() => (cleanupCandidate.value = null))
  }
  function removeCleanupShip(typeId: number) {
    cleanupShips.value = cleanupShips.value.filter((ship) => ship.type_id !== typeId)
  }
  async function saveCleanup() {
    if (!cleanupSeason.value) return
    cleanupSaving.value = true
    try {
      await atApi.setCleanupShips(
        cleanupSeason.value.id,
        cleanupShips.value.map((ship) => ship.type_id)
      )
      ElMessage.success(t('at.common.saved'))
      cleanupVisible.value = false
    } finally {
      cleanupSaving.value = false
    }
  }
  async function searchUsers(keyword: string) {
    userSearchLoading.value = true
    try {
      userOptions.value = await atApi.searchContributionUsers(keyword.trim())
    } finally {
      userSearchLoading.value = false
    }
  }
  function userOptionLabel(item: Api.AT.ContributionUserOption) {
    const identity =
      item.nickname || item.tournament_character_name || t('at.contributions.unnamedUser')
    return `${identity}${item.tournament_character_name && item.tournament_character_name !== identity ? ` · ${item.tournament_character_name}` : ''} (#${item.id})`
  }
  async function saveManual() {
    if (!manualForm.user_id) return ElMessage.warning(t('at.contributions.userRequired'))
    await atApi.manualContribution({ ...manualForm, user_id: manualForm.user_id })
    ElMessage.success(t('at.common.saved'))
    manualForm.reason = ''
    await loadBoard()
  }

  onMounted(async () => {
    await loadReference()
    seasonId.value = seasons.value.find((season) => season.is_current)?.id
    manualForm.season_id = seasonId.value ?? 0
    await Promise.all([loadBoard(), loadLedger()])
  })
  watch(seasonId, () => {
    boardPagination.current = 1
    ledgerPagination.current = 1
    loadLedger()
  })
</script>

<style scoped>
  .manage-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }
  .cleanup-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .cleanup-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
  }
  .cleanup-item img {
    width: 32px;
    height: 32px;
    border-radius: 5px;
  }
</style>
