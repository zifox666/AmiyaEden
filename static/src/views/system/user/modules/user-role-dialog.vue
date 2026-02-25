<template>
  <ElDialog
    v-model="dialogVisible"
    title="修改角色"
    width="400px"
    align-center
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="80px">
      <ElFormItem label="用户">
        <div class="flex items-center gap-2">
          <ElAvatar :size="32" :src="userData?.avatar" />
          <span>{{ userData?.nickname || '未命名' }}</span>
        </div>
      </ElFormItem>
      <ElFormItem label="当前角色">
        <ElTag :type="getRoleType(userData?.role || 'guest')" size="small">
          {{ getRoleLabel(userData?.role || 'guest') }}
        </ElTag>
      </ElFormItem>
      <ElFormItem label="新角色" prop="role">
        <ElSelect v-model="formData.role" placeholder="请选择角色" style="width: 100%">
          <ElOption
            v-for="item in roleOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确认修改</ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'

  interface Props {
    visible: boolean
    userData?: Partial<Api.SystemManage.UserListItem>
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', role: string): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const formRef = ref<FormInstance>()

  const formData = reactive({
    role: ''
  })

  const rules: FormRules = {
    role: [{ required: true, message: '请选择角色', trigger: 'change' }]
  }

  const roleOptions = [
    { label: '超级管理员', value: 'super_admin' },
    { label: '管理员', value: 'admin' },
    { label: '已认证用户', value: 'user' },
    { label: '访客', value: 'guest' }
  ]

  const ROLE_LABELS: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    user: '已认证用户',
    guest: '访客'
  }

  const ROLE_TYPES: Record<string, string> = {
    super_admin: 'danger',
    admin: 'warning',
    user: 'success',
    guest: 'info'
  }

  const getRoleLabel = (role: string) => ROLE_LABELS[role] || role
  const getRoleType = (role: string) => (ROLE_TYPES[role] || 'info') as any

  // 弹窗打开时初始化
  watch(
    () => props.visible,
    (val) => {
      if (val && props.userData) {
        formData.role = props.userData.role || ''
        nextTick(() => formRef.value?.clearValidate())
      }
    }
  )

  const handleSubmit = async () => {
    if (!formRef.value) return
    await formRef.value.validate((valid) => {
      if (valid) {
        emit('submit', formData.role)
      }
    })
  }
</script>
