<!-- 角色管理页面 - 系统内置角色（只读） -->
<template>
  <div class="art-full-height">
    <ElCard class="art-table-card" shadow="never">
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
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetRoleList } from '@/api/system-manage'
  import { ElTag } from 'element-plus'

  defineOptions({ name: 'Role' })

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetRoleList,
      apiParams: {
        current: 1,
        size: 10
      },
      columnsFactory: () => [
        {
          prop: 'roleId',
          label: '序号',
          width: 80
        },
        {
          prop: 'roleName',
          label: '角色名称',
          minWidth: 140
        },
        {
          prop: 'roleCode',
          label: '角色标识',
          minWidth: 140,
          formatter: (row) => {
            const typeMap: Record<string, string> = {
              super_admin: 'danger',
              admin: 'warning',
              user: 'success',
              guest: 'info'
            }
            return h(
              ElTag,
              { type: (typeMap[row.roleCode] || 'info') as any, effect: 'plain', size: 'small' },
              () => row.roleCode
            )
          }
        },
        {
          prop: 'description',
          label: '角色描述',
          minWidth: 260,
          showOverflowTooltip: true
        },
        {
          prop: 'enabled',
          label: '状态',
          width: 100,
          formatter: (row) => {
            return h(
              ElTag,
              { type: row.enabled ? 'success' : 'warning', size: 'small' },
              () => (row.enabled ? '启用' : '禁用')
            )
          }
        }
      ]
    }
  })
</script>
