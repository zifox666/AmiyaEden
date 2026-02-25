<template>
  <ElDialog
    v-model="dialogVisible"
    title="分配角色"
    width="480px"
    align-center
    @open="onOpen"
  >
    <ElForm label-width="80px">
      <ElFormItem label="用户">
        <div class="flex items-center gap-2">
          <ElAvatar :size="32" :src="userData?.avatar" />
          <span>{{ userData?.nickname || '未命名' }}</span>
        </div>
      </ElFormItem>
      <ElFormItem label="角色">
        <ElCheckboxGroup v-model="selectedRoleIds">
          <ElCheckbox
            v-for="role in allRoles"
            :key="role.id"
            :label="role.id"
            :disabled="role.code === 'super_admin' && !isSuperAdmin"
          >
            {{ role.name }}
            <span v-if="role.is_system" class="text-xs text-gray-400 ml-1">(系统)</span>
          </ElCheckbox>
        </ElCheckboxGroup>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">确认</ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetAllRoles, fetchGetUserRoles, fetchSetUserRoles } from '@/api/system-manage'
  import { useUserStore } from '@/store/modules/user'

  interface Props {
    visible: boolean
    userData?: Partial<Api.SystemManage.UserListItem>
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'saved'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const userStore = useUserStore()
  const isSuperAdmin = computed(() => userStore.info?.roles?.includes('super_admin'))

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const allRoles = ref<Api.SystemManage.RoleItem[]>([])
  const selectedRoleIds = ref<number[]>([])
  const submitting = ref(false)

  const onOpen = async () => {
    try {
      allRoles.value = await fetchGetAllRoles()
      if (props.userData?.id) {
        const userRoles = await fetchGetUserRoles(props.userData.id)
        selectedRoleIds.value = userRoles.map((r) => r.id)
      }
    } catch (err) {
      console.error('加载角色数据失败:', err)
    }
  }

  const handleSubmit = async () => {
    if (!props.userData?.id) return
    submitting.value = true
    try {
      await fetchSetUserRoles(props.userData.id, selectedRoleIds.value)
      ElMessage.success('角色分配成功')
      dialogVisible.value = false
      emit('saved')
    } catch (err) {
      console.error('角色分配失败:', err)
    } finally {
      submitting.value = false
    }
  }
</script>
