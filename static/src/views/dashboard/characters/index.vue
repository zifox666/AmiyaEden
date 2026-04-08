<!-- EVE 人物管理页面 —— 绑定/解绑人物、切换主人物 -->
<template>
  <div class="w-full h-full p-0 bg-transparent border-none shadow-none">
    <div class="art-card-sm p-6 mb-4">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="max-w-2xl">
          <h2 class="text-lg font-medium">{{ $t('characters.profile.title') }}</h2>
          <p class="mt-1 text-sm text-g-500">{{ $t('characters.profile.subTitle') }}</p>
        </div>
        <ElTag :type="profileComplete ? 'success' : 'warning'" effect="light" round>
          {{
            profileComplete
              ? $t('characters.profile.completed')
              : $t('characters.profile.incomplete')
          }}
        </ElTag>
      </div>

      <ElAlert
        :type="profileComplete ? 'success' : 'warning'"
        :closable="false"
        class="mt-4"
        :title="
          profileComplete
            ? $t('characters.profile.completedHint')
            : $t('characters.profile.requiredHint')
        "
      />

      <ElForm
        ref="profileFormRef"
        :model="profileForm"
        :rules="profileRules"
        label-position="top"
        class="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3"
      >
        <ElFormItem :label="$t('characters.profile.nickname')" prop="nickname">
          <ElInput
            v-model="profileForm.nickname"
            :maxlength="20"
            clearable
            show-word-limit
            :placeholder="$t('characters.profile.nicknamePlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('characters.profile.qq')" prop="qq">
          <ElInput
            v-model="profileForm.qq"
            :maxlength="20"
            clearable
            show-word-limit
            :placeholder="$t('characters.profile.qqPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('characters.profile.discordId')" prop="discordId">
          <ElInput
            v-model="profileForm.discordId"
            :maxlength="20"
            clearable
            show-word-limit
            :placeholder="$t('characters.profile.discordPlaceholder')"
          />
        </ElFormItem>
      </ElForm>

      <div class="flex justify-end mt-2">
        <ElButton type="primary" :loading="profileSaving" @click="handleSaveProfile">
          {{ $t('characters.profile.save') }}
        </ElButton>
      </div>
    </div>

    <div class="art-card-sm p-6">
      <!-- 页头 -->
      <div class="flex items-center justify-between mb-6">
        <div>
          <h1 class="text-xl font-medium">{{ $t('characters.title') }}</h1>
          <p class="mt-1 text-sm text-g-500">{{ $t('characters.subTitle') }}</p>
        </div>
        <ElButton type="primary" :loading="bindLoading" @click="handleBind">
          <ArtSvgIcon icon="ri:add-line" class="mr-1" />
          {{ $t('characters.bind') }}
        </ElButton>
      </div>

      <ElAlert
        v-if="showTokenHealthAlert"
        type="error"
        :closable="false"
        show-icon
        class="mb-4"
        :title="$t('characters.tokenHealth.title')"
        :description="$t('characters.tokenHealth.requiredHint')"
      />

      <!-- 人物列表 -->
      <div v-loading="loading" class="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="char in characters"
          :key="char.character_id"
          class="relative flex items-center gap-4 p-4 rounded-lg border transition-all"
          :class="
            isPrimary(char)
              ? 'border-primary bg-primary/5 shadow-sm'
              : 'border-g-300 hover:border-g-400'
          "
        >
          <!-- 主人物徽标 -->
          <div
            v-if="isPrimary(char)"
            class="absolute -top-2 -right-2 flex items-center gap-0.5 px-2 py-0.5 text-xs font-medium text-white bg-primary rounded-full shadow"
          >
            <ArtSvgIcon icon="ri:star-fill" :size="12" />
            {{ $t('characters.primary') }}
          </div>

          <!-- 头像 -->
          <img
            :src="char.portrait_url"
            :alt="char.character_name"
            class="w-14 h-14 rounded-full object-cover border-2"
            :class="[
              isPrimary(char) ? 'border-primary' : 'border-g-200',
              char.token_invalid ? 'border-red-500 grayscale-0 opacity-100' : ''
            ]"
          />

          <!-- 信息 -->
          <div class="flex-1 min-w-0">
            <h3 class="text-base font-medium truncate">
              {{ char.character_name }}
              <span v-if="char.token_invalid">{{ t('characters.tokenInvalid') }}</span>
            </h3>
            <p class="mt-0.5 text-xs text-g-500">ID: {{ char.character_id }}</p>
            <p class="mt-0.5 text-xs text-g-400 truncate" :title="char.scopes">
              {{ scopeSummary(char) }}
            </p>
          </div>

          <!-- 操作按钮 -->
          <div class="flex flex-col gap-1.5">
            <ElButton
              v-if="!isPrimary(char)"
              size="small"
              type="primary"
              plain
              :loading="switchingId === char.character_id"
              @click="handleSetPrimary(char)"
            >
              {{ $t('characters.setPrimary') }}
            </ElButton>
            <ElButton
              v-if="characters.length > 1"
              size="small"
              type="danger"
              plain
              :loading="unbindingId === char.character_id"
              @click="handleUnbind(char)"
            >
              {{ $t('characters.unbind') }}
            </ElButton>
            <!-- 军团KM监控（管理员可见） -->
            <template v-if="canManageCorpKm">
              <ElTag v-if="hasCorpKmScope(char)" type="success" size="small" effect="light">
                {{ $t('characters.corpKm.enabled') }}
              </ElTag>
              <ElTooltip v-else :content="$t('characters.corpKm.tooltip')" placement="top">
                <ElButton size="small" type="warning" plain @click="handleEnableCorpKm">
                  {{ $t('characters.corpKm.enable') }}
                </ElButton>
              </ElTooltip>
            </template>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <ElEmpty v-if="!loading && characters.length === 0" :description="$t('characters.empty')" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import {
    fetchMyCharacters,
    getEveBindURL,
    hasInvalidCharacterToken as hasInvalidCharacterTokenInList,
    hasInvalidPrimaryCharacterToken as hasInvalidPrimaryCharacterTokenInList,
    isUserProfileComplete,
    setPrimaryCharacter,
    updateMyProfile,
    unbindCharacter
  } from '@/api/auth'
  import { fetchGetUserInfo } from '@/api/auth'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'Characters' })

  const { t } = useI18n()
  const userStore = useUserStore()

  const CORP_KM_SCOPE = 'esi-killmails.read_corporation_killmails.v1'

  const canManageCorpKm = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((r) => ['super_admin', 'admin'].includes(r))
  })

  const hasCorpKmScope = (char: Api.Auth.EveCharacter) =>
    char.scopes?.split(' ').includes(CORP_KM_SCOPE) ?? false

  const loading = ref(false)
  const bindLoading = ref(false)
  const profileSaving = ref(false)
  const switchingId = ref<number | null>(null)
  const unbindingId = ref<number | null>(null)
  const profileFormRef = ref<FormInstance>()
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const primaryCharacterId = ref<number>(0)
  const profileForm = reactive({
    nickname: '',
    qq: '',
    discordId: ''
  })
  const profileComplete = computed(() => isUserProfileComplete(userStore.getUserInfo))
  const enforceCharacterESIRestriction = computed(
    () => userStore.getUserInfo.enforceCharacterESIRestriction !== false
  )
  const hasInvalidCharacterToken = computed(() => hasInvalidCharacterTokenInList(characters.value))
  const hasInvalidPrimaryCharacterToken = computed(() =>
    hasInvalidPrimaryCharacterTokenInList(primaryCharacterId.value, characters.value)
  )
  const showTokenHealthAlert = computed(
    () =>
      hasInvalidPrimaryCharacterToken.value ||
      (enforceCharacterESIRestriction.value && hasInvalidCharacterToken.value)
  )

  const getTextLength = (value: string) => Array.from(value.trim()).length

  const profileRules = computed<FormRules>(() => ({
    nickname: [
      {
        validator: (_rule, value: string, callback: (error?: Error) => void) => {
          const nickname = value.trim()
          if (!nickname) {
            callback(new Error(t('characters.profile.validation.nicknameRequired')))
            return
          }
          if (getTextLength(nickname) > 20) {
            callback(new Error(t('characters.profile.validation.nicknameLength')))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    qq: [
      {
        validator: (_rule, value: string, callback: (error?: Error) => void) => {
          const qq = value.trim()
          const discordId = profileForm.discordId.trim()
          if (getTextLength(qq) > 20) {
            callback(new Error(t('characters.profile.validation.qqLength')))
            return
          }
          if (qq && !/^\d+$/.test(qq)) {
            callback(new Error(t('characters.profile.validation.qqDigits')))
            return
          }
          if (!qq && !discordId) {
            callback(new Error(t('characters.profile.validation.contactRequired')))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    discordId: [
      {
        validator: (_rule, value: string, callback: (error?: Error) => void) => {
          const discordId = value.trim()
          const qq = profileForm.qq.trim()
          if (getTextLength(discordId) > 20) {
            callback(new Error(t('characters.profile.validation.discordLength')))
            return
          }
          if (!discordId && !qq) {
            callback(new Error(t('characters.profile.validation.contactRequired')))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }))

  const syncProfileForm = () => {
    profileForm.nickname = userStore.getUserInfo.nickname ?? ''
    profileForm.qq = userStore.getUserInfo.qq ?? ''
    profileForm.discordId = userStore.getUserInfo.discordId ?? ''
  }

  /** 判断是否为主人物 */
  const isPrimary = (char: Api.Auth.EveCharacter) => {
    return char.character_id === primaryCharacterId.value
  }

  /** Scope 简述 */
  const scopeSummary = (char: Api.Auth.EveCharacter) => {
    if (!char.scopes) return t('characters.noScopes')
    const list = char.scopes.split(' ').filter(Boolean)
    return `${list.length} ${t('characters.scopeCount')}`
  }

  /** 加载人物列表 */
  const loadCharacters = async () => {
    loading.value = true
    try {
      characters.value = await fetchMyCharacters()
      // 同步主人物 ID
      primaryCharacterId.value = userStore.getUserInfo.primaryCharacterId ?? 0
    } finally {
      loading.value = false
    }
  }

  /** 刷新用户信息（切换主人物后头像/昵称变化） */
  const refreshUserInfo = async () => {
    const info = await fetchGetUserInfo()
    userStore.setUserInfo(info)
    primaryCharacterId.value = info.primaryCharacterId ?? 0
    syncProfileForm()
  }

  const handleSaveProfile = async () => {
    await profileFormRef.value?.validate()

    profileSaving.value = true
    try {
      await updateMyProfile({
        nickname: profileForm.nickname.trim(),
        qq: profileForm.qq.trim(),
        discord_id: profileForm.discordId.trim()
      })
      await refreshUserInfo()
      ElMessage.success(t('characters.profile.saveSuccess'))
    } finally {
      profileSaving.value = false
    }
  }

  /** 绑定新人物 */
  const handleBind = async () => {
    bindLoading.value = true
    try {
      const url = await getEveBindURL()
      window.location.href = url
    } catch {
      bindLoading.value = false
      ElMessage.error(t('characters.bindFailed'))
    }
  }

  /** 启用军团KM监控 */
  const handleEnableCorpKm = async () => {
    try {
      const url = await getEveBindURL([CORP_KM_SCOPE])
      window.location.href = url
    } catch {
      ElMessage.error(t('characters.corpKm.enableFailed'))
    }
  }

  /** 设置主人物 */
  const handleSetPrimary = async (char: Api.Auth.EveCharacter) => {
    switchingId.value = char.character_id
    try {
      await setPrimaryCharacter(char.character_id)
      ElMessage.success(t('characters.setPrimarySuccess', { name: char.character_name }))
      await refreshUserInfo()
      await loadCharacters()
    } catch {
      ElMessage.error(t('characters.setPrimaryFailed'))
    } finally {
      switchingId.value = null
    }
  }

  /** 解绑人物 */
  const handleUnbind = async (char: Api.Auth.EveCharacter) => {
    try {
      await ElMessageBox.confirm(
        t('characters.unbindConfirm', { name: char.character_name }),
        t('common.tips'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )
    } catch {
      return // 用户取消
    }

    unbindingId.value = char.character_id
    try {
      await unbindCharacter(char.character_id)
      ElMessage.success(t('characters.unbindSuccess', { name: char.character_name }))
      await refreshUserInfo()
      await loadCharacters()
    } catch {
      ElMessage.error(t('characters.unbindFailed'))
    } finally {
      unbindingId.value = null
    }
  }

  watch(
    () => [
      userStore.getUserInfo.nickname,
      userStore.getUserInfo.qq,
      userStore.getUserInfo.discordId
    ],
    () => {
      syncProfileForm()
    },
    { immediate: true }
  )

  onMounted(loadCharacters)
</script>
