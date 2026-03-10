<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    @reset="handleReset"
    @search="handleSearch"
  >
    <template #fetchBtn>
      <ElButton type="primary" :loading="fetching" :icon="Download" @click="handleFetch">
        {{ t('alliancePap.fetchLatest') }}
      </ElButton>
      <ArtExcelImport @import-success="handleImport">
        {{ $t('alliancePap.importBtn') }}
      </ArtExcelImport>
    </template>
  </ArtSearchBar>
</template>

<script setup lang="ts">
  import { Download } from '@element-plus/icons-vue'
  import { ElButton } from 'element-plus'
  import { useI18n } from 'vue-i18n'

  interface Props {
    modelValue: Record<string, any>
    fetching?: boolean
  }
  interface Emits {
    (e: 'update:modelValue', value: Record<string, any>): void
    (e: 'search', params: Record<string, any>): void
    (e: 'reset'): void
    (e: 'fetch'): void
  }

  const props = withDefaults(defineProps<Props>(), { fetching: false })
  const emit = defineEmits<Emits>()
  const { t } = useI18n()

  const searchBarRef = ref()
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  const formItems = computed(() => [
    {
      label: t('alliancePap.selectMonth'),
      key: 'month',
      type: 'date',
      props: {
        type: 'month',
        format: 'YYYY-MM',
        valueFormat: 'YYYY-MM',
        clearable: false
      }
    },
    {
      key: 'fetchBtn',
      label: ''
    }
  ])

  function handleReset() {
    emit('reset')
  }

  async function handleSearch() {
    emit('search', formData.value)
  }

  function handleFetch() {
    emit('fetch')
  }

  function handleImport(rows: Record<string, unknown>[]) {
    emit('import', rows)
  }
</script>
