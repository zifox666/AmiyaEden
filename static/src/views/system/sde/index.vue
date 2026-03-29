<!-- 系统管理 - SDE 数据管理 -->
<template>
  <div class="sde-manage-page">
    <ElCard shadow="never">
      <template #header>
        <h2 class="section-title">{{ $t('system.sdeManage.title') }}</h2>
      </template>

      <div v-loading="loading">
        <!-- 版本信息 -->
        <ElDescriptions :column="1" border style="max-width: 560px; margin-bottom: 24px">
          <ElDescriptionsItem :label="$t('system.sdeManage.currentVersion')">
            <ElTag v-if="version" type="success">{{ version.version }}</ElTag>
            <ElTag v-else type="info">{{ $t('system.sdeManage.noVersion') }}</ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="version" :label="$t('system.sdeManage.note')">
            {{ version.note || '—' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem v-if="version" :label="$t('system.sdeManage.importedAt')">
            {{ version.created_at }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <!-- 操作区 -->
        <ElButton
          type="primary"
          :loading="updating"
          v-auth="'system:sde:update'"
          @click="handleUpdate"
        >
          {{ $t('system.sdeManage.triggerUpdate') }}
        </ElButton>

        <p class="tip-text">{{ $t('system.sdeManage.tip') }}</p>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import {
    ElCard,
    ElDescriptions,
    ElDescriptionsItem,
    ElTag,
    ElButton,
    ElMessage
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { fetchSdeVersionAdmin, triggerSdeUpdate } from '@/api/sde'

  defineOptions({ name: 'SdeManage' })

  const { t } = useI18n()

  const loading = ref(false)
  const updating = ref(false)
  const version = ref<Api.Sde.SdeVersion | null>(null)

  async function loadVersion() {
    loading.value = true
    try {
      const res = await fetchSdeVersionAdmin()
      version.value = res ?? null
    } finally {
      loading.value = false
    }
  }

  async function handleUpdate() {
    updating.value = true
    try {
      const res = await triggerSdeUpdate()
      ElMessage.success(t('system.sdeManage.updateSuccess', { version: res.version }))
      await loadVersion()
    } catch {
      // http 层已统一处理错误弹窗
    } finally {
      updating.value = false
    }
  }

  onMounted(() => {
    loadVersion()
  })
</script>

<style scoped>
  .tip-text {
    margin-top: 12px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
</style>
