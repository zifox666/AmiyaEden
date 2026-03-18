<template>
  <div class="art-card h-128 p-5 mb-5 max-sm:mb-4">
    <div class="art-card-header">
      <div class="title">
        <h4>{{ title }}</h4>
        <p>
          {{ $t('console.papChart.recentMonths', { count: chartData.length }) }}
        </p>
      </div>
    </div>
    <div v-if="chartData.length > 0" class="mt-2">
      <ArtBarChart
        height="13rem"
        :data="chartValues"
        :xAxisData="chartLabels"
        barWidth="50%"
        :showAxisLine="false"
      />
    </div>
    <div v-else class="flex-cc h-[calc(100%-40px)] text-g-500 text-sm">
      {{ $t('console.papChart.empty') }}
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()
  const props = defineProps<{
    title: string
    data: Api.Dashboard.PapMonthly[]
  }>()

  /** 按时间正序排列的数据 */
  const chartData = computed(() => {
    return [...props.data].sort((a, b) => {
      if (a.year !== b.year) return a.year - b.year
      return a.month - b.month
    })
  })

  const chartLabels = computed(() => {
    return chartData.value.map((d) => t('console.papChart.monthLabel', { month: d.month }))
  })

  const chartValues = computed(() => {
    return chartData.value.map((d) => d.total_pap)
  })
</script>
