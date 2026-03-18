<template>
  <ElDialog
    v-model="dialogVisible"
    :title="$t('roleUi.userRoleDialog.title')"
    width="480px"
    align-center
    @open="onOpen"
  >
    <ElForm label-width="80px">
      <ElFormItem :label="$t('common.user')">
        <div class="flex items-center gap-2">
          <ElAvatar :size="32" :src="userData?.avatar" />
          <span>{{ userData?.nickname || $t('roleUi.userRoleDialog.unnamed') }}</span>
        </div>
      </ElFormItem>
      <ElFormItem :label="$t('common.role')">
        <ElCheckboxGroup v-model="selectedRoleIds">
          <ElCheckbox
            v-for="role in allRoles"
            :key="role.id"
            :label="role.id"
            :disabled="role.code === 'super_admin' && !isSuperAdmin"
          >
            {{ role.name }}
            <span v-if="role.is_system" class="text-xs text-gray-400 ml-1">
              {{ $t('roleUi.userRoleDialog.systemTag') }}
            </span>
          </ElCheckbox>
        </ElCheckboxGroup>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
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
  const { t } = useI18n()

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
      console.error(t('roleUi.userRoleDialog.loadFailed'), err)
    }
  }

  const handleSubmit = async () => {
    if (!props.userData?.id) return
    submitting.value = true
    try {
      await fetchSetUserRoles(props.userData.id, selectedRoleIds.value)
      ElMessage.success(t('roleUi.userRoleDialog.saveSuccess'))
      dialogVisible.value = false
      emit('saved')
    } catch (err) {
      console.error(t('roleUi.userRoleDialog.saveFailed'), err)
    } finally {
      submitting.value = false
    }
  }
</script>
