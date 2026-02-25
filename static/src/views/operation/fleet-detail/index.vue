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
              <ElTag v-if="fleet" :type="importanceType(fleet.importance)" size="small" effect="dark">
                {{ $t(`fleet.importance.${fleet.importance}`) }}
              </ElTag>
            </div>
            <div class="flex gap-2">
              <ElButton size="small" @click="loadAll">
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
          <ElDescriptionsItem :label="$t('fleet.fields.papCount')">{{ fleet.pap_count }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.esiFleetId')">{{ fleet.esi_fleet_id ?? '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.startAt')">{{ formatTime(fleet.start_at) }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.endAt')">{{ formatTime(fleet.end_at) }}</ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ formatTime(fleet.created_at) }}</ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('fleet.fields.description')" :span="3">
            {{ fleet.description || '-' }}
          </ElDescriptionsItem>
        </ElDescriptions>
      </ElCard>

      <!-- 舰队成员 -->
      <ElCard class="mt-4" shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="card-title">{{ $t('fleet.members.title') }}</span>
            <ElButton type="primary" size="small" :loading="syncLoading" @click="handleSyncESI">
              <el-icon class="mr-1"><Refresh /></el-icon>
              {{ $t('fleet.members.syncESI') }}
            </ElButton>
          </div>
        </template>
        <ElTable v-loading="membersLoading" :data="members" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="character_name" :label="$t('fleet.members.characterName')" min-width="160" />
          <ElTableColumn prop="character_id" label="角色 ID" width="120" align="center" />
          <ElTableColumn prop="ship_type_id" :label="$t('fleet.members.shipType')" width="120" align="center">
            <template #default="{ row }">
              {{ row.ship_type_id ?? '-' }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="solar_system_id" label="星系ID" width="120" align="center">
            <template #default="{ row }">
              {{ row.solar_system_id ?? '-' }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="joined_at" :label="$t('fleet.members.joinedAt')" width="180">
            <template #default="{ row }">
              {{ formatTime(row.joined_at) }}
            </template>
          </ElTableColumn>
        </ElTable>
        <ElEmpty v-if="!membersLoading && members.length === 0" :description="$t('fleet.members.empty')" />
      </ElCard>

      <!-- PAP 发放 -->
      <ElCard class="mt-4" shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="card-title">{{ $t('fleet.pap.title') }}</span>
            <ElButton type="success" size="small" :loading="papIssueLoading" @click="handleIssuePap">
              {{ $t('fleet.pap.issue') }}
            </ElButton>
          </div>
        </template>
        <ElTable v-loading="papLoading" :data="papLogs" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="character_id" label="角色 ID" width="120" align="center" />
          <ElTableColumn prop="pap_count" :label="$t('fleet.pap.count')" width="120" align="center">
            <template #default="{ row }">
              <ElTag type="success" size="small">+{{ row.pap_count }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="issued_by" :label="$t('fleet.pap.issuedBy')" width="120" align="center" />
          <ElTableColumn prop="created_at" :label="$t('fleet.pap.issuedAt')" min-width="180">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </ElTableColumn>
        </ElTable>
        <ElEmpty v-if="!papLoading && papLogs.length === 0" :description="$t('fleet.pap.empty')" />
      </ElCard>

      <!-- 邀请链接 -->
      <ElCard class="mt-4" shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="card-title">{{ $t('fleet.invite.title') }}</span>
            <ElButton type="primary" size="small" :loading="inviteCreateLoading" @click="handleCreateInvite">
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
    ElEmpty,
    ElMessageBox
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import {
    fetchFleetDetail,
    fetchFleetMembers,
    syncESIFleetMembers,
    fetchFleetPapLogs,
    issuePap,
    fetchFleetInvites,
    createFleetInvite,
    deactivateFleetInvite,
    refreshFleetESI
  } from '@/api/fleet'

  defineOptions({ name: 'FleetDetail' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()
  const fleetId = computed(() => route.params.id as string)

  // ---- 数据 ----
  const fleet = ref<Api.Fleet.FleetItem | null>(null)
  const members = ref<Api.Fleet.FleetMember[]>([])
  const papLogs = ref<Api.Fleet.PapLog[]>([])
  const invites = ref<Api.Fleet.FleetInvite[]>([])

  // ---- 加载状态 ----
  const fleetLoading = ref(false)
  const membersLoading = ref(false)
  const papLoading = ref(false)
  const invitesLoading = ref(false)
  const syncLoading = ref(false)
  const papIssueLoading = ref(false)
  const inviteCreateLoading = ref(false)

  // ---- 等级样式 ----
  const IMPORTANCE_MAP: Record<string, string> = { strat_op: 'danger', cta: 'warning', other: 'info' }
  const importanceType = (v: string) => (IMPORTANCE_MAP[v] || 'info') as any

  // ---- 时间格式化 ----
  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  // ---- 返回列表 ----
  const goBack = () => router.push({ name: 'Fleets' })

  // ---- 加载数据 ----
  const loadFleet = async () => {
    fleetLoading.value = true
    try {
      // 先尝试刷新 ESI fleet id，如果失败（非 FC 或 ESI 异常）就简单读取详情
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

  const loadMembers = async () => {
    membersLoading.value = true
    try {
      members.value = (await fetchFleetMembers(fleetId.value)) ?? []
    } catch {
      members.value = []
    } finally {
      membersLoading.value = false
    }
  }

  const loadPap = async () => {
    papLoading.value = true
    try {
      papLogs.value = (await fetchFleetPapLogs(fleetId.value)) ?? []
    } catch {
      papLogs.value = []
    } finally {
      papLoading.value = false
    }
  }

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

  const loadAll = () => {
    loadFleet()
    loadMembers()
    loadPap()
    loadInvites()
  }

  // ---- ESI 同步成员 ----
  const handleSyncESI = async () => {
    syncLoading.value = true
    try {
      await syncESIFleetMembers(fleetId.value)
      ElMessage.success(t('fleet.members.syncSuccess'))
      loadMembers()
    } catch (e) {
      console.error('Sync ESI members error:', e)
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
      loadPap()
    } catch (e) {
      console.error('Issue PAP error:', e)
    } finally {
      papIssueLoading.value = false
    }
  }

  // ---- 生成邀请链接 ----
  const handleCreateInvite = async () => {
    inviteCreateLoading.value = true
    try {
      await createFleetInvite(fleetId.value)
      ElMessage.success(t('fleet.invite.createSuccess'))
      loadInvites()
    } catch (e) {
      console.error('Create invite error:', e)
    } finally {
      inviteCreateLoading.value = false
    }
  }

  // ---- 复制邀请链接 ----
  const copyInviteLink = async (invite: Api.Fleet.FleetInvite) => {
    const link = `${window.location.origin}/#/operation/join?code=${invite.code}`
    try {
      await navigator.clipboard.writeText(link)
      ElMessage.success(t('fleet.invite.copied'))
    } catch {
      // 降级：选中 code 文本
      ElMessage.info(invite.code)
    }
  }

  // ---- 禁用邀请链接 ----
  const handleDeactivateInvite = async (invite: Api.Fleet.FleetInvite) => {
    try {
      await ElMessageBox.confirm(t('fleet.invite.deactivateConfirm'), t('fleet.invite.deactivate'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      })
    } catch {
      return
    }
    try {
      await deactivateFleetInvite(invite.id)
      ElMessage.success(t('fleet.invite.deactivateSuccess'))
      loadInvites()
    } catch (e) {
      console.error('Deactivate invite error:', e)
    }
  }

  // ---- 初始化 ----
  onMounted(loadAll)
</script>

<style scoped>
  .card-title {
    font-size: 15px;
    font-weight: 500;
  }
</style>
