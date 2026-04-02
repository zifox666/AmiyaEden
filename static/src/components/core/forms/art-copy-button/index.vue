<template>
  <ElTooltip :content="tooltipText" placement="top" effect="dark">
    <ElButton
      class="art-copy-button"
      size="small"
      text
      :icon="CopyDocument"
      :disabled="isDisabled"
      :aria-label="resolvedAriaLabel"
      @click="handleClick"
    />
  </ElTooltip>
</template>

<script setup lang="ts">
  import { CopyDocument } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import { useClipboardCopy } from '@/hooks/core/useClipboardCopy'

  defineOptions({ name: 'ArtCopyButton' })

  const props = withDefaults(
    defineProps<{
      text?: string | number | null
      tooltip?: string
      ariaLabel?: string
      disabled?: boolean
    }>(),
    {
      text: '',
      tooltip: '',
      ariaLabel: '',
      disabled: false
    }
  )

  const { t } = useI18n()
  const { copyText } = useClipboardCopy()

  const tooltipText = computed(() => props.tooltip || t('common.copy'))
  const resolvedAriaLabel = computed(() => props.ariaLabel || tooltipText.value)
  const normalizedText = computed(() => String(props.text ?? '').trim())
  const isDisabled = computed(() => props.disabled || !normalizedText.value)

  const handleClick = () => {
    if (isDisabled.value) return
    void copyText(normalizedText.value)
  }
</script>

<style scoped>
  .art-copy-button {
    flex-shrink: 0;
  }
</style>
