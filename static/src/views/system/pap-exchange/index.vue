<!-- 系统管理 - PAP 兑换汇率配置 -->
<template>
  <div class="pap-exchange-page">
    <ElCard shadow="never" class="mb-4">
      <template #header>
        <h2 class="section-title">{{ t('papExchange.fcSection') }}</h2>
      </template>

      <ElForm label-width="150px" style="max-width: 680px" v-loading="loading">
        <ElFormItem :label="t('papExchange.fcSalary')">
          <div class="field-block">
            <div class="field-row">
              <ElInputNumber
                v-model="form.fc_salary"
                :min="0"
                :precision="2"
                :step="10"
                :controls="false"
                style="width: 180px"
              />
              <span class="text-sm text-secondary">{{ t('papExchange.fcSalaryUnit') }}</span>
            </div>
            <div class="form-hint">{{ t('papExchange.fcSalaryHint') }}</div>
          </div>
        </ElFormItem>

        <ElFormItem :label="t('papExchange.fcSalaryMonthlyLimit')">
          <div class="field-block">
            <div class="field-row">
              <ElInputNumber
                v-model="form.fc_salary_monthly_limit"
                :min="0"
                :precision="0"
                :step="1"
                :controls="false"
                style="width: 180px"
              />
              <span class="text-sm text-secondary">
                {{ t('papExchange.fcSalaryMonthlyLimitUnit') }}
              </span>
            </div>
            <div class="form-hint">{{ t('papExchange.fcSalaryMonthlyLimitHint') }}</div>
          </div>
        </ElFormItem>

        <ElFormItem>
          <ElButton
            v-auth="'system:pap:exchange'"
            type="primary"
            :loading="saving"
            @click="handleSave"
          >
            {{ t('common.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <ElCard shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="section-title">{{ t('papExchange.ratesSection') }}</h2>
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

      <ElAlert class="mb-4" :title="t('papExchange.tip')" type="info" :closable="false" show-icon />

      <ElTable :data="form.rates" v-loading="loading" border style="width: 100%">
        <ElTableColumn prop="display_name" :label="t('papExchange.columns.type')" width="200" />
        <ElTableColumn :label="t('papExchange.columns.rate')" min-width="280">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.rate"
              :min="0.01"
              :precision="2"
              :step="1"
              :controls="false"
              style="width: 160px"
            />
            <span class="ml-2 text-sm text-secondary">
              {{ t('papExchange.columns.rateUnit') }}
            </span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="updated_at" :label="t('papExchange.columns.updatedAt')" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
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
    ElForm,
    ElFormItem,
    ElTable,
    ElTableColumn,
    ElInputNumber,
    ElMessage
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { formatTime } from '@utils/common'
  import {
    fetchPAPExchangeConfig,
    updatePAPExchangeConfig,
    type PAPExchangeConfig
  } from '@/api/pap-exchange'

  defineOptions({ name: 'PAPExchange' })

  const { t } = useI18n()

  const loading = ref(false)
  const saving = ref(false)
  const form = reactive<PAPExchangeConfig>({
    rates: [],
    fc_salary: 400,
    fc_salary_monthly_limit: 5
  })

  async function loadRates() {
    loading.value = true
    try {
      const config = await fetchPAPExchangeConfig()
      form.rates = config.rates
      form.fc_salary = config.fc_salary
      form.fc_salary_monthly_limit = config.fc_salary_monthly_limit
    } catch {
      ElMessage.error(t('papExchange.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  async function handleSave() {
    saving.value = true
    try {
      const config = await updatePAPExchangeConfig({
        fc_salary: form.fc_salary,
        fc_salary_monthly_limit: form.fc_salary_monthly_limit,
        rates: form.rates.map(({ pap_type, display_name, rate }) => ({
          pap_type,
          display_name,
          rate
        }))
      })
      form.rates = config.rates
      form.fc_salary = config.fc_salary
      form.fc_salary_monthly_limit = config.fc_salary_monthly_limit
      ElMessage.success(t('papExchange.saveSuccess'))
    } catch {
      ElMessage.error(t('papExchange.saveFailed'))
    } finally {
      saving.value = false
    }
  }

  onMounted(loadRates)
</script>

<style scoped>
  .section-title {
    font-size: 15px;
    font-weight: 600;
    margin: 0;
  }

  .field-block {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .form-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .text-secondary {
    color: var(--el-text-color-secondary);
  }
</style>
