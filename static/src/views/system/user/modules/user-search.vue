<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    :rules="rules"
    @reset="handleReset"
    @search="handleSearch"
  >
  </ArtSearchBar>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { fetchGetRoleDefinitions } from '@/api/system-manage'
  import { useEnterSearch } from '@/hooks/core/useEnterSearch'

  interface Props {
    modelValue: Record<string, any>
  }
  interface Emits {
    (e: 'update:modelValue', value: Record<string, any>): void
    (e: 'search', params: Record<string, any>): void
    (e: 'reset'): void
  }
  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()
  const { t } = useI18n()
  const { createEnterSearchHandler } = useEnterSearch()

  // 表单数据双向绑定
  const searchBarRef = ref()
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  // 校验规则
  const rules = {}
  const roleDefinitions = ref<Api.SystemManage.RoleDefinition[]>([])

  // 状态选项
  const statusOptions = [
    { label: t('userAdmin.status.active'), value: 1 },
    { label: t('userAdmin.status.disabled'), value: 0 }
  ]
  const roleOptions = computed(() =>
    roleDefinitions.value.map((role) => ({
      label: t(`userAdmin.roles.${role.code}`),
      value: role.code
    }))
  )

  // 表单配置
  const formItems = computed(() => [
    {
      label: t('userAdmin.search.keyword'),
      key: 'keyword',
      type: 'input',
      props: {
        placeholder: t('userAdmin.search.keywordPlaceholder'),
        clearable: true,
        onKeyup: createEnterSearchHandler(handleSearch)
      }
    },
    {
      label: t('common.status'),
      key: 'status',
      type: 'select',
      props: {
        placeholder: t('userAdmin.search.statusPlaceholder'),
        options: statusOptions,
        clearable: true
      }
    },
    {
      label: t('common.role'),
      key: 'role',
      type: 'select',
      props: {
        placeholder: t('userAdmin.search.rolePlaceholder'),
        options: roleOptions.value,
        clearable: true
      }
    }
  ])

  onMounted(async () => {
    try {
      roleDefinitions.value = await fetchGetRoleDefinitions()
    } catch (error) {
      console.error(error)
    }
  })

  // 事件
  function handleReset() {
    emit('reset')
  }

  async function handleSearch() {
    await searchBarRef.value.validate()
    emit('search', formData.value)
  }
</script>
