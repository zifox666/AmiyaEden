<template>
  <el-select
    v-model="selectedId"
    filterable
    remote
    clearable
    :remote-method="handleSearch"
    :loading="loading"
    :placeholder="placeholder || $t('sde.search.placeholder')"
    :remote-show-suffix="true"
    @change="handleChange"
    @clear="handleClear"
  >
    <el-option
      v-for="item in options"
      :key="`${item.category}-${item.id}`"
      :label="item.name"
      :value="item.id"
    >
      <div class="sde-search-option">
        <img
          v-if="item.category === 'type' && props.showTypeIcon"
          :src="`https://images.evetech.net/types/${item.id}/icon?size=32`"
          class="sde-search-option__icon"
          loading="lazy"
        />
        <el-icon
          v-else-if="item.category !== 'type'"
          class="sde-search-option__icon sde-search-option__icon--char"
        >
          <User />
        </el-icon>
        <span class="sde-search-option__name">{{ item.name }}</span>
        <span v-if="item.group_name" class="sde-search-option__group">{{ item.group_name }}</span>
      </div>
    </el-option>
  </el-select>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { User } from '@element-plus/icons-vue'
  import { fuzzySearch } from '@/api/sde'
  import { useUserStore } from '@/store/modules/user'

  useI18n()
  const userStore = useUserStore()

  interface Props {
    /** 需要搜索的分类 ID 列表（仅搜索这些分类） */
    categoryIds?: number[]
    /** 排除的分类 ID 列表（不搜索这些分类） */
    excludeCategoryIds?: number[]
    /** 是否同时搜索成员名称 */
    searchMember?: boolean
    /** 最大返回数量 */
    limit?: number
    /** 搜索语言 */
    language?: string
    /** 占位文本 */
    placeholder?: string
    /** 初始选项列表，用于编辑时回显已选项 */
    initialOptions?: Api.Sde.FuzzySearchItem[]
    /** 是否展示 type 图标 */
    showTypeIcon?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    categoryIds: () => [],
    excludeCategoryIds: () => [],
    searchMember: false,
    limit: 20,
    language: 'zh',
    placeholder: '',
    initialOptions: () => [],
    showTypeIcon: true
  })

  const selectedId = defineModel<number | null>('modelValue', { default: null })

  const emit = defineEmits<{
    select: [item: Api.Sde.FuzzySearchItem | null]
  }>()

  const loading = ref(false)
  const options = ref<Api.Sde.FuzzySearchItem[]>([...props.initialOptions])

  // 当 initialOptions 变化时（如打开编辑弹框）同步更新
  watch(
    () => props.initialOptions,
    (v) => {
      if (v.length) options.value = [...v]
    }
  )

  let debounceTimer: ReturnType<typeof setTimeout> | null = null

  function handleSearch(keyword: string) {
    if (debounceTimer) clearTimeout(debounceTimer)

    if (!keyword.trim()) {
      options.value = []
      return
    }

    debounceTimer = setTimeout(async () => {
      loading.value = true
      try {
        const res = await fuzzySearch({
          keyword: keyword.trim(),
          language: userStore.language || props.language,
          category_ids: props.categoryIds.length ? props.categoryIds : undefined,
          exclude_category_ids: props.excludeCategoryIds.length
            ? props.excludeCategoryIds
            : undefined,
          limit: props.limit,
          search_member: props.searchMember
        })
        options.value = res ?? []
      } catch {
        options.value = []
      } finally {
        loading.value = false
      }
    }, 300)
  }

  function handleChange(val: number | null) {
    const item = options.value.find((o) => o.id === val) ?? null
    emit('select', item)
  }

  function handleClear() {
    options.value = []
    emit('select', null)
  }

  // 当 modelValue 从外部被清空时同步清理 options
  watch(selectedId, (v) => {
    if (v == null) {
      options.value = []
    }
  })
</script>

<style scoped lang="scss">
  .sde-search-option {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 2px 0;

    &__icon {
      width: 24px;
      height: 24px;
      border-radius: 4px;
      flex-shrink: 0;

      &--char {
        font-size: 18px;
        color: var(--el-text-color-secondary);
      }
    }

    &__name {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    &__group {
      flex-shrink: 0;
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }
</style>
