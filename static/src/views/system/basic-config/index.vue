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
    ElMessage
  } from 'element-plus'
  import { useSysConfigStore } from '@/store/modules/sys-config'
  import { fetchSeatConfig, updateSeatConfig } from '@/api/sys-config'

  defineOptions({ name: 'BasicConfig' })

  const { t } = useI18n()
  const sysConfigStore = useSysConfigStore()

  const loadingConfig = ref(false)
  const saving = ref(false)

  const form = reactive<Api.SysConfig.BasicConfig>({
    corp_id: sysConfigStore.config.corp_id,
    site_title: sysConfigStore.config.site_title
  })

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
  })
</script>

<style scoped>
  .section-title {
    font-size: 15px;
    font-weight: 600;
    margin: 0;
  }
</style>
