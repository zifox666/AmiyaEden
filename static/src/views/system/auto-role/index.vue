<!-- ESI 自动权限映射管理 -->
<template>
  <div class="art-full-height">
    <!-- 页头操作区 -->
    <ElCard class="mb-4" shadow="never">
      <div class="flex items-center justify-between">
        <div>
          <div class="text-base font-semibold">{{ t('autoRolePage.title') }}</div>
          <div class="mt-1 text-sm text-gray-500">
            {{ t('autoRolePage.description') }}<br />
            {{ t('autoRolePage.directorHint') }}
          </div>
        </div>
        <ElButton type="primary" :loading="syncLoading" @click="handleTriggerSync">
          {{ t('autoRolePage.triggerSync') }}
        </ElButton>
      </div>
    </ElCard>

    <ElTabs v-model="activeTab" type="border-card">
      <!-- ─── Tab 1：军团职位映射 ─── -->
      <ElTabPane :label="t('autoRolePage.tabs.esiRole')" name="esi-role">
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm text-gray-500">
            {{ t('autoRolePage.descriptions.esiRole') }}
          </span>
          <ElButton type="primary" :icon="Plus" @click="openEsiRoleDialog">
            {{ t('autoRolePage.addMapping') }}
          </ElButton>
        </div>

        <ElTable v-loading="esiRoleLoading" :data="esiRoleMappings" border stripe>
          <ElTableColumn :label="t('autoRolePage.columns.index')" type="index" width="60" />
          <ElTableColumn :label="t('autoRolePage.columns.esiRole')" prop="esi_role" min-width="200">
            <template #default="{ row }">
              <ElTag size="small" type="warning" effect="plain">{{ row.esi_role }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('autoRolePage.columns.mappedRole')" min-width="200">
            <template #default="{ row }">
              <ElTag size="small" :type="getRoleTagType(row.role_code)" effect="dark">
                {{ row.role_name || row.role_code }}
              </ElTag>
              <span class="ml-1 text-xs text-gray-400">{{ row.role_code }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('common.createdAt')" prop="created_at" width="180">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('common.operation')" width="100" fixed="right">
            <template #default="{ row }">
              <ElPopconfirm
                :title="t('autoRolePage.deleteConfirm')"
                :confirm-button-text="t('common.delete')"
                :cancel-button-text="t('common.cancel')"
                @confirm="handleDeleteEsiRole(row.id)"
              >
                <template #reference>
                  <ElButton size="small" type="danger" plain>{{ t('common.delete') }}</ElButton>
                </template>
              </ElPopconfirm>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <!-- ─── Tab 2：头衔映射 ─── -->
      <ElTabPane :label="t('autoRolePage.tabs.title')" name="title">
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm text-gray-500">
            {{ t('autoRolePage.descriptions.title') }}
          </span>
          <ElButton type="primary" :icon="Plus" @click="openTitleDialog">
            {{ t('autoRolePage.addMapping') }}
          </ElButton>
        </div>

        <ElTable v-loading="titleLoading" :data="titleMappings" border stripe>
          <ElTableColumn :label="t('autoRolePage.columns.index')" type="index" width="60" />
          <ElTableColumn
            :label="t('autoRolePage.columns.corporationId')"
            prop="corporation_id"
            width="160"
          />
          <ElTableColumn :label="t('autoRolePage.columns.titleId')" prop="title_id" width="100" />
          <ElTableColumn
            :label="t('autoRolePage.columns.titleName')"
            prop="title_name"
            min-width="180"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <span v-if="row.title_name" class="text-orange-400">{{ row.title_name }}</span>
              <span v-else class="text-gray-400">—</span>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('autoRolePage.columns.mappedRole')" min-width="180">
            <template #default="{ row }">
              <ElTag size="small" :type="getRoleTagType(row.role_code)" effect="dark">
                {{ row.role_name || row.role_code }}
              </ElTag>
              <span class="ml-1 text-xs text-gray-400">{{ row.role_code }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('common.createdAt')" prop="created_at" width="180">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('common.operation')" width="100" fixed="right">
            <template #default="{ row }">
              <ElPopconfirm
                :title="t('autoRolePage.deleteConfirm')"
                :confirm-button-text="t('common.delete')"
                :cancel-button-text="t('common.cancel')"
                @confirm="handleDeleteTitle(row.id)"
              >
                <template #reference>
                  <ElButton size="small" type="danger" plain>{{ t('common.delete') }}</ElButton>
                </template>
              </ElPopconfirm>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <!-- ─── Tab 3：同步日志 ─── -->
      <ElTabPane :label="t('autoRolePage.tabs.log')" name="log">
        <ElTable v-loading="logLoading" :data="logs" border stripe>
          <ElTableColumn :label="t('autoRolePage.columns.index')" type="index" width="60" />
          <ElTableColumn :label="t('autoRolePage.columns.userId')" prop="user_id" width="100" />
          <ElTableColumn :label="t('autoRolePage.columns.username')" prop="username" min-width="140" show-overflow-tooltip />
          <ElTableColumn :label="t('autoRolePage.columns.roleName')" min-width="160">
            <template #default="{ row }">
              <ElTag size="small" :type="getRoleTagType(row.role_code)" effect="dark">
                {{ row.role_name || row.role_code }}
              </ElTag>
              <span class="ml-1 text-xs text-gray-400">{{ row.role_code }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('autoRolePage.columns.action')" width="100">
            <template #default="{ row }">
              <ElTag size="small" :type="row.action === 'add' ? 'success' : 'danger'" effect="plain">
                {{ t(`autoRolePage.actions.${row.action}`) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('common.createdAt')" prop="created_at" width="180">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </ElTableColumn>
        </ElTable>
        <div class="flex justify-end mt-3">
          <ElPagination
            v-model:current-page="logPage"
            :page-size="logSize"
            :total="logTotal"
            layout="total, prev, pager, next"
            background
            @current-change="onLogPageChange"
          />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- ─── 新增 ESI 角色映射对话框 ─── -->
    <ElDialog
      v-model="esiRoleDialogVisible"
      :title="t('autoRolePage.createEsiRoleTitle')"
      width="460px"
      destroy-on-close
    >
      <ElForm
        ref="esiRoleFormRef"
        :model="esiRoleForm"
        :rules="esiRoleFormRules"
        label-width="100px"
      >
        <ElFormItem :label="t('autoRolePage.fields.esiRole')" prop="esi_role">
          <ElSelect
            v-model="esiRoleForm.esi_role"
            :placeholder="t('autoRolePage.placeholders.esiRole')"
            filterable
            style="width: 100%"
          >
            <ElOption v-for="role in allEsiRoles" :key="role" :label="role" :value="role" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('autoRolePage.fields.systemRole')" prop="role_id">
          <ElSelect
            v-model="esiRoleForm.role_id"
            :placeholder="t('autoRolePage.placeholders.systemRole')"
            filterable
            style="width: 100%"
          >
            <ElOption
              v-for="role in allSystemRoles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            >
              <span>{{ role.name }}</span>
              <span class="ml-2 text-xs text-gray-400">{{ role.code }}</span>
            </ElOption>
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="esiRoleDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="esiRoleSubmitting" @click="handleCreateEsiRole">
          {{ t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>

    <!-- ─── 新增头衔映射对话框 ─── -->
    <ElDialog
      v-model="titleDialogVisible"
      :title="t('autoRolePage.createTitleTitle')"
      width="480px"
      destroy-on-close
    >
      <ElForm ref="titleFormRef" :model="titleForm" :rules="titleFormRules" label-width="100px">
        <ElFormItem :label="t('autoRolePage.fields.title')" prop="title_key">
          <ElSelect
            v-model="titleForm.title_key"
            filterable
            :placeholder="t('autoRolePage.placeholders.title')"
            style="width: 100%"
            @change="onTitleKeyChange"
          >
            <ElOption
              v-for="t in allCorpTitles"
              :key="`${t.corporation_id}_${t.title_id}`"
              :label="t.title_name || $t('autoRolePage.titleFallback', { id: t.title_id })"
              :value="`${t.corporation_id}_${t.title_id}`"
            >
              <div class="flex items-center justify-between">
                <span>{{
                  t.title_name || $t('autoRolePage.titleFallback', { id: t.title_id })
                }}</span>
                <span class="ml-3 text-xs text-gray-400">
                  {{ $t('autoRolePage.corpPrefix', { id: t.corporation_id }) }}
                </span>
              </div>
            </ElOption>
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('autoRolePage.fields.systemRole')" prop="role_id">
          <ElSelect
            v-model="titleForm.role_id"
            :placeholder="t('autoRolePage.placeholders.systemRole')"
            filterable
            style="width: 100%"
          >
            <ElOption
              v-for="role in allSystemRoles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            >
              <span>{{ role.name }}</span>
              <span class="ml-2 text-xs text-gray-400">{{ role.code }}</span>
            </ElOption>
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="titleDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="titleSubmitting" @click="handleCreateTitle">{{
          t('common.confirm')
        }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { Plus } from '@element-plus/icons-vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import {
    fetchGetEsiRoleMappings,
    fetchCreateEsiRoleMapping,
    fetchDeleteEsiRoleMapping,
    fetchGetEsiTitleMappings,
    fetchCreateEsiTitleMapping,
    fetchDeleteEsiTitleMapping,
    fetchGetAllEsiRoles,
    fetchGetCorpTitles,
    fetchGetAllRoles,
    fetchTriggerAutoRoleSync,
    fetchGetAutoRoleLogs
  } from '@/api/system-manage'

  defineOptions({ name: 'AutoRole' })
  const { t } = useI18n()

  type EsiRoleMapping = Api.SystemManage.EsiRoleMapping
  type EsiTitleMapping = Api.SystemManage.EsiTitleMapping
  type RoleItem = Api.SystemManage.RoleItem
  type CorpTitleInfo = Api.SystemManage.CorpTitleInfo
  type AutoRoleLog = Api.SystemManage.AutoRoleLog

  // ─── Tab ───
  const activeTab = ref('esi-role')

  // ─── 角色标签颜色 ───
  const CODE_TYPE: Record<string, string> = {
    super_admin: 'danger',
    admin: 'warning',
    srp: '',
    fc: '',
    user: 'success',
    guest: 'info'
  }
  function getRoleTagType(code: string) {
    return (CODE_TYPE[code] ?? '') as any
  }

  // ─── 日期格式化 ───
  function formatDate(dateStr: string) {
    if (!dateStr) return '—'
    return new Date(dateStr).toLocaleString(undefined, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  // ─── 基础数据 ───
  const allEsiRoles = ref<string[]>([])
  const allSystemRoles = ref<RoleItem[]>([])
  const allCorpTitles = ref<CorpTitleInfo[]>([])

  async function loadBaseData() {
    const [esiRoles, systemRoles, corpTitles] = await Promise.all([
      fetchGetAllEsiRoles(),
      fetchGetAllRoles(),
      fetchGetCorpTitles()
    ])
    allEsiRoles.value = esiRoles
    // 过滤掉 super_admin（文档规定不可映射）
    allSystemRoles.value = systemRoles.filter((r) => r.code !== 'super_admin')
    allCorpTitles.value = corpTitles
  }

  // ─── ESI 角色映射 ───
  const esiRoleMappings = ref<EsiRoleMapping[]>([])
  const esiRoleLoading = ref(false)

  async function loadEsiRoleMappings() {
    esiRoleLoading.value = true
    try {
      esiRoleMappings.value = await fetchGetEsiRoleMappings()
    } finally {
      esiRoleLoading.value = false
    }
  }

  async function handleDeleteEsiRole(id: number) {
    try {
      await fetchDeleteEsiRoleMapping(id)
      ElMessage.success(t('autoRolePage.deleteSuccess'))
      await loadEsiRoleMappings()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('autoRolePage.deleteFailed'))
    }
  }

  // ─── 新增 ESI 角色映射 ───
  const esiRoleDialogVisible = ref(false)
  const esiRoleSubmitting = ref(false)
  const esiRoleFormRef = ref<FormInstance>()
  const esiRoleForm = reactive({
    esi_role: '',
    role_id: undefined as number | undefined
  })
  const esiRoleFormRules: FormRules = {
    esi_role: [{ required: true, message: t('autoRolePage.rules.esiRole'), trigger: 'change' }],
    role_id: [{ required: true, message: t('autoRolePage.rules.systemRole'), trigger: 'change' }]
  }

  function openEsiRoleDialog() {
    esiRoleForm.esi_role = ''
    esiRoleForm.role_id = undefined
    esiRoleDialogVisible.value = true
  }

  async function handleCreateEsiRole() {
    if (!esiRoleFormRef.value) return
    await esiRoleFormRef.value.validate()
    esiRoleSubmitting.value = true
    try {
      await fetchCreateEsiRoleMapping({
        esi_role: esiRoleForm.esi_role,
        role_id: esiRoleForm.role_id!
      })
      ElMessage.success(t('autoRolePage.mappingCreated'))
      esiRoleDialogVisible.value = false
      await loadEsiRoleMappings()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('autoRolePage.createFailed'))
    } finally {
      esiRoleSubmitting.value = false
    }
  }

  // ─── 头衔映射 ───
  const titleMappings = ref<EsiTitleMapping[]>([])
  const titleLoading = ref(false)

  async function loadTitleMappings() {
    titleLoading.value = true
    try {
      titleMappings.value = await fetchGetEsiTitleMappings()
    } finally {
      titleLoading.value = false
    }
  }

  async function handleDeleteTitle(id: number) {
    try {
      await fetchDeleteEsiTitleMapping(id)
      ElMessage.success(t('autoRolePage.deleteSuccess'))
      await loadTitleMappings()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('autoRolePage.deleteFailed'))
    }
  }

  // ─── 新增头衔映射 ───
  const titleDialogVisible = ref(false)
  const titleSubmitting = ref(false)
  const titleFormRef = ref<FormInstance>()
  const titleForm = reactive({
    title_key: '',
    corporation_id: 0,
    title_id: 0,
    title_name: '',
    role_id: undefined as number | undefined
  })
  const titleFormRules: FormRules = {
    title_key: [{ required: true, message: t('autoRolePage.rules.title'), trigger: 'change' }],
    role_id: [{ required: true, message: t('autoRolePage.rules.systemRole'), trigger: 'change' }]
  }

  function onTitleKeyChange(key: string) {
    const t = allCorpTitles.value.find((t) => `${t.corporation_id}_${t.title_id}` === key)
    if (t) {
      titleForm.corporation_id = t.corporation_id
      titleForm.title_id = t.title_id
      titleForm.title_name = t.title_name
    }
  }

  function openTitleDialog() {
    titleForm.title_key = ''
    titleForm.corporation_id = 0
    titleForm.title_id = 0
    titleForm.title_name = ''
    titleForm.role_id = undefined
    titleDialogVisible.value = true
  }

  async function handleCreateTitle() {
    if (!titleFormRef.value) return
    await titleFormRef.value.validate()
    titleSubmitting.value = true
    try {
      await fetchCreateEsiTitleMapping({
        corporation_id: titleForm.corporation_id,
        title_id: titleForm.title_id,
        title_name: titleForm.title_name || undefined,
        role_id: titleForm.role_id!
      })
      ElMessage.success(t('autoRolePage.mappingCreated'))
      titleDialogVisible.value = false
      await loadTitleMappings()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('autoRolePage.createFailed'))
    } finally {
      titleSubmitting.value = false
    }
  }

  // ─── 手动触发同步 ───
  const syncLoading = ref(false)

  async function handleTriggerSync() {
    syncLoading.value = true
    try {
      await fetchTriggerAutoRoleSync()
      ElMessage.success(t('autoRolePage.syncTriggered'))
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('autoRolePage.syncFailed'))
    } finally {
      syncLoading.value = false
    }
  }

  // ─── 日志 ───
  const logs = ref<AutoRoleLog[]>([])
  const logLoading = ref(false)
  const logTotal = ref(0)
  const logPage = ref(1)
  const logSize = ref(20)

  async function loadLogs() {
    logLoading.value = true
    try {
      const res = await fetchGetAutoRoleLogs({ current: logPage.value, size: logSize.value })
      logs.value = res.list ?? []
      logTotal.value = res.total ?? 0
    } finally {
      logLoading.value = false
    }
  }

  function onLogPageChange(page: number) {
    logPage.value = page
    loadLogs()
  }

  // ─── 初始化 ───
  onMounted(() => {
    Promise.all([loadBaseData(), loadEsiRoleMappings(), loadTitleMappings(), loadLogs()])
  })
</script>
