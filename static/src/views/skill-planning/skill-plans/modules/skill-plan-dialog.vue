<template>
  <ElDialog
    :model-value="visible"
    :title="dialogTitle"
    width="760px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
  >
    <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="110px">
      <ElFormItem :label="$t('skillPlan.fields.title')" prop="title">
        <ElInput
          v-model="formData.title"
          :placeholder="$t('skillPlan.fields.titlePlaceholder')"
          maxlength="256"
          show-word-limit
        />
      </ElFormItem>

      <ElFormItem :label="$t('skillPlan.fields.description')">
        <ElInput
          v-model="formData.description"
          type="textarea"
          :rows="3"
          :placeholder="$t('skillPlan.fields.descriptionPlaceholder')"
        />
      </ElFormItem>

      <ElFormItem :label="$t('skillPlan.fields.ship')">
        <div class="skill-plan-ship-field">
          <SdeSearchSelect
            v-model="formData.shipTypeId"
            :category-ids="[6]"
            :initial-options="formData.shipInitialOption ? [formData.shipInitialOption] : []"
            :placeholder="$t('skillPlan.fields.shipPlaceholder')"
            style="width: 100%"
            @select="onShipSelect"
          />
          <div class="skill-plan-ship-hint">
            {{ $t('skillPlan.fields.shipHint') }}
          </div>
          <div v-if="formData.shipTypeId" class="skill-plan-ship-preview">
            <img
              :src="getShipIconUrl(formData.shipTypeId, 64)"
              class="skill-plan-ship-preview__icon"
              alt=""
              loading="lazy"
            />
            <div class="skill-plan-ship-preview__meta">
              <strong>{{ formData.shipName || String(formData.shipTypeId) }}</strong>
              <span v-if="formData.shipGroupName">{{ formData.shipGroupName }}</span>
            </div>
          </div>
        </div>
      </ElFormItem>

      <ElFormItem :label="$t('skillPlan.fields.skillsText')">
        <ElInput
          v-model="formData.skillsText"
          type="textarea"
          :rows="10"
          :placeholder="$t('skillPlan.fields.skillsTextPlaceholder')"
        />
        <div class="skills-text-hint">
          <div>{{ $t('skillPlan.fields.skillsTextGuide') }}</div>
          {{ $t('skillPlan.fields.skillsTextHint') }}
        </div>
      </ElFormItem>

      <ElDivider content-position="left">{{ $t('skillPlan.skillListTitle') }}</ElDivider>

      <ElAlert
        v-if="usesTextInput"
        type="info"
        :closable="false"
        class="mb-3"
        :title="$t('skillPlan.textInputActive')"
      />

      <template v-if="!usesTextInput">
        <div v-for="(skill, index) in formData.skills" :key="skill.uid" class="skill-row">
          <div class="skill-row__main">
            <SdeSearchSelect
              v-model="skill.skill_type_id"
              :category-ids="[16]"
              :initial-options="skill.initialOption ? [skill.initialOption] : []"
              :placeholder="$t('skillPlan.fields.skillPlaceholder')"
              :show-type-icon="false"
              style="width: 100%"
              @select="onSkillSelect(index, $event)"
            />

            <ElSelect v-model="skill.required_level" style="width: 160px">
              <ElOption
                v-for="level in 5"
                :key="level"
                :label="$t('skillPlan.level', { level })"
                :value="level"
              />
            </ElSelect>
          </div>

          <div class="skill-row__meta">
            <span class="skill-row__group">{{ skill.group_name || '-' }}</span>
            <ElButton type="danger" link :icon="Delete" @click="removeSkill(index)">
              {{ $t('common.delete') }}
            </ElButton>
          </div>
        </div>

        <ElButton type="primary" plain :icon="Plus" @click="addSkill">
          {{ $t('skillPlan.addSkill') }}
        </ElButton>
      </template>

      <template v-else>
        <div class="skill-row__main">
          <span class="skill-row__group">{{ $t('skillPlan.manualInputHint') }}</span>
        </div>
      </template>
    </ElForm>

    <template #footer>
      <ElButton @click="emit('update:visible', false)">{{ $t('common.cancel') }}</ElButton>
      <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">
        {{ $t('common.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { createSkillPlan, updateSkillPlan } from '@/api/skill-plan'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'
  import { useUserStore } from '@/store/modules/user'
  import { Delete, Plus } from '@element-plus/icons-vue'
  import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'

  interface SkillFormItem {
    uid: string
    skill_type_id: number | null
    skill_name: string
    group_name: string
    required_level: number
    initialOption: Api.Sde.FuzzySearchItem | null
  }

  interface Props {
    visible: boolean
    editing?: Api.SkillPlan.SkillPlanDetail | null
  }

  const props = withDefaults(defineProps<Props>(), {
    editing: null
  })

  const emit = defineEmits<{
    'update:visible': [visible: boolean]
    success: [plan: Api.SkillPlan.SkillPlanDetail]
  }>()

  const { t } = useI18n()
  const userStore = useUserStore()

  const formRef = ref<FormInstance>()
  const submitLoading = ref(false)

  const formData = reactive({
    title: '',
    description: '',
    shipTypeId: null as number | null,
    shipName: '',
    shipGroupName: '',
    shipInitialOption: null as Api.Sde.FuzzySearchItem | null,
    skillsText: '',
    skills: [] as SkillFormItem[]
  })

  const formRules: FormRules = {
    title: [
      {
        required: true,
        message: t('skillPlan.fields.titlePlaceholder'),
        trigger: 'blur'
      }
    ]
  }

  const dialogTitle = computed(() => (props.editing ? t('skillPlan.edit') : t('skillPlan.create')))
  const usesTextInput = computed(() => formData.skillsText.trim().length > 0)

  function buildSkillItem(
    skill?: Api.SkillPlan.SkillRequirement | null,
    fallbackLevel = 1
  ): SkillFormItem {
    return {
      uid: `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
      skill_type_id: skill?.skill_type_id ?? null,
      skill_name: skill?.skill_name ?? '',
      group_name: skill?.group_name ?? '',
      required_level: skill?.required_level ?? fallbackLevel,
      initialOption: skill
        ? {
            id: skill.skill_type_id,
            name: skill.skill_name,
            group_id: 0,
            group_name: skill.group_name,
            category: 'type'
          }
        : null
    }
  }

  function resetForm() {
    formData.title = props.editing?.title ?? ''
    formData.description = props.editing?.description ?? ''
    formData.shipTypeId = props.editing?.ship_type_id ?? null
    formData.shipName = props.editing?.ship_name ?? ''
    formData.shipGroupName = ''
    formData.shipInitialOption = props.editing?.ship_type_id
      ? {
          id: props.editing.ship_type_id,
          name: props.editing.ship_name || String(props.editing.ship_type_id),
          group_id: 0,
          group_name: '',
          category: 'type'
        }
      : null
    formData.skillsText = props.editing?.skills?.length
      ? props.editing.skills
          .map((skill) => `${skill.skill_name} ${skill.required_level}`)
          .join('\n')
      : ''
    formData.skills = props.editing?.skills?.length
      ? props.editing.skills.map((skill) => buildSkillItem(skill))
      : [buildSkillItem()]
  }

  function addSkill() {
    formData.skills.push(buildSkillItem(null))
  }

  function removeSkill(index: number) {
    if (formData.skills.length === 1) {
      formData.skills.splice(0, 1, buildSkillItem(null))
      return
    }
    formData.skills.splice(index, 1)
  }

  function onSkillSelect(index: number, item: Api.Sde.FuzzySearchItem | null) {
    const target = formData.skills[index]
    if (!target) return

    target.skill_type_id = item?.id ?? null
    target.skill_name = item?.name ?? ''
    target.group_name = item?.group_name ?? ''
    target.initialOption = item
  }

  function onShipSelect(item: Api.Sde.FuzzySearchItem | null) {
    formData.shipTypeId = item?.id ?? null
    formData.shipName = item?.name ?? ''
    formData.shipGroupName = item?.group_name ?? ''
    formData.shipInitialOption = item
  }

  function getShipIconUrl(shipTypeId: number, size = 64) {
    return `https://images.evetech.net/types/${shipTypeId}/icon?size=${size}`
  }

  function buildPayload(): Api.SkillPlan.CreateSkillPlanParams {
    return {
      title: formData.title.trim(),
      description: formData.description.trim(),
      ship_type_id: formData.shipTypeId ?? undefined,
      skills_text: formData.skillsText.trim() || undefined,
      skills: usesTextInput.value
        ? undefined
        : formData.skills.map((skill) => ({
            skill_type_id: skill.skill_type_id ?? 0,
            required_level: skill.required_level
          }))
    }
  }

  function validateSkills(payload: Api.SkillPlan.CreateSkillPlanParams) {
    if (payload.skills_text?.trim()) {
      return
    }

    if (!payload.skills?.length) {
      throw new Error(t('skillPlan.noSkills'))
    }

    const seen = new Set<number>()
    for (const skill of payload.skills) {
      if (!skill.skill_type_id) {
        throw new Error(t('skillPlan.fields.skillPlaceholder'))
      }
      if (seen.has(skill.skill_type_id)) {
        throw new Error(t('skillPlan.duplicateSkill'))
      }
      seen.add(skill.skill_type_id)
    }
  }

  async function handleSubmit() {
    const valid = await formRef.value?.validate().catch(() => false)
    if (!valid) return

    const payload = buildPayload()

    try {
      validateSkills(payload)
    } catch (error: any) {
      ElMessage.warning(error?.message ?? t('common.tips'))
      return
    }

    submitLoading.value = true
    try {
      const lang = userStore.language || 'zh'
      const result = props.editing
        ? await updateSkillPlan(props.editing.id, payload, lang)
        : await createSkillPlan(payload, lang)

      ElMessage.success(t(props.editing ? 'skillPlan.updateSuccess' : 'skillPlan.createSuccess'))
      emit('success', result)
    } catch (error: any) {
      ElMessage.error(error?.message ?? t('httpMsg.requestFailed'))
    } finally {
      submitLoading.value = false
    }
  }

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return
      resetForm()
      nextTick(() => {
        formRef.value?.clearValidate()
      })
    },
    { immediate: true }
  )

  watch(
    () => props.editing,
    () => {
      if (props.visible) {
        resetForm()
      }
    }
  )
</script>

<style scoped lang="scss">
  .skill-row {
    border: 1px solid var(--el-border-color-light);
    border-radius: 12px;
    padding: 12px;
    margin-bottom: 12px;
    background: var(--el-fill-color-extra-light);
  }

  .skill-row__main {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .skill-row__meta {
    margin-top: 8px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .skill-row__group {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }

  .skills-text-hint {
    margin-top: 6px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.5;
    white-space: pre-wrap;
  }

  .skill-plan-ship-field {
    width: 100%;
  }

  .skill-plan-ship-hint {
    margin-top: 6px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.5;
  }

  .skill-plan-ship-preview {
    margin-top: 10px;
    display: inline-flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-radius: 12px;
    background: var(--el-fill-color-extra-light);
    border: 1px solid var(--el-border-color-lighter);
  }

  .skill-plan-ship-preview__icon {
    width: 48px;
    height: 48px;
    border-radius: 10px;
    flex-shrink: 0;
  }

  .skill-plan-ship-preview__meta {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .skill-plan-ship-preview__meta strong {
    color: var(--el-text-color-primary);
    line-height: 1.2;
  }

  .skill-plan-ship-preview__meta span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.2;
  }

  @media (max-width: 768px) {
    .skill-row__main {
      flex-direction: column;
      align-items: stretch;
    }
  }
</style>
