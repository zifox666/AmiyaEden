<!-- 用户管理页面 -->
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <UserSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></UserSearch>

    <ElCard v-if="isSuperAdmin" class="mb-4" shadow="never">
      <div class="user-page__config-card">
        <div>
          <div class="user-page__config-title">
            {{ t('userAdmin.characterEsiRestriction.title') }}
          </div>
          <p class="user-page__config-description">
            {{ t('userAdmin.characterEsiRestriction.description') }}
          </p>
        </div>

        <div class="user-page__config-actions">
          <ElTag :type="characterEsiRestrictionEnabled ? 'danger' : 'info'" effect="light" round>
            {{
              characterEsiRestrictionEnabled
                ? t('userAdmin.characterEsiRestriction.enabled')
                : t('userAdmin.characterEsiRestriction.disabled')
            }}
          </ElTag>
          <ElSwitch
            v-model="characterEsiRestrictionEnabled"
            :loading="characterEsiRestrictionLoading || characterEsiRestrictionSaving"
            :before-change="handleCharacterEsiRestrictionToggle"
            :active-text="t('userAdmin.characterEsiRestriction.switchLabel')"
            inline-prompt
          />
        </div>
      </div>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        :default-sort="{ prop: 'last_login_at', order: 'descending' }"
        :expand-row-keys="expandedUserIds"
        :row-class-name="userRowClassName"
        :row-key="getUserRowKey"
        visual-variant="ledger"
        @row-click="handleRowClick"
        @expand-change="handleExpandChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <UserManageDialog
        v-model:visible="manageDialogVisible"
        :user-data="currentUserData"
        :can-edit-profile="currentUserCanEditProfile"
        :can-edit-contacts="currentUserCanEditContacts"
        :can-edit-roles="currentUserCanEditRoles"
        @saved="refreshData"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtCopyButton from '@/components/core/forms/art-copy-button/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { formatTime } from '@utils/common'
  import { fetchGetUserList, fetchDeleteUser, fetchImpersonateUser } from '@/api/system-manage'
  import { fetchGetUserInfo } from '@/api/auth'
  import { buildEveCharacterPortraitUrl } from '@/utils/eve-image'
  import {
    fetchCharacterESIRestrictionConfig,
    updateCharacterESIRestrictionConfig
  } from '@/api/sys-config'
  import { useUserStore } from '@/store/modules/user'
  import UserSearch from './modules/user-search.vue'
  import UserManageDialog from './modules/user-manage-dialog.vue'
  import { ElTag, ElMessage, ElMessageBox, ElAvatar, ElEmpty } from 'element-plus'
  import type { TableColumnCtx } from 'element-plus'

  defineOptions({ name: 'User' })
  const { t } = useI18n()

  type UserListItem = Api.SystemManage.UserListItem

  const userStore = useUserStore()

  // 是否超级管理员（仅超管可使用模拟登录）
  const isSuperAdmin = computed(() => userStore.info?.roles?.includes('super_admin'))
  const currentUserId = computed(() => userStore.info?.userId)

  // 弹窗相关
  const manageDialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})
  const characterEsiRestrictionLoading = ref(false)
  const characterEsiRestrictionSaving = ref(false)
  const characterEsiRestrictionEnabled = ref(true)

  // 搜索表单
  const searchForm = ref({
    keyword: undefined,
    status: undefined,
    role: undefined
  })
  const expandedUserIds = ref<string[]>([])
  const skillPointFormatter = new Intl.NumberFormat('en-US')

  // 职权显示配置
  const ROLE_CONFIG: Record<string, { type: string; text: string }> = {
    super_admin: { type: 'danger', text: t('userAdmin.roles.super_admin') },
    admin: { type: 'warning', text: t('userAdmin.roles.admin') },
    srp: { type: 'success', text: t('userAdmin.roles.srp') },
    senior_fc: { type: 'warning', text: t('userAdmin.roles.senior_fc') },
    fc: { type: 'warning', text: t('userAdmin.roles.fc') },
    captain: { type: 'primary', text: t('userAdmin.roles.captain') },
    mentor: { type: 'success', text: t('userAdmin.roles.mentor') },
    welfare: { type: 'primary', text: t('userAdmin.roles.welfare') },
    user: { type: 'success', text: t('userAdmin.roles.user') },
    guest: { type: 'info', text: t('userAdmin.roles.guest') }
  }

  // 状态显示配置
  const STATUS_CONFIG: Record<number, { type: string; text: string }> = {
    1: { type: 'success', text: t('userAdmin.status.active') },
    0: { type: 'danger', text: t('userAdmin.status.disabled') }
  }

  const getRoleConfig = (role: string) => ROLE_CONFIG[role] || { type: 'info', text: role }
  const getStatusConfig = (status: number) =>
    STATUS_CONFIG[status] || { type: 'info', text: t('userAdmin.status.unknown') }
  const getDisplayRoles = (row: Partial<UserListItem>) => {
    return row.roles?.length ? row.roles : ['guest']
  }
  const getContactEntries = (row: UserListItem) => {
    const contacts = [
      row.qq
        ? {
            label: t('characters.profile.qq'),
            value: row.qq
          }
        : null,
      row.discord_id
        ? {
            label: t('characters.profile.discordId'),
            value: row.discord_id
          }
        : null
    ]

    return contacts.filter((entry): entry is { label: string; value: string } => entry !== null)
  }
  const isSuperAdminUser = (row: Partial<UserListItem>) =>
    getDisplayRoles(row).includes('super_admin')
  const isSelfUser = (row: Partial<UserListItem>) =>
    currentUserId.value != null && row.id === currentUserId.value
  const isProtectedUser = (row: Partial<UserListItem>) =>
    getDisplayRoles(row).some((role) => ['super_admin', 'admin'].includes(role))
  const canEditProfile = (row: Partial<UserListItem>) => isSuperAdmin.value || !isProtectedUser(row)
  const canEditContacts = (row: Partial<UserListItem>) =>
    isSuperAdmin.value && !isSuperAdminUser(row)
  const canEditRoles = (row: Partial<UserListItem>) => {
    if (isSuperAdmin.value) {
      return true
    }

    if (isSuperAdminUser(row)) {
      return false
    }

    if (isSelfUser(row)) {
      return true
    }

    return true
  }
  const canManageUser = (row: Partial<UserListItem>) => canEditProfile(row) || canEditRoles(row)
  const currentUserCanEditProfile = computed(() => canEditProfile(currentUserData.value))
  const currentUserCanEditContacts = computed(() => canEditContacts(currentUserData.value))
  const currentUserCanEditRoles = computed(() => canEditRoles(currentUserData.value))
  const canDeleteUser = (row: UserListItem) => isSuperAdmin.value || !isProtectedUser(row)
  const formatSkillPoints = (value: number) => skillPointFormatter.format(value ?? 0)
  const getUserRowKey = (row: Record<string, any>) => String(row.id)
  const getUserCharacters = (row: UserListItem) => row.characters ?? []
  const userRowClassName = () => 'user-row--expandable'
  const getTokenHealthConfig = (tokenInvalid: boolean) =>
    tokenInvalid
      ? { type: 'danger', text: t('userAdmin.characters.tokenExpired') }
      : { type: 'success', text: t('userAdmin.characters.tokenValid') }
  const getSeatUrl = (characterId: number) =>
    `https://seat.winterco.space/character/view/sheet/${characterId}`

  const renderUserCharacters = (row: UserListItem) => {
    const characters = getUserCharacters(row)
    if (characters.length === 0) {
      return h('div', { class: 'user-characters-panel user-characters-panel--empty' }, [
        h(ElEmpty, {
          description: t('userAdmin.characters.empty'),
          imageSize: 72
        })
      ])
    }

    return h('div', { class: 'user-characters-panel' }, [
      h('div', { class: 'user-characters-panel__header' }, [
        h('div', { class: 'user-characters-panel__avatar-spacer' }),
        h(
          'div',
          { class: 'user-characters-panel__header-cell' },
          t('userAdmin.characters.character')
        ),
        h(
          'div',
          { class: 'user-characters-panel__header-cell' },
          t('userAdmin.characters.characterIdLabel')
        ),
        h(
          'div',
          { class: 'user-characters-panel__header-cell' },
          t('userAdmin.characters.tokenHealth')
        ),
        h('div', { class: 'user-characters-panel__header-cell' }, t('SeAT')),
        h(
          'div',
          { class: 'user-characters-panel__header-cell' },
          t('userAdmin.characters.totalSkillPointsLabel')
        )
      ]),
      ...characters.map((character) => {
        const tokenHealth = getTokenHealthConfig(character.token_invalid)

        return h('div', { key: character.character_id, class: 'user-characters-panel__row' }, [
          h('div', { class: 'user-characters-panel__avatar-cell' }, [
            h(ElAvatar, {
              src: buildEveCharacterPortraitUrl(character.character_id, 30),
              size: 30,
              class: 'user-character-avatar'
            })
          ]),
          h('div', { class: 'user-characters-panel__cell user-characters-panel__cell--name' }, [
            h('div', { class: 'user-character-name-row' }, [
              h('span', { class: 'user-character-name' }, character.character_name),
              h(ArtCopyButton, { text: character.character_name })
            ])
          ]),
          h('div', { class: 'user-characters-panel__cell' }, [
            h('span', { class: 'user-character-badge' }, String(character.character_id))
          ]),
          h('div', { class: 'user-characters-panel__cell user-characters-panel__cell--token' }, [
            h(
              ElTag,
              { type: tokenHealth.type as any, size: 'small', effect: 'light' },
              () => tokenHealth.text
            )
          ]),
          h('div', { class: 'user-characters-panel__cell user-characters-panel__cell--seat' }, [
            h(
              'a',
              {
                class: 'user-character-link',
                href: getSeatUrl(character.character_id),
                target: '_blank',
                rel: 'noreferrer noopener',
                title: getSeatUrl(character.character_id)
              },
              getSeatUrl(character.character_id)
            )
          ]),
          h('div', { class: 'user-characters-panel__cell user-characters-panel__cell--sp' }, [
            h('span', { class: 'user-character-sp' }, formatSkillPoints(character.total_sp))
          ])
        ])
      })
    ])
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetUserList,
      apiParams: {
        current: 1,
        size: 200,
        keyword: searchForm.value.keyword,
        status: searchForm.value.status,
        role: searchForm.value.role
      },
      columnsFactory: () => [
        {
          type: 'expand',
          width: 52,
          formatter: (row: UserListItem) => () => renderUserCharacters(row)
        },
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'userInfo',
          label: t('userAdmin.table.userInfo'),
          width: 240,
          formatter: (row) => {
            return h('div', { class: 'flex items-center gap-2' }, [
              h(ElAvatar, {
                size: 36,
                src: buildEveCharacterPortraitUrl(row.primary_character_id, 36),
                class: 'flex-shrink-0'
              }),
              h('div', {}, [
                h('p', { class: 'font-medium text-sm' }, row.nickname || t('userAdmin.unnamed')),
                h('p', { class: 'text-xs text-gray-400' }, `ID: ${row.id}`)
              ])
            ])
          }
        },
        {
          prop: 'roles',
          label: t('common.role'),
          width: 220,
          formatter: (row) => {
            return h(
              'div',
              { class: 'flex flex-wrap gap-1' },
              getDisplayRoles(row).map((role) => {
                const cfg = getRoleConfig(role)
                return h(ElTag, { type: cfg.type as any, size: 'small' }, () => cfg.text)
              })
            )
          }
        },
        {
          prop: 'contact',
          label: t('userAdmin.table.contact'),
          width: 220,
          formatter: (row) => {
            const contacts = getContactEntries(row)
            if (contacts.length === 0) return '-'

            return h(
              'div',
              { class: 'flex flex-col gap-1 text-xs' },
              contacts.map((contact) =>
                h('div', { class: 'leading-5' }, [
                  h('span', { class: 'text-gray-500' }, `${contact.label}: `),
                  h('span', { class: 'font-medium text-gray-700' }, contact.value)
                ])
              )
            )
          }
        },
        {
          prop: 'status',
          label: t('common.status'),
          width: 100,
          formatter: (row) => {
            const cfg = getStatusConfig(row.status)
            return h(ElTag, { type: cfg.type as any, size: 'small' }, () => cfg.text)
          }
        },
        {
          prop: 'last_login_at',
          label: t('userAdmin.table.lastLogin'),
          width: 180,
          sortable: true,
          formatter: (row) => formatTime(row.last_login_at)
        },
        {
          prop: 'last_login_ip',
          label: t('userAdmin.table.loginIp'),
          width: 140,
          formatter: (row) => row.last_login_ip || '-'
        },
        {
          prop: 'created_at',
          label: t('userAdmin.table.registeredAt'),
          width: 180,
          sortable: true,
          formatter: (row) => formatTime(row.created_at)
        },
        {
          prop: 'operation',
          label: t('common.operation'),
          width: 160,
          fixed: 'right',
          formatter: (row) =>
            h('div', { class: 'user-operations flex gap-2' }, [
              isSuperAdmin.value &&
                h(ArtButtonTable, {
                  icon: 'ri:user-follow-line',
                  iconClass: 'bg-purple/12 text-purple',
                  title: t('userAdmin.impersonate'),
                  onClick: () => impersonateUser(row)
                }),
              canManageUser(row) &&
                h(ArtButtonTable, {
                  type: 'edit',
                  title: t('userAdmin.manageDialog.title'),
                  onClick: () => showManageDialog(row)
                }),
              canDeleteUser(row) &&
                h(ArtButtonTable, {
                  type: 'delete',
                  onClick: () => deleteUser(row)
                })
            ])
        }
      ]
    }
  })

  watch(
    data,
    (rows) => {
      const visibleUserIds = new Set((rows as UserListItem[]).map((row) => String(row.id)))
      expandedUserIds.value = expandedUserIds.value.filter((id) => visibleUserIds.has(id))
    },
    { immediate: true }
  )

  const handleExpandChange = (_row: UserListItem, expandedRows: UserListItem[]) => {
    expandedUserIds.value = expandedRows.map((item) => String(item.id))
  }

  const syncCharacterEsiRestrictionStoreState = (enabled: boolean) => {
    userStore.setUserInfo({
      ...(userStore.getUserInfo as Api.Auth.UserInfo),
      enforceCharacterESIRestriction: enabled
    })
  }

  const loadCharacterEsiRestrictionConfig = async () => {
    if (!isSuperAdmin.value) {
      return
    }

    characterEsiRestrictionLoading.value = true
    try {
      const config = await fetchCharacterESIRestrictionConfig()
      characterEsiRestrictionEnabled.value = config.enforce_character_esi_restriction
      syncCharacterEsiRestrictionStoreState(config.enforce_character_esi_restriction)
    } catch {
      ElMessage.error(t('userAdmin.characterEsiRestriction.loadFailed'))
    } finally {
      characterEsiRestrictionLoading.value = false
    }
  }

  const handleCharacterEsiRestrictionToggle = async () => {
    const nextValue = !characterEsiRestrictionEnabled.value

    characterEsiRestrictionSaving.value = true
    try {
      await updateCharacterESIRestrictionConfig({
        enforce_character_esi_restriction: nextValue
      })
      syncCharacterEsiRestrictionStoreState(nextValue)
      ElMessage.success(t('userAdmin.characterEsiRestriction.saveSuccess'))
      return true
    } catch {
      ElMessage.error(t('userAdmin.characterEsiRestriction.saveFailed'))
      return false
    } finally {
      characterEsiRestrictionSaving.value = false
    }
  }

  onMounted(loadCharacterEsiRestrictionConfig)

  const handleRowClick = (
    row: UserListItem,
    column: TableColumnCtx<UserListItem>,
    event: MouseEvent
  ) => {
    const target = event.target as HTMLElement | null
    if (
      column.type === 'expand' ||
      column.property === 'operation' ||
      target?.closest('.user-operations')
    ) {
      return
    }

    const rowKey = String(row.id)

    if (expandedUserIds.value.includes(rowKey)) {
      expandedUserIds.value = expandedUserIds.value.filter((id) => id !== rowKey)
      return
    }

    expandedUserIds.value = [...expandedUserIds.value, rowKey]
  }

  /** 搜索 */
  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, { current: 1 }, params)
    getData()
  }

  /** 打开用户编辑弹窗 */
  const showManageDialog = (row: UserListItem): void => {
    if (!canManageUser(row)) {
      ElMessage.error(t('userAdmin.editProtectedDenied'))
      return
    }
    currentUserData.value = row
    nextTick(() => {
      manageDialogVisible.value = true
    })
  }

  /** 删除用户 */
  const deleteUser = (row: UserListItem): void => {
    if (!canDeleteUser(row)) {
      ElMessage.error(t('userAdmin.deleteProtectedDenied'))
      return
    }
    ElMessageBox.confirm(
      t('userAdmin.deleteConfirm', { name: row.nickname || row.id }),
      t('common.tips'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'error'
      }
    )
      .then(async () => {
        try {
          await fetchDeleteUser(row.id)
          ElMessage.success(t('userAdmin.deleteSuccess'))
          refreshData()
        } catch (error) {
          console.error(t('userAdmin.deleteFailed'), error)
        }
      })
      .catch(() => {})
  }

  /** 模拟以指定用户登录 */
  const impersonateUser = (row: UserListItem): void => {
    ElMessageBox.confirm(
      t('userAdmin.impersonateConfirm', { name: row.nickname || row.id }),
      t('common.tips'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
      .then(async () => {
        try {
          const res = await fetchImpersonateUser(row.id)
          userStore.setToken(res.token)
          userStore.setLoginStatus(true)
          const userInfo = await fetchGetUserInfo()
          userStore.setUserInfo(userInfo)
          ElMessage.success(t('userAdmin.impersonateSuccess', { name: row.nickname || row.id }))
          window.location.href = '/'
        } catch (error: any) {
          ElMessage.error(error?.message ?? t('userAdmin.impersonateFailed'))
        }
      })
      .catch(() => {})
  }
</script>

<style lang="scss">
  .user-page {
    &__config-card {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      flex-wrap: wrap;
    }

    &__config-title {
      font-size: 16px;
      font-weight: 600;
      color: var(--el-text-color-primary);
    }

    &__config-description {
      margin-top: 6px;
      max-width: 720px;
      font-size: 13px;
      line-height: 1.6;
      color: var(--el-text-color-secondary);
    }

    &__config-actions {
      display: flex;
      align-items: center;
      gap: 12px;
      min-height: 40px;
    }

    .user-characters-panel {
      width: 100%;
      margin: 4px 0 8px 56px;
      overflow: hidden;
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 14px;
      background: var(--el-bg-color-overlay);
    }

    .user-characters-panel--empty {
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 96px;
      padding: 12px;
    }

    .user-characters-panel__header,
    .user-characters-panel__row {
      width: 100%;
      display: grid;
      grid-template-columns:
        56px minmax(0, 0.18fr) minmax(25px, 0.12fr) minmax(72px, 0.16fr)
        minmax(50px, 0.24fr) minmax(86px, 0.3fr);
      column-gap: 12px;
      align-items: center;
    }

    .user-characters-panel__header {
      padding: 8px 14px;
      background: var(--el-fill-color-light);
      border-bottom: 1px solid var(--el-border-color-lighter);
      font-size: 11px;
      font-weight: 700;
      letter-spacing: 0.03em;
      text-transform: uppercase;
      color: var(--el-text-color-secondary);
    }

    .user-characters-panel__avatar-spacer {
      width: 100%;
    }

    .user-characters-panel__row {
      padding: 8px 14px;
      border-bottom: 1px solid var(--el-border-color-lighter);
      transition:
        background-color 0.18s ease,
        transform 0.18s ease;
    }

    .user-characters-panel__row:last-child {
      border-bottom: 0;
    }

    .user-characters-panel__row:hover {
      background: rgba(64, 158, 255, 0.04);
    }

    .user-characters-panel__avatar-cell {
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .user-character-avatar {
      border: 1px solid var(--el-border-color-lighter);
      box-shadow: 0 4px 12px rgba(15, 23, 42, 0.08);
    }

    .user-characters-panel__cell {
      min-width: 0;
      display: flex;
      align-items: center;
      font-size: 12px;
      line-height: 1.25;
      color: var(--el-text-color-primary);
    }

    .user-characters-panel__cell--name {
      font-size: 13px;
      font-weight: 600;
    }

    .user-character-name {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .user-character-name-row {
      min-width: 0;
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .user-character-name-row :deep(.art-copy-button) {
      flex-shrink: 0;
    }

    .user-character-badge,
    .user-character-sp {
      display: inline-flex;
      align-items: center;
      min-height: 22px;
      padding: 0 8px;
      border-radius: 999px;
      background: var(--el-fill-color);
      color: var(--el-text-color-secondary);
      white-space: nowrap;
    }

    .user-character-sp {
      background: rgba(103, 194, 58, 0.14);
      color: var(--el-color-success);
      font-weight: 600;
      letter-spacing: 0.01em;
    }

    .user-character-link {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-size: 12px;
      line-height: 1.2;
      color: var(--el-color-primary);
      text-decoration: none;
    }

    .user-character-link:hover {
      text-decoration: underline;
    }

    .user-characters-panel__cell--token {
      min-width: 0;
    }

    .user-row--expandable > td {
      cursor: pointer;
    }

    .user-row--expandable .user-operations,
    .user-row--expandable .user-operations * {
      cursor: default;
    }
  }

  @media (max-width: 768px) {
    .user-page {
      .user-characters-panel {
        margin-left: 8px;
      }

      .user-characters-panel__header,
      .user-characters-panel__row {
        grid-template-columns: 56px minmax(0, 1fr);
        row-gap: 8px;
      }

      .user-characters-panel__header-cell:nth-child(n + 3),
      .user-characters-panel__row .user-characters-panel__cell:nth-child(n + 3) {
        grid-column: 1 / -1;
      }

      .user-characters-panel__header-cell:nth-child(n + 3) {
        display: none;
      }
    }
  }
</style>
