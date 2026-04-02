<template>
  <div class="skill-plan-check-page art-full-height">
    <ElCard shadow="never" class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar__info">
          <div class="toolbar__title">{{ $t('skillPlanCheck.title') }}</div>
          <div class="toolbar__subtitle">{{ $t('skillPlanCheck.subtitle') }}</div>
        </div>

        <div class="toolbar__actions">
          <ElButton @click="openPlanDialog">
            {{ $t('skillPlanCheck.selectPlans') }}
          </ElButton>
          <ElButton @click="openCharacterDialog">
            {{ $t('skillPlanCheck.selectCharacters') }}
          </ElButton>
          <ElButton
            type="primary"
            :loading="running"
            :disabled="!selectedCharacterIds.length"
            @click="runCheck"
          >
            {{ $t('skillPlanCheck.startCheck') }}
          </ElButton>
        </div>
      </div>

      <div class="selected-characters">
        <span class="selected-characters__label">{{
          $t('skillPlanCheck.selectedCharacters')
        }}</span>
        <template v-if="selectedCharacters.length">
          <div class="selected-characters__list">
            <ElTag
              v-for="character in selectedCharacters"
              :key="character.character_id"
              effect="light"
              round
            >
              {{ character.character_name }}
            </ElTag>
          </div>
        </template>
        <span v-else class="selected-characters__empty">{{
          $t('skillPlanCheck.noCharactersSelected')
        }}</span>
      </div>

      <div class="selected-plans">
        <span class="selected-plans__label">{{ $t('skillPlanCheck.selectedPlans') }}</span>
        <template v-if="selectedPlanItems.length">
          <div class="selected-plans__list">
            <ElTag v-for="plan in selectedPlanItems" :key="plan.id" effect="light" round>
              {{ plan.title }}
            </ElTag>
          </div>
        </template>
        <span v-else class="selected-plans__empty">{{ $t('skillPlanCheck.noPlansSelected') }}</span>
      </div>
    </ElCard>

    <ElCard shadow="never" class="result-card">
      <div class="result-card__header">
        <div>
          <div class="result-card__title">{{ $t('skillPlanCheck.resultTitle') }}</div>
          <div class="result-card__subtitle">{{ $t('skillPlanCheck.resultSubtitle') }}</div>
        </div>
        <ElTag type="info" effect="light">
          {{ $t('skillPlanCheck.planCount', { count: result?.plan_count ?? 0 }) }}
        </ElTag>
      </div>

      <div v-loading="running" class="result-panel">
        <ElEmpty
          v-if="!result?.characters?.length"
          :description="$t('skillPlanCheck.emptyResult')"
          :image-size="84"
        />

        <ElCollapse v-else v-model="activeCharacterNames" class="character-collapse">
          <ElCollapseItem
            v-for="character in result.characters"
            :key="character.character_id"
            :name="String(character.character_id)"
          >
            <template #title>
              <div class="character-header">
                <div class="character-header__identity">
                  <ElAvatar :src="character.portrait_url" :size="34" />
                  <div>
                    <div class="character-header__name">{{ character.character_name }}</div>
                    <div class="character-header__meta">
                      {{
                        $t('skillPlanCheck.characterSummary', {
                          completed: character.completed_plans,
                          total: character.total_plans
                        })
                      }}
                    </div>
                  </div>
                </div>
                <ElTag
                  :type="
                    character.completed_plans === character.total_plans ? 'success' : 'warning'
                  "
                  effect="light"
                >
                  {{ character.completed_plans }}/{{ character.total_plans }}
                </ElTag>
              </div>
            </template>

            <ElCollapse v-model="activePlanNames[character.character_id]" class="plan-collapse">
              <ElCollapseItem
                v-for="plan in character.plans"
                :key="plan.plan_id"
                :name="String(plan.plan_id)"
              >
                <template #title>
                  <ElTooltip
                    :disabled="!plan.plan_description"
                    placement="top-start"
                    effect="light"
                  >
                    <template #content>
                      <div class="plan-header__description-tooltip">
                        {{ plan.plan_description }}
                      </div>
                    </template>
                    <div class="plan-header">
                      <div class="plan-header__left">
                        <img
                          v-if="plan.ship_type_id"
                          :src="getShipImageUrl(plan.ship_type_id)"
                          :alt="plan.plan_title"
                          class="plan-header__icon"
                          loading="lazy"
                          @error="onShipImageError"
                        />
                        <span class="plan-header__title">{{ plan.plan_title }}</span>
                        <span v-if="plan.plan_description" class="plan-header__info">i</span>
                        <ElTag
                          :type="plan.fully_satisfied ? 'success' : 'danger'"
                          effect="light"
                          size="small"
                        >
                          {{
                            plan.fully_satisfied
                              ? $t('skillPlanCheck.planCompleted')
                              : $t('skillPlanCheck.planIncomplete')
                          }}
                        </ElTag>
                      </div>

                      <div class="plan-header__right">
                        <span class="plan-header__ratio">
                          {{ plan.matched_skills }}/{{ plan.total_skills }}
                        </span>
                        <ElTooltip
                          v-if="!plan.fully_satisfied && plan.missing_skills.length"
                          placement="top"
                          effect="light"
                        >
                          <template #content>
                            <div class="missing-tooltip">
                              <div
                                v-for="skill in plan.missing_skills"
                                :key="`${plan.plan_id}-${skill.skill_type_id}`"
                                class="missing-tooltip__item"
                              >
                                <span class="missing-tooltip__name">{{ skill.skill_name }}</span>
                                <span class="missing-tooltip__level">
                                  {{
                                    $t('skillPlanCheck.requiredVsCurrent', {
                                      required: skill.required_level,
                                      current: skill.current_level
                                    })
                                  }}
                                </span>
                              </div>
                            </div>
                          </template>
                          <span class="plan-header__missing">
                            {{
                              $t('skillPlanCheck.missingCount', {
                                count: plan.missing_skills.length
                              })
                            }}
                          </span>
                        </ElTooltip>
                      </div>
                    </div>
                  </ElTooltip>
                </template>

                <div class="plan-body">
                  <div class="plan-body__summary">
                    {{
                      plan.fully_satisfied
                        ? $t('skillPlanCheck.planCompleteSummary')
                        : $t('skillPlanCheck.planIncompleteSummary', {
                            matched: plan.matched_skills,
                            total: plan.total_skills
                          })
                    }}
                  </div>

                  <div v-if="plan.missing_skills.length" class="missing-skills">
                    <div class="missing-skills__title">{{
                      $t('skillPlanCheck.missingSkillsTitle')
                    }}</div>
                    <ul class="missing-skills__list">
                      <li
                        v-for="skill in plan.missing_skills"
                        :key="`${character.character_id}-${plan.plan_id}-${skill.skill_type_id}`"
                      >
                        <span class="missing-skills__name-row">
                          <span class="name">{{ skill.skill_name }}</span>
                          <ArtCopyButton :text="skill.skill_name" />
                        </span>
                        <span class="meta">
                          {{
                            $t('skillPlanCheck.requiredVsCurrent', {
                              required: skill.required_level,
                              current: skill.current_level
                            })
                          }}
                        </span>
                      </li>
                    </ul>
                  </div>
                </div>
              </ElCollapseItem>
            </ElCollapse>
          </ElCollapseItem>
        </ElCollapse>
      </div>
    </ElCard>

    <ElDialog
      v-model="characterDialogVisible"
      :title="$t('skillPlanCheck.selectCharacters')"
      width="640px"
      destroy-on-close
    >
      <div class="character-dialog">
        <ElCheckboxGroup v-model="draftCharacterIds" class="character-dialog__list">
          <ElCheckbox
            v-for="character in characters"
            :key="character.character_id"
            :label="character.character_id"
            class="character-option"
          >
            <div class="character-option__content">
              <ElAvatar :src="character.portrait_url" :size="32" />
              <span>{{ character.character_name }}</span>
            </div>
          </ElCheckbox>
        </ElCheckboxGroup>

        <ElEmpty
          v-if="!characters.length"
          :description="$t('skillPlanCheck.noAvailableCharacters')"
          :image-size="72"
        />
      </div>

      <template #footer>
        <ElButton @click="characterDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="savingSelection" @click="saveSelection">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>

    <ElDialog
      v-model="planDialogVisible"
      :title="$t('skillPlanCheck.selectPlans')"
      width="640px"
      destroy-on-close
    >
      <div class="character-dialog">
        <ElCheckboxGroup v-model="draftPlanIds" class="character-dialog__list">
          <ElCheckbox
            v-for="plan in allPlans"
            :key="plan.id"
            :label="plan.id"
            class="character-option"
          >
            <div class="character-option__content">
              <span>{{ plan.title }}</span>
            </div>
          </ElCheckbox>
        </ElCheckboxGroup>

        <ElEmpty
          v-if="!allPlans.length"
          :description="$t('skillPlanCheck.noAvailablePlans')"
          :image-size="72"
        />
      </div>

      <template #footer>
        <ElButton @click="planDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="savingPlanSelection" @click="savePlanSelection">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import ArtCopyButton from '@/components/core/forms/art-copy-button/index.vue'
  import { fetchMyCharacters } from '@/api/auth'
  import {
    fetchSkillPlanCheckSelection,
    fetchSkillPlanCheckPlanSelection,
    fetchSkillPlanList,
    runSkillPlanCompletionCheck,
    saveSkillPlanCheckSelection,
    saveSkillPlanCheckPlanSelection
  } from '@/api/skill-plan'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'SkillPlanCompletionCheck' })

  const { t } = useI18n()
  const userStore = useUserStore()

  const currentLang = computed(() => userStore.language || 'zh')
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterIds = ref<number[]>([])
  const draftCharacterIds = ref<number[]>([])
  const result = ref<Api.SkillPlan.CompletionCheckResult | null>(null)
  const characterDialogVisible = ref(false)
  const savingSelection = ref(false)
  const running = ref(false)
  const activeCharacterNames = ref<string[]>([])
  const activePlanNames = ref<Record<number, string[]>>({})

  // Plan selection state
  const allPlans = ref<Api.SkillPlan.SkillPlanListItem[]>([])
  const selectedPlanIds = ref<number[]>([])
  const draftPlanIds = ref<number[]>([])
  const planDialogVisible = ref(false)
  const savingPlanSelection = ref(false)

  function getShipImageUrl(shipTypeId: number, size = 64) {
    return `https://images.evetech.net/types/${shipTypeId}/render?size=${size}`
  }

  function onShipImageError(event: Event) {
    const image = event.target as HTMLImageElement | null
    if (!image) return

    if (image.src.includes('/render')) {
      image.src = image.src.replace('/render', '/icon')
      return
    }

    image.style.display = 'none'
  }

  const selectedCharacters = computed(() => {
    const selectedSet = new Set(selectedCharacterIds.value)
    return characters.value.filter((character) => selectedSet.has(character.character_id))
  })

  const selectedPlanItems = computed(() => {
    const selectedSet = new Set(selectedPlanIds.value)
    return allPlans.value.filter((plan) => selectedSet.has(plan.id))
  })

  function syncDraftSelection() {
    draftCharacterIds.value = [...selectedCharacterIds.value]
  }

  function openCharacterDialog() {
    syncDraftSelection()
    characterDialogVisible.value = true
  }

  async function loadCharacters() {
    characters.value = (await fetchMyCharacters()) ?? []
  }

  async function loadSavedSelection() {
    const saved = await fetchSkillPlanCheckSelection()
    const availableIDs = new Set(characters.value.map((character) => character.character_id))
    selectedCharacterIds.value = (saved.character_ids ?? []).filter((id) => availableIDs.has(id))
    syncDraftSelection()
  }

  async function saveSelection() {
    savingSelection.value = true
    try {
      const payload = {
        character_ids: [...draftCharacterIds.value]
      }
      const saved = await saveSkillPlanCheckSelection(payload)
      selectedCharacterIds.value = saved.character_ids ?? []
      syncDraftSelection()
      characterDialogVisible.value = false
      ElMessage.success(t('skillPlanCheck.selectionSaved'))
    } catch (error: any) {
      ElMessage.error(error?.message ?? t('httpMsg.requestFailed'))
    } finally {
      savingSelection.value = false
    }
  }

  async function loadAllPlans() {
    try {
      const res = await fetchSkillPlanList({ current: 1, size: 200 })
      allPlans.value = res?.list ?? []
    } catch {
      allPlans.value = []
    }
  }

  async function loadSavedPlanSelection() {
    const saved = await fetchSkillPlanCheckPlanSelection()
    const availableIDs = new Set(allPlans.value.map((plan) => plan.id))
    selectedPlanIds.value = (saved.plan_ids ?? []).filter((id) => availableIDs.has(id))
    draftPlanIds.value = [...selectedPlanIds.value]
  }

  function openPlanDialog() {
    draftPlanIds.value = [...selectedPlanIds.value]
    planDialogVisible.value = true
  }

  async function savePlanSelection() {
    savingPlanSelection.value = true
    try {
      const saved = await saveSkillPlanCheckPlanSelection({
        plan_ids: [...draftPlanIds.value]
      })
      selectedPlanIds.value = saved.plan_ids ?? []
      draftPlanIds.value = [...selectedPlanIds.value]
      planDialogVisible.value = false
      ElMessage.success(t('skillPlanCheck.planSelectionSaved'))
    } catch (error: any) {
      ElMessage.error(error?.message ?? t('httpMsg.requestFailed'))
    } finally {
      savingPlanSelection.value = false
    }
  }

  function expandResults(data: Api.SkillPlan.CompletionCheckResult) {
    activeCharacterNames.value = data.characters.map((character) => String(character.character_id))
    activePlanNames.value = Object.fromEntries(
      data.characters.map((character) => [character.character_id, []])
    )
  }

  async function runCheck() {
    if (!selectedCharacterIds.value.length) {
      ElMessage.warning(t('skillPlanCheck.selectCharactersFirst'))
      return
    }

    running.value = true
    try {
      const data = await runSkillPlanCompletionCheck({
        character_ids: selectedCharacterIds.value,
        language: currentLang.value
      })
      result.value = data
      expandResults(data)
    } catch (error: any) {
      ElMessage.error(error?.message ?? t('httpMsg.requestFailed'))
    } finally {
      running.value = false
    }
  }

  onMounted(async () => {
    await loadCharacters()
    await loadSavedSelection()
    await loadAllPlans()
    await loadSavedPlanSelection()
  })
</script>

<style scoped lang="scss">
  .skill-plan-check-page {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .toolbar-card,
  .result-card {
    border-radius: 16px;
  }

  .result-card {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;

    :deep(.el-card__body) {
      flex: 1;
      min-height: 0;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
  }

  .toolbar {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
  }

  .toolbar__title,
  .result-card__title {
    font-size: 18px;
    font-weight: 700;
    color: var(--el-text-color-primary);
  }

  .toolbar__subtitle,
  .result-card__subtitle,
  .character-header__meta,
  .plan-body__summary {
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }

  .toolbar__actions {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .selected-characters,
  .selected-plans {
    margin-top: 14px;
    display: flex;
    gap: 10px;
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .selected-characters__label,
  .selected-plans__label {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    padding-top: 4px;
  }

  .selected-characters__list,
  .selected-plans__list {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .selected-characters__empty,
  .selected-plans__empty {
    color: var(--el-text-color-placeholder);
    font-size: 13px;
  }

  .result-panel {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .result-card__header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: center;
    margin-bottom: 12px;
  }

  .character-header,
  .plan-header {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .character-header__identity,
  .plan-header__left,
  .plan-header__right {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .character-header__name,
  .plan-header__title {
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .plan-header__info {
    width: 18px;
    height: 18px;
    border-radius: 999px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 700;
    color: var(--el-color-primary);
    background: color-mix(in srgb, var(--el-color-primary) 12%, transparent);
    flex-shrink: 0;
  }

  .plan-header__icon {
    width: 28px;
    height: 28px;
    border-radius: 8px;
    flex-shrink: 0;
  }

  .plan-header__right {
    margin-left: auto;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .plan-header__description-tooltip {
    max-width: 320px;
    white-space: pre-wrap;
    line-height: 1.5;
  }

  .plan-header__ratio {
    font-weight: 700;
    color: var(--el-text-color-primary);
  }

  .plan-header__missing {
    color: var(--el-color-danger);
    font-size: 12px;
  }

  .plan-collapse,
  .character-collapse {
    border-top: none;
  }

  .plan-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 4px 2px 8px;
  }

  .missing-skills {
    background: color-mix(in srgb, var(--el-color-danger) 6%, transparent);
    border: 1px solid color-mix(in srgb, var(--el-color-danger) 16%, transparent);
    border-radius: 12px;
    padding: 12px 14px;
  }

  .missing-skills__title {
    font-weight: 600;
    color: var(--el-color-danger);
    margin-bottom: 8px;
  }

  .missing-skills__list {
    margin: 0;
    padding-left: 18px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .missing-skills__name-row {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    max-width: 100%;
  }

  .missing-skills__name-row :deep(.art-copy-button) {
    flex-shrink: 0;
  }

  .missing-skills__list .name,
  .missing-tooltip__name {
    color: var(--el-color-danger);
    font-weight: 600;
  }

  .missing-skills__list .meta,
  .missing-tooltip__level {
    margin-left: 6px;
    color: var(--el-text-color-secondary);
  }

  .missing-tooltip {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-width: 320px;
  }

  .character-dialog__list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .character-option {
    padding: 10px 12px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 12px;
    margin-right: 0;
    width: 100%;
  }

  .character-option:hover {
    border-color: var(--el-color-primary-light-5);
    background: color-mix(in srgb, var(--el-color-primary) 4%, transparent);
  }

  .character-option :deep(.el-checkbox__label) {
    width: 100%;
    padding-left: 10px;
  }

  .character-option__content {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  @media (max-width: 768px) {
    .toolbar,
    .result-card__header,
    .character-header,
    .plan-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .plan-header__right {
      margin-left: 0;
      justify-content: flex-start;
    }
  }
</style>
