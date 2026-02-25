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
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- 角色编辑弹窗 -->
      <UserRoleDialog
        v-model:visible="dialogVisible"
        :user-data="currentUserData"
        @submit="handleRoleSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetUserList, fetchDeleteUser, fetchUpdateUserRole } from '@/api/system-manage'
  import UserSearch from './modules/user-search.vue'
  import UserRoleDialog from './modules/user-role-dialog.vue'
  import { ElTag, ElMessageBox, ElAvatar } from 'element-plus'

  defineOptions({ name: 'User' })

  type UserListItem = Api.SystemManage.UserListItem

  // 弹窗相关
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})

  // 搜索表单
  const searchForm = ref({
    nickname: undefined,
    status: undefined,
    role: undefined
  })

  // 角色显示配置
  const ROLE_CONFIG: Record<string, { type: string; text: string }> = {
    super_admin: { type: 'danger', text: '超级管理员' },
    admin: { type: 'warning', text: '管理员' },
    user: { type: 'success', text: '已认证用户' },
    guest: { type: 'info', text: '访客' }
  }

  // 状态显示配置
  const STATUS_CONFIG: Record<number, { type: string; text: string }> = {
    1: { type: 'success', text: '正常' },
    0: { type: 'danger', text: '禁用' }
  }

  const getRoleConfig = (role: string) => ROLE_CONFIG[role] || { type: 'info', text: role }
  const getStatusConfig = (status: number) =>
    STATUS_CONFIG[status] || { type: 'info', text: '未知' }

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
        size: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        {
          prop: 'userInfo',
          label: '用户信息',
          width: 240,
          formatter: (row) => {
            return h('div', { class: 'flex items-center gap-2' }, [
              h(ElAvatar, {
                size: 36,
                src: row.avatar,
                class: 'flex-shrink-0'
              }),
              h('div', {}, [
                h('p', { class: 'font-medium text-sm' }, row.nickname || '未命名'),
                h('p', { class: 'text-xs text-gray-400' }, `ID: ${row.id}`)
              ])
            ])
          }
        },
        {
          prop: 'role',
          label: '角色',
          width: 140,
          formatter: (row) => {
            const cfg = getRoleConfig(row.role)
            return h(ElTag, { type: cfg.type as any, size: 'small' }, () => cfg.text)
          }
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row) => {
            const cfg = getStatusConfig(row.status)
            return h(ElTag, { type: cfg.type as any, size: 'small' }, () => cfg.text)
          }
        },
        {
          prop: 'last_login_at',
          label: '最后登录',
          width: 180,
          sortable: true,
          formatter: (row) => row.last_login_at || '-'
        },
        {
          prop: 'last_login_ip',
          label: '登录IP',
          width: 140,
          formatter: (row) => row.last_login_ip || '-'
        },
        {
          prop: 'created_at',
          label: '注册时间',
          width: 180,
          sortable: true
        },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row) =>
            h('div', [
              h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showRoleDialog(row)
              }),
              h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deleteUser(row)
              })
            ])
        }
      ]
    }
  })

  /** 搜索 */
  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  /** 打开角色编辑弹窗 */
  const showRoleDialog = (row: UserListItem): void => {
    currentUserData.value = row
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /** 角色修改提交 */
  const handleRoleSubmit = async (role: string) => {
    try {
      const id = currentUserData.value.id
      if (!id) return
      await fetchUpdateUserRole(id, role)
      ElMessage.success('角色修改成功')
      dialogVisible.value = false
      refreshData()
    } catch (error) {
      console.error('角色修改失败:', error)
    }
  }

  /** 删除用户 */
  const deleteUser = (row: UserListItem): void => {
    ElMessageBox.confirm(
      `确定要删除用户「${row.nickname || row.id}」吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error'
      }
    )
      .then(async () => {
        try {
          await fetchDeleteUser(row.id)
          ElMessage.success('删除成功')
          refreshData()
        } catch (error) {
          console.error('删除失败:', error)
        }
      })
      .catch(() => {})
  }
</script>
