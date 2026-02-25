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
  import { Loading, CircleCheckFilled, CircleCloseFilled } from '@element-plus/icons-vue'
  import { useUserStore } from '@/store/modules/user'
  import { fetchGetUserInfo } from '@/api/auth'

  defineOptions({ name: 'AuthCallback' })

  type Status = 'loading' | 'success' | 'error'

  const status = ref<Status>('loading')
  const errMsg = ref('')
  const userName = ref('')
  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const goLogin = () => router.replace({ name: 'Login' })

  onMounted(async () => {
    const token = route.query.token as string
    const error = route.query.error as string

    // EVE 端拒绝授权
    if (error) {
      status.value = 'error'
      errMsg.value = (route.query.error_description as string) || error
      return
    }

    if (!token) {
      status.value = 'error'
      errMsg.value = '未收到登录令牌，请重试'
      return
    }

    try {
      // 1. 存储 token 并设置登录状态
      userStore.setToken(token)
      userStore.setLoginStatus(true)

      // 2. 拉取用户信息
      const userInfo = await fetchGetUserInfo()
      userStore.setUserInfo(userInfo)
      userStore.checkAndClearWorktabs()

      userName.value = userInfo.userName
      status.value = 'success'

      // 3. 短暂停留后跳转
      const redirect = (route.query.redirect as string) || '/'
      setTimeout(() => {
        router.replace(redirect)
      }, 1200)
    } catch (err: any) {
      // token 无效或请求失败
      userStore.setToken('')
      userStore.setLoginStatus(false)
      status.value = 'error'
      errMsg.value = err?.message ?? '登录验证失败，请重试'
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
