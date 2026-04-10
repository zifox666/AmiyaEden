<template>
  <div class="newbro-select-page">
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <div class="text-lg font-semibold">{{ t('newbro.select.title') }}</div>
          <div class="text-sm text-gray-500 mt-1">{{ t('newbro.select.subtitle') }}</div>
        </div>
        <ElButton class="min-w-[120px]" :disabled="loading" @click="loadData">{{
          $t('common.refresh')
        }}</ElButton>
      </div>
    </ElCard>

    <ElCard shadow="never" class="mb-4">
      <ElTabs v-model="activeTab">
        <ElTabPane :label="t('newbro.select.overviewTab')" name="overview">
          <div class="space-y-4">
            <ElCard shadow="never">
              <template #header>
                <span>{{ t('newbro.select.currentCaptain') }}</span>
              </template>

              <ElEmpty
                v-if="!state?.current_affiliation"
                :description="t('newbro.select.noCurrentCaptain')"
                :image-size="48"
                class="compact-empty"
              />

              <div v-else class="flex items-center justify-between gap-4 flex-wrap">
                <div class="flex items-center gap-3">
                  <ElAvatar
                    :src="
                      buildEveCharacterPortraitUrl(
                        state.current_affiliation.captain_character_id,
                        48
                      )
                    "
                    :size="48"
                  />
                  <div>
                    <div class="font-medium">
                      {{ state.current_affiliation.captain_character_name }}
                    </div>
                    <div class="text-sm text-gray-500">
                      {{ t('newbro.select.affiliationStarted') }}:
                      {{ formatDateTime(state.current_affiliation.started_at) }}
                    </div>
                  </div>
                </div>
                <div class="flex items-center gap-3 flex-wrap">
                  <ElButton
                    size="small"
                    type="danger"
                    :disabled="endingAffiliation"
                    class="min-w-[150px]"
                    @click="endAffiliation"
                  >
                    {{ t('newbro.select.endAffiliationButton') }}
                  </ElButton>
                  <ElTag type="info" effect="light">
                    #{{ state.current_affiliation.captain_user_id }}
                  </ElTag>
                </div>
              </div>
            </ElCard>

            <ElCard shadow="never">
              <template #header>
                <span>{{ t('newbro.select.captainList') }}</span>
              </template>

              <ElEmpty
                v-if="!captains.length && !loading"
                :description="t('newbro.select.noCaptains')"
                :image-size="72"
              />

              <div v-else class="grid grid-cols-1 xl:grid-cols-2 gap-4">
                <ElCard v-for="captain in captains" :key="captain.captain_user_id" shadow="hover">
                  <div class="flex items-center justify-between gap-4">
                    <div class="flex items-center gap-3">
                      <ElAvatar
                        :src="buildEveCharacterPortraitUrl(captain.captain_character_id, 44)"
                        :size="44"
                      />
                      <div>
                        <div class="font-medium">{{ captain.captain_character_name }}</div>
                        <div class="text-sm text-gray-500">
                          {{ t('newbro.select.nicknameLabel') }}:
                          {{ captain.captain_nickname || '-' }}
                        </div>
                        <div class="text-xs text-gray-500">
                          {{ t('newbro.select.activeNewbros') }}: {{ captain.active_newbro_count }}
                        </div>
                        <div class="text-xs text-gray-400">
                          {{ t('newbro.select.lastOnline') }}:
                          {{ formatDateTime(captain.last_online_at) }}
                        </div>
                      </div>
                    </div>
                    <ElButton
                      type="primary"
                      class="min-w-[150px]"
                      :disabled="
                        !state?.is_currently_newbro ||
                        submittingCaptainId === captain.captain_user_id
                      "
                      @click="selectCaptain(captain.captain_user_id)"
                    >
                      {{
                        state?.current_affiliation?.captain_user_id === captain.captain_user_id
                          ? t('newbro.select.currentSelection')
                          : t('newbro.select.chooseCaptain')
                      }}
                    </ElButton>
                  </div>
                </ElCard>
              </div>
            </ElCard>
          </div>
        </ElTabPane>

        <ElTabPane :label="t('newbro.select.historyTab')" name="history">
          <ElCard shadow="never">
            <template #header>
              <span>{{ t('newbro.select.recentHistory') }}</span>
            </template>

            <ArtTable
              :loading="historyLoading"
              :data="historyData"
              :columns="historyColumns"
              :pagination="historyPagination"
              visual-variant="ledger"
              @pagination:size-change="historyHandleSizeChange"
              @pagination:current-change="historyHandleCurrentChange"
            />
          </ElCard>
        </ElTabPane>
      </ElTabs>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import {
    fetchMyNewbroAffiliation,
    fetchMyAffiliationHistory,
    fetchNewbroCaptains,
    fetchSelectCaptain,
    fetchEndAffiliation
  } from '@/api/newbro'
  import { useTable } from '@/hooks/core/useTable'
  import { useNewbroEligibility } from '@/hooks/newbro/useNewbroEligibility'
  import { useNewbroFormatters } from '@/hooks/newbro/useNewbroFormatters'
  import { buildEveCharacterPortraitUrl } from '@/utils/eve-image'

  defineOptions({ name: 'NewbroSelectCaptain' })

  const { t } = useI18n()
  const { redirectIfIneligible } = useNewbroEligibility()
  const { formatDateTime } = useNewbroFormatters()

  // ─── History tab ───
  const {
    columns: historyColumns,
    data: historyData,
    loading: historyLoading,
    pagination: historyPagination,
    handleSizeChange: historyHandleSizeChange,
    handleCurrentChange: historyHandleCurrentChange,
    getData: loadHistoryData
  } = useTable({
    core: {
      apiFn: fetchMyAffiliationHistory,
      apiParams: { current: 1, size: 200 },
      immediate: false,
      columnsFactory: () => [
        {
          prop: 'captain_character_name',
          label: t('newbro.common.captain'),
          minWidth: 180
        },
        {
          prop: 'started_at',
          label: t('newbro.common.startedAt'),
          width: 180,
          formatter: (row: Api.Newbro.AffiliationSummary) => formatDateTime(row.started_at)
        },
        {
          prop: 'ended_at',
          label: t('newbro.common.endedAt'),
          width: 180,
          formatter: (row: Api.Newbro.AffiliationSummary) =>
            row.ended_at ? formatDateTime(row.ended_at) : '-'
        }
      ]
    }
  })

  const historyLoaded = ref(false)

  const loading = ref(false)
  const submittingCaptainId = ref<number | null>(null)
  const endingAffiliation = ref(false)
  const activeTab = ref('overview')
  const captains = ref<Api.Newbro.CaptainCandidate[]>([])
  const state = ref<Api.Newbro.MyAffiliationResponse | null>(null)

  const loadData = async () => {
    loading.value = true
    try {
      state.value = await fetchMyNewbroAffiliation()
      if (await redirectIfIneligible(state.value)) {
        captains.value = []
        return
      }
      captains.value = await fetchNewbroCaptains()
    } catch (error) {
      console.error('Failed to load newbro selection data', error)
      state.value = null
      captains.value = []
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      loading.value = false
    }
  }

  const selectCaptain = async (captainUserId: number) => {
    submittingCaptainId.value = captainUserId
    try {
      await fetchSelectCaptain({ captain_user_id: captainUserId })
      ElMessage.success(t('newbro.select.chooseSuccess'))
      await loadData()
    } catch (error) {
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
      console.error('Failed to select captain', error)
    } finally {
      submittingCaptainId.value = null
    }
  }

  const endAffiliation = async () => {
    if (!state.value?.current_affiliation) {
      return
    }
    endingAffiliation.value = true
    try {
      await fetchEndAffiliation()
      ElMessage.success(t('newbro.select.endAffiliationSuccess'))
      await loadData()
    } catch (error) {
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
      console.error('Failed to end affiliation', error)
    } finally {
      endingAffiliation.value = false
    }
  }

  watch(activeTab, (tab) => {
    if (tab === 'history' && !historyLoaded.value) {
      historyLoaded.value = true
      loadHistoryData()
    }
  })

  onMounted(() => {
    loadData()
  })
</script>

<style scoped>
  :deep(.compact-empty.el-empty) {
    padding: 8px 0;
  }
</style>
