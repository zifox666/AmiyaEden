<!-- 工作台页面 -->
<template>
  <div v-loading="loading">
    <CardList :cards="dashboardData?.cards" />

    <ElRow :gutter="20">
      <ElCol :sm="24" :md="24" :lg="12">
        <FleetList :fleets="dashboardData?.fleets ?? []" />
      </ElCol>
      <ElCol :sm="24" :md="12" :lg="6">
        <PapChart title="联盟 PAP" :data="dashboardData?.pap_stats?.alliance ?? []" />
      </ElCol>
      <ElCol :sm="24" :md="12" :lg="6">
        <PapChart title="内部 PAP" :data="dashboardData?.pap_stats?.internal ?? []" />
      </ElCol>
    </ElRow>

    <ElRow :gutter="20">
      <ElCol :span="24">
        <SrpList :list="dashboardData?.srp_list ?? []" />
      </ElCol>
    </ElRow>
  </div>
</template>

<script setup lang="ts">
  import { fetchDashboard } from '@/api/dashboard'
  import CardList from './modules/card-list.vue'
  import FleetList from './modules/fleet-list.vue'
  import PapChart from './modules/pap-chart.vue'
  import SrpList from './modules/srp-list.vue'

  defineOptions({ name: 'Console' })

  const loading = ref(false)
  const dashboardData = ref<Api.Dashboard.DashboardResult>()

  const loadDashboard = async () => {
    loading.value = true
    try {
      dashboardData.value = await fetchDashboard()
    } catch (e) {
      console.error('加载工作台数据失败', e)
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    loadDashboard()
  })
</script>
