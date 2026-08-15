<template>
  <div class="art-card p-5">
    <div class="flex-b mb-4"
      ><div
        ><h2 class="text-xl">{{ t('at.recruitment.title') }}</h2
        ><p class="text-sm text-g-500 mt-1">{{ t('at.recruitment.description') }}</p></div
      ><ElButton type="primary" @click="openQuestionnaire">{{
        t('at.recruitment.createQuestionnaire')
      }}</ElButton
      ><ElButton v-if="isAdmin" type="danger" plain @click="privacyCleanup">{{
        t('at.recruitment.privacyCleanup')
      }}</ElButton></div
    >
    <ElTabs>
      <ElTabPane :label="t('at.recruitment.questionnaires')">
        <ArtTable
          :data="pagedQuestionnaires"
          :loading="loading"
          :columns="questionnaireColumns"
          :pagination="questionnairePagination"
          @pagination:size-change="(size) => changePageSize(questionnairePagination, size)"
          @pagination:current-change="questionnairePagination.current = $event"
        >
          <template #published="{ row }"
            ><ElTag :type="row.published ? 'success' : 'info'">{{
              row.published ? t('at.common.yes') : t('at.common.no')
            }}</ElTag></template
          >
          <template #operation="{ row }">
            <div class="flex items-center gap-2">
              <ArtButtonTable
                type="edit"
                :title="row.published ? t('at.recruitment.cloneToEdit') : t('at.common.edit')"
                @click="editQuestionnaire(row.id)"
              />
              <ArtButtonTable
                icon="ri:eye-line"
                icon-class="bg-theme/12 text-theme"
                :title="t('at.recruitment.preview')"
                @click="previewQuestionnaire(row.id)"
              />
              <ArtButtonTable
                v-if="row.published"
                icon="ri:link"
                icon-class="bg-success/12 text-success"
                :title="t('at.recruitment.createLink')"
                @click="openLink(row.id)"
              />
            </div>
          </template>
        </ArtTable>
      </ElTabPane>
      <ElTabPane :label="t('at.recruitment.links')">
        <ArtTable
          :data="pagedLinks"
          :loading="loading"
          :columns="linkColumns"
          :pagination="linkPagination"
          @pagination:size-change="(size) => changePageSize(linkPagination, size)"
          @pagination:current-change="linkPagination.current = $event"
        >
          <template #active="{ row }"
            ><ElSwitch :model-value="row.active" @change="toggleLink(row.id, Boolean($event))"
          /></template>
        </ArtTable>
      </ElTabPane>
      <ElTabPane :label="t('at.recruitment.applications')">
        <ArtTable
          :data="pagedApplications"
          :loading="loading"
          :columns="applicationColumns"
          :pagination="applicationPagination"
          @pagination:size-change="(size) => changePageSize(applicationPagination, size)"
          @pagination:current-change="applicationPagination.current = $event"
        >
          <template #status="{ row }"
            ><ElTag>{{ t(`at.recruitment.status.${row.status}`) }}</ElTag></template
          >
          <template #operation="{ row }"
            ><ArtButtonTable
              icon="ri:eye-line"
              icon-class="bg-theme/12 text-theme"
              :title="t('at.common.details')"
              @click="openApplication(row.id)"
          /></template>
        </ArtTable>
      </ElTabPane>
    </ElTabs>

    <ElDialog
      v-model="questionnaireVisible"
      :title="
        editingQuestionnaireId
          ? t('at.recruitment.editQuestionnaire')
          : t('at.recruitment.createQuestionnaire')
      "
      width="1180px"
      top="4vh"
      class="questionnaire-designer-dialog"
    >
      <div class="designer-shell">
        <aside class="designer-sidebar">
          <h3>{{ t('at.recruitment.systemFields') }}</h3>
          <p class="designer-hint">{{ t('at.recruitment.systemFieldsHint') }}</p>
          <div class="system-field-list">
            <div v-for="item in systemFields" :key="item" class="system-field-item">
              <ArtSvgIcon icon="ri:checkbox-circle-line" />
              <span>{{ t(`at.recruitment.systemFieldItems.${item}`) }}</span>
            </div>
          </div>
          <ElDivider />
          <h3>{{ t('at.recruitment.questionTypes') }}</h3>
          <div class="question-type-list">
            <ElButton v-for="type in fieldTypes" :key="type" plain @click="addField(type)">
              <ArtSvgIcon :icon="fieldTypeIcon(type)" />
              {{ t(`at.recruitment.fieldTypes.${type}`) }}
            </ElButton>
          </div>
        </aside>

        <main class="designer-canvas">
          <section class="questionnaire-heading">
            <ElInput
              v-model="questionnaire.title"
              size="large"
              :placeholder="t('at.recruitment.titlePlaceholder')"
            />
            <ElInput
              v-model="questionnaire.description"
              type="textarea"
              :rows="3"
              :placeholder="t('at.recruitment.introductionPlaceholder')"
            />
          </section>

          <section
            v-for="(field, index) in questionnaire.fields"
            :key="field.field_key"
            class="question-card"
          >
            <div class="question-card-header">
              <span class="question-number">{{ String(index + 1).padStart(2, '0') }}</span>
              <ElInput
                v-model="field.label"
                :placeholder="t('at.recruitment.fieldLabel')"
                class="question-title-input"
              />
              <ElSelect v-model="field.type" class="question-type-select">
                <ElOption
                  v-for="type in fieldTypes"
                  :key="type"
                  :label="t(`at.recruitment.fieldTypes.${type}`)"
                  :value="type"
                />
                <ElOption
                  v-if="!fieldTypes.includes(field.type)"
                  :label="t(`at.recruitment.fieldTypes.${field.type}`)"
                  :value="field.type"
                />
              </ElSelect>
            </div>
            <ElInput
              v-model="field.description"
              :placeholder="t('at.recruitment.questionDescriptionPlaceholder')"
              class="mt-3"
            />

            <div class="question-preview">
              <ElInput
                v-if="field.type === 'text'"
                disabled
                :placeholder="t('at.recruitment.answerPlaceholder')"
              />
              <ElInput
                v-else-if="field.type === 'textarea'"
                type="textarea"
                :rows="3"
                disabled
                :placeholder="t('at.recruitment.answerPlaceholder')"
              />
              <template v-else-if="['radio', 'checkbox', 'select'].includes(field.type)">
                <div
                  v-for="(_, optionIndex) in field.options ?? []"
                  :key="optionIndex"
                  class="option-row"
                >
                  <ArtSvgIcon
                    :icon="
                      field.type === 'checkbox'
                        ? 'ri:checkbox-blank-line'
                        : field.type === 'radio'
                          ? 'ri:radio-button-line'
                          : 'ri:arrow-down-s-line'
                    "
                  />
                  <ElInput
                    :model-value="field.options?.[optionIndex] ?? ''"
                    :placeholder="t('at.recruitment.option', { index: optionIndex + 1 })"
                    @update:model-value="setOption(field, optionIndex, String($event))"
                  />
                  <ElButton link type="danger" @click="removeOption(field, optionIndex)"
                    ><ArtSvgIcon icon="ri:close-line"
                  /></ElButton>
                </div>
                <ElButton link type="primary" @click="addOption(field)"
                  ><ArtSvgIcon icon="ri:add-circle-line" />{{
                    t('at.recruitment.addOption')
                  }}</ElButton
                >
              </template>
              <ElInputNumber v-else-if="field.type === 'number'" disabled />
              <ElDatePicker v-else-if="field.type === 'date'" disabled class="w-full" />
              <ElButton v-else-if="field.type === 'image'" disabled>{{
                t('at.recruitment.uploadImage')
              }}</ElButton>
            </div>

            <div class="question-actions">
              <ElCheckbox v-model="field.required">{{ t('at.recruitment.required') }}</ElCheckbox>
              <div class="flex items-center gap-1">
                <ElButton
                  link
                  :disabled="index === 0"
                  :title="t('at.recruitment.moveUp')"
                  @click="moveField(index, -1)"
                  ><ArtSvgIcon icon="ri:arrow-up-line"
                /></ElButton>
                <ElButton
                  link
                  :disabled="index === questionnaire.fields.length - 1"
                  :title="t('at.recruitment.moveDown')"
                  @click="moveField(index, 1)"
                  ><ArtSvgIcon icon="ri:arrow-down-line"
                /></ElButton>
                <ElButton
                  link
                  :title="t('at.recruitment.duplicateQuestion')"
                  @click="duplicateField(index)"
                  ><ArtSvgIcon icon="ri:file-copy-line"
                /></ElButton>
                <ElButton
                  link
                  type="danger"
                  :title="t('at.common.delete')"
                  @click="questionnaire.fields.splice(index, 1)"
                  ><ArtSvgIcon icon="ri:delete-bin-line"
                /></ElButton>
              </div>
            </div>
          </section>
          <ElEmpty
            v-if="questionnaire.fields.length === 0"
            :description="t('at.recruitment.addQuestionHint')"
          />
        </main>
      </div>
      <template #footer>
        <ElButton @click="questionnaireVisible = false">{{ t('at.common.cancel') }}</ElButton>
        <ElButton :loading="savingQuestionnaire" @click="saveQuestionnaire(false)">{{
          t('at.recruitment.saveDraft')
        }}</ElButton>
        <ElButton type="primary" :loading="savingQuestionnaire" @click="saveQuestionnaire(true)">{{
          t('at.recruitment.saveAndPublish')
        }}</ElButton>
      </template>
    </ElDialog>
    <ElDialog
      v-model="previewVisible"
      :title="previewDetail?.questionnaire.title ?? t('at.recruitment.preview')"
      width="720px"
    >
      <template v-if="previewDetail">
        <p class="text-g-500 mb-5">{{ previewDetail.questionnaire.description }}</p>
        <ElForm label-position="top" disabled>
          <ElFormItem
            v-for="field in previewDetail.fields"
            :key="field.field_key"
            :label="field.label"
            :required="field.required"
          >
            <p v-if="field.description" class="w-full text-sm text-g-500 mb-2">{{
              field.description
            }}</p>
            <ElInput v-if="field.type === 'text'" />
            <ElInput v-else-if="field.type === 'textarea'" type="textarea" :rows="3" />
            <ElInputNumber v-else-if="field.type === 'number'" />
            <ElDatePicker v-else-if="field.type === 'date'" class="w-full" />
            <ElSelect v-else-if="field.type === 'select'" class="w-full">
              <ElOption
                v-for="option in field.options ?? []"
                :key="option"
                :label="option"
                :value="option"
              />
            </ElSelect>
            <ElRadioGroup v-else-if="field.type === 'radio'">
              <ElRadio v-for="option in field.options ?? []" :key="option" :value="option">{{
                option
              }}</ElRadio>
            </ElRadioGroup>
            <ElCheckboxGroup v-else-if="field.type === 'checkbox'">
              <ElCheckbox v-for="option in field.options ?? []" :key="option" :value="option">{{
                option
              }}</ElCheckbox>
            </ElCheckboxGroup>
            <ElButton v-else-if="field.type === 'image'" disabled>{{
              t('at.recruitment.uploadImage')
            }}</ElButton>
          </ElFormItem>
        </ElForm>
      </template>
    </ElDialog>
    <ElDialog v-model="linkVisible" :title="t('at.recruitment.createLink')"
      ><ElForm label-position="top"
        ><ElFormItem :label="t('at.recruitment.expiresAt')"
          ><ElDatePicker
            v-model="linkForm.expires_at"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            class="w-full" /></ElFormItem
        ><ElFormItem :label="t('at.recruitment.maxSubmissions')"
          ><ElInputNumber v-model="linkForm.max_submissions" :min="0" /></ElFormItem></ElForm
      ><ElAlert v-if="createdURL" type="success" :title="createdURL" :closable="false" /><template
        #footer
        ><ElButton type="primary" @click="createLink">{{
          t('at.common.create')
        }}</ElButton></template
      ></ElDialog
    >
    <ElDialog
      v-model="applicationVisible"
      :title="t('at.recruitment.applicationDetail')"
      width="860px"
    >
      <template v-if="applicationDetail">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem :label="t('at.recruitment.mainCharacter')">{{
            applicationDetail.application.main_character_name
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="t('at.recruitment.characterId')">{{
            applicationDetail.application.main_character_id
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="t('at.recruitment.corporationId')">{{
            applicationDetail.application.corporation_id
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="t('at.common.status')">{{
            applicationDetail.application.status
          }}</ElDescriptionsItem>
        </ElDescriptions>
        <h3 class="mt-4 mb-2">{{ t('at.recruitment.answers') }}</h3>
        <ElTable :data="applicationDetail.answers"
          ><ElTableColumn :label="t('at.recruitment.fieldLabel')"
            ><template #default="{ row }">{{ answerLabel(row.field_key) }}</template></ElTableColumn
          ><ElTableColumn :label="t('at.recruitment.answer')"
            ><template #default="{ row }"
              ><ElImage
                v-if="typeof row.value === 'string' && row.value.startsWith('/uploads/')"
                style="width: 120px; height: 80px"
                fit="cover"
                :src="row.value"
                :preview-src-list="[row.value]"
              /><span v-else>{{ JSON.stringify(row.value) }}</span></template
            ></ElTableColumn
          ></ElTable
        >
        <h3 class="mt-4 mb-2">{{ t('at.recruitment.skills') }}</h3>
        <ElTable :data="applicationDetail.skills" max-height="260"
          ><ElTableColumn prop="skill_id" label="TypeID" /><ElTableColumn
            prop="trained_level"
            :label="t('at.recruitment.trainedLevel')" /><ElTableColumn
            prop="skillpoints_in_skill"
            :label="t('at.recruitment.skillpoints')"
        /></ElTable>
        <ElForm v-if="canReview" label-position="top" class="mt-4">
          <ElFormItem :label="t('at.common.status')"
            ><ElSelect v-model="reviewForm.status" class="w-full"
              ><ElOption
                v-for="status in reviewStatuses"
                :key="status"
                :label="t(`at.recruitment.status.${status}`)"
                :value="status" /></ElSelect
          ></ElFormItem>
          <ElFormItem :label="t('at.common.reason')"
            ><ElInput v-model="reviewForm.reason" type="textarea"
          /></ElFormItem>
          <ElButton type="primary" @click="reviewApplication">{{
            t('at.recruitment.review')
          }}</ElButton>
        </ElForm>
      </template>
    </ElDialog>
  </div>
</template>
<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import type { ColumnOption } from '@/types'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { recruitmentApi } from '@/api/at'
  import { useUserStore } from '@/store/modules/user'
  const { t } = useI18n(),
    questionnaires = ref<Api.AT.Questionnaire[]>([]),
    links = ref<Api.AT.RecruitmentLink[]>([]),
    applications = ref<Api.AT.RecruitmentApplication[]>([]),
    questionnaireVisible = ref(false),
    linkVisible = ref(false),
    selectedQuestionnaire = ref(0),
    createdURL = ref(''),
    editingQuestionnaireId = ref<number | null>(null),
    savingQuestionnaire = ref(false),
    previewVisible = ref(false),
    previewDetail = ref<Api.AT.QuestionnaireDetail>(),
    loading = ref(false)
  const questionnairePagination = reactive({ current: 1, size: 10, total: 0 })
  const linkPagination = reactive({ current: 1, size: 10, total: 0 })
  const applicationPagination = reactive({ current: 1, size: 10, total: 0 })
  const userStore = useUserStore()
  const isAdmin = computed(() =>
    (userStore.info.roles ?? []).some((role) => ['admin', 'super_admin'].includes(role))
  )
  const canReview = computed(() =>
    (userStore.info.roles ?? []).some((role) =>
      ['recruiter', 'admin', 'super_admin'].includes(role)
    )
  )
  const applicationVisible = ref(false)
  const applicationDetail = ref<Api.AT.RecruitmentApplicationDetail>()
  const reviewStatuses = ['reviewing', 'approved', 'rejected', 'draft']
  const reviewForm = reactive({ status: 'reviewing', reason: '' })
  const fieldTypes: Api.AT.RecruitmentField['type'][] = [
    'text',
    'textarea',
    'radio',
    'checkbox',
    'select'
  ]
  const systemFields = [
    'account',
    'profile',
    'tournamentCharacter',
    'mainCharacter',
    'publicInfo',
    'skills'
  ] as const
  const questionnaire = reactive<Api.AT.QuestionnaireInput>({
      title: '',
      description: '',
      fields: []
    }),
    linkForm = reactive<{ max_submissions: number; expires_at: string | null }>({
      max_submissions: 0,
      expires_at: null
    })
  const pagedQuestionnaires = computed(() =>
    paginate(questionnaires.value, questionnairePagination)
  )
  const pagedLinks = computed(() => paginate(links.value, linkPagination))
  const pagedApplications = computed(() => paginate(applications.value, applicationPagination))
  const questionnaireColumns = computed<ColumnOption[]>(() => [
    { prop: 'title', label: t('at.common.name'), minWidth: 240 },
    { prop: 'version', label: t('at.scoreRules.version'), width: 110 },
    { prop: 'published', label: t('at.recruitment.published'), width: 120, useSlot: true },
    { prop: 'operation', label: t('at.common.actions'), width: 180, fixed: 'right', useSlot: true }
  ])
  const linkColumns = computed<ColumnOption[]>(() => [
    { prop: 'id', label: '#', width: 90 },
    { prop: 'questionnaire_id', label: t('at.recruitment.questionnaireId'), minWidth: 180 },
    { prop: 'submission_count', label: t('at.recruitment.submissions'), width: 130 },
    { prop: 'active', label: t('at.common.status'), width: 130, useSlot: true }
  ])
  const applicationColumns = computed<ColumnOption[]>(() => [
    { prop: 'main_character_name', label: t('at.recruitment.mainCharacter'), minWidth: 220 },
    { prop: 'main_character_id', label: t('at.recruitment.characterId'), minWidth: 150 },
    { prop: 'status', label: t('at.common.status'), width: 130, useSlot: true },
    { prop: 'submitted_at', label: t('at.recruitment.submittedAt'), minWidth: 190 },
    { prop: 'operation', label: t('at.common.actions'), width: 100, fixed: 'right', useSlot: true }
  ])
  function paginate<T>(items: T[], pagination: { current: number; size: number; total: number }) {
    const start = (pagination.current - 1) * pagination.size
    return items.slice(start, start + pagination.size)
  }
  watch(
    () => questionnaires.value.length,
    (total) => (questionnairePagination.total = total),
    { immediate: true }
  )
  watch(
    () => links.value.length,
    (total) => (linkPagination.total = total),
    { immediate: true }
  )
  watch(
    () => applications.value.length,
    (total) => (applicationPagination.total = total),
    { immediate: true }
  )
  function changePageSize(pagination: { current: number; size: number }, size: number) {
    pagination.size = size
    pagination.current = 1
  }
  function createFieldKey() {
    return `question_${crypto.randomUUID().replaceAll('-', '').slice(0, 16)}`
  }
  const addField = (type: Api.AT.RecruitmentField['type'] = 'text') => {
    questionnaire.fields.push({
      field_key: createFieldKey(),
      type,
      label: '',
      description: '',
      required: false,
      sort: questionnaire.fields.length,
      options: ['radio', 'checkbox', 'select'].includes(type) ? ['', ''] : []
    })
  }
  function fieldTypeIcon(type: Api.AT.RecruitmentField['type']) {
    return (
      (
        {
          text: 'ri:text',
          textarea: 'ri:align-left',
          radio: 'ri:radio-button-line',
          checkbox: 'ri:checkbox-line',
          select: 'ri:list-check-3'
        } as Record<string, string>
      )[type] ?? 'ri:question-line'
    )
  }
  function addOption(field: Api.AT.RecruitmentField) {
    field.options ??= []
    field.options.push('')
  }
  function removeOption(field: Api.AT.RecruitmentField, index: number) {
    field.options?.splice(index, 1)
  }
  function setOption(field: Api.AT.RecruitmentField, index: number, value: string) {
    field.options ??= []
    field.options[index] = value
  }
  function moveField(index: number, offset: number) {
    const target = index + offset
    if (target < 0 || target >= questionnaire.fields.length) return
    const [field] = questionnaire.fields.splice(index, 1)
    questionnaire.fields.splice(target, 0, field)
  }
  function duplicateField(index: number) {
    const source = questionnaire.fields[index]
    questionnaire.fields.splice(index + 1, 0, {
      field_key: createFieldKey(),
      type: source.type,
      label: source.label,
      description: source.description,
      required: source.required,
      sort: index + 1,
      options: [...(source.options ?? [])]
    })
  }
  function openQuestionnaire() {
    editingQuestionnaireId.value = null
    Object.assign(questionnaire, { title: '', description: '', fields: [] })
    addField()
    questionnaireVisible.value = true
  }
  async function editQuestionnaire(id: number) {
    const detail = await recruitmentApi.getQuestionnaire(id)
    const cloning = detail.questionnaire.published
    editingQuestionnaireId.value = cloning ? null : id
    Object.assign(questionnaire, {
      title: detail.questionnaire.title,
      description: detail.questionnaire.description,
      fields: detail.fields.map((field) => ({
        field_key: cloning ? createFieldKey() : field.field_key,
        type: field.type,
        label: field.label,
        description: field.description ?? '',
        required: field.required,
        sort: field.sort,
        options: [...(field.options ?? [])]
      }))
    })
    if (cloning) ElMessage.info(t('at.recruitment.clonedDraftHint'))
    questionnaireVisible.value = true
  }
  async function previewQuestionnaire(id: number) {
    previewDetail.value = await recruitmentApi.getQuestionnaire(id)
    previewVisible.value = true
  }
  async function load() {
    loading.value = true
    try {
      ;[questionnaires.value, links.value, applications.value] = await Promise.all([
        recruitmentApi.listQuestionnaires(),
        recruitmentApi.listLinks(),
        recruitmentApi.listApplications()
      ])
    } finally {
      loading.value = false
    }
  }
  async function saveQuestionnaire(publish: boolean) {
    questionnaire.fields.forEach((field, index) => (field.sort = index))
    savingQuestionnaire.value = true
    try {
      if (editingQuestionnaireId.value) {
        await recruitmentApi.updateQuestionnaire(
          editingQuestionnaireId.value,
          questionnaire,
          publish
        )
      } else {
        const created = await recruitmentApi.createQuestionnaire(questionnaire)
        if (publish) {
          await recruitmentApi.updateQuestionnaire(created.questionnaire.id, questionnaire, true)
        }
      }
      questionnaireVisible.value = false
      ElMessage.success(t('at.common.saved'))
      await load()
    } finally {
      savingQuestionnaire.value = false
    }
  }
  function openLink(id: number) {
    selectedQuestionnaire.value = id
    createdURL.value = ''
    linkVisible.value = true
  }
  async function createLink() {
    const result = await recruitmentApi.createLink({
      questionnaire_id: selectedQuestionnaire.value,
      expires_at: linkForm.expires_at,
      max_submissions: linkForm.max_submissions
    })
    createdURL.value = `${location.origin}/#/recruitment/${result.token}`
    await navigator.clipboard.writeText(createdURL.value)
    ElMessage.success(t('at.recruitment.linkCopied'))
    await load()
  }
  async function toggleLink(id: number, active: boolean) {
    await recruitmentApi.setLinkActive(id, active)
    await load()
  }
  async function openApplication(id: number) {
    applicationDetail.value = await recruitmentApi.getApplication(id)
    reviewForm.status = applicationDetail.value.application.status
    reviewForm.reason = ''
    applicationVisible.value = true
  }
  function answerLabel(fieldKey: string) {
    return (
      applicationDetail.value?.fields.find((field) => field.field_key === fieldKey)?.label ??
      fieldKey
    )
  }
  async function reviewApplication() {
    if (!applicationDetail.value) return
    await recruitmentApi.reviewApplication(
      applicationDetail.value.application.id,
      reviewForm.status,
      reviewForm.reason
    )
    ElMessage.success(t('at.common.saved'))
    await openApplication(applicationDetail.value.application.id)
    await load()
  }
  async function privacyCleanup() {
    const reason = window.prompt(t('at.recruitment.cleanupReason'))
    if (!reason) return
    await ElMessageBox.confirm(
      t('at.recruitment.cleanupConfirm'),
      t('at.recruitment.privacyCleanup'),
      { type: 'warning' }
    )
    const result = await recruitmentApi.privacyCleanup(reason)
    ElMessage.success(t('at.recruitment.cleaned', { count: result.applications_cleaned }))
  }
  onMounted(load)
</script>

<style scoped>
  .designer-shell {
    display: grid;
    grid-template-columns: 230px minmax(0, 1fr);
    gap: 20px;
    min-height: 650px;
    max-height: 72vh;
  }

  .designer-sidebar {
    padding: 18px;
    overflow-y: auto;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-fill-color-light);
  }

  .designer-sidebar h3 {
    margin-bottom: 8px;
    font-weight: 600;
  }
  .designer-hint {
    margin-bottom: 12px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.5;
  }
  .system-field-list {
    display: flex;
    flex-direction: column;
    gap: 9px;
  }
  .system-field-item {
    display: flex;
    align-items: center;
    gap: 7px;
    color: var(--el-text-color-regular);
    font-size: 13px;
  }
  .system-field-item :deep(svg) {
    color: var(--el-color-success);
  }
  .question-type-list {
    display: grid;
    grid-template-columns: 1fr;
    gap: 8px;
  }
  .question-type-list .el-button {
    justify-content: flex-start;
    margin-left: 0;
  }
  .designer-canvas {
    overflow-y: auto;
    padding: 2px 8px 20px 2px;
  }
  .questionnaire-heading {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 20px;
    margin-bottom: 16px;
    border-radius: 12px;
    background: var(--el-fill-color-light);
  }
  .question-card {
    padding: 20px;
    margin-bottom: 14px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-bg-color);
    transition:
      border-color 0.2s,
      box-shadow 0.2s;
  }
  .question-card:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 8px 20px rgb(0 0 0 / 5%);
  }
  .question-card-header {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .question-number {
    color: var(--el-color-primary);
    font-size: 18px;
    font-weight: 700;
  }
  .question-title-input {
    flex: 1;
  }
  .question-type-select {
    width: 150px;
  }
  .question-preview {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 16px;
    margin-top: 14px;
    border-radius: 9px;
    background: var(--el-fill-color-lighter);
  }
  .option-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .question-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 13px;
    margin-top: 13px;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  @media (max-width: 900px) {
    .designer-shell {
      grid-template-columns: 1fr;
      max-height: none;
    }
    .designer-sidebar {
      max-height: 260px;
    }
    .question-card-header {
      align-items: stretch;
      flex-wrap: wrap;
    }
    .question-title-input {
      min-width: calc(100% - 48px);
    }
    .question-type-select {
      width: 100%;
    }
  }
</style>
