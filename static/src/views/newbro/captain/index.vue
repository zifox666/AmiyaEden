<template>
  <div class="newbro-captain-page art-full-height">
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <div class="text-lg font-semibold">{{ t('newbro.captain.title') }}</div>
          <div class="text-sm text-gray-500 mt-1">{{ t('newbro.captain.subtitle') }}</div>
        </div>
        <ElButton class="min-w-[120px]" :disabled="isRefreshing" @click="reloadActiveTab">{{
          $t('common.refresh')
        }}</ElButton>
      </div>
    </ElCard>

    <ElTabs v-model="activeTab">
      <ElTabPane :label="t('newbro.captain.playersTitle')" name="players">
        <ElCard shadow="never" class="mb-4">
          <template #header>
            <div class="flex items-center justify-between gap-4 flex-wrap">
              <span>{{ t('newbro.captain.playersTitle') }}</span>
              <ElRadioGroup
                v-model="playerStatus"
                @change="
                  () => {
                    playerPage.current = 1
                    loadPlayers()
                  }
                "
              >
                <ElRadioButton label="all">{{ t('newbro.captain.playersAll') }}</ElRadioButton>
                <ElRadioButton label="active">{{
                  t('newbro.captain.playersActive')
                }}</ElRadioButton>
                <ElRadioButton label="historical">{{
                  t('newbro.captain.playersHistorical')
                }}</ElRadioButton>
              </ElRadioGroup>
            </div>
          </template>

          <ElTable :data="players" v-loading="loadingPlayers" stripe border>
            <ElTableColumn
              prop="player_character_name"
              :label="t('newbro.common.mainCharacter')"
              min-width="180"
            />
            <ElTableColumn
              prop="player_nickname"
              :label="t('newbro.common.nickname')"
              min-width="160"
            >
              <template #default="{ row }">
                <span :class="row.player_nickname ? '' : 'text-gray-400'">
                  {{ row.player_nickname || '-' }}
                </span>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="started_at" :label="t('newbro.common.startedAt')" width="180">
              <template #default="{ row }">{{ formatDateTime(row.started_at) }}</template>
            </ElTableColumn>
            <ElTableColumn prop="ended_at" :label="t('newbro.common.endedAt')" width="180">
              <template #default="{ row }">{{
                row.ended_at ? formatDateTime(row.ended_at) : '-'
              }}</template>
            </ElTableColumn>
            <ElTableColumn
              prop="attributed_bounty_total"
              :label="t('newbro.captain.totalBounty')"
              width="160"
            >
              <template #default="{ row }">{{ formatIsk(row.attributed_bounty_total) }}</template>
            </ElTableColumn>
            <ElTableColumn :label="$t('common.operation')" width="180" fixed="right">
              <template #default="{ row }">
                <ElButton
                  v-if="!row.ended_at"
                  type="danger"
                  size="small"
                  class="min-w-[150px]"
                  :disabled="endingPlayerId === row.player_user_id"
                  @click="endPlayerAffiliation(row.player_user_id)"
                >
                  {{ t('newbro.captain.endAffiliationButton') }}
                </ElButton>
                <span v-else class="text-gray-400">-</span>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="flex justify-end mt-4">
            <ElPagination
              background
              layout="prev, pager, next"
              :current-page="playerPage.current"
              :page-size="playerPage.size"
              :total="playerPage.total"
              @current-change="
                (page: number) => {
                  playerPage.current = page
                  loadPlayers()
                }
              "
            />
          </div>
        </ElCard>
      </ElTabPane>

      <ElTabPane :label="t('newbro.captain.enrollTitle')" name="enroll">
        <ElCard shadow="never" class="mb-4">
          <template #header>
            <div class="flex items-center justify-between gap-4 flex-wrap">
              <div>
                <div class="font-medium">{{ t('newbro.captain.enrollTitle') }}</div>
                <div class="text-sm text-gray-500 mt-1">
                  {{ t('newbro.captain.enrollSubtitle') }}
                </div>
              </div>
              <div class="flex items-center gap-3 flex-wrap">
                <ElInput
                  v-model="eligibleKeyword"
                  clearable
                  style="width: 320px"
                  :placeholder="t('newbro.captain.enrollSearchPlaceholder')"
                  @keyup.enter="handleEligibleSearch"
                />
                <ElButton class="min-w-[110px]" @click="handleEligibleSearch">{{
                  $t('common.search')
                }}</ElButton>
              </div>
            </div>
          </template>

          <div class="text-sm text-gray-500 mb-4">
            {{ t('newbro.captain.enrollSearchHint') }}
          </div>

          <ElEmpty
            v-if="!eligiblePlayers.length && !loadingEligiblePlayers"
            :description="t('newbro.captain.enrollEmpty')"
            :image-size="72"
          />

          <ElTable v-else :data="eligiblePlayers" v-loading="loadingEligiblePlayers" stripe border>
            <ElTableColumn
              prop="player_character_name"
              :label="t('newbro.common.mainCharacter')"
              min-width="220"
            >
              <template #default="{ row }">
                <div class="flex items-center gap-3">
                  <ElAvatar
                    :src="buildEveCharacterPortraitUrl(row.player_character_id, 40)"
                    :size="40"
                  />
                  <div>
                    <div class="font-medium">{{ row.player_character_name }}</div>
                    <div class="text-sm text-gray-500">
                      {{ t('newbro.common.nickname') }}: {{ row.player_nickname || '-' }}
                    </div>
                  </div>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn :label="t('newbro.captain.enrollCurrentCaptain')" min-width="260">
              <template #default="{ row }">
                <div v-if="row.current_affiliation">
                  <div class="font-medium">
                    {{ row.current_affiliation.captain_character_name }}
                  </div>
                  <div class="text-sm text-gray-500">
                    {{ t('newbro.common.nickname') }}:
                    {{ row.current_affiliation.captain_nickname || '-' }}
                  </div>
                  <div class="text-xs text-gray-400">
                    {{ t('newbro.captain.enrollSwitchingFrom') }}:
                    {{ formatDateTime(row.current_affiliation.started_at) }}
                  </div>
                </div>
                <span v-else class="text-gray-400">{{ t('newbro.captain.enrollNoCaptain') }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn :label="$t('common.operation')" width="140" fixed="right">
              <template #default="{ row }">
                <ElButton
                  type="primary"
                  class="min-w-[120px]"
                  :disabled="submittingPlayerId === row.player_user_id"
                  @click="enrollPlayer(row.player_user_id)"
                >
                  {{ t('newbro.captain.enrollButton') }}
                </ElButton>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="flex justify-end mt-4">
            <ElPagination
              background
              layout="prev, pager, next"
              :current-page="eligiblePage.current"
              :page-size="eligiblePage.size"
              :total="eligiblePage.total"
              @current-change="
                (page: number) => {
                  eligiblePage.current = page
                  loadEligiblePlayers()
                }
              "
            />
          </div>
        </ElCard>
      </ElTabPane>

      <ElTabPane :label="t('newbro.captain.attributionsTitle')" name="attributions">
        <div v-if="overview" class="grid grid-cols-2 xl:grid-cols-4 gap-4 mb-4">
          <ElCard shadow="never">
            <div class="text-sm text-gray-500">{{ t('newbro.captain.activePlayers') }}</div>
            <div class="text-2xl font-semibold mt-2">{{ overview.active_player_count }}</div>
          </ElCard>
          <ElCard shadow="never">
            <div class="text-sm text-gray-500">{{ t('newbro.captain.historicalPlayers') }}</div>
            <div class="text-2xl font-semibold mt-2">{{ overview.historical_player_count }}</div>
          </ElCard>
          <ElCard shadow="never">
            <div class="text-sm text-gray-500">{{ t('newbro.captain.totalBounty') }}</div>
            <div class="text-2xl font-semibold mt-2">{{
              formatIsk(overview.attributed_bounty_total)
            }}</div>
          </ElCard>
          <ElCard shadow="never">
            <div class="text-sm text-gray-500">{{ t('newbro.captain.recordCount') }}</div>
            <div class="text-2xl font-semibold mt-2">{{ overview.attribution_record_count }}</div>
          </ElCard>
        </div>

        <ElCard shadow="never">
          <template #header>
            <span>{{ t('newbro.captain.attributionsTitle') }}</span>
          </template>

          <div class="grid grid-cols-2 xl:grid-cols-4 gap-4 mb-4">
            <ElCard shadow="never">
              <div class="text-xs text-gray-500">{{ t('newbro.captain.totalBounty') }}</div>
              <div class="text-lg font-semibold mt-2">{{
                formatIsk(attributionSummary.attributed_bounty_total)
              }}</div>
            </ElCard>
            <ElCard shadow="never">
              <div class="text-xs text-gray-500">{{ t('newbro.captain.recordCount') }}</div>
              <div class="text-lg font-semibold mt-2">{{ attributionSummary.record_count }}</div>
            </ElCard>
            <ElCard shadow="never">
              <div class="text-xs text-gray-500">{{ t('newbro.captain.totalRewardedValue') }}</div>
              <div class="text-lg font-semibold mt-2">{{
                formatCredit(rewardSummary.total_credited_value)
              }}</div>
            </ElCard>
            <ElCard shadow="never">
              <div class="text-xs text-gray-500">{{
                t('newbro.captain.rewardSettlementCount')
              }}</div>
              <div class="text-lg font-semibold mt-2">{{ rewardSummary.settlement_count }}</div>
              <div class="text-xs text-gray-400 mt-2">
                {{ t('newbro.captain.lastRewardProcessedAt') }}:
                {{ formatDateTime(rewardSummary.last_processed_at) }}
              </div>
            </ElCard>
          </div>

          <ElTable :data="attributions" v-loading="loadingAttributions" stripe border>
            <ElTableColumn
              prop="player_character_name"
              :label="t('newbro.common.player')"
              min-width="160"
            />
            <ElTableColumn
              prop="captain_character_name"
              :label="t('newbro.common.captain')"
              min-width="160"
            />
            <ElTableColumn prop="ref_type" :label="t('newbro.common.refType')" width="160" />
            <ElTableColumn prop="system_id" :label="t('newbro.common.systemId')" width="120" />
            <ElTableColumn prop="amount" :label="t('newbro.common.amount')" width="160">
              <template #default="{ row }">{{ formatIsk(row.amount) }}</template>
            </ElTableColumn>
            <ElTableColumn prop="journal_at" :label="t('newbro.common.journalAt')" width="180">
              <template #default="{ row }">{{ formatDateTime(row.journal_at) }}</template>
            </ElTableColumn>
            <ElTableColumn prop="processed_at" :label="t('newbro.common.processedAt')" width="180">
              <template #default="{ row }">{{ formatDateTime(row.processed_at) }}</template>
            </ElTableColumn>
          </ElTable>

          <div class="flex justify-end mt-4">
            <ElPagination
              background
              layout="prev, pager, next"
              :current-page="attributionPage.current"
              :page-size="attributionPage.size"
              :total="attributionPage.total"
              @current-change="
                (page: number) => {
                  attributionPage.current = page
                  loadAttributions()
                }
              "
            />
          </div>
        </ElCard>

        <ElCard shadow="never" class="mt-4">
          <template #header>
            <span>{{ t('newbro.captain.rewardHistoryTitle') }}</span>
          </template>

          <ElTable :data="rewardSettlements" v-loading="loadingRewards" stripe border>
            <ElTableColumn prop="processed_at" :label="t('newbro.common.processedAt')" width="180">
              <template #default="{ row }">{{ formatDateTime(row.processed_at) }}</template>
            </ElTableColumn>
            <ElTableColumn
              prop="attribution_count"
              :label="t('newbro.captain.rewardAttributionCount')"
              width="160"
            />
            <ElTableColumn
              prop="attributed_isk_total"
              :label="t('newbro.captain.rewardAttributedTotal')"
              width="180"
            >
              <template #default="{ row }">{{ formatIsk(row.attributed_isk_total) }}</template>
            </ElTableColumn>
            <ElTableColumn
              prop="bonus_rate"
              :label="t('newbro.captain.rewardBonusRate')"
              width="140"
            >
              <template #default="{ row }">{{ formatPercentage(row.bonus_rate) }}</template>
            </ElTableColumn>
            <ElTableColumn
              prop="credited_value"
              :label="t('newbro.captain.rewardCreditedValue')"
              width="160"
            >
              <template #default="{ row }">{{ formatCredit(row.credited_value) }}</template>
            </ElTableColumn>
          </ElTable>

          <div class="flex justify-end mt-4">
            <ElPagination
              background
              layout="prev, pager, next"
              :current-page="rewardPage.current"
              :page-size="rewardPage.size"
              :total="rewardPage.total"
              @current-change="
                (page: number) => {
                  rewardPage.current = page
                  loadRewards()
                }
              "
            />
          </div>
        </ElCard>
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import {
    fetchCaptainAttributions,
    fetchCaptainEligiblePlayers,
    fetchCaptainEnrollPlayer,
    fetchCaptainEndAffiliation,
    fetchCaptainOverview,
    fetchCaptainPlayers,
    fetchCaptainRewardSettlements
  } from '@/api/newbro'
  import { buildEveCharacterPortraitUrl } from '@/utils/eve-image'
  import { useNewbroFormatters } from '@/hooks/newbro/useNewbroFormatters'

  defineOptions({ name: 'NewbroCaptainDashboard' })

  const { t } = useI18n()
  const { formatDateTime, formatIsk, formatCredit, formatPercentage } = useNewbroFormatters()

  const loadingOverview = ref(false)
  const loadingPlayers = ref(false)
  const loadingEligiblePlayers = ref(false)
  const loadingAttributions = ref(false)
  const loadingRewards = ref(false)
  const overviewLoaded = ref(false)
  const playersLoaded = ref(false)
  const eligiblePlayersLoaded = ref(false)
  const attributionsLoaded = ref(false)
  const rewardsLoaded = ref(false)
  const activeTab = ref('players')
  const overview = ref<Api.Newbro.CaptainOverview | null>(null)
  const players = ref<Api.Newbro.CaptainPlayerListItem[]>([])
  const eligiblePlayers = ref<Api.Newbro.CaptainEligiblePlayerListItem[]>([])
  const attributions = ref<Api.Newbro.CaptainAttributionItem[]>([])
  const rewardSettlements = ref<Api.Newbro.CaptainRewardSettlementItem[]>([])
  const eligibleKeyword = ref('')
  const submittingPlayerId = ref<number | null>(null)
  const endingPlayerId = ref<number | null>(null)
  const playerStatus = ref<Api.Newbro.CaptainPlayerStatus>('active')
  const attributionSummary = ref<Api.Newbro.CaptainAttributionSummary>({
    attributed_bounty_total: 0,
    record_count: 0
  })
  const rewardSummary = ref<Api.Newbro.CaptainRewardSummary>({
    settlement_count: 0,
    total_credited_value: 0,
    last_processed_at: null
  })
  const playerPage = reactive({ current: 1, size: 20, total: 0 })
  const eligiblePage = reactive({ current: 1, size: 10, total: 0 })
  const attributionPage = reactive({ current: 1, size: 20, total: 0 })
  const rewardPage = reactive({ current: 1, size: 20, total: 0 })
  const isRefreshing = computed(
    () =>
      loadingOverview.value ||
      loadingPlayers.value ||
      loadingEligiblePlayers.value ||
      loadingAttributions.value ||
      loadingRewards.value
  )

  const loadOverview = async () => {
    loadingOverview.value = true
    try {
      overview.value = await fetchCaptainOverview()
      overviewLoaded.value = true
    } finally {
      loadingOverview.value = false
    }
  }

  const loadPlayers = async () => {
    loadingPlayers.value = true
    try {
      const data = await fetchCaptainPlayers({
        current: playerPage.current,
        size: playerPage.size,
        status: playerStatus.value
      })
      players.value = data.list
      playerPage.total = data.total
      playersLoaded.value = true
    } finally {
      loadingPlayers.value = false
    }
  }

  const loadAttributions = async () => {
    loadingAttributions.value = true
    try {
      const data = await fetchCaptainAttributions({
        current: attributionPage.current,
        size: attributionPage.size
      })
      attributions.value = data.list
      attributionSummary.value = data.summary
      attributionPage.total = data.total
      attributionsLoaded.value = true
    } finally {
      loadingAttributions.value = false
    }
  }

  const loadRewards = async () => {
    loadingRewards.value = true
    try {
      const data = await fetchCaptainRewardSettlements({
        current: rewardPage.current,
        size: rewardPage.size
      })
      rewardSettlements.value = data.list
      rewardSummary.value = data.summary
      rewardPage.total = data.total
      rewardsLoaded.value = true
    } finally {
      loadingRewards.value = false
    }
  }

  const loadEligiblePlayers = async () => {
    loadingEligiblePlayers.value = true
    try {
      const data = await fetchCaptainEligiblePlayers({
        current: eligiblePage.current,
        size: eligiblePage.size,
        keyword: eligibleKeyword.value || undefined
      })
      eligiblePlayers.value = data.list
      eligiblePage.total = data.total
      eligiblePlayersLoaded.value = true
    } finally {
      loadingEligiblePlayers.value = false
    }
  }

  const loadAttributionTabData = async () => {
    await Promise.all([loadOverview(), loadAttributions(), loadRewards()])
  }

  const ensurePlayersLoaded = async () => {
    if (!playersLoaded.value) {
      await loadPlayers()
    }
  }

  const ensureEligiblePlayersLoaded = async () => {
    if (!eligiblePlayersLoaded.value) {
      await loadEligiblePlayers()
    }
  }

  const ensureAttributionTabLoaded = async () => {
    if (!overviewLoaded.value || !attributionsLoaded.value || !rewardsLoaded.value) {
      await loadAttributionTabData()
    }
  }

  const handleEligibleSearch = async () => {
    eligiblePage.current = 1
    await loadEligiblePlayers()
  }

  const enrollPlayer = async (playerUserId: number) => {
    submittingPlayerId.value = playerUserId
    try {
      await fetchCaptainEnrollPlayer({ player_user_id: playerUserId })
      ElMessage.success(t('newbro.captain.enrollSuccess'))
      await Promise.all([loadOverview(), loadPlayers(), loadEligiblePlayers()])
    } catch (error) {
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
      console.error('Failed to enroll player', error)
    } finally {
      submittingPlayerId.value = null
    }
  }

  const endPlayerAffiliation = async (playerUserId: number) => {
    endingPlayerId.value = playerUserId
    try {
      await fetchCaptainEndAffiliation({ player_user_id: playerUserId })
      ElMessage.success(t('newbro.captain.endAffiliationSuccess'))
      await Promise.all([loadOverview(), loadPlayers(), loadEligiblePlayers()])
    } catch (error) {
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
      console.error('Failed to end player affiliation', error)
    } finally {
      endingPlayerId.value = null
    }
  }

  const reloadActiveTab = async () => {
    if (activeTab.value === 'players') {
      await loadPlayers()
      return
    }

    if (activeTab.value === 'enroll') {
      await loadEligiblePlayers()
      return
    }

    await loadAttributionTabData()
  }

  watch(activeTab, (value) => {
    if (value === 'players') {
      void ensurePlayersLoaded()
    }

    if (value === 'enroll') {
      void ensureEligiblePlayersLoaded()
    }

    if (value === 'attributions') {
      void ensureAttributionTabLoaded()
    }
  })

  onMounted(() => {
    void ensurePlayersLoaded()
  })
</script>
