<!-- 系统管理 - PAP 兑换汇率配置 -->
<template>
  <div class="pap-exchange-page art-full-height">
    <ElCard shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium">{{ t('papExchange.title') }}</span>
          <ElButton
            v-auth="'system:pap:exchange'"
            type="primary"
            :loading="saving"
            @click="handleSave"
          >
            {{ t('common.save') }}
          </ElButton>
        </div>
      </template>

      <ElAlert class="mb-4" :title="t('papExchange.tip')" type="info" :closable="false" />

      <ElTable :data="rates" :loading="loading" border style="width: 100%">
        <ElTableColumn prop="display_name" :label="t('papExchange.columns.type')" width="200" />
        <ElTableColumn :label="t('papExchange.columns.rate')" min-width="260">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.rate"
              :min="0.01"
              :precision="2"
              :step="1"
              :controls="false"
              style="width: 160px"
            />
            <span class="ml-2 text-gray-400 text-sm">{{ t('papExchange.columns.rateUnit') }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="updated_at" :label="t('papExchange.columns.updatedAt')" width="180">
          <template #default="{ row }">
            {{ row.updated_at ? new Date(row.updated_at).toLocaleString() : '-' }}
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import {
    ElCard,
    ElButton,
    ElAlert,
    ElTable,
    ElTableColumn,
    ElInputNumber,
    ElMessage
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { fetchPAPTypeRates, updatePAPTypeRates, type PAPTypeRate } from '@/api/pap-exchange'

  defineOptions({ name: 'PAPExchange' })

  const { t } = useI18n()

  const loading = ref(false)
  const saving = ref(false)
  const rates = ref<PAPTypeRate[]>([])

  async function loadRates() {
    loading.value = true
    try {
      rates.value = await fetchPAPTypeRates()
    } catch {
      ElMessage.error(t('papExchange.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  async function handleSave() {
    saving.value = true
    try {
      rates.value = await updatePAPTypeRates(
        rates.value.map(({ pap_type, display_name, rate }) => ({ pap_type, display_name, rate }))
      )
      ElMessage.success(t('papExchange.saveSuccess'))
    } catch {
      ElMessage.error(t('papExchange.saveFailed'))
    } finally {
      saving.value = false
    }
  }

  onMounted(loadRates)
</script>
