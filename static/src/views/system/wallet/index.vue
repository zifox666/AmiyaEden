<!-- 系统钱包管理页面 -->
<template>
  <div class="wallet-admin-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ElTabs v-model="activeTab">
        <!-- 钱包列表 -->
        <ElTabPane :label="$t('walletAdmin.tabs.wallets')" name="wallets">
          <WalletList
            ref="walletListRef"
            @adjust="handleAdjust"
            @view-transactions="handleViewTransactions"
          />
        </ElTabPane>

        <!-- 流水查询 -->
        <ElTabPane :label="$t('walletAdmin.tabs.transactions')" name="transactions">
          <WalletTransactions ref="walletTxRef" />
        </ElTabPane>

        <!-- 操作日志 -->
        <ElTabPane :label="$t('walletAdmin.tabs.logs')" name="logs">
          <WalletLogs />
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <!-- 调整余额弹窗 -->
    <ElDialog
      v-model="adjustDialogVisible"
      :title="$t('walletAdmin.adjustTitle')"
      width="480px"
      destroy-on-close
    >
      <ElForm ref="adjustFormRef" :model="adjustForm" :rules="adjustRules" label-width="100px">
        <ElFormItem :label="$t('walletAdmin.fields.targetUserId')" prop="target_uid">
          <ElInputNumber
            v-model="adjustForm.target_uid"
            :min="1"
            :controls="false"
            style="width: 100%"
          />
        </ElFormItem>
        <ElFormItem :label="$t('walletAdmin.fields.action')" prop="action">
          <ElRadioGroup v-model="adjustForm.action">
            <ElRadio value="add">{{ $t('walletAdmin.actions.add') }}</ElRadio>
            <ElRadio value="deduct">{{ $t('walletAdmin.actions.deduct') }}</ElRadio>
            <ElRadio value="set">{{ $t('walletAdmin.actions.set') }}</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem :label="$t('walletAdmin.fields.amount')" prop="amount">
          <ElInputNumber
            v-model="adjustForm.amount"
            :min="0.01"
            :precision="2"
            :controls="false"
            style="width: 100%"
          />
        </ElFormItem>
        <ElFormItem :label="$t('walletAdmin.fields.reason')" prop="reason">
          <ElInput
            v-model="adjustForm.reason"
            type="textarea"
            :rows="3"
            :placeholder="$t('walletAdmin.placeholders.reason')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="adjustDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="adjustLoading" @click="submitAdjust">{{
          $t('common.confirm')
        }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    ElCard,
    ElTabs,
    ElTabPane,
    ElButton,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElRadioGroup,
    ElRadio,
    ElMessage,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { adminAdjustWallet } from '@/api/sys-wallet'
  import WalletList from './modules/wallet-list.vue'
  import WalletTransactions from './modules/wallet-transactions.vue'
  import WalletLogs from './modules/wallet-logs.vue'

  defineOptions({ name: 'SystemWallet' })
  const { t } = useI18n()

  // ── Tab ──
  const activeTab = ref('wallets')

  // ── 子模块 refs ──
  const walletListRef = ref<InstanceType<typeof WalletList>>()
  const walletTxRef = ref<InstanceType<typeof WalletTransactions>>()

  // ── 来自钱包列表的事件 ──
  const handleViewTransactions = (userId: number) => {
    activeTab.value = 'transactions'
    nextTick(() => walletTxRef.value?.filterByUser(userId))
  }

  const handleAdjust = (userId: number, action: 'add' | 'deduct' | 'set') => {
    showAdjustDialog(userId, action)
  }

  // ══════════════════════════════════════════
  //  调整余额弹窗
  // ══════════════════════════════════════════
  const adjustDialogVisible = ref(false)
  const adjustLoading = ref(false)
  const adjustFormRef = ref<FormInstance>()

  const adjustForm = reactive<Api.SysWallet.AdjustParams>({
    target_uid: 0,
    action: 'add',
    amount: 0,
    reason: ''
  })

  const adjustRules: FormRules = {
    target_uid: [
      { required: true, message: t('walletAdmin.validation.targetUserId'), trigger: 'blur' }
    ],
    action: [{ required: true, message: t('walletAdmin.validation.action'), trigger: 'change' }],
    amount: [{ required: true, message: t('walletAdmin.validation.amount'), trigger: 'blur' }],
    reason: [{ required: true, message: t('walletAdmin.validation.reason'), trigger: 'blur' }]
  }

  const showAdjustDialog = (userId = 0, action: 'add' | 'deduct' | 'set' = 'add') => {
    adjustForm.target_uid = userId
    adjustForm.action = action
    adjustForm.amount = 0
    adjustForm.reason = ''
    adjustDialogVisible.value = true
  }

  const submitAdjust = async () => {
    if (!adjustFormRef.value) return
    await adjustFormRef.value.validate()

    adjustLoading.value = true
    try {
      await adminAdjustWallet(adjustForm)
      ElMessage.success(t('walletAdmin.messages.adjustSuccess'))
      adjustDialogVisible.value = false
      walletListRef.value?.refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('walletAdmin.messages.actionFailed'))
    } finally {
      adjustLoading.value = false
    }
  }
</script>
