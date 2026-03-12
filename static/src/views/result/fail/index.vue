<template>
  <ArtResultPage
    type="fail"
    :title="$t('resultPage.fail.title')"
    :message="$t('resultPage.fail.message')"
    iconCode="ri:close-fill"
  >
    <template #content>
      <p>{{ $t('resultPage.fail.content') }}</p>
      <p>
        <ArtSvgIcon icon="ri:close-circle-line" class="text-red-500 mr-1" />
        <span>{{ $t('resultPage.fail.frozen') }}</span>
      </p>
      <p>
        <ArtSvgIcon icon="ri:close-circle-line" class="text-red-500 mr-1" />
        <span>{{ $t('resultPage.fail.ineligible') }}</span>
      </p>
    </template>
    <template #buttons>
      <ElButton type="primary" v-ripple>{{ $t('resultPage.fail.back') }}</ElButton>
      <ElButton v-ripple>{{ $t('resultPage.fail.view') }}</ElButton>
    </template>
  </ArtResultPage>
</template>

<script setup lang="ts">
  defineOptions({ name: 'ResultFail' })
</script>
