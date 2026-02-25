<!-- 角色菜单权限分配对话框 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="分配菜单权限"
    width="520px"
    align-center
    destroy-on-close
    @open="onOpen"
  >
    <div class="mb-3 flex items-center justify-between">
      <span class="text-sm text-gray-500">角色：{{ roleName }}</span>
      <div class="flex gap-2">
        <ElButton size="small" @click="toggleExpandAll">
          {{ expandAll ? '全部收起' : '全部展开' }}
        </ElButton>
        <ElButton size="small" @click="toggleSelectAll">
          {{ selectAll ? '取消全选' : '全部选择' }}
        </ElButton>
      </div>
    </div>

    <ElScrollbar height="60vh">
      <ElTree
        ref="treeRef"
        v-loading="treeLoading"
        :data="menuTree"
        show-checkbox
        node-key="id"
        :default-expand-all="expandAll"
        :props="treeProps"
        :default-checked-keys="checkedKeys"
      />
    </ElScrollbar>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetMenuTree, fetchGetRoleMenus, fetchSetRoleMenus } from '@/api/system-manage'

  interface Props {
    visible: boolean
    roleId?: number
    roleName?: string
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'saved'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (v) => emit('update:visible', v)
  })

  const treeRef = ref()
  const menuTree = ref<Api.SystemManage.MenuItem[]>([])
  const checkedKeys = ref<number[]>([])
  const treeLoading = ref(false)
  const saving = ref(false)
  const expandAll = ref(true)
  const selectAll = ref(false)

  const treeProps = {
    children: 'children',
    label: (data: any) => data.title || data.name
  }

  const onOpen = async () => {
    treeLoading.value = true
    try {
      const [tree, roleMenuIds] = await Promise.all([
        fetchGetMenuTree(),
        props.roleId ? fetchGetRoleMenus(props.roleId) : Promise.resolve([])
      ])
      menuTree.value = tree

      // ElTree needs leaf-only keys for check-strictly=false (default)
      await nextTick()
      const leafKeys = getLeafKeys(tree, new Set(roleMenuIds))
      checkedKeys.value = leafKeys
    } catch (err) {
      console.error('加载菜单树失败:', err)
    } finally {
      treeLoading.value = false
    }
  }

  /** Get only leaf node IDs from the checked set to avoid parent auto-check */
  function getLeafKeys(nodes: Api.SystemManage.MenuItem[], checkedSet: Set<number>): number[] {
    const result: number[] = []
    for (const node of nodes) {
      if (node.children?.length) {
        result.push(...getLeafKeys(node.children, checkedSet))
      } else if (checkedSet.has(node.id)) {
        result.push(node.id)
      }
    }
    return result
  }

  /** Collect all node IDs recursively */
  function getAllKeys(nodes: Api.SystemManage.MenuItem[]): number[] {
    const keys: number[] = []
    for (const node of nodes) {
      keys.push(node.id)
      if (node.children?.length) {
        keys.push(...getAllKeys(node.children))
      }
    }
    return keys
  }

  const toggleExpandAll = () => {
    const tree = treeRef.value
    if (!tree) return
    const nodes = (tree as any).store.nodesMap
    Object.values(nodes).forEach((node: any) => {
      node.expanded = !expandAll.value
    })
    expandAll.value = !expandAll.value
  }

  const toggleSelectAll = () => {
    const tree = treeRef.value
    if (!tree) return
    if (!selectAll.value) {
      tree.setCheckedKeys(getAllKeys(menuTree.value))
    } else {
      tree.setCheckedKeys([])
    }
    selectAll.value = !selectAll.value
  }

  const handleSave = async () => {
    if (!props.roleId || !treeRef.value) return
    saving.value = true
    try {
      // Get both checked and half-checked (parent) keys
      const checked = treeRef.value.getCheckedKeys(false) as number[]
      const halfChecked = treeRef.value.getHalfCheckedKeys() as number[]
      const allMenuIds = [...checked, ...halfChecked]

      await fetchSetRoleMenus(props.roleId, allMenuIds)
      ElMessage.success('权限保存成功')
      dialogVisible.value = false
      emit('saved')
    } catch (err) {
      console.error('权限保存失败:', err)
    } finally {
      saving.value = false
    }
  }
</script>
