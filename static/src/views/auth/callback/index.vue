<!-- EVE Online SSO 回调页面 - 处理 OAuth 跳转回来后的 token -->
<template>
  <div class="callback-page">
    <div class="callback-card">
      <template v-if="status === 'loading'">
        <div class="icon-wrap spin">
          <el-icon :size="48"><Loading /></el-icon>
        </div>
        <p class="title">{{ $t('login.callback.loading') }}</p>
        <p class="sub">{{ $t('login.callback.loadingSub') }}</p>
      </template>

      <template v-else-if="status === 'transfer'">
        <div class="icon-wrap warn">
          <el-icon :size="48"><WarningFilled /></el-icon>
        </div>
        <p class="title">{{ $t('login.callback.transferTitle') }}</p>
        <p class="sub">{{ $t('login.callback.transferSub', { name: pendingCharacterName }) }}</p>
        <div class="flex gap-3 mt-6">
          <ElButton :loading="transferLoading" type="primary" @click="handleConfirmTransfer">
            {{ $t('login.callback.transferConfirm') }}
          </ElButton>
          <ElButton @click="handleCancelTransfer">
            {{ $t('common.cancel') }}
          </ElButton>
        </div>
      </template>

      <template v-else-if="status === 'success'">
        <div class="icon-wrap success">
          <el-icon :size="48"><CircleCheckFilled /></el-icon>
        </div>
        <p class="title">{{ $t('login.callback.success') }}</p>
        <p class="sub">{{ $t('login.callback.successSub', { name: userName }) }}</p>
      </template>

      <template v-else-if="status === 'error'">
        <div class="icon-wrap error">
          <el-icon :size="48"><CircleCloseFilled /></el-icon>
        </div>
        <p class="title">{{ $t('login.callback.error') }}</p>
        <p class="sub error-msg">{{ errMsg }}</p>
        <ElButton type="primary" class="mt-6" @click="goLogin">
          {{ $t('login.callback.retry') }}
        </ElButton>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
  import {
    Loading,
    CircleCheckFilled,
    CircleCloseFilled,
    WarningFilled
  } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import { useUserStore } from '@/store/modules/user'
  import { fetchGetUserInfo, confirmCharacterTransfer } from '@/api/auth'

  defineOptions({ name: 'AuthCallback' })

  type Status = 'loading' | 'transfer' | 'success' | 'error'

  const status = ref<Status>('loading')
  const errMsg = ref('')
  const userName = ref('')
  const pendingCharacterName = ref('')
  const pendingTransferToken = ref('')
  const transferLoading = ref(false)
  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const goLogin = () => router.replace({ name: 'Login' })

  /** 完成登录流程（存 token、拉用户信息、跳转） */
  const finishLogin = async (token: string) => {
    userStore.setToken(token)
    userStore.setLoginStatus(true)

    const userInfo = await fetchGetUserInfo()
    userStore.setUserInfo(userInfo)
    userStore.checkAndClearWorktabs()

    userName.value = userInfo.userName
    status.value = 'success'

    const redirect = (route.query.redirect as string) || '/'
    setTimeout(() => {
      router.replace(redirect)
    }, 1200)
  }

  /** 用户确认迁移 */
  const handleConfirmTransfer = async () => {
    transferLoading.value = true
    try {
      const result = await confirmCharacterTransfer(pendingTransferToken.value)
      // 用返回的新 token 完成登录
      await finishLogin(result.token)
    } catch (err: any) {
      status.value = 'error'
      errMsg.value = err?.message ?? t('login.callback.transferFailed')
    } finally {
      transferLoading.value = false
    }
  }

  /** 用户取消迁移，跳回角色管理页 */
  const handleCancelTransfer = () => {
    router.replace('/dashboard/characters')
  }

  onMounted(async () => {
    const token = route.query.token as string
    const error = route.query.error as string
    const pendingTransfer = route.query.pending_transfer as string
    const characterName = route.query.character_name as string

    // EVE 端拒绝授权
    if (error) {
      status.value = 'error'
      errMsg.value = (route.query.error_description as string) || error
      return
    }

    // 角色已绑定到其他账号，需要用户确认迁移
    if (pendingTransfer) {
      pendingTransferToken.value = pendingTransfer
      pendingCharacterName.value = characterName || ''
      status.value = 'transfer'
      return
    }

    if (!token) {
      status.value = 'error'
      errMsg.value = t('authCallback.missingToken')
      return
    }

    try {
      await finishLogin(token)
    } catch (err: any) {
      userStore.setToken('')
      userStore.setLoginStatus(false)
      status.value = 'error'
      errMsg.value = err?.message ?? t('authCallback.verifyFailed')
    }
  })
</script>

<style lang="scss" scoped>
  .callback-page {
    @apply flex items-center justify-center w-full h-screen bg-default;
  }

  .callback-card {
    @apply flex flex-col items-center gap-3 p-12 rounded-2xl shadow-lg bg-default-box
           w-[420px] max-w-[90vw];

    .icon-wrap {
      @apply mb-2;

      &.spin :deep(.el-icon) {
        animation: spin 1s linear infinite;
      }

      &.success :deep(.el-icon) {
        color: var(--el-color-success);
      }

      &.error :deep(.el-icon) {
        color: var(--el-color-danger);
      }

      &.warn :deep(.el-icon) {
        color: var(--el-color-warning);
      }
    }

    .title {
      @apply text-xl font-semibold text-g-800;
    }

    .sub {
      @apply text-sm text-g-500 text-center;
    }

    .error-msg {
      @apply text-red-400;
    }
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
