<template>
  <div class="art-card h-128 p-5 mb-5 max-sm:mb-4">
    <div class="art-card-header">
      <div class="title">
        <h4>{{ title }}</h4>
        <p>
          近
          <span class="text-theme font-medium">{{ chartData.length }}</span>
          个月
        </p>
      </div>
    </div>
    <div v-if="chartData.length > 0" class="mt-2">
      <ArtBarChart
        height="calc(100% - 56px)"
        :data="chartValues"
        :xAxisData="chartLabels"
        barWidth="50%"
        :showAxisLine="false"
      />
    </div>
    <div v-else class="flex-cc h-[calc(100%-40px)] text-g-500 text-sm">
      暂无 PAP 数据
    </div>
  </div>
</template>

<script setup lang="ts">
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
    return chartData.value.map((d) => `${d.month}月`)
  })

  const chartValues = computed(() => {
    return chartData.value.map((d) => d.total_pap)
  })
</script>
