<!-- EVE Online SSO 鐧诲綍椤?-->
<template>
  <div class="flex w-full h-screen">
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h3 class="title">{{ $t('login.title') }}</h3>
          <p class="sub-title">{{ $t('login.subTitle') }}</p>

          <div class="sso-section">
            <!-- EVE Online Logo -->
            <div class="eve-logo-wrap">
              <img
                src="https://web.ccpgamescdn.com/eveonlineassets/developers/eve-sso-login-white-large.png"
                alt="Sign In with EVE Online"
                class="eve-logo"
                @click="handleEveLogin"
                :class="{ loading: loading }"
              />
            </div>

            <!-- SeAT Login -->
            <el-button
              v-if="seatEnabled"
              type="primary"
              size="large"
              class="sso-btn seat-btn"
              :loading="seatLoading"
              @click="handleSeatLogin"
            >
              {{ $t('login.seatBtnText') }}
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { getEveSSOLoginURL, checkSeatEnabled, getSeatLoginURL } from '@/api/auth'

  defineOptions({ name: 'Login' })

  const loading = ref(false)
  const seatLoading = ref(false)
  const seatEnabled = ref(false)

  onMounted(async () => {
    try {
      const data = await checkSeatEnabled()
      seatEnabled.value = data.enabled
    } catch {
      // SeAT 不可用时静默忽略
    }
  })

  const handleEveLogin = async () => {
    loading.value = true
    try {
      const url = await getEveSSOLoginURL()
      window.location.href = url
    } catch {
      loading.value = false
    }
  }

  const handleSeatLogin = async () => {
    seatLoading.value = true
    try {
      const url = await getSeatLoginURL()
      window.location.href = url
    } catch {
      seatLoading.value = false
    }
  }
</script>

<style scoped>
  @import './style.css';
</style>

<style lang="scss" scoped>
  .sso-section {
    @apply flex flex-col items-center gap-6 mt-12;
  }

  .eve-logo-wrap {
    @apply flex justify-center;

    .eve-logo {
      @apply h-12 object-contain opacity-80 transition-opacity duration-200;
      cursor: pointer;

      &:hover:not(.loading) {
        @apply opacity-100;
      }

      &.loading {
        @apply opacity-50 cursor-not-allowed;
      }
    }
  }

  .sso-btn {
    @apply h-12 text-base font-medium tracking-wide;
    max-width: 320px;

    .btn-icon {
      @apply h-5 mr-2 object-contain;
    }

    &.seat-btn {
      @apply w-full;
      max-width: 320px;
    }
  }

  .hint-text {
    @apply text-xs text-g-400 text-center max-w-xs;
  }
</style>
