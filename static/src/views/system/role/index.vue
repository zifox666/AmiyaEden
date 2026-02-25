<!-- 角色管理页面 - 新 RBAC -->
<template>
  <div class="art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #default>
          <ElButton type="primary" :icon="Plus" @click="openCreateDialog">新增角色</ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- 创建/编辑角色对话框 -->
    <ElDialog v-model="dialogVisible" :title="editingRole ? '编辑角色' : '新增角色'" width="480px" destroy-on-close>
      <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <ElFormItem label="角色编码" prop="code">
          <ElInput
            v-model="formData.code"
            placeholder="如 custom_role"
            :disabled="!!editingRole"
          />
        </ElFormItem>
        <ElFormItem label="角色名称" prop="name">
          <ElInput v-model="formData.name" placeholder="角色显示名称" />
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="formData.description" type="textarea" :rows="3" placeholder="角色描述（可选）" />
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber v-model="formData.sort" :min="0" :max="999" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDialog>

    <!-- 菜单权限分配 -->
    <RolePermissionDialog
      v-model:visible="permVisible"
      :role-id="permRoleId"
      :role-name="permRoleName"
      @saved="refreshData"
    />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import {
    fetchGetRoleList,
    fetchCreateRole,
    fetchUpdateRole,
    fetchDeleteRole
  } from '@/api/system-manage'
  import RolePermissionDialog from './modules/role-permission-dialog.vue'
  import { ElTag, ElButton, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
  import { Plus, Setting } from '@element-plus/icons-vue'

  defineOptions({ name: 'Role' })

  type RoleItem = Api.SystemManage.RoleItem

  // ─── 角色编码颜色映射 ───
  const CODE_TYPE: Record<string, string> = {
    super_admin: 'danger',
    admin: 'warning',
    srp: '',
    fc: '',
    user: 'success',
    guest: 'info'
  }

  // ─── 表格 ───
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
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        {
          prop: 'code',
          label: '角色编码',
          width: 160,
          formatter: (row: RoleItem) =>
            h(
              ElTag,
              {
                type: (CODE_TYPE[row.code] ?? '') as any,
                effect: row.is_system ? 'dark' : 'plain',
                size: 'small'
              },
              () => row.code
            )
        },
        { prop: 'name', label: '名称', width: 160 },
        { prop: 'description', label: '描述', minWidth: 200, showOverflowTooltip: true },
        { prop: 'sort', label: '排序', width: 80 },
        {
          prop: 'is_system',
          label: '系统角色',
          width: 100,
          formatter: (row: RoleItem) =>
            h(ElTag, { type: row.is_system ? 'danger' : 'info', size: 'small' }, () =>
              row.is_system ? '是' : '否'
            )
        },
        {
          prop: 'actions',
          label: '操作',
          width: 200,
          fixed: 'right',
          formatter: (row: RoleItem) =>
            h('div', { class: 'flex gap-1' }, [
              h(ElButton, { size: 'small', type: 'warning', icon: Setting, onClick: () => openPermDialog(row) }, () => '权限'),
              h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
              h(ArtButtonTable, {
                type: 'delete',
                disabled: row.is_system,
                onClick: () => handleDelete(row)
              })
            ])
        }
      ]
    }
  })

  // ─── 创建 / 编辑 ───
  const dialogVisible = ref(false)
  const submitLoading = ref(false)
  const formRef = ref<FormInstance>()
  const editingRole = ref<RoleItem | null>(null)

  const formData = reactive({
    code: '',
    name: '',
    description: '',
    sort: 0
  })

  const formRules: FormRules = {
    code: [
      { required: true, message: '请输入角色编码', trigger: 'blur' },
      { pattern: /^[a-z][a-z0-9_]*$/, message: '小写字母开头，仅含字母/数字/下划线', trigger: 'blur' }
    ],
    name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }]
  }

  function resetForm() {
    formData.code = ''
    formData.name = ''
    formData.description = ''
    formData.sort = 0
    editingRole.value = null
  }

  function openCreateDialog() {
    resetForm()
    dialogVisible.value = true
  }

  function openEditDialog(row: RoleItem) {
    editingRole.value = row
    formData.code = row.code
    formData.name = row.name
    formData.description = row.description
    formData.sort = row.sort
    dialogVisible.value = true
  }

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitLoading.value = true
    try {
      if (editingRole.value) {
        await fetchUpdateRole(editingRole.value.id, {
          name: formData.name,
          description: formData.description,
          sort: formData.sort
        })
        ElMessage.success('更新成功')
      } else {
        await fetchCreateRole({
          code: formData.code,
          name: formData.name,
          description: formData.description,
          sort: formData.sort
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? '操作失败')
    } finally {
      submitLoading.value = false
    }
  }

  // ─── 删除 ───
  async function handleDelete(row: RoleItem) {
    if (row.is_system) return
    await ElMessageBox.confirm(`确定要删除角色「${row.name}」吗？`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    try {
      await fetchDeleteRole(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? '删除失败')
    }
  }

  // ─── 权限分配 ───
  const permVisible = ref(false)
  const permRoleId = ref<number>()
  const permRoleName = ref('')

  function openPermDialog(row: RoleItem) {
    permRoleId.value = row.id
    permRoleName.value = row.name
    permVisible.value = true
  }
</script>
