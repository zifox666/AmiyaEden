<template>
  <div class="application-page">
    <div class="application-shell">
      <ElSkeleton :loading="loading" :rows="10">
        <template v-if="context">
          <header class="application-header">
            <ElTag effect="dark" round>{{ t('at.recruitment.applyTitle') }}</ElTag>
            <h1>{{ context.title }}</h1>
            <p>{{ context.description }}</p>
          </header>

          <div class="application-steps">
            <div :class="['application-step', { completed: isLoggedIn }]">
              <span>1</span
              ><div
                ><strong>{{ t('at.recruitment.stepAccount') }}</strong
                ><small>{{ t('at.recruitment.stepAccountHint') }}</small></div
              >
            </div>
            <div :class="['application-step', { completed: profileComplete && skillsReady }]">
              <span>2</span
              ><div
                ><strong>{{ t('at.recruitment.stepIdentity') }}</strong
                ><small>{{ t('at.recruitment.stepIdentityHint') }}</small></div
              >
            </div>
            <div :class="['application-step', { completed: Boolean(submittedApplication) }]">
              <span>3</span
              ><div
                ><strong>{{ t('at.recruitment.stepQuestionnaire') }}</strong
                ><small>{{ t('at.recruitment.stepQuestionnaireHint') }}</small></div
              >
            </div>
          </div>

          <section v-if="!isLoggedIn" class="login-card">
            <ArtSvgIcon icon="ri:shield-user-line" class="login-icon" />
            <h2>{{ t('at.recruitment.loginFirstTitle') }}</h2>
            <p>{{ t('at.recruitment.loginFirstHint') }}</p>
            <ElButton type="primary" size="large" @click="goLogin">{{
              t('at.recruitment.loginOrRegister')
            }}</ElButton>
            <small>{{ t('at.recruitment.oauthReturnHint') }}</small>
          </section>

          <template v-else>
            <section class="system-section">
              <div class="section-title-row">
                <div
                  ><h2>{{ t('at.recruitment.systemIdentity') }}</h2
                  ><p>{{ t('at.recruitment.systemIdentityHint') }}</p></div
                >
                <ElTag :type="profileComplete && skillsReady ? 'success' : 'warning'">
                  {{
                    profileComplete && skillsReady
                      ? t('at.recruitment.identityReady')
                      : t('at.recruitment.identityIncomplete')
                  }}
                </ElTag>
              </div>

              <div class="identity-status-grid">
                <div class="identity-status"
                  ><ArtSvgIcon icon="ri:user-follow-line" /><span>{{
                    t('at.recruitment.oauthAccount')
                  }}</span
                  ><ElTag type="success" size="small">{{ t('at.common.ready') }}</ElTag></div
                >
                <div class="identity-status"
                  ><ArtSvgIcon icon="ri:spaceship-line" /><span>{{
                    t('at.recruitment.mainCharacter')
                  }}</span
                  ><ElTag :type="profile?.main_character_id ? 'success' : 'danger'" size="small">{{
                    profile?.main_character_name || t('at.recruitment.notSet')
                  }}</ElTag></div
                >
                <div class="identity-status"
                  ><ArtSvgIcon icon="ri:bar-chart-box-line" /><span>{{
                    t('at.recruitment.skills')
                  }}</span
                  ><ElTag :type="skillsReady ? 'success' : 'warning'" size="small">{{
                    skillsReady
                      ? t('at.recruitment.skillsReady')
                      : t('at.recruitment.skillsPending')
                  }}</ElTag></div
                >
              </div>

              <ElForm label-position="top" class="identity-form">
                <ElRow :gutter="16">
                  <ElCol :xs="24" :sm="12"
                    ><ElFormItem :label="t('atProfile.nickname')"
                      ><ElInput v-model="profileForm.nickname" /></ElFormItem
                  ></ElCol>
                  <ElCol :xs="24" :sm="12"
                    ><ElFormItem :label="t('atProfile.qq')"
                      ><ElInput v-model="profileForm.qq" inputmode="numeric" /></ElFormItem
                  ></ElCol>
                  <ElCol :xs="24" :sm="12"
                    ><ElFormItem :label="t('atProfile.tournamentName')"
                      ><ElInput v-model="profileForm.tournament_character_name" /></ElFormItem
                  ></ElCol>
                  <ElCol :xs="24" :sm="12"
                    ><ElFormItem :label="t('atProfile.tournamentId')"
                      ><ElInputNumber
                        v-model="profileForm.tournament_character_id"
                        :min="1"
                        :precision="0"
                        class="!w-full" /></ElFormItem
                  ></ElCol>
                  <ElCol :span="24"
                    ><ElFormItem :label="t('atProfile.primaryCharacter')">
                      <ElSelect
                        :model-value="profile?.main_character_id || undefined"
                        :disabled="
                          profile?.main_character_source === 'seat' || characters.length < 2
                        "
                        class="w-full"
                        @change="switchPrimary"
                      >
                        <ElOption
                          v-for="character in characters"
                          :key="character.character_id"
                          :label="`${character.character_name} (${character.character_id})`"
                          :value="character.character_id"
                        />
                      </ElSelect>
                      <p class="field-hint">{{
                        profile?.main_character_source === 'seat'
                          ? t('atProfile.seatPrimaryHint')
                          : t('atProfile.primaryHint')
                      }}</p>
                    </ElFormItem></ElCol
                  >
                </ElRow>
                <div class="flex gap-2">
                  <ElButton type="primary" :loading="savingProfile" @click="saveIdentity(true)">{{
                    t('at.recruitment.saveIdentity')
                  }}</ElButton>
                  <ElButton :loading="checkingSkills" @click="checkSkills(true)">{{
                    t('at.recruitment.checkSkills')
                  }}</ElButton>
                  <ElButton link @click="router.push('/profile')">{{
                    t('at.recruitment.openProfile')
                  }}</ElButton>
                </div>
              </ElForm>
            </section>

            <section class="questions-section">
              <div class="section-title-row"
                ><div
                  ><h2>{{ t('at.recruitment.customQuestions') }}</h2
                  ><p>{{ t('at.recruitment.customQuestionsHint') }}</p></div
                ></div
              >
              <ElForm label-position="top">
                <div
                  v-for="(field, index) in context.fields"
                  :key="field.field_key"
                  class="answer-card"
                >
                  <div class="answer-title"
                    ><span>{{ String(index + 1).padStart(2, '0') }}</span
                    ><strong>{{ field.label }}</strong
                    ><i v-if="field.required">*</i></div
                  >
                  <p v-if="field.description" class="answer-description">{{ field.description }}</p>
                  <ElInput
                    v-if="field.type === 'text'"
                    v-model="scalarAnswers[field.field_key]"
                    :placeholder="t('at.recruitment.answerPlaceholder')"
                  />
                  <ElInput
                    v-else-if="field.type === 'textarea'"
                    v-model="scalarAnswers[field.field_key]"
                    type="textarea"
                    :rows="4"
                    :placeholder="t('at.recruitment.answerPlaceholder')"
                  />
                  <ElInput
                    v-else-if="field.type === 'number'"
                    v-model="scalarAnswers[field.field_key]"
                    type="number"
                  />
                  <ElDatePicker
                    v-else-if="field.type === 'date'"
                    v-model="scalarAnswers[field.field_key]"
                    class="w-full"
                  />
                  <ElSelect
                    v-else-if="field.type === 'select'"
                    v-model="scalarAnswers[field.field_key]"
                    class="w-full"
                    ><ElOption
                      v-for="option in field.options ?? []"
                      :key="option"
                      :label="option"
                      :value="option"
                  /></ElSelect>
                  <ElRadioGroup
                    v-else-if="field.type === 'radio'"
                    v-model="scalarAnswers[field.field_key]"
                    class="choice-group"
                    ><ElRadio
                      v-for="option in field.options ?? []"
                      :key="option"
                      :value="option"
                      border
                      >{{ option }}</ElRadio
                    ></ElRadioGroup
                  >
                  <ElCheckboxGroup
                    v-else-if="field.type === 'checkbox'"
                    v-model="multipleAnswers[field.field_key]"
                    class="choice-group"
                    ><ElCheckbox
                      v-for="option in field.options ?? []"
                      :key="option"
                      :value="option"
                      border
                      >{{ option }}</ElCheckbox
                    ></ElCheckboxGroup
                  >
                  <div v-else-if="field.type === 'image'">
                    <ElUpload
                      :show-file-list="false"
                      accept="image/jpeg,image/png,image/gif,image/webp"
                      :http-request="({ file }) => uploadImage(field.field_key, file as File)"
                    >
                      <ElButton :loading="uploadingField === field.field_key">{{
                        t('at.recruitment.uploadImage')
                      }}</ElButton>
                    </ElUpload>
                    <ElImage
                      v-if="scalarAnswers[field.field_key]"
                      class="mt-3 rounded"
                      style="width: 180px; height: 120px"
                      fit="cover"
                      :src="scalarAnswers[field.field_key]"
                      :preview-src-list="[scalarAnswers[field.field_key]]"
                    />
                  </div>
                </div>
                <ElEmpty
                  v-if="context.fields.length === 0"
                  :description="t('at.recruitment.noCustomQuestions')"
                />
              </ElForm>
            </section>

            <section class="submit-section">
              <template v-if="!submittedApplication">
                <p>{{ t('at.recruitment.submitConsent') }}</p>
                <ElButton type="primary" size="large" :loading="submitting" @click="submit">{{
                  t('at.recruitment.submit')
                }}</ElButton>
              </template>
              <ElResult
                v-else
                icon="success"
                :title="t('at.recruitment.submitted')"
                :sub-title="t(`at.recruitment.status.${submittedApplication.status}`)"
              >
                <template #extra
                  ><ElButton
                    v-if="['submitted', 'reviewing'].includes(submittedApplication.status)"
                    type="danger"
                    plain
                    @click="withdraw"
                    >{{ t('at.recruitment.withdraw') }}</ElButton
                  ></template
                >
              </ElResult>
            </section>
          </template>
        </template>
      </ElSkeleton>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import { recruitmentApi } from '@/api/at'
  import { fetchProfile, updateProfile } from '@/api/profile'
  import { fetchMyCharacters, setPrimaryCharacter } from '@/api/auth'
  import { fetchInfoSkills } from '@/api/eve-info'
  import { useUserStore } from '@/store/modules/user'

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()
  const token = computed(() => String(route.params.token))
  const isLoggedIn = computed(() => userStore.isLogin && Boolean(userStore.accessToken))
  const context = ref<Api.AT.RecruitmentContext>()
  const profile = ref<Api.Profile.Profile>()
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const profileForm = reactive<Api.Profile.UpdateRequest>({
    qq: '',
    nickname: '',
    tournament_character_name: '',
    tournament_character_id: null,
    ui_language: 'zh',
    game_language: 'en'
  })
  const scalarAnswers = reactive<Record<string, string>>({})
  const multipleAnswers = reactive<Record<string, string[]>>({})
  const loading = ref(false)
  const submitting = ref(false)
  const savingProfile = ref(false)
  const checkingSkills = ref(false)
  const skillsReady = ref(false)
  const submittedApplication = ref<Api.AT.RecruitmentApplication>()
  const uploadingField = ref('')
  const profileComplete = computed(
    () =>
      /^\d{5,20}$/.test(profileForm.qq) &&
      Boolean(
        profileForm.nickname.trim() &&
        profileForm.tournament_character_name.trim() &&
        profileForm.tournament_character_id &&
        profile.value?.main_character_id
      )
  )

  function goLogin() {
    router.push({ path: '/auth/login', query: { redirect: route.fullPath } })
  }
  async function load() {
    loading.value = true
    try {
      context.value = await recruitmentApi.publicContext(token.value)
      for (const field of context.value.fields) {
        if (field.type === 'checkbox') multipleAnswers[field.field_key] = []
        else scalarAnswers[field.field_key] = ''
      }
      if (isLoggedIn.value) await loadIdentity()
    } finally {
      loading.value = false
    }
  }
  async function loadIdentity() {
    const [profileData, boundCharacters, applications] = await Promise.all([
      fetchProfile(),
      fetchMyCharacters(),
      recruitmentApi.listMyApplications()
    ])
    profile.value = profileData
    characters.value = boundCharacters
    Object.assign(profileForm, {
      qq: profileData.qq,
      nickname: profileData.nickname,
      tournament_character_name: profileData.tournament_character_name,
      tournament_character_id: profileData.tournament_character_id,
      ui_language: profileData.ui_language,
      game_language: profileData.game_language
    })
    submittedApplication.value = applications.find(
      (item) => item.link_id === context.value?.link_id
    )
    await checkSkills(false)
  }
  async function saveIdentity(showMessage = true) {
    savingProfile.value = true
    try {
      profile.value = await updateProfile({ ...profileForm })
      if (showMessage) ElMessage.success(t('at.recruitment.identitySaved'))
    } finally {
      savingProfile.value = false
    }
  }
  async function switchPrimary(characterId: number) {
    await setPrimaryCharacter(characterId)
    profile.value = await fetchProfile()
    await checkSkills(false)
    ElMessage.success(t('atProfile.primaryUpdated'))
  }
  async function checkSkills(showMessage = true) {
    if (!profile.value?.main_character_id) {
      skillsReady.value = false
      return
    }
    checkingSkills.value = true
    try {
      const result = await fetchInfoSkills({
        character_id: profile.value.main_character_id,
        language: userStore.gameLanguage || 'en'
      })
      skillsReady.value = result.skill_count > 0
      if (showMessage)
        ElMessage[skillsReady.value ? 'success' : 'warning'](
          t(skillsReady.value ? 'at.recruitment.skillsReady' : 'at.recruitment.skillsPending')
        )
    } catch {
      skillsReady.value = false
    } finally {
      checkingSkills.value = false
    }
  }
  async function submit() {
    if (!profileComplete.value) return ElMessage.warning(t('at.recruitment.completeIdentityFirst'))
    submitting.value = true
    try {
      await saveIdentity(false)
      if (!skillsReady.value) await checkSkills(false)
      if (!skillsReady.value) return ElMessage.warning(t('at.recruitment.skillsRequired'))
      submittedApplication.value = await recruitmentApi.submit(token.value, {
        ...scalarAnswers,
        ...multipleAnswers
      })
      ElMessage.success(t('at.recruitment.submitted'))
    } finally {
      submitting.value = false
    }
  }
  async function uploadImage(fieldKey: string, file: File) {
    uploadingField.value = fieldKey
    try {
      const result = await recruitmentApi.uploadImage(file)
      scalarAnswers[fieldKey] = result.url
      ElMessage.success(t('at.recruitment.imageUploaded'))
    } finally {
      uploadingField.value = ''
    }
  }
  async function withdraw() {
    if (!submittedApplication.value) return
    await recruitmentApi.withdraw(submittedApplication.value.id)
    submittedApplication.value.status = 'withdrawn'
    ElMessage.success(t('at.recruitment.withdrawn'))
  }
  onMounted(load)
</script>

<style scoped>
  .application-page {
    min-height: 100vh;
    padding: 36px 16px;
    background:
      radial-gradient(circle at top left, var(--el-color-primary-light-9), transparent 38%),
      var(--el-bg-color-page);
  }
  .application-shell {
    max-width: 920px;
    margin: 0 auto;
    padding: 30px;
    border-radius: 18px;
    background: var(--el-bg-color);
    box-shadow: 0 18px 50px rgb(0 0 0 / 8%);
  }
  .application-header {
    padding: 14px 8px 28px;
    text-align: center;
  }
  .application-header h1 {
    margin: 14px 0 8px;
    font-size: 30px;
    font-weight: 700;
  }
  .application-header p,
  .section-title-row p {
    color: var(--el-text-color-secondary);
  }
  .application-steps {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
    margin-bottom: 24px;
  }
  .application-step {
    display: flex;
    gap: 10px;
    padding: 13px;
    border-radius: 10px;
    background: var(--el-fill-color-light);
  }
  .application-step > span {
    display: grid;
    width: 28px;
    height: 28px;
    place-items: center;
    flex: none;
    border-radius: 50%;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color);
  }
  .application-step.completed > span {
    color: white;
    background: var(--el-color-success);
  }
  .application-step div {
    display: flex;
    flex-direction: column;
  }
  .application-step small {
    margin-top: 3px;
    color: var(--el-text-color-secondary);
  }
  .login-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 56px 24px;
    text-align: center;
    border: 1px dashed var(--el-border-color);
    border-radius: 14px;
  }
  .login-icon {
    font-size: 52px;
    color: var(--el-color-primary);
  }
  .login-card h2,
  .section-title-row h2 {
    font-size: 20px;
    font-weight: 650;
  }
  .login-card p,
  .login-card small {
    max-width: 580px;
    color: var(--el-text-color-secondary);
  }
  .system-section,
  .questions-section,
  .submit-section {
    padding: 24px;
    margin-top: 18px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 14px;
  }
  .section-title-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 20px;
  }
  .identity-status-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
    margin-bottom: 22px;
  }
  .identity-status {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px;
    border-radius: 9px;
    background: var(--el-fill-color-light);
  }
  .identity-status span {
    flex: 1;
    font-size: 13px;
  }
  .field-hint {
    margin-top: 5px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
  .answer-card {
    padding: 20px;
    margin-bottom: 14px;
    border-radius: 12px;
    background: var(--el-fill-color-light);
  }
  .answer-title {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
  }
  .answer-title span {
    color: var(--el-color-primary);
    font-weight: 700;
  }
  .answer-title i {
    color: var(--el-color-danger);
    font-style: normal;
  }
  .answer-description {
    margin: -4px 0 13px 28px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
  .choice-group {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  .choice-group :deep(.el-radio),
  .choice-group :deep(.el-checkbox) {
    width: 100%;
    height: auto;
    min-height: 40px;
    margin-right: 0;
  }
  .submit-section {
    text-align: center;
  }
  .submit-section > p {
    margin-bottom: 14px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
  @media (max-width: 720px) {
    .application-shell {
      padding: 18px;
    }
    .application-steps,
    .identity-status-grid {
      grid-template-columns: 1fr;
    }
    .application-header h1 {
      font-size: 24px;
    }
  }
</style>
