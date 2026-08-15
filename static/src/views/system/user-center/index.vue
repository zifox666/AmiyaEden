<template>
  <div class="art-card p-6 max-w-3xl mx-auto">
    <div class="mb-6">
      <h1 class="text-xl font-medium">{{ t('atProfile.title') }}</h1>
      <p class="mt-2 text-sm text-g-500">{{ t('atProfile.description') }}</p>
    </div>

    <ElSkeleton :loading="loading" animated :rows="7">
      <ElForm ref="formRef" :model="form" :rules="rules" label-position="top">
        <ElAlert class="mb-5" type="info" :closable="false">
          <template #title>
            {{ t('atProfile.mainCharacter') }}:
            {{ profile?.main_character_name || '-' }}
            <span v-if="profile?.main_character_id">({{ profile.main_character_id }})</span>
            · {{ t(`atProfile.source.${profile?.main_character_source || 'none'}`) }}
          </template>
        </ElAlert>

        <ElFormItem :label="t('atProfile.primaryCharacter')">
          <ElSelect
            :model-value="profile?.main_character_id || undefined"
            class="w-full"
            :disabled="profile?.main_character_source === 'seat' || characters.length < 2"
            :loading="switchingPrimary"
            @change="switchPrimary"
          >
            <ElOption
              v-for="character in characters"
              :key="character.character_id"
              :label="`${character.character_name} (${character.character_id})`"
              :value="character.character_id"
            />
          </ElSelect>
          <div class="mt-2 text-xs text-g-500">
            {{
              profile?.main_character_source === 'seat'
                ? t('atProfile.seatPrimaryHint')
                : t('atProfile.primaryHint')
            }}
          </div>
        </ElFormItem>

        <ElRow :gutter="20">
          <ElCol :xs="24" :sm="12">
            <ElFormItem :label="t('atProfile.nickname')" prop="nickname">
              <ElInput v-model="form.nickname" maxlength="128" show-word-limit />
            </ElFormItem>
          </ElCol>
          <ElCol :xs="24" :sm="12">
            <ElFormItem :label="t('atProfile.qq')" prop="qq">
              <ElInput v-model="form.qq" inputmode="numeric" />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="20">
          <ElCol :xs="24" :sm="12">
            <ElFormItem :label="t('atProfile.tournamentName')" prop="tournament_character_name">
              <ElInput v-model="form.tournament_character_name" maxlength="128" />
            </ElFormItem>
          </ElCol>
          <ElCol :xs="24" :sm="12">
            <ElFormItem :label="t('atProfile.tournamentId')" prop="tournament_character_id">
              <ElInputNumber
                v-model="form.tournament_character_id"
                :min="1"
                :precision="0"
                controls-position="right"
                class="!w-full"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="20">
          <ElCol :xs="24" :sm="12">
            <ElFormItem :label="t('atProfile.uiLanguage')" prop="ui_language">
              <ElSelect v-model="form.ui_language" class="w-full">
                <ElOption :label="t('atProfile.languages.zh')" value="zh" />
                <ElOption :label="t('atProfile.languages.en')" value="en" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :xs="24" :sm="12">
            <ElFormItem :label="t('atProfile.gameLanguage')" prop="game_language">
              <ElSelect v-model="form.game_language" class="w-full">
                <ElOption :label="t('atProfile.languages.zh')" value="zh" />
                <ElOption :label="t('atProfile.languages.en')" value="en" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElAlert class="mb-5" type="warning" :closable="false" :title="t('atProfile.eftEnglish')" />
        <ElButton type="primary" :loading="saving" @click="save">{{
          t('atProfile.save')
        }}</ElButton>
      </ElForm>
    </ElSkeleton>
  </div>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { fetchProfile, updateProfile } from '@/api/profile'
  import { fetchGetUserInfo, fetchMyCharacters, setPrimaryCharacter } from '@/api/auth'
  import { useUserStore } from '@/store/modules/user'
  import { LanguageEnum } from '@/enums/appEnum'

  defineOptions({ name: 'UserCenter' })

  const { t, locale } = useI18n()
  const userStore = useUserStore()
  const loading = ref(true)
  const saving = ref(false)
  const switchingPrimary = ref(false)
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const formRef = ref<FormInstance>()
  const profile = ref<Api.Profile.Profile>()
  const form = reactive<Api.Profile.UpdateRequest>({
    qq: '',
    nickname: '',
    tournament_character_name: '',
    tournament_character_id: null,
    ui_language: 'zh',
    game_language: 'en'
  })
  const rules = computed<FormRules>(() => ({
    qq: [
      { required: true, message: t('atProfile.validation.qqRequired'), trigger: 'blur' },
      { pattern: /^[0-9]{5,20}$/, message: t('atProfile.validation.qqFormat'), trigger: 'blur' }
    ],
    nickname: [{ required: true, message: t('atProfile.validation.nickname'), trigger: 'blur' }],
    tournament_character_name: [
      { required: true, message: t('atProfile.validation.tournamentName'), trigger: 'blur' }
    ],
    tournament_character_id: [
      { required: true, message: t('atProfile.validation.tournamentId'), trigger: 'change' }
    ]
  }))

  async function load() {
    loading.value = true
    try {
      const [data, boundCharacters] = await Promise.all([fetchProfile(), fetchMyCharacters()])
      profile.value = data
      characters.value = boundCharacters
      Object.assign(form, {
        qq: data.qq,
        nickname: data.nickname,
        tournament_character_name: data.tournament_character_name,
        tournament_character_id: data.tournament_character_id,
        ui_language: data.ui_language,
        game_language: data.game_language
      })
    } finally {
      loading.value = false
    }
  }

  async function switchPrimary(characterId: number) {
    if (!characterId || characterId === profile.value?.main_character_id) return
    switchingPrimary.value = true
    try {
      await setPrimaryCharacter(characterId)
      userStore.setUserInfo(await fetchGetUserInfo())
      await load()
      ElMessage.success(t('atProfile.primaryUpdated'))
    } finally {
      switchingPrimary.value = false
    }
  }

  async function save() {
    if (!(await formRef.value?.validate().catch(() => false))) return
    saving.value = true
    try {
      profile.value = await updateProfile({ ...form })
      userStore.setLanguage(form.ui_language as LanguageEnum)
      userStore.setGameLanguage(form.game_language as LanguageEnum)
      locale.value = form.ui_language
      ElMessage.success(t('atProfile.saved'))
    } finally {
      saving.value = false
    }
  }

  onMounted(load)
</script>
