<!-- 系统管理 - 基础配置 -->
<template>
  <div class="basic-config-page">
    <ElCard shadow="never">
      <template #header>
        <h2 class="section-title">{{ $t('system.basicConfig.title') }}</h2>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        label-width="120px"
        style="max-width: 680px"
        v-loading="loadingConfig"
      >
        <ElFormItem :label="$t('system.basicConfig.corpId')" prop="corp_id">
          <ElInputNumber
            v-model="form.corp_id"
            :min="1"
            :controls="false"
            style="width: 220px"
            :placeholder="$t('system.basicConfig.corpIdPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('system.basicConfig.siteTitle')" prop="site_title">
          <ElInput
            v-model="form.site_title"
            clearable
            :placeholder="$t('system.basicConfig.siteTitlePlaceholder')"
          />
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :loading="saving" @click="handleSave">
            {{ $t('system.basicConfig.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 服务器更新 -->
    <ElCard shadow="never" class="mt-4">
      <template #header>
        <h2 class="section-title">{{ $t('system.serverUpdate.title') }}</h2>
      </template>

      <ElDescriptions :column="2" border class="mb-4" v-loading="checkingUpdate">
        <ElDescriptionsItem :label="$t('system.serverUpdate.currentVersion')">
          <ElTag type="info">{{ updateInfo?.current_version ?? '-' }}</ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('system.serverUpdate.latestVersion')">
          <ElTag :type="updateInfo?.has_update ? 'warning' : 'success'">
            {{ updateInfo?.latest_version ?? '-' }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('system.serverUpdate.backendSize')">
          {{ updateInfo ? formatBytes(updateInfo.download_size) : '-' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('system.serverUpdate.frontendSize')">
          {{ updateInfo ? formatBytes(updateInfo.frontend_download_size) : '-' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem
          v-if="updateInfo?.release_notes"
          :label="$t('system.serverUpdate.releaseNotes')"
          :span="2"
        >
          <pre class="release-notes">{{ updateInfo.release_notes }}</pre>
        </ElDescriptionsItem>
      </ElDescriptions>

      <ElSpace>
        <ElButton :loading="checkingUpdate" @click="handleCheckUpdate">
          {{ $t('system.serverUpdate.checkUpdate') }}
        </ElButton>
        <ElButton
          v-if="updateInfo?.has_update"
          type="danger"
          :loading="upgrading"
          @click="handleUpgrade"
        >
          {{ $t('system.serverUpdate.upgradeBackend') }}
        </ElButton>
        <ElButton
          v-if="updateInfo?.has_update"
          type="warning"
          :loading="upgradingFrontend"
          @click="handleFrontendUpgrade"
        >
          {{ $t('system.serverUpdate.upgradeFrontend') }}
        </ElButton>
      </ElSpace>

      <ElAlert
        v-if="upgrading"
        class="mt-4"
        type="warning"
        :title="$t('system.serverUpdate.upgradingTip')"
        show-icon
        :closable="false"
      />
    </ElCard>

    <!-- SeAT 配置 -->
    <ElCard shadow="never" class="mt-4">
      <template #header>
        <h2 class="section-title">{{ $t('system.seatConfig.title') }}</h2>
      </template>

      <ElForm
        :model="seatForm"
        label-width="140px"
        style="max-width: 680px"
        v-loading="loadingSeat"
      >
        <ElFormItem :label="$t('system.seatConfig.enabled')">
          <ElSwitch v-model="seatForm.enabled" />
        </ElFormItem>

        <ElFormItem :label="$t('system.seatConfig.baseUrl')">
          <ElInput
            v-model="seatForm.base_url"
            clearable
            :placeholder="$t('system.seatConfig.baseUrlPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('system.seatConfig.clientId')">
          <ElInput
            v-model="seatForm.client_id"
            clearable
            :placeholder="$t('system.seatConfig.clientIdPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('system.seatConfig.clientSecret')">
          <ElInput
            v-model="seatForm.client_secret"
            clearable
            type="password"
            show-password
            :placeholder="$t('system.seatConfig.clientSecretPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('system.seatConfig.callbackUrl')">
          <ElInput
            v-model="seatForm.callback_url"
            clearable
            :placeholder="$t('system.seatConfig.callbackUrlPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('system.seatConfig.scopes')">
          <ElInput
            v-model="seatForm.scopes"
            clearable
            :placeholder="$t('system.seatConfig.scopesPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :loading="savingSeat" @click="handleSaveSeat">
            {{ $t('system.seatConfig.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import {
    ElCard,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElButton,
    ElSwitch,
    ElMessage,
    ElDescriptions,
    ElDescriptionsItem,
    ElTag,
    ElSpace,
    ElAlert
  } from 'element-plus'
  import { useSysConfigStore } from '@/store/modules/sys-config'
  import {
    fetchSeatConfig,
    updateSeatConfig,
    checkServerUpdate,
    performServerUpgrade,
    performFrontendUpgrade
  } from '@/api/sys-config'

  defineOptions({ name: 'BasicConfig' })

  const { t } = useI18n()
  const sysConfigStore = useSysConfigStore()

  const loadingConfig = ref(false)
  const saving = ref(false)

  const form = reactive<Api.SysConfig.BasicConfig>({
    corp_id: sysConfigStore.config.corp_id,
    site_title: sysConfigStore.config.site_title
  })

  // ─── 服务器更新 ───
  const checkingUpdate = ref(false)
  const upgrading = ref(false)
  const upgradingFrontend = ref(false)
  const updateInfo = ref<Api.SysConfig.ServerUpdateInfo | null>(null)

  const formatBytes = (bytes: number): string => {
    if (!bytes) return '-'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  }

  const handleCheckUpdate = async () => {
    checkingUpdate.value = true
    try {
      updateInfo.value = await checkServerUpdate()
      if (!updateInfo.value.has_update) {
        ElMessage.success(t('system.serverUpdate.alreadyLatest'))
      }
    } catch {
      /* empty */
    } finally {
      checkingUpdate.value = false
    }
  }

  const handleUpgrade = async () => {
    upgrading.value = true
    try {
      await performServerUpgrade()
      ElMessage.success(t('system.serverUpdate.upgradeStarted'))
    } catch {
      upgrading.value = false
    }
    // keep upgrading=true: server is restarting, button stays disabled
  }

  const handleFrontendUpgrade = async () => {
    upgradingFrontend.value = true
    try {
      await performFrontendUpgrade()
      ElMessage.success(t('system.serverUpdate.frontendUpgradeSuccess'))
    } catch {
      /* empty */
    } finally {
      upgradingFrontend.value = false
    }
  }

  // ─── SeAT 配置 ───
  const loadingSeat = ref(false)
  const savingSeat = ref(false)
  const seatForm = reactive({
    enabled: false,
    base_url: '',
    client_id: '',
    client_secret: '',
    callback_url: '',
    scopes: ''
  })

  const loadConfig = async () => {
    loadingConfig.value = true
    try {
      await sysConfigStore.ensureLoaded()
      form.corp_id = sysConfigStore.config.corp_id
      form.site_title = sysConfigStore.config.site_title
    } catch {
      /* empty */
    } finally {
      loadingConfig.value = false
    }
  }

  const loadSeatConfig = async () => {
    loadingSeat.value = true
    try {
      const data = await fetchSeatConfig()
      seatForm.enabled = data.enabled
      seatForm.base_url = data.base_url
      seatForm.client_id = data.client_id
      seatForm.client_secret = data.client_secret
      seatForm.callback_url = data.callback_url
      seatForm.scopes = data.scopes
    } catch {
      /* empty */
    } finally {
      loadingSeat.value = false
    }
  }

  const handleSave = async () => {
    saving.value = true
    try {
      await sysConfigStore.updateConfig({
        corp_id: form.corp_id,
        site_title: form.site_title
      })
      ElMessage.success(t('system.basicConfig.saveSuccess'))
    } catch {
      /* empty */
    } finally {
      saving.value = false
    }
  }

  const handleSaveSeat = async () => {
    savingSeat.value = true
    try {
      await updateSeatConfig({
        enabled: seatForm.enabled ? 'true' : 'false',
        base_url: seatForm.base_url,
        client_id: seatForm.client_id,
        client_secret: seatForm.client_secret,
        callback_url: seatForm.callback_url,
        scopes: seatForm.scopes
      })
      ElMessage.success(t('system.seatConfig.saveSuccess'))
    } catch {
      /* empty */
    } finally {
      savingSeat.value = false
    }
  }

  onMounted(() => {
    loadConfig()
    loadSeatConfig()
    handleCheckUpdate()
  })
</script>

<style scoped>
  .section-title {
    font-size: 15px;
    font-weight: 600;
    margin: 0;
  }

  .release-notes {
    white-space: pre-wrap;
    word-break: break-word;
    font-family: inherit;
    font-size: 13px;
    margin: 0;
    max-height: 200px;
    overflow-y: auto;
  }
</style>
