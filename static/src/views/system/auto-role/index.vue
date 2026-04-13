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
          <ElTableColumn :label="t('autoRolePage.columns.onlyMainChar')" width="120" align="center">
            <template #default="{ row }">
              <ElTag size="small" :type="row.only_main_char ? 'primary' : 'info'" effect="plain">
                {{
                  row.only_main_char
                    ? t('autoRolePage.onlyMainChar.yes')
                    : t('autoRolePage.onlyMainChar.no')
                }}
              </ElTag>
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
          <ElTableColumn :label="t('autoRolePage.columns.onlyMainChar')" width="120" align="center">
            <template #default="{ row }">
              <ElTag size="small" :type="row.only_main_char ? 'primary' : 'info'" effect="plain">
                {{
                  row.only_main_char
                    ? t('autoRolePage.onlyMainChar.yes')
                    : t('autoRolePage.onlyMainChar.no')
                }}
              </ElTag>
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

      <!-- ─── Tab 3：SeAT 分组映射 ─── -->
      <ElTabPane :label="t('autoRolePage.tabs.seatRole')" name="seat-role">
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm text-gray-500">
            {{ t('autoRolePage.descriptions.seatRole') }}
          </span>
          <ElButton type="primary" :icon="Plus" @click="openSeatRoleDialog">
            {{ t('autoRolePage.addMapping') }}
          </ElButton>
        </div>

        <ElTable v-loading="seatRoleLoading" :data="seatRoleMappings" border stripe>
          <ElTableColumn :label="t('autoRolePage.columns.index')" type="index" width="60" />
          <ElTableColumn
            :label="t('autoRolePage.columns.seatRole')"
            prop="seat_role"
            min-width="200"
          >
            <template #default="{ row }">
              <ElTag size="small" type="success" effect="plain">{{ row.seat_role }}</ElTag>
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
                @confirm="handleDeleteSeatRole(row.id)"
              >
                <template #reference>
                  <ElButton size="small" type="danger" plain>{{ t('common.delete') }}</ElButton>
                </template>
              </ElPopconfirm>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <!-- ─── Tab 4：同步日志 ─── -->
      <ElTabPane :label="t('autoRolePage.tabs.log')" name="log">
        <ElTable v-loading="logLoading" :data="logs" border stripe>
          <ElTableColumn :label="t('autoRolePage.columns.index')" type="index" width="60" />
          <ElTableColumn :label="t('autoRolePage.columns.userId')" prop="user_id" width="100" />
          <ElTableColumn
            :label="t('autoRolePage.columns.username')"
            prop="username"
            min-width="140"
            show-overflow-tooltip
          />
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
              <ElTag
                size="small"
                :type="row.action === 'add' ? 'success' : 'danger'"
                effect="plain"
              >
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

      <!-- Tab 4: 准入配置 -->
      <ElTabPane :label="t('autoRolePage.tabs.allowList')" name="allow-list">
        <!-- auto_role 列表 -->
        <ElCard class="mb-4" shadow="never">
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-semibold">{{ t('autoRolePage.allowList.autoRoleTitle') }}</span>
              <div class="flex items-center gap-3">
                <span class="text-sm text-gray-500">{{ t('autoRolePage.allowList.onlyMainChar') }}</span>
                <ElSwitch
                  v-model="allowListConfig.auto_role_only_main_char"
                  @change="saveAllowListOnlyMainCharConfig"
                />
                <ElButton type="primary" :icon="Plus" @click="openAllowDialog('auto_role')">
                  {{ t('autoRolePage.allowList.addEntity') }}
                </ElButton>
              </div>
            </div>
          </template>
          <p class="text-sm text-gray-400 mb-3">{{ t('autoRolePage.allowList.autoRoleDesc') }}</p>
          <ElTable :data="autoRoleEntities" v-loading="autoRoleLoading" border>
            <ElTableColumn :label="t('autoRolePage.allowList.columns.type')" width="110">
              <template #default="{ row }">
                <ElTag :type="row.entity_type === 'alliance' ? 'warning' : undefined">
                  {{ t(`autoRolePage.allowList.entityTypes.${row.entity_type}`) }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn :label="t('autoRolePage.allowList.columns.name')" prop="entity_name" />
            <ElTableColumn
              :label="t('autoRolePage.allowList.columns.entityId')"
              prop="entity_id"
              width="130"
            />
            <ElTableColumn
              :label="t('autoRolePage.allowList.columns.createdAt')"
              prop="created_at"
              width="160"
            >
              <template #default="{ row }">{{
                row.created_at?.slice(0, 19).replace('T', ' ')
              }}</template>
            </ElTableColumn>
            <ElTableColumn :label="t('common.actions')" width="90" align="center">
              <template #default="{ row }">
                <ElPopconfirm
                  :title="t('autoRolePage.allowList.removeConfirm')"
                  @confirm="handleRemoveAllowedEntity(row.id, 'auto_role')"
                >
                  <template #reference>
                    <ElButton type="danger" link size="small">{{ t('common.delete') }}</ElButton>
                  </template>
                </ElPopconfirm>
              </template>
            </ElTableColumn>
          </ElTable>
        </ElCard>

        <!-- basic_access 列表 -->
        <ElCard shadow="never">
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-semibold">{{ t('autoRolePage.allowList.basicAccessTitle') }}</span>
              <div class="flex items-center gap-3">
                <span class="text-sm text-gray-500">{{ t('autoRolePage.allowList.onlyMainChar') }}</span>
                <ElSwitch
                  v-model="allowListConfig.basic_access_only_main_char"
                  @change="saveAllowListOnlyMainCharConfig"
                />
                <ElButton type="primary" :icon="Plus" @click="openAllowDialog('basic_access')">
                  {{ t('autoRolePage.allowList.addEntity') }}
                </ElButton>
              </div>
            </div>
          </template>
          <p class="text-sm text-gray-400 mb-3">{{
            t('autoRolePage.allowList.basicAccessDesc')
          }}</p>
          <ElTable :data="basicAccessEntities" v-loading="basicAccessLoading" border>
            <ElTableColumn :label="t('autoRolePage.allowList.columns.type')" width="110">
              <template #default="{ row }">
                <ElTag :type="row.entity_type === 'alliance' ? 'warning' : undefined">
                  {{ t(`autoRolePage.allowList.entityTypes.${row.entity_type}`) }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn :label="t('autoRolePage.allowList.columns.name')" prop="entity_name" />
            <ElTableColumn
              :label="t('autoRolePage.allowList.columns.entityId')"
              prop="entity_id"
              width="130"
            />
            <ElTableColumn
              :label="t('autoRolePage.allowList.columns.createdAt')"
              prop="created_at"
              width="160"
            >
              <template #default="{ row }">{{
                row.created_at?.slice(0, 19).replace('T', ' ')
              }}</template>
            </ElTableColumn>
            <ElTableColumn :label="t('common.actions')" width="90" align="center">
              <template #default="{ row }">
                <ElPopconfirm
                  :title="t('autoRolePage.allowList.removeConfirm')"
                  @confirm="handleRemoveAllowedEntity(row.id, 'basic_access')"
                >
                  <template #reference>
                    <ElButton type="danger" link size="small">{{ t('common.delete') }}</ElButton>
                  </template>
                </ElPopconfirm>
              </template>
            </ElTableColumn>
          </ElTable>
        </ElCard>
      </ElTabPane>
    </ElTabs>

    <!-- ─── 新增 EVE 实体对话框（准入名单） ─── -->
    <ElDialog
      v-model="allowDialogVisible"
      :title="t('autoRolePage.allowList.addEntity')"
      width="520px"
      destroy-on-close
    >
      <ElInput
        v-model="allowSearchQuery"
        :placeholder="t('autoRolePage.allowList.searchPlaceholder')"
        clearable
        @input="onAllowSearch"
      />
      <div class="mt-3" style="max-height: 320px; overflow-y: auto">
        <div v-if="allowSearchLoading" class="text-center text-gray-400 py-4">
          {{ t('common.loading') }}
        </div>
        <div
          v-else-if="allowSearchResults.length === 0 && allowSearchQuery"
          class="text-center text-gray-400 py-4"
        >
          {{ t('autoRolePage.allowList.noResults') }}
        </div>
        <div
          v-for="item in allowSearchResults"
          :key="item.id"
          class="flex items-center gap-3 p-2 rounded cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
          :class="{ 'ring-2 ring-primary': allowSelected?.id === item.id }"
          @click="allowSelected = item"
        >
          <img :src="item.image" class="w-8 h-8 rounded" alt="" />
          <div class="flex-1 min-w-0">
            <div class="text-sm font-medium truncate">{{ item.name }}</div>
            <div class="text-xs text-gray-400">{{ item.id }}</div>
          </div>
          <ElTag size="small" :type="item.type === 'alliance' ? 'warning' : undefined">
            {{ t(`autoRolePage.allowList.entityTypes.${item.type}`) }}
          </ElTag>
        </div>
      </div>
      <template #footer>
        <ElButton @click="allowDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton
          type="primary"
          :disabled="!allowSelected"
          :loading="allowAdding"
          @click="handleAddAllowedEntity"
        >
          {{ t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>

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
        <ElFormItem :label="t('autoRolePage.fields.onlyMainChar')">
          <ElSwitch v-model="esiRoleForm.only_main_char" />
          <span class="ml-2 text-xs text-gray-400">{{ t('autoRolePage.onlyMainChar.hint') }}</span>
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
        <ElFormItem :label="t('autoRolePage.fields.onlyMainChar')">
          <ElSwitch v-model="titleForm.only_main_char" />
          <span class="ml-2 text-xs text-gray-400">{{ t('autoRolePage.onlyMainChar.hint') }}</span>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="titleDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="titleSubmitting" @click="handleCreateTitle">{{
          t('common.confirm')
        }}</ElButton>
      </template>
    </ElDialog>

    <!-- ─── 新增 SeAT 分组映射对话框 ─── -->
    <ElDialog
      v-model="seatRoleDialogVisible"
      :title="t('autoRolePage.createSeatRoleTitle')"
      width="460px"
      destroy-on-close
    >
      <ElForm
        ref="seatRoleFormRef"
        :model="seatRoleForm"
        :rules="seatRoleFormRules"
        label-width="110px"
      >
        <ElFormItem :label="t('autoRolePage.fields.seatRole')" prop="seat_role">
          <ElSelect
            v-model="seatRoleForm.seat_role"
            :placeholder="t('autoRolePage.placeholders.seatRole')"
            filterable
            allow-create
            style="width: 100%"
          >
            <ElOption v-for="role in allSeatRoles" :key="role" :label="role" :value="role" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('autoRolePage.fields.systemRole')" prop="role_id">
          <ElSelect
            v-model="seatRoleForm.role_id"
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
        <ElButton @click="seatRoleDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="seatRoleSubmitting" @click="handleCreateSeatRole">
          {{ t('common.confirm') }}
        </ElButton>
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
    fetchGetAutoRoleLogs,
    fetchGetAllowedEntities,
    fetchAddAllowedEntity,
    fetchRemoveAllowedEntity,
    fetchSearchEveEntities,
    fetchGetAllowListOnlyMainCharConfig,
    fetchUpdateAllowListOnlyMainCharConfig,
    fetchGetAllSeatRoles,
    fetchGetSeatRoleMappings,
    fetchCreateSeatRoleMapping,
    fetchDeleteSeatRoleMapping
  } from '@/api/system-manage'

  defineOptions({ name: 'AutoRole' })
  const { t } = useI18n()

  type EsiRoleMapping = Api.SystemManage.EsiRoleMapping
  type EsiTitleMapping = Api.SystemManage.EsiTitleMapping
  type SeatRoleMapping = Api.SystemManage.SeatRoleMapping
  type RoleItem = Api.SystemManage.RoleItem
  type CorpTitleInfo = Api.SystemManage.CorpTitleInfo
  type AutoRoleLog = Api.SystemManage.AutoRoleLog
  type AllowedEntity = Api.SystemManage.AllowedEntity
  type ZkbSearchResult = Api.SystemManage.ZkbSearchResult

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
  const allSeatRoles = ref<string[]>([])

  async function loadBaseData() {
    const [esiRoles, systemRoles, corpTitles, seatRoles] = await Promise.all([
      fetchGetAllEsiRoles(),
      fetchGetAllRoles(),
      fetchGetCorpTitles(),
      fetchGetAllSeatRoles()
    ])
    allEsiRoles.value = esiRoles
    // 过滤掉 super_admin（文档规定不可映射）
    allSystemRoles.value = systemRoles.filter((r) => r.code !== 'super_admin')
    allCorpTitles.value = corpTitles
    allSeatRoles.value = seatRoles
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
    role_id: undefined as number | undefined,
    only_main_char: true
  })
  const esiRoleFormRules: FormRules = {
    esi_role: [{ required: true, message: t('autoRolePage.rules.esiRole'), trigger: 'change' }],
    role_id: [{ required: true, message: t('autoRolePage.rules.systemRole'), trigger: 'change' }]
  }

  function openEsiRoleDialog() {
    esiRoleForm.esi_role = ''
    esiRoleForm.role_id = undefined
    esiRoleForm.only_main_char = true
    esiRoleDialogVisible.value = true
  }

  async function handleCreateEsiRole() {
    if (!esiRoleFormRef.value) return
    await esiRoleFormRef.value.validate()
    esiRoleSubmitting.value = true
    try {
      await fetchCreateEsiRoleMapping({
        esi_role: esiRoleForm.esi_role,
        role_id: esiRoleForm.role_id!,
        only_main_char: esiRoleForm.only_main_char
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
    role_id: undefined as number | undefined,
    only_main_char: true
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
    titleForm.only_main_char = true
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
        role_id: titleForm.role_id!,
        only_main_char: titleForm.only_main_char
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

  // ─── SeAT 分组映射 ───
  const seatRoleMappings = ref<SeatRoleMapping[]>([])
  const seatRoleLoading = ref(false)

  async function loadSeatRoleMappings() {
    seatRoleLoading.value = true
    try {
      seatRoleMappings.value = await fetchGetSeatRoleMappings()
    } finally {
      seatRoleLoading.value = false
    }
  }

  async function handleDeleteSeatRole(id: number) {
    try {
      await fetchDeleteSeatRoleMapping(id)
      ElMessage.success(t('autoRolePage.deleteSuccess'))
      await loadSeatRoleMappings()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('autoRolePage.deleteFailed'))
    }
  }

  // ─── 新增 SeAT 分组映射 ───
  const seatRoleDialogVisible = ref(false)
  const seatRoleSubmitting = ref(false)
  const seatRoleFormRef = ref<FormInstance>()
  const seatRoleForm = reactive({
    seat_role: '',
    role_id: undefined as number | undefined
  })
  const seatRoleFormRules: FormRules = {
    seat_role: [{ required: true, message: t('autoRolePage.rules.seatRole'), trigger: 'change' }],
    role_id: [{ required: true, message: t('autoRolePage.rules.systemRole'), trigger: 'change' }]
  }

  function openSeatRoleDialog() {
    seatRoleForm.seat_role = ''
    seatRoleForm.role_id = undefined
    seatRoleDialogVisible.value = true
  }

  async function handleCreateSeatRole() {
    if (!seatRoleFormRef.value) return
    await seatRoleFormRef.value.validate()
    seatRoleSubmitting.value = true
    try {
      await fetchCreateSeatRoleMapping({
        seat_role: seatRoleForm.seat_role,
        role_id: seatRoleForm.role_id!
      })
      ElMessage.success(t('autoRolePage.mappingCreated'))
      seatRoleDialogVisible.value = false
      await loadSeatRoleMappings()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('autoRolePage.createFailed'))
    } finally {
      seatRoleSubmitting.value = false
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

  // ─── 准入名单 ───
  const autoRoleEntities = ref<AllowedEntity[]>([])
  const basicAccessEntities = ref<AllowedEntity[]>([])
  const autoRoleLoading = ref(false)
  const basicAccessLoading = ref(false)

  // 准入名单"仅主角色"配置
  const allowListConfig = ref<Api.SystemManage.AllowListOnlyMainCharConfig>({
    auto_role_only_main_char: false,
    basic_access_only_main_char: false
  })

  async function loadAllowListOnlyMainCharConfig() {
    const res = await fetchGetAllowListOnlyMainCharConfig()
    if (res) allowListConfig.value = res
  }

  async function saveAllowListOnlyMainCharConfig() {
    await fetchUpdateAllowListOnlyMainCharConfig(allowListConfig.value)
  }

  // 搜索对话框
  const allowDialogVisible = ref(false)
  const allowDialogListType = ref<'auto_role' | 'basic_access'>('auto_role')
  const allowSearchQuery = ref('')
  const allowSearchResults = ref<ZkbSearchResult[]>([])
  const allowSearchLoading = ref(false)
  const allowSelected = ref<ZkbSearchResult | null>(null)
  const allowAdding = ref(false)

  let allowSearchTimer: ReturnType<typeof setTimeout> | null = null

  async function loadAutoRoleEntities() {
    autoRoleLoading.value = true
    try {
      const res = await fetchGetAllowedEntities('auto_role')
      autoRoleEntities.value = res ?? []
    } finally {
      autoRoleLoading.value = false
    }
  }

  async function loadBasicAccessEntities() {
    basicAccessLoading.value = true
    try {
      const res = await fetchGetAllowedEntities('basic_access')
      basicAccessEntities.value = res ?? []
    } finally {
      basicAccessLoading.value = false
    }
  }

  async function handleRemoveAllowedEntity(id: number, listType: 'auto_role' | 'basic_access') {
    try {
      await fetchRemoveAllowedEntity(listType, id)
      ElMessage.success(t('autoRolePage.allowList.removeSuccess'))
      if (listType === 'auto_role') loadAutoRoleEntities()
      else loadBasicAccessEntities()
    } catch {
      ElMessage.error(t('autoRolePage.allowList.removeFailed'))
    }
  }

  function openAllowDialog(listType: 'auto_role' | 'basic_access') {
    allowDialogListType.value = listType
    allowDialogVisible.value = true
    allowSearchQuery.value = ''
    allowSearchResults.value = []
    allowSelected.value = null
  }

  function onAllowSearch() {
    if (allowSearchTimer) clearTimeout(allowSearchTimer)
    allowSearchTimer = setTimeout(async () => {
      const q = allowSearchQuery.value.trim()
      if (!q) {
        allowSearchResults.value = []
        return
      }
      allowSearchLoading.value = true
      try {
        const res = await fetchSearchEveEntities(q)
        allowSearchResults.value = res ?? []
      } finally {
        allowSearchLoading.value = false
      }
    }, 400)
  }

  async function handleAddAllowedEntity() {
    if (!allowSelected.value) return
    allowAdding.value = true
    try {
      await fetchAddAllowedEntity(allowDialogListType.value, {
        entity_id: allowSelected.value.id,
        entity_type: allowSelected.value.type,
        entity_name: allowSelected.value.name
      })
      ElMessage.success(t('autoRolePage.allowList.addSuccess'))
      allowDialogVisible.value = false
      if (allowDialogListType.value === 'auto_role') loadAutoRoleEntities()
      else loadBasicAccessEntities()
    } catch {
      ElMessage.error(t('autoRolePage.allowList.addFailed'))
    } finally {
      allowAdding.value = false
    }
  }

  // ─── 初始化 ───
  onMounted(() => {
    Promise.all([
      loadBaseData(),
      loadEsiRoleMappings(),
      loadTitleMappings(),
      loadSeatRoleMappings(),
      loadLogs(),
      loadAutoRoleEntities(),
      loadBasicAccessEntities(),
      loadAllowListOnlyMainCharConfig()
    ])
  })
</script>
