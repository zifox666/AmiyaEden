<template>
  <!-- 月度归档弹窗 -->
  <ElDialog v-model="visible" :title="t('alliancePap.settle.title')" width="400px" destroy-on-close>
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px" class="mt-2">
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
    </ElForm>

    <!-- 归档结果 -->
    <ElAlert
      v-if="settleResult"
      class="mt-2"
      :title="t('alliancePap.settle.settleSuccess')"
      type="success"
      :closable="false"
    />

    <template #footer>
      <ElButton @click="visible = false">{{ t('common.cancel') }}</ElButton>
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
    ElDatePicker,
    ElAlert,
    ElButton,
    ElMessage,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { settleAlliancePAPMonth, type SettleMonthResult } from '@/api/alliance-pap'

  const { t } = useI18n()

  const visible = defineModel<boolean>({ required: true })

  const formRef = ref<FormInstance>()
  const settling = ref(false)
  const settleResult = ref<SettleMonthResult | null>(null)

  const now = new Date()
  const defaultMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`

  const form = reactive({
    settle_month: defaultMonth
  })

  const rules: FormRules = {
    settle_month: [
      { required: true, message: t('alliancePap.settle.settleMonthRequired'), trigger: 'change' }
    ]
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
        month: Number(monthStr)
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
