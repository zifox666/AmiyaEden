<!-- 表格按钮 -->
<template>
  <div
    :class="[
      'inline-flex items-center justify-center min-w-8 h-8 px-2.5 text-sm c-p rounded-md align-middle',
      buttonClass
    ]"
    :style="buttonStyle"
    @click="handleClick"
  >
    <ArtSvgIcon v-if="loading" icon="ri:loader-4-line" class="animate-spin" />
    <ArtSvgIcon v-else-if="iconContent" :icon="iconContent" />
    <span v-if="label">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
  defineOptions({ name: 'ArtButtonTable' })

  interface Props {
    /** 按钮类型 */
    type?: 'add' | 'edit' | 'delete' | 'more' | 'view'
    /** 按钮图标 */
    icon?: string
    /** 按钮样式类 */
    iconClass?: string
    /** icon 颜色 */
    iconColor?: string
    /** 按钮背景色 */
    buttonBgColor?: string
    /** Element Plus 按钮类型（使用 --el-color-* 配色） */
    elType?: 'primary' | 'success' | 'warning' | 'danger' | 'info'
    /** 文本内容 */
    label?: string
    /** 加载状态 */
    loading?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {})

  const emit = defineEmits<{
    (e: 'click'): void
  }>()

  // 默认按钮配置
  const defaultButtons = {
    add: { icon: 'ri:add-fill', class: 'bg-theme/12 text-theme' },
    edit: { icon: 'ri:pencil-line', class: 'bg-secondary/12 text-secondary' },
    delete: { icon: 'ri:delete-bin-5-line', class: 'bg-error/12 text-error' },
    view: { icon: 'ri:eye-line', class: 'bg-info/12 text-info' },
    more: { icon: 'ri:more-2-fill', class: '' }
  } as const

  // 获取图标内容
  const iconContent = computed(() => {
    return props.icon || (props.type ? defaultButtons[props.type]?.icon : '') || ''
  })

  // 获取按钮样式类
  const buttonClass = computed(() => {
    if (props.elType) return ''
    return props.iconClass || (props.type ? defaultButtons[props.type]?.class : '') || ''
  })

  // 获取按钮内联样式
  const buttonStyle = computed(() => {
    if (props.elType) {
      return {
        backgroundColor: `var(--el-color-${props.elType})`,
        color: '#fff'
      }
    }
    return {
      backgroundColor: props.buttonBgColor,
      color: props.iconColor
    }
  })

  const handleClick = () => {
    if (props.loading) return
    emit('click')
  }
</script>
