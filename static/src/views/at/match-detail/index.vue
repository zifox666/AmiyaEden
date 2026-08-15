<template>
  <div class="p-4 md:p-6" v-loading="loading">
    <div v-if="detail" class="space-y-4">
      <ElCard shadow="never">
        <div class="flex-b flex-wrap gap-3"
          ><div
            ><div class="text-sm text-g-500">UUID {{ detail.match.uuid }}</div
            ><h1 class="text-2xl mt-1"
              >{{ detail.match.team_a_name }} vs {{ detail.match.team_b_name }}</h1
            ><div class="mt-2"
              ><ElTag>{{ detail.match.format }}</ElTag
              ><ElTag class="ml-2">{{ detail.match.series }}</ElTag
              ><ElTag class="ml-2" type="primary">{{
                t(`at.status.${detail.match.status}`)
              }}</ElTag></div
            ></div
          >
          <div v-if="canManage" class="flex gap-2 flex-wrap"
            ><ElButton v-if="nextStatus" type="primary" @click="advance">{{
              nextStatusLabel
            }}</ElButton
            ><ElButton
              v-if="currentRound?.status === 'in_progress'"
              type="warning"
              @click="finishCurrent"
              >{{ t('at.matches.finishAndAnalyze') }}</ElButton
            ><ElButton
              v-if="detail.match.status === 'in_progress'"
              type="danger"
              plain
              @click="endMatch"
              >{{ t('at.matches.endMatch') }}</ElButton
            ></div
          >
        </div>
      </ElCard>

      <ElAlert
        v-if="currentRound?.status === 'in_progress'"
        type="info"
        :closable="false"
        :title="t('at.matches.autoRoundHint')"
      />
      <ElAlert
        v-if="detail.match.status === 'in_progress'"
        type="warning"
        :closable="false"
        :title="t('at.matches.opponentRosterWarning')"
      />

      <ElCard shadow="never">
        <template #header
          ><div class="flex-b"
            ><div
              ><strong>{{ t('at.matches.positions') }}</strong
              ><div class="text-sm text-g-500 mt-1">{{
                detail.match.allow_self_assign
                  ? t('at.matches.selfAssignEnabled')
                  : t('at.matches.selfAssignDisabled')
              }}</div></div
            ><span>{{ setDetail?.score.total_points || 0 }} / 200</span></div
          ></template
        >
        <ElTable :data="positions" row-key="slotOrder">
          <ElTableColumn prop="slotOrder" :label="t('at.sets.position')" width="80" />
          <ElTableColumn :label="t('at.sets.shipFit')" min-width="230"
            ><template #default="{ row }"
              ><div class="flex items-center gap-2"
                ><img :src="shipIcon(row.shipTypeId)" class="ship-icon" /><div
                  ><strong>{{ row.shipName }}</strong
                  ><div class="text-xs text-g-500">{{ positionLabel(row) }}</div></div
                ></div
              ></template
            ></ElTableColumn
          >
          <ElTableColumn :label="t('at.sets.ewarTags')" min-width="170"
            ><template #default="{ row }"
              ><ElSelect v-model="row.ewarTags" multiple :disabled="!canManage"
                ><ElOption
                  v-for="tag in row.ewarOptions"
                  :key="tag"
                  :label="tag"
                  :value="tag" /></ElSelect></template
          ></ElTableColumn>
          <ElTableColumn :label="t('at.sets.bonusTags')" min-width="170"
            ><template #default="{ row }"
              ><ElSelect v-model="row.bonusTags" multiple :disabled="!canManage"
                ><ElOption
                  v-for="tag in row.bonusOptions"
                  :key="tag"
                  :label="tag"
                  :value="tag" /></ElSelect></template
          ></ElTableColumn>
          <ElTableColumn :label="t('at.matches.member')" min-width="220"
            ><template #default="{ row }">
              <ElSelect
                v-if="canManage"
                v-model="row.userId"
                filterable
                remote
                clearable
                :remote-method="searchUsers"
                :loading="searchingUsers"
                class="w-full"
                ><ElOption v-for="x in userOptions" :key="x.id" :label="userLabel(x)" :value="x.id"
              /></ElSelect>
              <span v-else>{{ registrationLabel(row.registration) }}</span>
            </template></ElTableColumn
          >
          <ElTableColumn :label="t('at.common.status')" width="110"
            ><template #default="{ row }"
              ><ElTag v-if="row.registration">{{ row.registration.status }}</ElTag
              ><span v-else>-</span></template
            ></ElTableColumn
          >
          <ElTableColumn :label="t('at.common.actions')" width="190"
            ><template #default="{ row }"
              ><ElButton
                v-if="canManage"
                link
                type="primary"
                :disabled="!row.userId"
                @click="assign(row)"
                >{{ t('at.common.save') }}</ElButton
              ><ElButton
                v-else-if="canSelfSelect(row)"
                link
                type="primary"
                @click="selfSelect(row)"
                >{{ t('at.matches.choosePosition') }}</ElButton
              ><ElButton v-if="canCopy(row)" link @click="copyPosition(row)">{{
                t('at.matches.copyMyEFT')
              }}</ElButton></template
            ></ElTableColumn
          >
        </ElTable>
      </ElCard>

      <ElCard shadow="never">
        <template #header
          ><strong>{{ t('at.matches.rounds') }}</strong></template
        >
        <ElTable :data="detail.rounds"
          ><ElTableColumn
            prop="round_no"
            :label="t('at.matches.round')"
            width="100"
          /><ElTableColumn
            prop="attempt"
            :label="t('at.matches.attempt')"
            width="100"
          /><ElTableColumn prop="status" :label="t('at.common.status')" /><ElTableColumn
            prop="result"
            :label="t('at.matches.result')"
          /><ElTableColumn :label="t('at.common.actions')" width="320"
            ><template #default="{ row }"
              ><ElButton
                v-if="canManage && row.status !== 'void' && row.status !== 'confirmed'"
                link
                @click="openRoster(row.id)"
                >{{ t('at.matches.opponentRoster') }}</ElButton
              ><ElButton
                v-if="canManage && row.status === 'br_review'"
                link
                type="success"
                @click="buildBR(row.id)"
                >{{ t('at.matches.buildBR') }}</ElButton
              ><ElDropdown
                v-if="canManage && row.status === 'br_review'"
                @command="confirm(row.id, $event)"
                ><ElButton link type="primary">{{ t('at.matches.confirmResult') }}</ElButton
                ><template #dropdown
                  ><ElDropdownMenu
                    ><ElDropdownItem command="a_win">{{ t('at.matches.teamAWin') }}</ElDropdownItem
                    ><ElDropdownItem command="b_win">{{ t('at.matches.teamBWin') }}</ElDropdownItem
                    ><ElDropdownItem command="draw">{{
                      t('at.matches.draw')
                    }}</ElDropdownItem></ElDropdownMenu
                  ></template
                ></ElDropdown
              ><ElButton
                v-if="row.status === 'br_review' || row.status === 'confirmed'"
                link
                @click="router.push(`/at/br/${row.id}`)"
                >BR</ElButton
              ></template
            ></ElTableColumn
          ></ElTable
        >
        <ElEmpty v-if="!detail.rounds.length" :description="t('at.matches.roundsAutomatic')" />
      </ElCard>
    </div>
    <ElDialog v-model="rosterVisible" :title="t('at.matches.opponentRoster')" width="780px">
      <ElAlert
        type="info"
        :closable="false"
        class="mb-3"
        :title="t('at.matches.opponentRosterHelp')"
      />
      <ElTable :data="opponentRows">
        <ElTableColumn prop="slot_order" :label="t('at.sets.position')" width="80" />
        <ElTableColumn :label="t('at.br.character')" min-width="220"
          ><template #default="{ row }"><ElInput v-model="row.character_name" /></template
        ></ElTableColumn>
        <ElTableColumn :label="t('at.km.ship')" min-width="320"
          ><template #default="{ row }"
            ><SdeSearchSelect
              v-model="row.ship_type_id"
              :initial-options="row.options"
              @select="(item) => (row.options = item ? [item] : [])" /></template
        ></ElTableColumn>
      </ElTable>
      <template #footer
        ><ElButton @click="rosterVisible = false">{{ t('at.common.cancel') }}</ElButton
        ><ElButton type="primary" @click="saveOpponentRoster">{{
          t('at.common.save')
        }}</ElButton></template
      >
    </ElDialog>
  </div>
</template>
<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { atApi } from '@/api/at'
  import { useUserStore } from '@/store/modules/user'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'
  const { t } = useI18n(),
    route = useRoute(),
    router = useRouter(),
    user = useUserStore()
  const loading = ref(false),
    detail = ref<Api.AT.MatchDetail>(),
    setDetail = ref<Api.AT.SetDetail>(),
    userOptions = ref<Api.AT.ContributionUserOption[]>([]),
    searchingUsers = ref(false),
    rosterVisible = ref(false),
    rosterRoundId = ref(0),
    ownRoster = ref<Api.AT.RoundRoster[]>([])
  type OpponentRow = Api.AT.RoundRoster & { options: Api.Sde.FuzzySearchItem[] }
  const opponentRows = ref<OpponentRow[]>([])
  const canManage = computed(() =>
    (user.info.roles || []).some((r) => ['fc', 'admin', 'super_admin'].includes(r))
  )
  type Position = {
    slotOrder: number
    shipTypeId: number
    shipName: string
    fitName: string
    ewarTags: string[]
    bonusTags: string[]
    ewarOptions: string[]
    bonusOptions: string[]
    userId: number | null
    registration: Api.AT.MatchDetail['registrations'][number] | null
  }
  const positions = ref<Position[]>([])
  const currentRound = computed(() =>
    [...(detail.value?.rounds || [])]
      .reverse()
      .find((x) => x.status === 'in_progress' || x.status === 'br_review')
  )
  const nextStatus = computed(
    () =>
      (
        ({ draft: 'recruiting', recruiting: 'ready', ready: 'in_progress' }) as Record<
          string,
          string
        >
      )[detail.value?.match.status || '']
  )
  const nextStatusLabel = computed(() =>
    nextStatus.value === 'in_progress' ? t('at.matches.startMatch') : t('at.matches.advance')
  )
  const shipIcon = (id: number) => `https://images.evetech.net/types/${id}/icon?size=64`
  const userLabel = (x: Api.AT.ContributionUserOption) =>
    x.tournament_character_name || x.nickname || `#${x.id}`
  const registrationLabel = (r: Position['registration']) =>
    r ? r.tournament_character_name || r.user_nickname || `#${r.user_id}` : '-'
  const positionLabel = (value: unknown) => {
    const r = value as Position
    return `${r.shipName}${[...r.ewarTags, ...r.bonusTags].length ? ' · ' + [...r.ewarTags, ...r.bonusTags].sort().join('+') : ''}`
  }
  async function load() {
    loading.value = true
    try {
      detail.value = await atApi.getMatchByUUID(String(route.params.uuid))
      setDetail.value = await atApi.getSet(detail.value.match.set_id)
      if (canManage.value && detail.value.registrations.length)
        userOptions.value = await atApi.searchMatchUsers('')
      positions.value = await Promise.all(
        setDetail.value.slots.map(async (b) => {
          const fit = b.slot.fit_id ? await atApi.getFit(b.slot.fit_id) : undefined
          const registration =
            detail.value!.registrations.find(
              (x) => x.slot_order === b.slot.slot_order && x.status !== 'rejected'
            ) || null
          return {
            slotOrder: b.slot.slot_order,
            shipTypeId: b.slot.ship_type_id,
            shipName: b.fit?.ship_name || fit?.fit.ship_name || `TypeID ${b.slot.ship_type_id}`,
            fitName: b.fit?.name || '',
            ewarTags: [...(registration?.ewar_tags || b.slot.ewar_tags || [])],
            bonusTags: [...(registration?.bonus_tags || b.slot.bonus_tags || [])],
            ewarOptions: fit?.ewar_tags || b.slot.ewar_tags || [],
            bonusOptions: fit?.bonus_tags || b.slot.bonus_tags || [],
            userId: registration?.user_id || null,
            registration
          }
        })
      )
    } finally {
      loading.value = false
    }
  }
  async function searchUsers(q: string) {
    searchingUsers.value = true
    try {
      userOptions.value = await atApi.searchMatchUsers(q)
    } finally {
      searchingUsers.value = false
    }
  }
  async function assign(value: unknown) {
    const row = value as Position
    if (!detail.value || !row.userId) return
    await atApi.assignMatchPosition(detail.value.match.id, {
      user_id: row.userId,
      slot_order: row.slotOrder,
      ewar_tags: row.ewarTags,
      bonus_tags: row.bonusTags
    })
    ElMessage.success(t('at.common.saved'))
    await load()
  }
  function canSelfSelect(value: unknown) {
    const row = value as Position
    return (
      !!detail.value?.match.allow_self_assign &&
      detail.value.match.status === 'recruiting' &&
      !row.registration
    )
  }
  async function selfSelect(value: unknown) {
    const row = value as Position
    if (!detail.value) return
    await atApi.selectMatchPosition(detail.value.match.id, {
      slot_order: row.slotOrder,
      ewar_tags: row.ewarTags,
      bonus_tags: row.bonusTags
    })
    ElMessage.success(t('at.matches.registered'))
    await load()
  }
  function canCopy(value: unknown) {
    const row = value as Position
    return !!row.registration && (row.registration.user_id === user.info.userId || canManage.value)
  }
  async function copyPosition(value: unknown) {
    const row = value as Position
    if (!detail.value) return
    const out = await atApi.getMatchPositionEFT(detail.value.match.id, row.slotOrder)
    await navigator.clipboard.writeText(out.eft)
    ElMessage.success(t('at.common.copied'))
  }
  async function advance() {
    if (!detail.value || !nextStatus.value) return
    await atApi.transitionMatch(detail.value.match.id, { status: nextStatus.value, reason: '' })
    await load()
  }
  async function finishCurrent() {
    if (!currentRound.value) return
    await atApi.finishRound(currentRound.value.id)
    await load()
  }
  async function endMatch() {
    if (!detail.value) return
    const { value } = await ElMessageBox.prompt(t('at.matches.endReason'), t('at.matches.endMatch'))
    await atApi.transitionMatch(detail.value.match.id, { status: 'early_ended', reason: value })
    await load()
  }
  async function buildBR(id: number) {
    await atApi.buildBR(id)
    ElMessage.success(t('at.matches.brBuilt'))
  }
  async function openRoster(roundId: number) {
    rosterRoundId.value = roundId
    const roster = await atApi.getRoundRoster(roundId)
    ownRoster.value = roster.filter((entry) => entry.side === 'A')
    const opponents = roster.filter((entry) => entry.side === 'B')
    const max = detail.value?.match.format === '10v10' ? 10 : 7
    opponentRows.value = Array.from({ length: max }, (_, index) => {
      const existing = opponents.find((entry) => entry.slot_order === index + 1)
      return {
        side: 'B',
        slot_order: index + 1,
        user_id: null,
        character_name: existing?.character_name || '',
        ship_type_id: existing?.ship_type_id || 0,
        options: existing?.ship_type_id
          ? [
              {
                id: existing.ship_type_id,
                name: `TypeID ${existing.ship_type_id}`,
                group_id: 0,
                group_name: '',
                category: 'type'
              }
            ]
          : []
      }
    })
    rosterVisible.value = true
  }
  async function saveOpponentRoster() {
    const opponents = opponentRows.value
      .filter((entry) => entry.character_name.trim() || entry.ship_type_id)
      .map((entry) => ({
        side: entry.side,
        slot_order: entry.slot_order,
        user_id: entry.user_id,
        character_name: entry.character_name,
        ship_type_id: entry.ship_type_id
      }))
    if (opponents.some((entry) => !entry.character_name.trim() || !entry.ship_type_id)) {
      ElMessage.warning(t('at.matches.opponentRosterIncompleteRow'))
      return
    }
    const own = ownRoster.value.filter((entry) => entry.character_name.trim() && entry.ship_type_id)
    await atApi.setRoundRoster(rosterRoundId.value, [...own, ...opponents])
    rosterVisible.value = false
    ElMessage.success(t('at.common.saved'))
  }
  async function confirm(id: number, result: unknown) {
    await atApi.confirmRound(id, result as 'a_win' | 'b_win' | 'draw')
    await load()
  }
  onMounted(load)
</script>
<style scoped>
  .space-y-4 > * + * {
    margin-top: 16px;
  }
  .ship-icon {
    width: 44px;
    height: 44px;
    border-radius: 6px;
  }
</style>
