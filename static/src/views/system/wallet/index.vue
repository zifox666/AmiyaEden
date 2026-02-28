<!-- 系统钱包管理页面 -->
<template>
  <div class="wallet-admin-page art-full-height">
    <!-- 搜索/操作栏 -->
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center gap-4 flex-wrap">
        <ElInput
          v-model="searchUserId"
          placeholder="用户 ID"
          clearable
          style="width: 160px"
          @clear="loadWallets"
        />
        <ElButton type="primary" @click="loadWallets">查询钱包</ElButton>
        <ElButton type="success" @click="showAdjustDialog()">手动调整余额</ElButton>
        <ElButton @click="activeTab = 'logs'">查看操作日志</ElButton>
      </div>
    </ElCard>

    <ElTabs v-model="activeTab" type="border-card">
      <!-- 钱包列表 -->
      <ElTabPane label="钱包列表" name="wallets">
        <ElTable v-loading="walletLoading" :data="wallets" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="user_id" label="用户 ID" width="120" align="center" />
          <ElTableColumn prop="balance" label="余额" width="180" align="right">
            <template #default="{ row }">
              <span :class="row.balance >= 0 ? 'text-green-600' : 'text-red-500'" class="font-bold">
                {{ formatISK(row.balance) }}
              </span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="updated_at" label="最后更新" width="200">
            <template #default="{ row }">
              {{ formatTime(row.updated_at) }}
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="280" align="center">
            <template #default="{ row }">
              <ElButton size="small" type="success" @click="showAdjustDialog(row.user_id, 'add')">
                增加
              </ElButton>
              <ElButton size="small" type="warning" @click="showAdjustDialog(row.user_id, 'deduct')">
                扣减
              </ElButton>
              <ElButton size="small" type="primary" @click="showUserTransactions(row.user_id)">
                流水
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>

        <div v-if="walletPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="walletPagination.current"
            v-model:page-size="walletPagination.size"
            :total="walletPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @size-change="() => { walletPagination.current = 1; loadWallets() }"
            @current-change="loadWallets"
          />
        </div>
      </ElTabPane>

      <!-- 流水查询 -->
      <ElTabPane label="流水查询" name="transactions">
        <div class="flex items-center gap-4 mb-4 flex-wrap">
          <ElInput
            v-model="txFilterUserId"
            placeholder="按用户 ID 筛选"
            clearable
            style="width: 160px"
          />
          <ElSelect v-model="txFilterRefType" placeholder="流水类型" clearable style="width: 160px">
            <ElOption label="全部" value="" />
            <ElOption label="PAP 奖励" value="pap_reward" />
            <ElOption label="管理员调整" value="admin_adjust" />
            <ElOption label="手动操作" value="manual" />
            <ElOption label="兑换消费" value="redeem" />
            <ElOption label="SRP 补损" value="srp_payout" />
            <ElOption label="商城购买" value="shop_purchase" />
          </ElSelect>
          <ElButton type="primary" @click="loadTransactions">查询</ElButton>
        </div>

        <ElTable v-loading="txLoading" :data="transactions" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="user_id" label="用户 ID" width="100" align="center" />
          <ElTableColumn prop="amount" label="金额" width="140" align="right">
            <template #default="{ row }">
              <span :class="row.amount >= 0 ? 'text-green-600' : 'text-red-500'" class="font-medium">
                {{ row.amount >= 0 ? '+' : '' }}{{ formatISK(row.amount) }}
              </span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="balance_after" label="余额" width="140" align="right">
            <template #default="{ row }">
              {{ formatISK(row.balance_after) }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="reason" label="原因" min-width="200" />
          <ElTableColumn prop="ref_type" label="类型" width="120" align="center">
            <template #default="{ row }">
              <ElTag size="small" :type="getRefTypeTag(row.ref_type)">{{ getRefTypeLabel(row.ref_type) }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="operator_id" label="操作人" width="100" align="center">
            <template #default="{ row }">
              {{ row.operator_id === 0 ? '系统' : `#${row.operator_id}` }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="created_at" label="时间" width="200">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </ElTableColumn>
        </ElTable>

        <div v-if="txPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="txPagination.current"
            v-model:page-size="txPagination.size"
            :total="txPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @size-change="() => { txPagination.current = 1; loadTransactions() }"
            @current-change="loadTransactions"
          />
        </div>
      </ElTabPane>

      <!-- 操作日志 -->
      <ElTabPane label="操作日志" name="logs">
        <div class="flex items-center gap-4 mb-4 flex-wrap">
          <ElInput
            v-model="logFilterTargetUid"
            placeholder="目标用户 ID"
            clearable
            style="width: 160px"
          />
          <ElInput
            v-model="logFilterOperatorId"
            placeholder="操作人 ID"
            clearable
            style="width: 160px"
          />
          <ElSelect v-model="logFilterAction" placeholder="操作类型" clearable style="width: 140px">
            <ElOption label="全部" value="" />
            <ElOption label="增加" value="add" />
            <ElOption label="扣减" value="deduct" />
            <ElOption label="设置" value="set" />
          </ElSelect>
          <ElButton type="primary" @click="loadLogs">查询</ElButton>
        </div>

        <ElTable v-loading="logLoading" :data="logs" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="operator_id" label="操作人" width="100" align="center">
            <template #default="{ row }">
              #{{ row.operator_id }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="target_uid" label="目标用户" width="100" align="center">
            <template #default="{ row }">
              #{{ row.target_uid }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="action" label="操作" width="100" align="center">
            <template #default="{ row }">
              <ElTag size="small" :type="getActionTag(row.action)">{{ getActionLabel(row.action) }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="amount" label="金额" width="140" align="right">
            <template #default="{ row }">
              {{ formatISK(row.amount) }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="before" label="操作前余额" width="140" align="right">
            <template #default="{ row }">
              {{ formatISK(row.before) }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="after" label="操作后余额" width="140" align="right">
            <template #default="{ row }">
              {{ formatISK(row.after) }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="reason" label="原因" min-width="200" />
          <ElTableColumn prop="created_at" label="时间" width="200">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </ElTableColumn>
        </ElTable>

        <div v-if="logPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="logPagination.current"
            v-model:page-size="logPagination.size"
            :total="logPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @size-change="() => { logPagination.current = 1; loadLogs() }"
            @current-change="loadLogs"
          />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- 调整余额弹窗 -->
    <ElDialog v-model="adjustDialogVisible" title="调整用户钱包" width="480px" destroy-on-close>
      <ElForm ref="adjustFormRef" :model="adjustForm" :rules="adjustRules" label-width="100px">
        <ElFormItem label="目标用户 ID" prop="target_uid">
          <ElInputNumber v-model="adjustForm.target_uid" :min="1" :controls="false" style="width: 100%" />
        </ElFormItem>
        <ElFormItem label="操作类型" prop="action">
          <ElRadioGroup v-model="adjustForm.action">
            <ElRadio value="add">增加</ElRadio>
            <ElRadio value="deduct">扣减</ElRadio>
            <ElRadio value="set">设为</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
        <ElFormItem label="金额" prop="amount">
          <ElInputNumber v-model="adjustForm.amount" :min="0.01" :precision="2" :controls="false" style="width: 100%" />
        </ElFormItem>
        <ElFormItem label="操作原因" prop="reason">
          <ElInput v-model="adjustForm.reason" type="textarea" :rows="3" placeholder="请说明操作原因（必填）" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="adjustDialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="adjustLoading" @click="submitAdjust">确认</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    ElCard,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElPagination,
    ElTabs,
    ElTabPane,
    ElInput,
    ElInputNumber,
    ElSelect,
    ElOption,
    ElDialog,
    ElForm,
    ElFormItem,
    ElRadioGroup,
    ElRadio,
    ElMessage,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import {
    adminListWallets,
    adminAdjustWallet,
    adminListTransactions,
    adminListWalletLogs
  } from '@/api/sys-wallet'

  defineOptions({ name: 'SystemWallet' })

  // ---- Tab ----
  const activeTab = ref('wallets')

  // ---- 格式化 ----
  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  // ---- 类型标签 ----
  const REF_TYPE_MAP: Record<string, { label: string; tag: string }> = {
    pap_reward: { label: 'PAP 奖励', tag: 'success' },
    admin_adjust: { label: '管理员调整', tag: 'warning' },
    manual: { label: '手动操作', tag: '' },
    redeem: { label: '兑换消费', tag: 'danger' },
    srp_payout: { label: 'SRP 补损', tag: 'primary' },
    shop_purchase: { label: '商城购买', tag: 'info' }
  }
  const getRefTypeLabel = (t: string) => REF_TYPE_MAP[t]?.label ?? t
  const getRefTypeTag = (t: string): any => REF_TYPE_MAP[t]?.tag ?? 'info'

  const ACTION_MAP: Record<string, { label: string; tag: string }> = {
    add: { label: '增加', tag: 'success' },
    deduct: { label: '扣减', tag: 'danger' },
    set: { label: '设置', tag: 'warning' }
  }
  const getActionLabel = (a: string) => ACTION_MAP[a]?.label ?? a
  const getActionTag = (a: string): any => ACTION_MAP[a]?.tag ?? 'info'

  // ═══════════════════════════════════════════
  //  钱包列表
  // ═══════════════════════════════════════════
  const wallets = ref<Api.SysWallet.Wallet[]>([])
  const walletLoading = ref(false)
  const searchUserId = ref('')
  const walletPagination = reactive({ current: 1, size: 20, total: 0 })

  const loadWallets = async () => {
    walletLoading.value = true
    try {
      const res = await adminListWallets({
        current: walletPagination.current,
        size: walletPagination.size
      })
      if (res) {
        wallets.value = res.list ?? []
        walletPagination.total = res.total ?? 0
      }
    } catch {
      wallets.value = []
    } finally {
      walletLoading.value = false
    }
  }

  // ═══════════════════════════════════════════
  //  流水查询
  // ═══════════════════════════════════════════
  const transactions = ref<Api.SysWallet.WalletTransaction[]>([])
  const txLoading = ref(false)
  const txFilterUserId = ref('')
  const txFilterRefType = ref('')
  const txPagination = reactive({ current: 1, size: 20, total: 0 })

  const showUserTransactions = (userId: number) => {
    txFilterUserId.value = String(userId)
    activeTab.value = 'transactions'
    nextTick(() => loadTransactions())
  }

  const loadTransactions = async () => {
    txLoading.value = true
    try {
      const data: Api.SysWallet.TransactionSearchParams = {
        current: txPagination.current,
        size: txPagination.size
      }
      if (txFilterUserId.value) data.user_id = Number(txFilterUserId.value)
      if (txFilterRefType.value) data.ref_type = txFilterRefType.value

      const res = await adminListTransactions(data)
      if (res) {
        transactions.value = res.list ?? []
        txPagination.total = res.total ?? 0
      }
    } catch {
      transactions.value = []
    } finally {
      txLoading.value = false
    }
  }

  // ═══════════════════════════════════════════
  //  操作日志
  // ═══════════════════════════════════════════
  const logs = ref<Api.SysWallet.WalletLog[]>([])
  const logLoading = ref(false)
  const logFilterTargetUid = ref('')
  const logFilterOperatorId = ref('')
  const logFilterAction = ref('')
  const logPagination = reactive({ current: 1, size: 20, total: 0 })

  const loadLogs = async () => {
    logLoading.value = true
    try {
      const data: Api.SysWallet.LogSearchParams = {
        current: logPagination.current,
        size: logPagination.size
      }
      if (logFilterTargetUid.value) data.target_uid = Number(logFilterTargetUid.value)
      if (logFilterOperatorId.value) data.operator_id = Number(logFilterOperatorId.value)
      if (logFilterAction.value) data.action = logFilterAction.value

      const res = await adminListWalletLogs(data)
      if (res) {
        logs.value = res.list ?? []
        logPagination.total = res.total ?? 0
      }
    } catch {
      logs.value = []
    } finally {
      logLoading.value = false
    }
  }

  // ═══════════════════════════════════════════
  //  调整余额弹窗
  // ═══════════════════════════════════════════
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
    target_uid: [{ required: true, message: '请输入目标用户 ID', trigger: 'blur' }],
    action: [{ required: true, message: '请选择操作类型', trigger: 'change' }],
    amount: [{ required: true, message: '请输入金额', trigger: 'blur' }],
    reason: [{ required: true, message: '请输入操作原因', trigger: 'blur' }]
  }

  const showAdjustDialog = (userId?: number, action?: 'add' | 'deduct' | 'set') => {
    adjustForm.target_uid = userId ?? 0
    adjustForm.action = action ?? 'add'
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
      ElMessage.success('余额调整成功')
      adjustDialogVisible.value = false
      loadWallets()
    } catch (e: any) {
      ElMessage.error(e?.message ?? '操作失败')
    } finally {
      adjustLoading.value = false
    }
  }

  // ---- 初始化 ----
  onMounted(() => {
    loadWallets()
  })
</script>

<style scoped>
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
