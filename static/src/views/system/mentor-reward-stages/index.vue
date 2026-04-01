<template>
  <div class="mentor-reward-stages-page art-full-height">
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <div class="text-lg font-semibold">{{ t('system.mentorRewardStages.title') }}</div>
          <div class="text-sm text-gray-500 mt-1">{{
            t('system.mentorRewardStages.subtitle')
          }}</div>
        </div>
        <div class="flex items-center gap-3 flex-wrap">
          <ElButton @click="addStage">{{ t('system.mentorRewardStages.addStage') }}</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">
            {{ t('system.mentorRewardStages.save') }}
          </ElButton>
          <ElButton type="success" :loading="processing" @click="handleRunProcess">
            {{ t('system.mentorRewardStages.runProcess') }}
          </ElButton>
        </div>
      </div>
    </ElCard>

    <ElAlert
      :title="t('system.mentorRewardStages.description')"
      type="info"
      :closable="false"
      show-icon
      class="mb-4"
    />

    <ElCard shadow="never" class="mb-4" v-loading="settingsLoading">
      <template #header>
        <div class="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <div class="text-base font-semibold">{{
              t('system.mentorRewardStages.eligibilityTitle')
            }}</div>
            <div class="text-sm text-gray-500 mt-1">{{
              t('system.mentorRewardStages.eligibilityDescription')
            }}</div>
          </div>
          <ElButton type="primary" :loading="settingsSaving" @click="handleSaveEligibility">
            {{ t('system.mentorRewardStages.saveEligibility') }}
          </ElButton>
        </div>
      </template>

      <ElForm label-width="220px" label-position="left" class="max-w-2xl">
        <ElFormItem :label="t('system.mentorRewardStages.maxCharacterSP')">
          <ElInputNumber
            v-model="mentorSettings.max_character_sp"
            :min="1"
            :step="1000000"
            :controls="false"
            step-strictly
            style="width: 100%"
          />
        </ElFormItem>
        <ElFormItem :label="t('system.mentorRewardStages.maxAccountAgeDays')">
          <ElInputNumber
            v-model="mentorSettings.max_account_age_days"
            :min="1"
            :step="1"
            :controls="false"
            step-strictly
            style="width: 100%"
          />
        </ElFormItem>
      </ElForm>
    </ElCard>

    <ElCard shadow="never" v-loading="loading">
      <ElEmpty
        v-if="!stages.length && !loading"
        :description="t('system.mentorRewardStages.empty')"
        :image-size="72"
      />

      <ElTable v-else :data="stages" stripe border row-key="local_id">
        <ElTableColumn :label="t('system.mentorRewardStages.stageOrder')" width="140">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.stage_order"
              :min="1"
              :step="1"
              :controls="false"
              step-strictly
              style="width: 100%"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('system.mentorRewardStages.stageName')" min-width="220">
          <template #default="{ row }">
            <ElInput v-model="row.name" />
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('system.mentorRewardStages.conditionType')" width="220">
          <template #default="{ row }">
            <ElSelect v-model="row.condition_type" style="width: 100%">
              <ElOption
                v-for="option in conditionOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </ElSelect>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('system.mentorRewardStages.threshold')" width="180">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.threshold"
              :min="1"
              :step="1"
              :controls="false"
              step-strictly
              style="width: 100%"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('system.mentorRewardStages.rewardAmount')" width="180">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.reward_amount"
              :min="1"
              :step="1"
              :controls="false"
              step-strictly
              style="width: 100%"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('system.mentorRewardStages.operation')" width="120" fixed="right">
          <template #default="{ $index }">
            <ElButton link type="danger" @click="removeStage($index)">
              {{ t('system.mentorRewardStages.remove') }}
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import {
	  fetchMentorSettings,
    fetchMentorRewardStages,
    runMentorRewardProcessing,
	  updateMentorSettings,
    updateMentorRewardStages
  } from '@/api/mentor'

  defineOptions({ name: 'MentorRewardStages' })

  type StageRow = Api.Mentor.RewardStageInput & {
    local_id: number
  }

  const { t } = useI18n()
  const numberFormatter = new Intl.NumberFormat('en-US', { maximumFractionDigits: 2 })

  const loading = ref(false)
  const saving = ref(false)
  const processing = ref(false)
  const settingsLoading = ref(false)
  const settingsSaving = ref(false)
  const stages = ref<StageRow[]>([])
  const mentorSettings = reactive<Api.Mentor.Settings>({
    max_character_sp: 4_000_000,
    max_account_age_days: 7
  })
  let nextLocalId = 1

  const conditionOptions = computed(() => [
    {
      label: t('newbro.mentorConditionTypes.skill_points'),
      value: 'skill_points'
    },
    {
      label: t('newbro.mentorConditionTypes.pap_count'),
      value: 'pap_count'
    },
    {
      label: t('newbro.mentorConditionTypes.days_active'),
      value: 'days_active'
    }
  ])

  function toStageRow(stage?: Api.Mentor.RewardStage | Api.Mentor.RewardStageInput): StageRow {
    return {
      local_id: nextLocalId++,
      stage_order: stage?.stage_order ?? stages.value.length + 1,
      name: stage?.name ?? '',
      condition_type: stage?.condition_type ?? 'skill_points',
      threshold: stage?.threshold ?? 1,
      reward_amount: stage?.reward_amount ?? 1
    }
  }

  async function loadStages() {
    loading.value = true
    try {
      const data = await fetchMentorRewardStages()
      stages.value = data.map((stage) => toStageRow(stage))
    } catch (error) {
      console.error('Failed to load mentor reward stages', error)
      stages.value = []
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      loading.value = false
    }
  }

  async function loadMentorSettings() {
    settingsLoading.value = true
    try {
      const data = await fetchMentorSettings()
      mentorSettings.max_character_sp = data.max_character_sp
      mentorSettings.max_account_age_days = data.max_account_age_days
    } catch (error) {
      console.error('Failed to load mentor settings', error)
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      settingsLoading.value = false
    }
  }

  function addStage() {
    const nextOrder =
      stages.value.reduce((maxOrder, stage) => Math.max(maxOrder, stage.stage_order), 0) + 1
    stages.value.push(
      toStageRow({
        stage_order: nextOrder,
        name: '',
        condition_type: 'skill_points',
        threshold: 1,
        reward_amount: 1
      })
    )
  }

  function removeStage(index: number) {
    stages.value.splice(index, 1)
  }

  function validateStages(rows: StageRow[]) {
    const stageOrders = new Set<number>()
    for (const row of rows) {
      if (!Number.isInteger(row.stage_order) || row.stage_order <= 0) {
        return t('system.mentorRewardStages.validation.stageOrder')
      }
      if (stageOrders.has(row.stage_order)) {
        return t('system.mentorRewardStages.validation.stageOrder')
      }
      stageOrders.add(row.stage_order)
      if (!row.name.trim()) {
        return t('system.mentorRewardStages.validation.name')
      }
      if (!row.condition_type) {
        return t('system.mentorRewardStages.validation.conditionType')
      }
      if (!Number.isInteger(row.threshold) || row.threshold <= 0) {
        return t('system.mentorRewardStages.validation.threshold')
      }
      if (!Number.isInteger(row.reward_amount) || row.reward_amount <= 0) {
        return t('system.mentorRewardStages.validation.rewardAmount')
      }
    }
    return ''
  }

  function validateMentorSettings(settings: Api.Mentor.Settings) {
    if (!Number.isInteger(settings.max_character_sp) || settings.max_character_sp <= 0) {
      return t('system.mentorRewardStages.validation.maxCharacterSP')
    }
    if (!Number.isInteger(settings.max_account_age_days) || settings.max_account_age_days <= 0) {
      return t('system.mentorRewardStages.validation.maxAccountAgeDays')
    }
    return ''
  }

  async function handleSaveEligibility() {
    const validationMessage = validateMentorSettings(mentorSettings)
    if (validationMessage) {
      ElMessage.warning(validationMessage)
      return
    }

    settingsSaving.value = true
    try {
      const data = await updateMentorSettings({
        max_character_sp: mentorSettings.max_character_sp,
        max_account_age_days: mentorSettings.max_account_age_days
      })
      mentorSettings.max_character_sp = data.max_character_sp
      mentorSettings.max_account_age_days = data.max_account_age_days
      ElMessage.success(t('system.mentorRewardStages.saveEligibilitySuccess'))
    } catch (error) {
      console.error('Failed to save mentor settings', error)
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      settingsSaving.value = false
    }
  }

  async function handleSave() {
    const validationMessage = validateStages(stages.value)
    if (validationMessage) {
      ElMessage.warning(validationMessage)
      return
    }

    saving.value = true
    try {
      const payload = {
        stages: [...stages.value]
          .sort((a, b) => a.stage_order - b.stage_order)
          .map(({ stage_order, name, condition_type, threshold, reward_amount }) => ({
            stage_order,
            name: name.trim(),
            condition_type,
            threshold,
            reward_amount
          }))
      }
      const data = await updateMentorRewardStages(payload)
      stages.value = data.map((stage) => toStageRow(stage))
      ElMessage.success(t('system.mentorRewardStages.saveSuccess'))
    } catch (error) {
      console.error('Failed to save mentor reward stages', error)
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      saving.value = false
    }
  }

  async function handleRunProcess() {
    processing.value = true
    try {
      const data = await runMentorRewardProcessing()
      ElMessage.success(
        t('system.mentorRewardStages.runProcessSuccess', {
          relationships: data.processed_relationships,
          rewards: data.rewards_distributed,
          total: numberFormatter.format(data.total_coin_awarded),
          graduated: data.graduated_count
        })
      )
    } catch (error) {
      console.error('Failed to process mentor rewards', error)
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      processing.value = false
    }
  }

  onMounted(() => {
    loadStages()
    loadMentorSettings()
  })
</script>
