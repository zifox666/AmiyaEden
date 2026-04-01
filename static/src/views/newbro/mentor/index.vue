<template>
  <div class="mentor-dashboard-page art-full-height">
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <div class="text-lg font-semibold">{{ t('newbro.mentor.title') }}</div>
          <div class="text-sm text-gray-500 mt-1">{{ t('newbro.mentor.subtitle') }}</div>
        </div>
        <ElButton class="min-w-[120px]" :disabled="isRefreshing" @click="reloadAll">{{
          $t('common.refresh')
        }}</ElButton>
      </div>
    </ElCard>

    <ElTabs v-model="activeTab">
      <ElTabPane :label="t('newbro.mentor.applicationsTab')" name="applications">
        <ElCard shadow="never">
          <template #header>
            <span>{{ t('newbro.mentor.applicationsTitle') }}</span>
          </template>

          <ElEmpty
            v-if="!applications.length && !loadingApplications"
            :description="t('newbro.mentor.pendingApplicationsEmpty')"
            :image-size="72"
          />

          <ElTable v-else :data="applications" v-loading="loadingApplications" stripe border>
            <ElTableColumn :label="t('newbro.common.mentee')" min-width="240">
              <template #default="{ row }">
                <div class="flex items-center gap-3">
                  <ElAvatar :src="row.mentee_portrait_url" :size="40" />
                  <div>
                    <div class="font-medium">{{ row.mentee_character_name }}</div>
                    <div class="text-sm text-gray-500">
                      {{ t('newbro.common.nickname') }}: {{ row.mentee_nickname || '-' }}
                    </div>
                    <div class="text-xs text-gray-500">
                      {{ t('newbro.mentor.qq') }}: {{ row.mentee_qq || '-' }}
                    </div>
                    <div v-if="row.mentee_discord_id" class="text-xs text-gray-500">
                      {{ t('newbro.mentor.discordId') }}: {{ row.mentee_discord_id || '-' }}
                    </div>
                  </div>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn :label="t('newbro.mentor.qq')" min-width="140">
              <template #default="{ row }">{{ row.mentee_qq || '-' }}</template>
            </ElTableColumn>
            <ElTableColumn :label="t('newbro.mentor.discordId')" min-width="160">
              <template #default="{ row }">{{ row.mentee_discord_id || '-' }}</template>
            </ElTableColumn>
            <ElTableColumn prop="mentee_total_sp" :label="t('newbro.mentor.totalSP')" width="160">
              <template #default="{ row }">{{ formatNumber(row.mentee_total_sp) }}</template>
            </ElTableColumn>
            <ElTableColumn prop="mentee_total_pap" :label="t('newbro.mentor.totalPap')" width="160">
              <template #default="{ row }">{{ formatMetric(row.mentee_total_pap) }}</template>
            </ElTableColumn>
            <ElTableColumn
              prop="mentee_days_active"
              :label="t('newbro.mentor.daysActive')"
              width="140"
            />
            <ElTableColumn :label="$t('common.operation')" width="200" fixed="right">
              <template #default="{ row }">
                <div class="flex items-center gap-2">
                  <ElButton
                    type="primary"
                    size="small"
                    :disabled="actioningRelationshipId === row.relationship_id"
                    @click="handleApplicationAction('accept', row.relationship_id)"
                  >
                    {{ t('newbro.mentor.accept') }}
                  </ElButton>
                  <ElButton
                    type="danger"
                    size="small"
                    :disabled="actioningRelationshipId === row.relationship_id"
                    @click="handleApplicationAction('reject', row.relationship_id)"
                  >
                    {{ t('newbro.mentor.reject') }}
                  </ElButton>
                </div>
              </template>
            </ElTableColumn>
          </ElTable>
        </ElCard>
      </ElTabPane>

      <ElTabPane :label="t('newbro.mentor.menteesTab')" name="mentees">
        <ElCard shadow="never">
          <template #header>
            <div class="flex items-center justify-between gap-4 flex-wrap">
              <span>{{ t('newbro.mentor.menteesTitle') }}</span>
              <div class="flex items-center gap-3 flex-wrap">
                <span class="text-sm text-gray-500">{{ t('newbro.mentor.statusFilter') }}</span>
                <ElSelect v-model="statusFilter" style="width: 160px" @change="handleStatusChange">
                  <ElOption
                    v-for="option in statusOptions"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </ElSelect>
              </div>
            </div>
          </template>

          <ElEmpty
            v-if="!mentees.length && !loadingMentees"
            :description="t('newbro.mentor.menteesEmpty')"
            :image-size="72"
          />

          <ElTable v-else :data="mentees" v-loading="loadingMentees" stripe border>
            <ElTableColumn :label="t('newbro.common.mentee')" min-width="240">
              <template #default="{ row }">
                <div class="flex items-center gap-3">
                  <ElAvatar :src="row.mentee_portrait_url" :size="40" />
                  <div>
                    <div class="font-medium">{{ row.mentee_character_name }}</div>
                    <div class="text-sm text-gray-500">
                      {{ t('newbro.common.nickname') }}: {{ row.mentee_nickname || '-' }}
                    </div>
                    <div class="text-xs text-gray-500">
                      {{ t('newbro.mentor.qq') }}: {{ row.mentee_qq || '-' }}
                    </div>
                    <div v-if="row.mentee_discord_id" class="text-xs text-gray-500">
                      {{ t('newbro.mentor.discordId') }}: {{ row.mentee_discord_id || '-' }}
                    </div>
                  </div>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="status" :label="t('newbro.mentor.statusFilter')" width="130">
              <template #default="{ row }">
                <ElTag :type="statusTagType(row.status)" effect="light">
                  {{ formatStatus(row.status) }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="mentee_total_sp" :label="t('newbro.mentor.totalSP')" width="160">
              <template #default="{ row }">{{ formatNumber(row.mentee_total_sp) }}</template>
            </ElTableColumn>
            <ElTableColumn prop="mentee_total_pap" :label="t('newbro.mentor.totalPap')" width="160">
              <template #default="{ row }">{{ formatMetric(row.mentee_total_pap) }}</template>
            </ElTableColumn>
            <ElTableColumn
              prop="mentee_days_active"
              :label="t('newbro.mentor.daysActive')"
              width="140"
            />
            <ElTableColumn :label="t('newbro.mentor.distributedStages')" min-width="180">
              <template #default="{ row }">
                <div v-if="row.distributed_stages.length" class="flex gap-2 flex-wrap">
                  <ElTag v-for="stage in row.distributed_stages" :key="stage" size="small">
                    {{ formatDistributedStageName(stage) }}
                  </ElTag>
                </div>
                <span v-else class="text-gray-400">{{
                  t('newbro.mentor.noDistributedStages')
                }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn
              prop="distributed_reward_amount"
              :label="t('newbro.mentor.distributedRewardAmount')"
              width="180"
            >
              <template #default="{ row }">{{
                formatNumber(row.distributed_reward_amount)
              }}</template>
            </ElTableColumn>
            <ElTableColumn prop="responded_at" :label="t('newbro.mentor.respondedAt')" width="180">
              <template #default="{ row }">{{ formatDateTime(row.responded_at) }}</template>
            </ElTableColumn>
          </ElTable>

          <div class="flex justify-end mt-4">
            <ElPagination
              background
              layout="total, prev, pager, next"
              :current-page="page.current"
              :page-size="page.size"
              :total="page.total"
              @current-change="handleCurrentChange"
            />
          </div>
        </ElCard>
      </ElTabPane>

      <ElTabPane :label="t('newbro.mentor.rewardStagesTab')" name="reward-stages">
        <ElCard shadow="never">
          <template #header>
            <div class="flex items-center justify-between gap-4 flex-wrap">
              <span>{{ t('newbro.mentor.rewardStagesTitle') }}</span>
              <span class="text-sm text-gray-500">{{
                t('newbro.mentor.rewardStagesDescription')
              }}</span>
            </div>
          </template>

          <ElEmpty
            v-if="!rewardStages.length && !loadingRewardStages"
            :description="t('newbro.mentor.rewardStagesEmpty')"
            :image-size="72"
          />

          <ElTable v-else :data="rewardStages" v-loading="loadingRewardStages" stripe border>
            <ElTableColumn
              prop="stage_order"
              :label="t('system.mentorRewardStages.stageOrder')"
              width="140"
            />
            <ElTableColumn
              prop="name"
              :label="t('system.mentorRewardStages.stageName')"
              min-width="220"
            />
            <ElTableColumn :label="t('system.mentorRewardStages.conditionType')" width="220">
              <template #default="{ row }">
                {{ formatConditionType(row.condition_type) }}
              </template>
            </ElTableColumn>
            <ElTableColumn
              prop="threshold"
              :label="t('system.mentorRewardStages.threshold')"
              width="180"
            >
              <template #default="{ row }">{{ formatNumber(row.threshold) }}</template>
            </ElTableColumn>
            <ElTableColumn
              prop="reward_amount"
              :label="t('system.mentorRewardStages.rewardAmount')"
              width="180"
            >
              <template #default="{ row }">{{ formatNumber(row.reward_amount) }}</template>
            </ElTableColumn>
          </ElTable>
        </ElCard>
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import {
    acceptMentorApplication,
    fetchMentorApplications,
    fetchMentorDashboardRewardStages,
    fetchMentorMentees,
    rejectMentorApplication
  } from '@/api/mentor'
  import { useNewbroFormatters } from '@/hooks/newbro/useNewbroFormatters'

  defineOptions({ name: 'NewbroMentorDashboard' })

  const { t } = useI18n()
  const { formatDateTime } = useNewbroFormatters()
  const numberFormatter = new Intl.NumberFormat('en-US', { maximumFractionDigits: 2 })

  const activeTab = ref('applications')
  const loadingApplications = ref(false)
  const loadingMentees = ref(false)
  const loadingRewardStages = ref(false)
  const actioningRelationshipId = ref<number | null>(null)
  const applications = ref<Api.Mentor.MenteeListItem[]>([])
  const mentees = ref<Api.Mentor.MenteeListItem[]>([])
  const rewardStages = ref<Api.Mentor.RewardStage[]>([])
  const statusFilter = ref<Api.Mentor.MenteeStatusFilter>('active')
  const page = reactive({ current: 1, size: 20, total: 0 })

  const statusOptions = computed(() => [
    { label: t('newbro.mentor.allStatuses'), value: 'all' },
    { label: formatStatus('active'), value: 'active' },
    { label: formatStatus('pending'), value: 'pending' },
    { label: formatStatus('rejected'), value: 'rejected' },
    { label: formatStatus('revoked'), value: 'revoked' },
    { label: formatStatus('graduated'), value: 'graduated' }
  ])

  const isRefreshing = computed(
    () =>
      loadingApplications.value ||
      loadingMentees.value ||
      loadingRewardStages.value ||
      actioningRelationshipId.value !== null
  )

  function formatStatus(status: Api.Mentor.MentorRelationshipStatus) {
    return t(`newbro.mentorStatus.${status}`)
  }

  function statusTagType(status: Api.Mentor.MentorRelationshipStatus) {
    switch (status) {
      case 'active':
        return 'success'
      case 'pending':
        return 'warning'
      case 'graduated':
        return 'primary'
      case 'rejected':
      case 'revoked':
        return 'info'
      default:
        return 'info'
    }
  }

  function formatNumber(value: number) {
    return numberFormatter.format(value || 0)
  }

  function formatMetric(value: number) {
    return numberFormatter.format(value || 0)
  }

  function formatConditionType(conditionType: Api.Mentor.RewardConditionType) {
    return t(`newbro.mentorConditionTypes.${conditionType}`)
  }

  function formatDistributedStageName(stageOrder: number) {
    return (
      rewardStages.value.find((stage) => stage.stage_order === stageOrder)?.name || `#${stageOrder}`
    )
  }

  async function loadApplications() {
    loadingApplications.value = true
    try {
      applications.value = await fetchMentorApplications()
    } catch (error) {
      console.error('Failed to load mentor applications', error)
      applications.value = []
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      loadingApplications.value = false
    }
  }

  async function loadRewardStages() {
    loadingRewardStages.value = true
    try {
      rewardStages.value = await fetchMentorDashboardRewardStages()
    } catch (error) {
      console.error('Failed to load mentor reward stages', error)
      rewardStages.value = []
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      loadingRewardStages.value = false
    }
  }

  async function loadMentees() {
    loadingMentees.value = true
    try {
      const data = await fetchMentorMentees({
        current: page.current,
        size: page.size,
        status: statusFilter.value
      })
      mentees.value = data.list
      page.total = data.total
    } catch (error) {
      console.error('Failed to load mentees', error)
      mentees.value = []
      page.total = 0
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      loadingMentees.value = false
    }
  }

  async function reloadAll() {
    await Promise.all([loadApplications(), loadMentees(), loadRewardStages()])
  }

  async function handleApplicationAction(action: 'accept' | 'reject', relationshipId: number) {
    actioningRelationshipId.value = relationshipId
    try {
      if (action === 'accept') {
        await acceptMentorApplication({ relationship_id: relationshipId })
        ElMessage.success(t('newbro.mentor.acceptSuccess'))
      } else {
        await rejectMentorApplication({ relationship_id: relationshipId })
        ElMessage.success(t('newbro.mentor.rejectSuccess'))
      }
      await reloadAll()
    } catch (error) {
      console.error('Failed to handle mentor application', error)
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      actioningRelationshipId.value = null
    }
  }

  async function handleStatusChange() {
    page.current = 1
    await loadMentees()
  }

  async function handleCurrentChange(value: number) {
    page.current = value
    await loadMentees()
  }

  onMounted(() => {
    reloadAll()
  })
</script>
