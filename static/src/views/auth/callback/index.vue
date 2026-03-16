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

      <template v-else-if="status === 'success'">
        <div class="icon-wrap success">
          <el-icon :size="48"><CircleCheckFilled /></el-icon>
        </div>
        <p class="title">{{ $t('login.callback.success') }}</p>
        <p class="sub">{{ $t('login.callback.successSub', { name: userName }) }}</p>
      </template>

      <template v-else-if="status === 'conflict'">
        <div class="icon-wrap warning">
          <el-icon :size="48"><WarningFilled /></el-icon>
        </div>
        <p class="title">{{ $t('login.callback.conflictTitle') }}</p>
        <p class="sub">{{ $t('login.callback.conflictDesc', { name: conflictCharName }) }}</p>
        <div class="flex gap-3 mt-6">
          <ElButton @click="goLogin">{{ $t('common.cancel') }}</ElButton>
          <ElButton type="primary" :loading="transferring" @click="doConfirmTransfer">
            {{ $t('login.callback.confirmTransfer') }}
          </ElButton>
        </div>
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
  import { Loading, CircleCheckFilled, CircleCloseFilled, WarningFilled } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import { useUserStore } from '@/store/modules/user'
  import { fetchGetUserInfo, confirmCharacterTransfer } from '@/api/auth'

  defineOptions({ name: 'AuthCallback' })

  type Status = 'loading' | 'success' | 'error' | 'conflict'

  const status = ref<Status>('loading')
  const errMsg = ref('')
  const userName = ref('')
  const conflictCharName = ref('')
  const transferToken = ref('')
  const transferring = ref(false)
  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const goLogin = () => router.replace({ name: 'Login' })

  /** 处理正常登录成功后的流程 */
  async function handleLoginSuccess(token: string) {
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

  /** 确认角色转移 */
  async function doConfirmTransfer() {
    transferring.value = true
    try {
      const data = await confirmCharacterTransfer(transferToken.value)
      await handleLoginSuccess(data.token)
    } catch (err: any) {
      status.value = 'error'
      errMsg.value = err?.message ?? t('authCallback.verifyFailed')
    } finally {
      transferring.value = false
    }
  }

  onMounted(async () => {
    const token = route.query.token as string
    const error = route.query.error as string
    const conflict = route.query.conflict as string

    // 角色转移冲突：显示确认弹窗
    if (conflict === 'true') {
      conflictCharName.value = (route.query.character_name as string) || ''
      transferToken.value = (route.query.transfer_token as string) || ''
      status.value = 'conflict'
      return
    }

    // EVE 端拒绝授权
    if (error) {
      status.value = 'error'
      errMsg.value = (route.query.error_description as string) || error
      return
    }

    if (!token) {
      status.value = 'error'
      errMsg.value = t('authCallback.missingToken')
      return
    }

    try {
      await handleLoginSuccess(token)
    } catch (err: any) {
      // token 无效或请求失败
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
           w-[380px] max-w-[90vw];

    .icon-wrap {
      @apply mb-2;

      &.spin :deep(.el-icon) {
        animation: spin 1s linear infinite;
      }

      &.success :deep(.el-icon) {
        color: var(--el-color-success);
      }

      &.warning :deep(.el-icon) {
        color: var(--el-color-warning);
      }

      &.error :deep(.el-icon) {
        color: var(--el-color-danger);
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
