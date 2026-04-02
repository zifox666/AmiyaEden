<template>
  <ElCard shadow="never">
    <template #header>
      <span>{{ t('newbro.mentor.responsibilitiesTitle') }}</span>
    </template>

    <div class="mb-3 text-sm text-gray-500">
      {{ t('newbro.mentor.responsibilitiesDescription') }}
    </div>

    <ul class="list-disc space-y-2 pl-5 text-sm leading-6 text-gray-700">
      <li v-for="item in responsibilityItems" :key="item">
        {{ item }}
      </li>
    </ul>
  </ElCard>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'MentorResponsibilitiesCard' })

  const { t } = useI18n()

  const responsibilityItems = computed(() => [
    t('newbro.mentor.responsibilityItems.registration'),
    t('newbro.mentor.responsibilityItems.basics'),
    t('newbro.mentor.responsibilityItems.pvp'),
    t('newbro.mentor.responsibilityItems.scope')
  ])
</script>
