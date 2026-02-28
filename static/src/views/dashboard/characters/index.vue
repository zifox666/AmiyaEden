<!-- EVE 角色管理页面 —— 绑定/解绑角色、切换主角色 -->
<template>
  <div class="w-full h-full p-0 bg-transparent border-none shadow-none">
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

      <!-- 角色列表 -->
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
          <!-- 主角色徽标 -->
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
              char.token_valid ? 'grayscale opacity-50' : ''
            ]"
          />

          <!-- 信息 -->
          <div class="flex-1 min-w-0">
            <h3 class="text-base font-medium truncate">
              {{ char.character_name }}
              <span v-if="char.token_valid">(已失效)</span>
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
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <ElEmpty v-if="!loading && characters.length === 0" :description="$t('characters.empty')" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import {
    fetchMyCharacters,
    getEveBindURL,
    setPrimaryCharacter,
    unbindCharacter
  } from '@/api/auth'
  import { fetchGetUserInfo } from '@/api/auth'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'Characters' })

  const { t } = useI18n()
  const userStore = useUserStore()

  const loading = ref(false)
  const bindLoading = ref(false)
  const switchingId = ref<number | null>(null)
  const unbindingId = ref<number | null>(null)
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const primaryCharacterId = ref<number>(0)

  /** 判断是否为主角色 */
  const isPrimary = (char: Api.Auth.EveCharacter) => {
    return char.character_id === primaryCharacterId.value
  }

  /** Scope 简述 */
  const scopeSummary = (char: Api.Auth.EveCharacter) => {
    if (!char.scopes) return t('characters.noScopes')
    const list = char.scopes.split(' ').filter(Boolean)
    return `${list.length} ${t('characters.scopeCount')}`
  }

  /** 加载角色列表 */
  const loadCharacters = async () => {
    loading.value = true
    try {
      characters.value = await fetchMyCharacters()
      // 同步主角色 ID
      primaryCharacterId.value = userStore.getUserInfo.primaryCharacterId ?? 0
    } finally {
      loading.value = false
    }
  }

  /** 刷新用户信息（切换主角色后头像/昵称变化） */
  const refreshUserInfo = async () => {
    const info = await fetchGetUserInfo()
    userStore.setUserInfo(info)
    primaryCharacterId.value = info.primaryCharacterId ?? 0
  }

  /** 绑定新角色 */
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

  /** 设置主角色 */
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

  /** 解绑角色 */
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

  onMounted(loadCharacters)
</script>
