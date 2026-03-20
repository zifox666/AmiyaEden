<template>
  <!-- PAP 兑换配置 & 月度结算弹窗 -->
  <ElDialog
    v-model="visible"
    :title="t('alliancePap.settle.title')"
    width="480px"
    destroy-on-close
    @open="handleOpen"
  >
    <!-- 兑换配置 -->
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="130px" class="mt-2">
      <ElFormItem :label="t('alliancePap.settle.enabled')">
        <ElSwitch v-model="form.enabled" />
      </ElFormItem>
      <ElFormItem :label="t('alliancePap.settle.walletPerPap')" prop="wallet_per_pap">
        <ElInputNumber
          v-model="form.wallet_per_pap"
          :min="0.01"
          :precision="2"
          :step="0.1"
          :controls="false"
          style="width: 180px"
        />
        <span class="ml-2 text-gray-400 text-sm">{{
          t('alliancePap.settle.walletPerPapUnit')
        }}</span>
      </ElFormItem>

      <ElDivider />

      <!-- 结算月份 -->
      <ElFormItem :label="t('alliancePap.settle.settleMonth')" prop="settle_month">
        <ElDatePicker
          v-model="form.settle_month"
          type="month"
          format="YYYY-MM"
          value-format="YYYY-MM"
          :clearable="false"
          style="width: 180px"
        />
      </ElFormItem>
      <ElFormItem :label="t('alliancePap.settle.walletConvert')">
        <ElSwitch v-model="form.wallet_convert" />
        <span class="ml-2 text-gray-400 text-xs">{{
          t('alliancePap.settle.walletConvertTip')
        }}</span>
      </ElFormItem>
    </ElForm>

    <!-- 结算结果 -->
    <ElAlert
      v-if="settleResult"
      class="mt-2"
      :title="t('alliancePap.settle.resultTitle')"
      type="success"
      :closable="false"
    >
      <template #default>
        <div class="text-sm mt-1 space-y-1">
          <div>{{ t('alliancePap.settle.resultUsers', { count: settleResult.total_users }) }}</div>
          <div>{{
            t('alliancePap.settle.resultSkipped', { count: settleResult.skipped_users })
          }}</div>
          <div v-if="form.wallet_convert">
            {{
              t('alliancePap.settle.resultWallet', { amount: settleResult.total_wallet.toFixed(2) })
            }}
          </div>
        </div>
      </template>
    </ElAlert>

    <template #footer>
      <ElButton @click="visible = false">{{ t('common.cancel') }}</ElButton>
      <ElButton type="warning" :loading="savingConfig" @click="handleSaveConfig">
        {{ t('alliancePap.settle.saveConfig') }}
      </ElButton>
      <ElButton type="primary" :loading="settling" @click="handleSettle">
        {{ t('alliancePap.settle.doSettle') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElSwitch,
    ElDatePicker,
    ElDivider,
    ElAlert,
    ElButton,
    ElMessage,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import {
    fetchPAPExchangeConfig,
    updatePAPExchangeConfig,
    settleAlliancePAPMonth,
    type SettleMonthResult
  } from '@/api/alliance-pap'

  const { t } = useI18n()

  const visible = defineModel<boolean>({ required: true })

  const formRef = ref<FormInstance>()
  const savingConfig = ref(false)
  const settling = ref(false)
  const settleResult = ref<SettleMonthResult | null>(null)

  const now = new Date()
  const defaultMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`

  const form = reactive({
    enabled: true,
    wallet_per_pap: 1,
    settle_month: defaultMonth,
    wallet_convert: true
  })

  const rules: FormRules = {
    wallet_per_pap: [
      {
        required: true,
        type: 'number',
        min: 0.01,
        message: t('alliancePap.settle.walletPerPapRequired'),
        trigger: 'blur'
      }
    ],
    settle_month: [
      { required: true, message: t('alliancePap.settle.settleMonthRequired'), trigger: 'change' }
    ]
  }

  async function handleOpen() {
    settleResult.value = null
    try {
      const cfg = await fetchPAPExchangeConfig()
      form.enabled = cfg.enabled
      form.wallet_per_pap = cfg.wallet_per_pap
    } catch {
      // 使用默认值
    }
  }

  async function handleSaveConfig() {
    const valid = await formRef.value?.validateField(['wallet_per_pap'])
    if (!valid) return
    savingConfig.value = true
    try {
      await updatePAPExchangeConfig({ enabled: form.enabled, wallet_per_pap: form.wallet_per_pap })
      ElMessage.success(t('alliancePap.settle.configSaved'))
    } catch (e: any) {
      ElMessage.error(e?.message || t('alliancePap.settle.configFailed'))
    } finally {
      savingConfig.value = false
    }
  }

  async function handleSettle() {
    const valid = await formRef.value?.validate()
    if (!valid) return
    const [yearStr, monthStr] = form.settle_month.split('-')
    settling.value = true
    settleResult.value = null
    try {
      const result = await settleAlliancePAPMonth({
        year: Number(yearStr),
        month: Number(monthStr),
        wallet_convert: form.wallet_convert
      })
      settleResult.value = result
      ElMessage.success(t('alliancePap.settle.settleSuccess'))
    } catch (e: any) {
      ElMessage.error(e?.message || t('alliancePap.settle.settleFailed'))
    } finally {
      settling.value = false
    }
  }
</script>
