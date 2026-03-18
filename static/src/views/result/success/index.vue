<template>
  <ArtResultPage
    type="success"
    :title="$t('resultPage.success.title')"
    :message="$t('resultPage.success.message')"
    iconCode="ri:check-fill"
  >
    <template #content>
      <p>{{ $t('resultPage.success.content') }}</p>
    </template>
    <template #buttons>
      <ElButton type="primary" v-ripple>{{ $t('resultPage.success.back') }}</ElButton>
      <ElButton v-ripple>{{ $t('resultPage.success.view') }}</ElButton>
      <ElButton v-ripple>{{ $t('resultPage.success.print') }}</ElButton>
    </template>
  </ArtResultPage>
</template>

<script setup lang="ts">
  defineOptions({ name: 'ResultSuccess' })
</script>
