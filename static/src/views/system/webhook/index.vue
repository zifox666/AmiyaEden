<!-- 系统管理 - Webhook 通知设置 -->
<template>
  <div class="webhook-page">
    <ElCard shadow="never">
      <template #header>
        <h2 class="section-title">{{ $t('webhook.title') }}</h2>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        label-width="120px"
        style="max-width: 680px"
        v-loading="loadingConfig"
      >
        <!-- 启用开关 -->
        <ElFormItem :label="$t('webhook.fields.enabled')">
          <ElSwitch v-model="form.enabled" />
        </ElFormItem>

        <!-- 类型 -->
        <ElFormItem :label="$t('webhook.fields.type')">
          <ElSelect v-model="form.type" style="width: 220px">
            <ElOption label="Discord" value="discord" />
            <ElOption label="飞书 (Feishu)" value="feishu" />
            <ElOption label="钉钉 (DingTalk)" value="dingtalk" />
            <ElOption label="OneBot v11 (QQ)" value="onebot" />
          </ElSelect>
        </ElFormItem>

        <!-- Webhook URL -->
        <ElFormItem :label="$t('webhook.fields.url')" prop="url">
          <ElInput
            v-model="form.url"
            :placeholder="
              form.type === 'onebot'
                ? $t('webhook.fields.obUrlPlaceholder')
                : $t('webhook.fields.urlPlaceholder')
            "
            clearable
          />
        </ElFormItem>

        <!-- OneBot 专属字段 -->
        <template v-if="form.type === 'onebot'">
          <ElFormItem :label="$t('webhook.fields.obTargetType')">
            <ElSelect v-model="form.ob_target_type" style="width: 160px">
              <ElOption :label="$t('webhook.fields.obGroup')" value="group" />
              <ElOption :label="$t('webhook.fields.obPrivate')" value="private" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem :label="$t('webhook.fields.obTargetId')">
            <ElInputNumber
              v-model="form.ob_target_id"
              :min="0"
              :controls="false"
              style="width: 220px"
              :placeholder="$t('webhook.fields.obTargetIdPlaceholder')"
            />
          </ElFormItem>
          <ElFormItem :label="$t('webhook.fields.obToken')">
            <ElInput
              v-model="form.ob_token"
              :placeholder="$t('webhook.fields.obTokenPlaceholder')"
              clearable
            />
          </ElFormItem>
        </template>

        <!-- 消息模板 -->
        <ElFormItem :label="$t('webhook.fields.template')">
          <div style="width: 100%">
            <ElInput
              v-model="form.fleet_template"
              type="textarea"
              :rows="8"
              :placeholder="$t('webhook.fields.templatePlaceholder')"
            />
            <div class="template-hint">
              {{ $t('webhook.fields.templateHint') }}: <code>{title}</code> <code>{fc_name}</code>
              <code>{importance}</code> <code>{pap_count}</code> <code>{start_at}</code>
              <code>{end_at}</code>
              <code>{description}</code>
            </div>
          </div>
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :loading="saving" @click="handleSave">
            {{ $t('common.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 测试发送 -->
    <ElCard shadow="never" class="mt-4">
      <template #header>
        <h3 class="section-title">{{ $t('webhook.test.title') }}</h3>
      </template>
      <ElForm label-width="120px" style="max-width: 680px">
        <ElFormItem :label="$t('webhook.test.type')">
          <ElSelect v-model="testForm.type" style="width: 220px">
            <ElOption label="Discord" value="discord" />
            <ElOption label="飞书 (Feishu)" value="feishu" />
            <ElOption label="钉钉 (DingTalk)" value="dingtalk" />
            <ElOption label="OneBot v11 (QQ)" value="onebot" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('webhook.test.url')">
          <ElInput
            v-model="testForm.url"
            :placeholder="
              testForm.type === 'onebot'
                ? $t('webhook.fields.obUrlPlaceholder')
                : $t('webhook.fields.urlPlaceholder')
            "
          />
        </ElFormItem>
        <template v-if="testForm.type === 'onebot'">
          <ElFormItem :label="$t('webhook.fields.obTargetType')">
            <ElSelect v-model="testForm.ob_target_type" style="width: 160px">
              <ElOption :label="$t('webhook.fields.obGroup')" value="group" />
              <ElOption :label="$t('webhook.fields.obPrivate')" value="private" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem :label="$t('webhook.fields.obTargetId')">
            <ElInputNumber
              v-model="testForm.ob_target_id"
              :min="0"
              :controls="false"
              style="width: 220px"
              :placeholder="$t('webhook.fields.obTargetIdPlaceholder')"
            />
          </ElFormItem>
          <ElFormItem :label="$t('webhook.fields.obToken')">
            <ElInput
              v-model="testForm.ob_token"
              :placeholder="$t('webhook.fields.obTokenPlaceholder')"
              clearable
            />
          </ElFormItem>
        </template>
        <ElFormItem :label="$t('webhook.test.content')">
          <ElInput
            v-model="testForm.content"
            :placeholder="$t('webhook.test.contentPlaceholder')"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="warning" :loading="testing" :disabled="!testForm.url" @click="handleTest">
            {{ $t('webhook.test.sendBtn') }}
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
    ElSwitch,
    ElSelect,
    ElOption,
    ElButton,
    ElMessage
  } from 'element-plus'
  import { fetchWebhookConfig, setWebhookConfig, testWebhook } from '@/api/webhook'

  defineOptions({ name: 'WebhookSettings' })

  const { t } = useI18n()

  const loadingConfig = ref(false)
  const saving = ref(false)
  const testing = ref(false)

  const form = reactive<Api.Webhook.Config>({
    url: '',
    enabled: false,
    type: 'discord',
    fleet_template: '',
    ob_target_type: 'group',
    ob_target_id: 0,
    ob_token: ''
  })

  const testForm = reactive({
    url: '',
    type: 'discord',
    content: '',
    ob_target_type: 'group' as 'group' | 'private',
    ob_target_id: 0,
    ob_token: ''
  })

  const loadConfig = async () => {
    loadingConfig.value = true
    try {
      const cfg = await fetchWebhookConfig()
      if (cfg) {
        form.url = cfg.url
        form.enabled = cfg.enabled
        form.type = cfg.type
        form.fleet_template = cfg.fleet_template
        form.ob_target_type = cfg.ob_target_type || 'group'
        form.ob_target_id = cfg.ob_target_id ?? 0
        form.ob_token = cfg.ob_token || ''
        testForm.url = cfg.url
        testForm.type = cfg.type
        testForm.ob_target_type = cfg.ob_target_type || 'group'
        testForm.ob_target_id = cfg.ob_target_id ?? 0
        testForm.ob_token = cfg.ob_token || ''
      }
    } catch {
      /* handled */
    } finally {
      loadingConfig.value = false
    }
  }

  const handleSave = async () => {
    saving.value = true
    try {
      await setWebhookConfig({ ...form })
      ElMessage.success(t('webhook.saveSuccess'))
    } catch {
      /* handled */
    } finally {
      saving.value = false
    }
  }

  const handleTest = async () => {
    if (!testForm.url) return
    testing.value = true
    try {
      await testWebhook({
        url: testForm.url,
        type: testForm.type,
        content: testForm.content,
        ob_target_type: testForm.ob_target_type,
        ob_target_id: testForm.ob_target_id,
        ob_token: testForm.ob_token
      })
      ElMessage.success(t('webhook.test.success'))
    } catch {
      /* handled */
    } finally {
      testing.value = false
    }
  }

  onMounted(() => {
    loadConfig()
  })
</script>

<style scoped>
  .section-title {
    font-size: 15px;
    font-weight: 600;
    margin: 0;
  }

  .template-hint {
    margin-top: 6px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.8;
  }

  .template-hint code {
    background: var(--el-fill-color);
    border-radius: 3px;
    padding: 1px 5px;
    margin: 0 2px;
    font-family: monospace;
    color: var(--el-color-primary);
  }
</style>
