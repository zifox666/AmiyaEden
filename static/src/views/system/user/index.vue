<!-- 用户管理页面 -->
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <UserSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></UserSearch>

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
        visual-variant="ledger"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- 角色分配弹窗 -->
      <UserRoleDialog
        v-model:visible="dialogVisible"
        :user-data="currentUserData"
        @saved="refreshData"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchGetUserList,
    fetchDeleteUser,
    fetchImpersonateUser,
    fetchGetRoleDefinitions
  } from '@/api/system-manage'
  import { fetchGetUserInfo } from '@/api/auth'
  import { useUserStore } from '@/store/modules/user'
  import UserSearch from './modules/user-search.vue'
  import UserRoleDialog from './modules/user-role-dialog.vue'
  import { ElTag, ElMessage, ElMessageBox, ElAvatar } from 'element-plus'

  defineOptions({ name: 'User' })
  const { t } = useI18n()

  type UserListItem = Api.SystemManage.UserListItem

  const userStore = useUserStore()

  // 是否超级管理员（仅超管可使用模拟登录）
  const isSuperAdmin = computed(() => userStore.info?.roles?.includes('super_admin'))

  // 弹窗相关
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})

  // 搜索表单
  const searchForm = ref({
    keyword: undefined,
    status: undefined
  })

  const DEFAULT_ROLE_PRIORITY: Record<string, number> = {
    super_admin: 100,
    admin: 90,
    srp: 80,
    senior_fc: 75,
    fc: 70,
    welfare: 50,
    captain: 30,
    user: 10,
    guest: 0
  }
  const rolePriorityMap = ref<Record<string, number>>({ ...DEFAULT_ROLE_PRIORITY })
  const rolePriorityLoaded = ref(false)
  let roleHydrationVersion = 0
  // 角色显示配置
  const ROLE_CONFIG: Record<string, { type: string; text: string }> = {
    super_admin: { type: 'danger', text: t('userAdmin.roles.super_admin') },
    admin: { type: 'warning', text: t('userAdmin.roles.admin') },
    srp: { type: 'success', text: t('userAdmin.roles.srp') },
    senior_fc: { type: 'warning', text: t('userAdmin.roles.senior_fc') },
    fc: { type: 'warning', text: t('userAdmin.roles.fc') },
    captain: { type: 'primary', text: t('userAdmin.roles.captain') },
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
  const getRolePriority = (role: string) =>
    rolePriorityMap.value[role] ?? DEFAULT_ROLE_PRIORITY[role] ?? -1
  const sortRoles = (roles: string[]) =>
    [...new Set(roles)].sort((a, b) => {
      const diff = getRolePriority(b) - getRolePriority(a)
      return diff !== 0 ? diff : a.localeCompare(b)
    })
  const getDisplayRoles = (row: UserListItem) => {
    const sortedRoles = sortRoles(row.roles ?? [])
    return sortedRoles.length > 0 ? sortedRoles : ['guest']
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
  const isProtectedUser = (row: UserListItem) =>
    getDisplayRoles(row).some((role) => ['super_admin', 'admin'].includes(role))
  const canEditUser = (row: UserListItem) => isSuperAdmin.value || !isProtectedUser(row)
  const canDeleteUser = (row: UserListItem) => isSuperAdmin.value || !isProtectedUser(row)

  const ensureRolePriorityMap = async () => {
    if (rolePriorityLoaded.value) return
    try {
      const defs = await fetchGetRoleDefinitions()
      rolePriorityMap.value = defs.reduce<Record<string, number>>(
        (acc, role) => {
          acc[role.code] = role.sort
          return acc
        },
        { ...DEFAULT_ROLE_PRIORITY }
      )
    } catch (error) {
      console.error('Failed to load role priorities', error)
      rolePriorityMap.value = { ...DEFAULT_ROLE_PRIORITY }
    } finally {
      rolePriorityLoaded.value = true
    }
  }

  const syncDisplayRoles = async (rows: UserListItem[]) => {
    if (rows.length === 0) return

    const hydrationVersion = ++roleHydrationVersion
    await ensureRolePriorityMap()

    if (hydrationVersion !== roleHydrationVersion) return
    data.value = rows.map((row) => ({
      ...row,
      roles: getDisplayRoles(row)
    }))
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
        status: searchForm.value.status
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'userInfo',
          label: t('userAdmin.table.userInfo'),
          width: 240,
          formatter: (row) => {
            return h('div', { class: 'flex items-center gap-2' }, [
              h(ElAvatar, {
                size: 36,
                src: row.avatar,
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
          formatter: (row) => row.last_login_at || '-'
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
          sortable: true
        },
        {
          prop: 'operation',
          label: t('common.operation'),
          width: 160,
          fixed: 'right',
          formatter: (row) =>
            h('div', { class: 'flex gap-2' }, [
              isSuperAdmin.value &&
                h(ArtButtonTable, {
                  icon: 'ri:user-follow-line',
                  iconClass: 'bg-purple/12 text-purple',
                  title: t('userAdmin.impersonate'),
                  onClick: () => impersonateUser(row)
                }),
              canEditUser(row) &&
                h(ArtButtonTable, {
                  type: 'edit',
                  onClick: () => showRoleDialog(row)
                }),
              canDeleteUser(row) &&
                h(ArtButtonTable, {
                  type: 'delete',
                  onClick: () => deleteUser(row)
                })
            ])
        }
      ]
    },
    transform: {
      dataTransformer: (rows) =>
        rows.map((row) => {
          const roles = getDisplayRoles(row as UserListItem)
          return {
            ...row,
            roles
          }
        })
    },
    hooks: {
      onSuccess: (rows) => {
        void syncDisplayRoles(rows as UserListItem[])
      }
    }
  })

  /** 搜索 */
  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  /** 打开角色编辑弹窗 */
  const showRoleDialog = (row: UserListItem): void => {
    if (!canEditUser(row)) {
      ElMessage.error(t('userAdmin.editProtectedDenied'))
      return
    }
    currentUserData.value = row
    nextTick(() => {
      dialogVisible.value = true
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
