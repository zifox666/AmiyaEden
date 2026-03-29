<!-- 系统管理 - 基础配置 -->
<template>
  <div class="basic-config-page">
    <ElCard shadow="never">
      <template #header>
        <h2 class="section-title">{{ $t('system.basicConfig.allowCorporations') }}</h2>
      </template>

      <ElForm
        :model="allowCorpsForm"
        label-width="120px"
        style="max-width: 680px"
        v-loading="loadingAllowCorpsConfig"
      >
        <ElFormItem
          :label="$t('system.basicConfig.allowCorporationsLabel')"
          prop="allow_corporations"
        >
          <ElInput
            v-model="allowCorpsInput"
            type="textarea"
            :rows="6"
            clearable
            :placeholder="$t('system.basicConfig.allowCorporationsPlaceholder')"
          />
          <div class="form-hint">
            {{ $t('system.basicConfig.allowCorporationsHint', SYSTEM_IDENTITY_I18N) }}
          </div>
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :loading="savingAllowCorps" @click="handleSaveAllowCorps">
            {{ $t('system.basicConfig.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <ElCard shadow="never" style="margin-top: 16px">
      <template #header>
        <h2 class="section-title">{{ $t('system.basicConfig.sdeConfig') }}</h2>
      </template>

      <ElForm
        :model="sdeForm"
        label-width="120px"
        style="max-width: 680px"
        v-loading="loadingSDEConfig"
      >
        <ElFormItem :label="$t('system.basicConfig.sdeApiKey')" prop="api_key">
          <ElInput
            v-model="sdeForm.api_key"
            clearable
            show-password
            :placeholder="$t('system.basicConfig.sdeApiKeyPlaceholder')"
            style="width: 400px"
          />
        </ElFormItem>

        <ElFormItem :label="$t('system.basicConfig.sdeProxy')" prop="proxy">
          <ElInput
            v-model="sdeForm.proxy"
            clearable
            :placeholder="$t('system.basicConfig.sdeProxyPlaceholder')"
            style="width: 400px"
          />
        </ElFormItem>

        <ElFormItem :label="$t('system.basicConfig.sdeDownloadUrl')" prop="download_url">
          <ElInput
            v-model="sdeForm.download_url"
            clearable
            :placeholder="$t('system.basicConfig.sdeDownloadUrlPlaceholder')"
            style="width: 500px"
          />
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :loading="savingSDE" @click="handleSaveSDE">
            {{ $t('system.basicConfig.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElCard, ElForm, ElFormItem, ElInput, ElButton, ElMessage } from 'element-plus'
  import {
    fetchSDEConfig,
    updateSDEConfig,
    fetchAllowCorporations,
    updateAllowCorporations
  } from '@/api/sys-config'
  import { SYSTEM_IDENTITY, SYSTEM_IDENTITY_I18N } from '@/constants/system-identity'

  defineOptions({ name: 'BasicConfig' })

  const { t } = useI18n()
  const loadingSDEConfig = ref(false)
  const savingSDE = ref(false)
  const loadingAllowCorpsConfig = ref(false)
  const savingAllowCorps = ref(false)
  const REQUIRED_ALLOW_CORPORATION_ID = SYSTEM_IDENTITY.corporationId

  const sdeForm = reactive<Api.SysConfig.SDEConfig>({
    api_key: '',
    proxy: '',
    download_url: ''
  })

  const allowCorpsForm = reactive<Api.SysConfig.AllowCorporationsConfig>({
    allow_corporations: []
  })

  const allowCorpsInput = ref('')

  const normalizeAllowCorporations = (corporations: number[]) => {
    const seen = new Set<number>([REQUIRED_ALLOW_CORPORATION_ID])
    return [
      REQUIRED_ALLOW_CORPORATION_ID,
      ...corporations.filter((corporationID) => {
        if (seen.has(corporationID)) {
          return false
        }
        seen.add(corporationID)
        return true
      })
    ]
  }

  const parseCorporationId = (value: string) => {
    if (!/^\d+$/.test(value)) {
      throw new Error(t('system.basicConfig.invalidCorpId'))
    }

    const corporationId = Number.parseInt(value, 10)
    if (!Number.isSafeInteger(corporationId) || corporationId <= 0) {
      throw new Error(t('system.basicConfig.invalidCorpId'))
    }

    return corporationId
  }

  const loadSDEConfig = async () => {
    loadingSDEConfig.value = true
    try {
      const res = await fetchSDEConfig()
      sdeForm.api_key = res.api_key
      sdeForm.proxy = res.proxy
      sdeForm.download_url = res.download_url
    } catch {
      /* empty */
    } finally {
      loadingSDEConfig.value = false
    }
  }

  const handleSaveSDE = async () => {
    savingSDE.value = true
    try {
      await updateSDEConfig({
        api_key: sdeForm.api_key,
        proxy: sdeForm.proxy,
        download_url: sdeForm.download_url
      })
      ElMessage.success(t('system.basicConfig.saveSuccess'))
    } catch {
      /* empty */
    } finally {
      savingSDE.value = false
    }
  }

  const loadAllowCorpsConfig = async () => {
    loadingAllowCorpsConfig.value = true
    try {
      const res = await fetchAllowCorporations()
      const corporations = normalizeAllowCorporations(res.allow_corporations)
      allowCorpsForm.allow_corporations = corporations
      allowCorpsInput.value = corporations.join('\n')
    } catch {
      /* empty */
    } finally {
      loadingAllowCorpsConfig.value = false
    }
  }

  const handleSaveAllowCorps = async () => {
    try {
      const lines = allowCorpsInput.value
        .split('\n')
        .map((line) => line.trim())
        .filter((line) => line !== '')
      const corps = normalizeAllowCorporations(lines.map(parseCorporationId))

      savingAllowCorps.value = true
      await updateAllowCorporations({ allow_corporations: corps })
      allowCorpsForm.allow_corporations = corps
      allowCorpsInput.value = corps.join('\n')
      ElMessage.success(t('system.basicConfig.saveSuccess'))
    } catch (error) {
      ElMessage.error(
        error instanceof Error && error.message
          ? error.message
          : t('system.basicConfig.invalidCorpId')
      )
    } finally {
      savingAllowCorps.value = false
    }
  }

  onMounted(() => {
    loadAllowCorpsConfig()
    loadSDEConfig()
  })
</script>

<style scoped>
  .section-title {
    font-size: 15px;
    font-weight: 600;
    margin: 0;
  }

  .form-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-top: 4px;
  }
</style>
