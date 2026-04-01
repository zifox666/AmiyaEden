<template>
  <div class="mentor-manage-page art-full-height">
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <div class="text-lg font-semibold">{{ t('newbro.mentorManage.title') }}</div>
          <div class="text-sm text-gray-500 mt-1">{{ t('newbro.mentorManage.subtitle') }}</div>
        </div>
        <ElButton
          class="min-w-[120px]"
          :disabled="loading || revokingId !== null"
          @click="loadData"
          >{{ $t('common.refresh') }}</ElButton
        >
      </div>
    </ElCard>

    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center gap-3 flex-wrap">
        <ElInput
          v-model="filters.keyword"
          clearable
          style="width: 280px"
          :placeholder="t('newbro.mentorManage.keyword')"
          @keyup.enter="handleSearch"
        />
        <ElSelect v-model="filters.status" style="width: 180px">
          <ElOption
            v-for="option in statusOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </ElSelect>
        <ElButton type="primary" @click="handleSearch">{{ $t('common.search') }}</ElButton>
        <ElButton @click="handleReset">{{ $t('common.reset') }}</ElButton>
      </div>
    </ElCard>

    <ElCard shadow="never">
      <template #header>
        <span>{{ t('newbro.mentorManage.relationshipList') }}</span>
      </template>

      <ElTable :data="rows" v-loading="loading" stripe border>
        <ElTableColumn :label="t('newbro.mentorManage.mentorColumn')" min-width="240">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <ElAvatar :src="row.mentor_portrait_url" :size="40" />
              <div>
                <div class="font-medium">{{ row.mentor_character_name }}</div>
                <div class="text-sm text-gray-500">
                  {{ t('newbro.mentorManage.mentorNickname') }}: {{ row.mentor_nickname || '-' }}
                </div>
              </div>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('newbro.mentorManage.menteeColumn')" min-width="240">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <ElAvatar :src="row.mentee_portrait_url" :size="40" />
              <div>
                <div class="font-medium">{{ row.mentee_character_name }}</div>
                <div class="text-sm text-gray-500">
                  {{ t('newbro.mentorManage.menteeNickname') }}: {{ row.mentee_nickname || '-' }}
                </div>
              </div>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('newbro.mentorManage.status')" width="140">
          <template #default="{ row }">
            <ElTag :type="statusTagType(row.status)" effect="light">
              {{ formatStatus(row.status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="applied_at" :label="t('newbro.mentorManage.appliedAt')" width="180">
          <template #default="{ row }">{{ formatDateTime(row.applied_at) }}</template>
        </ElTableColumn>
        <ElTableColumn
          prop="responded_at"
          :label="t('newbro.mentorManage.respondedAt')"
          width="180"
        >
          <template #default="{ row }">{{ formatDateTime(row.responded_at) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="revoked_at" :label="t('newbro.mentorManage.revokedAt')" width="180">
          <template #default="{ row }">{{ formatDateTime(row.revoked_at) }}</template>
        </ElTableColumn>
        <ElTableColumn
          prop="graduated_at"
          :label="t('newbro.mentorManage.graduatedAt')"
          width="180"
        >
          <template #default="{ row }">{{ formatDateTime(row.graduated_at) }}</template>
        </ElTableColumn>
        <ElTableColumn :label="$t('common.operation')" width="140" fixed="right">
          <template #default="{ row }">
            <ElButton
              v-if="canRevoke(row.status)"
              type="danger"
              size="small"
              :disabled="revokingId === row.id"
              @click="handleRevoke(row)"
            >
              {{ revokeActionLabel(row.status) }}
            </ElButton>
            <span v-else class="text-gray-400">-</span>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="flex justify-end mt-4">
        <ElPagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="page.current"
          :page-size="page.size"
          :page-sizes="[20, 50, 100]"
          :total="page.total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { fetchAdminMentorRelationships, revokeMentorRelationship } from '@/api/mentor'
  import { useNewbroFormatters } from '@/hooks/newbro/useNewbroFormatters'

  defineOptions({ name: 'MentorManage' })

  const { t } = useI18n()
  const { formatDateTime } = useNewbroFormatters()

  const loading = ref(false)
  const revokingId = ref<number | null>(null)
  const rows = ref<Api.Mentor.RelationshipView[]>([])
  const filters = reactive({
    keyword: '',
    status: 'all' as Api.Mentor.MenteeStatusFilter
  })
  const page = reactive({ current: 1, size: 20, total: 0 })

  const statusOptions = computed(() => [
    { label: t('newbro.mentorManage.allStatuses'), value: 'all' },
    { label: formatStatus('pending'), value: 'pending' },
    { label: formatStatus('active'), value: 'active' },
    { label: formatStatus('rejected'), value: 'rejected' },
    { label: formatStatus('revoked'), value: 'revoked' },
    { label: formatStatus('graduated'), value: 'graduated' }
  ])

  function formatStatus(status: Api.Mentor.MentorRelationshipStatus) {
    return t(`newbro.mentorStatus.${status}`)
  }

  function statusTagType(status: Api.Mentor.MentorRelationshipStatus) {
    switch (status) {
      case 'active':
        return 'success'
      case 'pending':
        return 'warning'
      case 'graduated':
        return 'primary'
      case 'rejected':
      case 'revoked':
        return 'info'
      default:
        return 'info'
    }
  }

  function canRevoke(status: Api.Mentor.MentorRelationshipStatus) {
    return status === 'pending' || status === 'active'
  }

  function revokeActionLabel(status: Api.Mentor.MentorRelationshipStatus) {
    return status === 'pending'
      ? t('newbro.mentorManage.cancelPending')
      : t('newbro.mentorManage.revoke')
  }

  function revokeActionSuccessMessage(status: Api.Mentor.MentorRelationshipStatus) {
    return status === 'pending'
      ? t('newbro.mentorManage.cancelPendingSuccess')
      : t('newbro.mentorManage.revokeSuccess')
  }

  async function loadData() {
    loading.value = true
    try {
      const data = await fetchAdminMentorRelationships({
        current: page.current,
        size: page.size,
        keyword: filters.keyword.trim() || undefined,
        status: filters.status === 'all' ? undefined : filters.status
      })
      rows.value = data.list
      page.total = data.total
    } catch (error) {
      console.error('Failed to load mentor relationships', error)
      rows.value = []
      page.total = 0
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      loading.value = false
    }
  }

  async function handleSearch() {
    page.current = 1
    await loadData()
  }

  async function handleReset() {
    filters.keyword = ''
    filters.status = 'all'
    page.current = 1
    await loadData()
  }

  async function handleCurrentChange(value: number) {
    page.current = value
    await loadData()
  }

  async function handleSizeChange(value: number) {
    page.size = value
    page.current = 1
    await loadData()
  }

  async function handleRevoke(row: Api.Mentor.RelationshipView) {
    try {
      await ElMessageBox.confirm(
        `${row.mentor_character_name} -> ${row.mentee_character_name}`,
        revokeActionLabel(row.status),
        { type: 'warning' }
      )
    } catch {
      return
    }

    revokingId.value = row.id
    try {
      await revokeMentorRelationship({ relationship_id: row.id })
      ElMessage.success(revokeActionSuccessMessage(row.status))
      await loadData()
    } catch (error) {
      console.error('Failed to revoke mentor relationship', error)
      ElMessage.error((error as Error)?.message || t('httpMsg.requestFailed'))
    } finally {
      revokingId.value = null
    }
  }

  onMounted(() => {
    loadData()
  })
</script>
