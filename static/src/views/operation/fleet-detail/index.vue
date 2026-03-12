<!-- 舰队详情页面 -->
<template>
  <div class="fleet-detail-page art-full-height">
    <div v-loading="fleetLoading">
      <!-- 舰队基本信息 -->
      <ElCard shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <ElButton link @click="goBack">
                <el-icon><ArrowLeft /></el-icon>
              </ElButton>
              <h2 class="text-lg font-medium">{{ fleet?.title || $t('fleet.fields.title') }}</h2>
              <ElTag
                v-if="fleet"
                :type="importanceType(fleet.importance)"
                size="small"
                effect="dark"
              >
                {{ $t(`fleet.importance.${fleet.importance}`) }}
              </ElTag>
            </div>
            <div class="flex gap-2">
              <ElButton size="small" @click="handleRefreshAll">
                <el-icon class="mr-1"><Refresh /></el-icon>
                {{ $t('common.refresh') }}
              </ElButton>
            </div>
          </div>
        </template>
        <ElDescriptions v-if="fleet" :column="3" border>
          <ElDescriptionsItem :label="$t('fleet.fields.fc')">{{
            fleet.fc_character_name
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.papCount')">{{
            fleet.pap_count
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.esiFleetId')">{{
            fleet.esi_fleet_id ?? '-'
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.startAt')">{{
            formatTime(fleet.start_at)
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.endAt')">{{
            formatTime(fleet.end_at)
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.createdAt')">{{
            formatTime(fleet.created_at)
          }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.description')" :span="3">
            {{ fleet.description || '-' }}
          </ElDescriptionsItem>
        </ElDescriptions>
      </ElCard>

      <!-- 成员 & PAP -->
      <ElCard class="art-table-card mt-4" shadow="never">
        <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
          <template #left>
            <ElButton type="primary" size="small" :loading="syncLoading" @click="handleSyncESI">
              <el-icon class="mr-1"><Refresh /></el-icon>
              {{ $t('fleet.members.syncESI') }}
            </ElButton>
            <ElButton
              type="success"
              size="small"
              :loading="papIssueLoading"
              @click="handleIssuePap"
            >
              {{ $t('fleet.pap.issue') }}
            </ElButton>
            <ElButton type="warning" size="small" :loading="pingLoading" @click="handlePing">
              {{ $t('fleet.ping.send') }}
            </ElButton>
          </template>
        </ArtTableHeader>

        <ArtTable
          :loading="loading"
          :data="data"
          :columns="columns"
          :pagination="pagination"
          @pagination:size-change="handleSizeChange"
          @pagination:current-change="handleCurrentChange"
        />
      </ElCard>

      <!-- 邀请链接 -->
      <ElCard class="mt-4" shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="card-title">{{ $t('fleet.invite.title') }}</span>
            <ElButton
              type="primary"
              size="small"
              :loading="inviteCreateLoading"
              @click="handleCreateInvite"
            >
              <el-icon class="mr-1"><Plus /></el-icon>
              {{ $t('fleet.invite.create') }}
            </ElButton>
          </div>
        </template>
        <ElTable v-loading="invitesLoading" :data="invites" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="code" :label="$t('fleet.invite.code')" min-width="260">
            <template #default="{ row }">
              <code class="text-xs bg-gray-100 px-2 py-0.5 rounded select-all">{{ row.code }}</code>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="$t('common.status')" width="100" align="center">
            <template #default="{ row }">
              <ElTag :type="row.active ? 'success' : 'info'" size="small">
                {{ row.active ? $t('fleet.invite.active') : $t('fleet.invite.inactive') }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="expires_at" :label="$t('fleet.invite.expiresAt')" width="180">
            <template #default="{ row }">
              {{ formatTime(row.expires_at) }}
            </template>
          </ElTableColumn>
          <ElTableColumn :label="$t('common.operation')" width="160" fixed="right" align="center">
            <template #default="{ row }">
              <ElButton type="primary" link size="small" @click="copyInviteLink(row)">
                {{ $t('fleet.invite.copyLink') }}
              </ElButton>
              <ElButton
                v-if="row.active"
                type="danger"
                link
                size="small"
                @click="handleDeactivateInvite(row)"
              >
                {{ $t('fleet.invite.deactivate') }}
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElCard>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ArrowLeft, Refresh, Plus } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElDescriptions,
    ElDescriptionsItem,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElMessageBox
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchFleetDetail,
    syncESIFleetMembers,
    issuePap,
    fetchFleetInvites,
    createFleetInvite,
    deactivateFleetInvite,
    refreshFleetESI,
    fetchMembersWithPap,
    pingFleet
  } from '@/api/fleet'
  import { useNameResolver } from '@/hooks'

  defineOptions({ name: 'FleetDetail' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()
  const fleetId = computed(() => route.params.id as string)
  const { getName, resolve: resolveNames } = useNameResolver()

  // ---- 舰队信息 ----
  const fleet = ref<Api.Fleet.FleetItem | null>(null)
  const fleetLoading = ref(false)

  const IMPORTANCE_MAP: Record<string, string> = {
    strat_op: 'danger',
    cta: 'warning',
    other: 'info'
  }
  const importanceType = (v: string) => (IMPORTANCE_MAP[v] || 'info') as any
  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')
  const goBack = () => router.push({ name: 'Fleets' })

  const loadFleet = async () => {
    fleetLoading.value = true
    try {
      try {
        const refreshed = await refreshFleetESI(fleetId.value)
        fleet.value = refreshed
      } catch {
        fleet.value = await fetchFleetDetail(fleetId.value)
      }
    } catch {
      fleet.value = null
    } finally {
      fleetLoading.value = false
    }
  }

  // ---- 成员 & PAP 表格 ----
  const syncLoading = ref(false)
  const papIssueLoading = ref(false)
  const pingLoading = ref(false)

  const apiFn = (params: { current: number; size: number }) =>
    fetchMembersWithPap(fleetId.value, params)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'character_name',
          label: t('fleet.members.characterName'),
          minWidth: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'ship_type_id',
          label: t('fleet.members.shipType'),
          width: 160,
          showOverflowTooltip: true,
          formatter: (row: Api.Fleet.MemberWithPap) => h('span', {}, getName(row.ship_type_id, '-'))
        },
        {
          prop: 'solar_system_id',
          label: t('fleet.members.solarSystem'),
          width: 140,
          showOverflowTooltip: true,
          formatter: (row: Api.Fleet.MemberWithPap) =>
            h('span', {}, getName(row.solar_system_id, '-'))
        },
        {
          prop: 'joined_at',
          label: t('fleet.members.joinedAt'),
          width: 180,
          formatter: (row: Api.Fleet.MemberWithPap) => h('span', {}, formatTime(row.joined_at))
        },
        {
          prop: 'pap_count',
          label: t('fleet.pap.count'),
          width: 120,
          align: 'center',
          formatter: (row: Api.Fleet.MemberWithPap) =>
            row.pap_count != null
              ? h(ElTag, { type: 'success', size: 'small' }, () => `+${row.pap_count}`)
              : h('span', { class: 'text-gray-400' }, '-')
        }
      ]
    }
  })

  watch(data, async (list) => {
    if (!list.length) return
    const typeIds = [...new Set(list.map((m) => m.ship_type_id).filter(Boolean))] as number[]
    const solarIds = [...new Set(list.map((m) => m.solar_system_id).filter(Boolean))] as number[]
    const ids: Record<string, number[]> = {}
    if (typeIds.length) ids.type = typeIds
    if (solarIds.length) ids.solar_system = solarIds
    if (Object.keys(ids).length) await resolveNames({ ids })
  })

  // ---- ESI 同步 ----
  const handleSyncESI = async () => {
    syncLoading.value = true
    try {
      await syncESIFleetMembers(fleetId.value)
      ElMessage.success(t('fleet.members.syncSuccess'))
      refreshData()
    } catch {
      /* handled */
    } finally {
      syncLoading.value = false
    }
  }

  // ---- 发放 PAP ----
  const handleIssuePap = async () => {
    try {
      await ElMessageBox.confirm(t('fleet.pap.issueConfirm'), t('fleet.pap.title'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      })
    } catch {
      return
    }
    papIssueLoading.value = true
    try {
      await issuePap(fleetId.value)
      ElMessage.success(t('fleet.pap.issueSuccess'))
      refreshData()
    } catch {
      /* handled */
    } finally {
      papIssueLoading.value = false
    }
  }

  // ---- 手动 Ping ----
  const handlePing = async () => {
    pingLoading.value = true
    try {
      await pingFleet(fleetId.value)
      ElMessage.success(t('fleet.ping.success'))
    } catch {
      /* handled */
    } finally {
      pingLoading.value = false
    }
  }

  // ---- 邀请链接 ----
  const invites = ref<Api.Fleet.FleetInvite[]>([])
  const invitesLoading = ref(false)
  const inviteCreateLoading = ref(false)

  const loadInvites = async () => {
    invitesLoading.value = true
    try {
      invites.value = (await fetchFleetInvites(fleetId.value)) ?? []
    } catch {
      invites.value = []
    } finally {
      invitesLoading.value = false
    }
  }

  const handleCreateInvite = async () => {
    inviteCreateLoading.value = true
    try {
      await createFleetInvite(fleetId.value)
      ElMessage.success(t('fleet.invite.createSuccess'))
      loadInvites()
    } catch {
      /* handled */
    } finally {
      inviteCreateLoading.value = false
    }
  }

  const copyInviteLink = async (invite: Api.Fleet.FleetInvite) => {
    const link = `${window.location.origin}/#/operation/join?code=${invite.code}`
    try {
      await navigator.clipboard.writeText(link)
      ElMessage.success(t('fleet.invite.copied'))
    } catch {
      ElMessage.info(invite.code)
    }
  }

  const handleDeactivateInvite = async (invite: Api.Fleet.FleetInvite) => {
    try {
      await ElMessageBox.confirm(
        t('fleet.invite.deactivateConfirm'),
        t('fleet.invite.deactivate'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning'
        }
      )
    } catch {
      return
    }
    try {
      await deactivateFleetInvite(invite.id)
      ElMessage.success(t('fleet.invite.deactivateSuccess'))
      loadInvites()
    } catch {
      /* handled */
    }
  }

  const handleRefreshAll = () => {
    loadFleet()
    refreshData()
    loadInvites()
  }

  // ---- 初始化 ----
  onMounted(() => {
    loadFleet()
    loadInvites()
  })
</script>

<style scoped>
  .card-title {
    font-size: 15px;
    font-weight: 500;
  }
</style>
