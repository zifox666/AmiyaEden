<template>
  <div class="mentor-select-page art-full-height">
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <div class="text-lg font-semibold">{{ t('newbro.selectMentor.title') }}</div>
          <div class="text-sm text-gray-500 mt-1">{{ t('newbro.selectMentor.subtitle') }}</div>
        </div>
        <ElButton class="min-w-[120px]" :disabled="loading" @click="loadData">{{
          $t('common.refresh')
        }}</ElButton>
      </div>
    </ElCard>

    <ElCard shadow="never" class="mb-4" v-loading="loading">
      <template #header>
        <span>{{ t('newbro.selectMentor.statusTitle') }}</span>
      </template>

      <ElAlert
        v-if="state"
        :title="
          state.is_eligible
            ? t('newbro.selectMentor.eligible')
            : t('newbro.selectMentor.ineligible')
        "
        :description="
          state.disqualified_reason
            ? t('newbro.selectMentor.disqualifiedReason', {
                reason: formatEligibilityReason(state.disqualified_reason)
              })
            : undefined
        "
        :type="state.is_eligible ? 'success' : 'warning'"
        show-icon
        :closable="false"
      />

      <ElCard shadow="never" class="mt-4">
        <template #header>
          <span>{{ t('newbro.selectMentor.currentRelationship') }}</span>
        </template>

        <ElEmpty
          v-if="!currentRelationship"
          :description="t('newbro.selectMentor.noCurrentRelationship')"
          :image-size="56"
        />

        <div v-else class="flex items-center justify-between gap-4 flex-wrap">
          <div class="flex items-center gap-3">
            <ElAvatar :src="currentRelationship.mentor_portrait_url" :size="48" />
            <div>
              <div class="font-medium">{{ currentRelationship.mentor_character_name }}</div>
              <div class="text-sm text-gray-500">
                {{ t('newbro.selectMentor.nicknameLabel') }}:
                {{ currentRelationship.mentor_nickname || '-' }}
              </div>
              <div class="text-xs text-gray-500 mt-1">
                {{ t('newbro.mentor.appliedAt') }}:
                {{ formatDateTime(currentRelationship.applied_at) }}
              </div>
              <div class="text-xs text-gray-500">
                {{ t('newbro.mentor.qq') }}: {{ currentRelationship.mentor_qq || '-' }}
              </div>
              <div class="text-xs text-gray-500">
                {{ t('newbro.mentor.discordId') }}:
                {{ currentRelationship.mentor_discord_id || '-' }}
              </div>
              <div class="text-xs text-gray-500" v-if="currentRelationship.responded_at">
                {{ t('newbro.mentor.respondedAt') }}:
                {{ formatDateTime(currentRelationship.responded_at) }}
              </div>
            </div>
          </div>
          <ElTag :type="statusTagType(currentRelationship.status)" effect="light">
            {{ formatRelationshipStatus(currentRelationship.status) }}
          </ElTag>
        </div>
      </ElCard>
    </ElCard>

    <ElCard shadow="never" v-if="!hasActiveRelationship">
      <template #header>
        <span>{{ t('newbro.selectMentor.mentorList') }}</span>
      </template>

      <ElEmpty
        v-if="!mentors.length && !loading"
        :description="t('newbro.selectMentor.noMentors')"
        :image-size="72"
      />

      <div v-else class="grid grid-cols-1 xl:grid-cols-2 gap-4">
        <ElCard v-for="mentor in mentors" :key="mentor.mentor_user_id" shadow="hover">
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-3 min-w-0">
              <ElAvatar :src="mentor.mentor_portrait_url" :size="44" />
              <div class="min-w-0">
                <div class="font-medium truncate">{{ mentor.mentor_character_name }}</div>
                <div class="text-sm text-gray-500 truncate">
                  {{ t('newbro.selectMentor.nicknameLabel') }}:
                  {{ mentor.mentor_nickname || '-' }}
                </div>
                <div class="text-xs text-gray-500">
                  {{ t('newbro.selectMentor.activeMentees') }}: {{ mentor.active_mentee_count }}
                </div>
                <div class="text-xs text-gray-500">
                  {{ t('newbro.mentor.qq') }}: {{ mentor.qq || '-' }}
                </div>
                <div class="text-xs text-gray-500">
                  {{ t('newbro.mentor.discordId') }}: {{ mentor.discord_id || '-' }}
                </div>
                <div class="text-xs text-gray-400">
                  {{ t('newbro.selectMentor.lastOnline') }}:
                  {{ formatDateTime(mentor.last_online_at) }}
                </div>
              </div>
            </div>

            <ElButton
              type="primary"
              class="min-w-[128px]"
              :disabled="
                !canApply || applyingMentorId === mentor.mentor_user_id || isCurrentMentor(mentor)
              "
              @click="submitApplication(mentor.mentor_user_id)"
            >
              {{ mentorActionLabel(mentor) }}
            </ElButton>
          </div>
        </ElCard>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { applyForMentor, fetchMentorCandidates, fetchMyMentorStatus } from '@/api/mentor'
  import { useNewbroFormatters } from '@/hooks/newbro/useNewbroFormatters'
  import { useMenuStore } from '@/store/modules/menu'

  defineOptions({ name: 'NewbroSelectMentor' })

  const { t } = useI18n()
  const router = useRouter()
  const menuStore = useMenuStore()
  const { formatDateTime } = useNewbroFormatters()

  const loading = ref(false)
  const applyingMentorId = ref<number | null>(null)
  const state = ref<Api.Mentor.MyStatusResponse | null>(null)
  const mentors = ref<Api.Mentor.MentorCandidate[]>([])

  const currentRelationship = computed(() => state.value?.current_relationship ?? null)
  const hasActiveRelationship = computed(() => currentRelationship.value?.status === 'active')
  const canApply = computed(() => Boolean(state.value?.is_eligible && !currentRelationship.value))

  const statusTagType = (status: Api.Mentor.MentorRelationshipStatus) => {
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

  const formatRelationshipStatus = (status: Api.Mentor.MentorRelationshipStatus) => {
    return t(`newbro.mentorStatus.${status}`)
  }

  const formatEligibilityReason = (reason: string) => {
    switch (reason) {
      case 'account_too_old':
        return t('newbro.selectMentor.ineligibleBecauseAccountTooOld')
      case 'skill_points_too_high':
        return t('newbro.selectMentor.ineligibleBecauseSkillPointsTooHigh')
      case 'no_characters':
        return t('newbro.selectMentor.ineligibleBecauseNoCharacters')
      default:
        return reason
    }
  }

  const isCurrentMentor = (mentor: Api.Mentor.MentorCandidate) => {
    return currentRelationship.value?.mentor_user_id === mentor.mentor_user_id
  }

  const mentorActionLabel = (mentor: Api.Mentor.MentorCandidate) => {
    if (!isCurrentMentor(mentor)) {
      return t('newbro.selectMentor.apply')
    }
    if (currentRelationship.value?.status === 'pending') {
      return t('newbro.selectMentor.applied')
    }
    return t('newbro.selectMentor.activeRelationship')
  }

  const loadData = async () => {
    loading.value = true
    try {
      state.value = await fetchMyMentorStatus()
      if (!state.value.is_eligible) {
        mentors.value = []
        ElMessage.warning(
          t('newbro.selectMentor.disqualifiedReason', {
            reason: formatEligibilityReason(state.value.disqualified_reason)
          })
        )
        const targetPath = menuStore.getHomePath()
        await router.replace(
          targetPath && targetPath !== router.currentRoute.value.path ? targetPath : '/'
        )
        return
      }
      mentors.value = await fetchMentorCandidates()
    } catch (error) {
      console.error('Failed to load mentor selection data', error)
      state.value = null
      mentors.value = []
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      loading.value = false
    }
  }

  const submitApplication = async (mentorUserId: number) => {
    if (!canApply.value) {
      return
    }
    applyingMentorId.value = mentorUserId
    try {
      await applyForMentor({ mentor_user_id: mentorUserId })
      ElMessage.success(t('newbro.selectMentor.applySuccess'))
      await loadData()
    } catch (error) {
      console.error('Failed to apply for mentor', error)
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      applyingMentorId.value = null
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
