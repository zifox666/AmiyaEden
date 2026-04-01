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
        <ElCheckboxGroup v-model="selectedRoleCodes">
          <ElCheckbox
            v-for="role in allRoles"
            :key="role.code"
            :label="role.code"
            :disabled="
              role.code === 'super_admin' ||
              (role.code === 'admin' && !(isSuperAdmin || isEditingSelf))
            "
          >
            {{ role.name }}
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
  import {
    fetchGetRoleDefinitions,
    fetchGetUserRoles,
    fetchSetUserRoles
  } from '@/api/system-manage'
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
  const currentUserId = computed(() => userStore.info?.userId)
  const isEditingSelf = computed(
    () => props.userData?.id != null && currentUserId.value === props.userData.id
  )

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const allRoles = ref<Api.SystemManage.RoleDefinition[]>([])
  const selectedRoleCodes = ref<string[]>([])
  const submitting = ref(false)

  watch(selectedRoleCodes, (codes) => {
    if (codes.length <= 1 || !codes.includes('guest')) return

    const nonGuestCodes = codes.filter((code) => code !== 'guest')
    selectedRoleCodes.value = nonGuestCodes.length > 0 ? nonGuestCodes : ['guest']
  })

  const onOpen = async () => {
    try {
      selectedRoleCodes.value = []
      allRoles.value = await fetchGetRoleDefinitions()
      if (props.userData?.id) {
        const userRoles = await fetchGetUserRoles(props.userData.id)
        selectedRoleCodes.value = userRoles.map((r) => r.code)
      }
    } catch (err) {
      console.error(t('roleUi.userRoleDialog.loadFailed'), err)
    }
  }

  const handleSubmit = async () => {
    if (!props.userData?.id) return
    submitting.value = true
    try {
      await fetchSetUserRoles(props.userData.id, selectedRoleCodes.value)
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
