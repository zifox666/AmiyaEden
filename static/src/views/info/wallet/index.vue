<template>
  <div class="info-wallet-page art-full-height">
    <!-- 角色切换器 -->
    <ElCard shadow="never" class="mb-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <span class="text-sm text-gray-500">{{ $t('info.selectCharacter') }}</span>
          <ElSelect
            v-model="selectedCharacterId"
            :placeholder="$t('info.selectCharacterPlaceholder')"
            @change="onCharacterChange"
            style="width: 240px"
          >
            <ElOption
              v-for="char in characters"
              :key="char.character_id"
              :value="char.character_id"
              :label="char.character_name"
            >
              <div class="flex items-center gap-2">
                <ElAvatar :src="char.portrait_url" :size="24" />
                <span>{{ char.character_name }}</span>
              </div>
            </ElOption>
          </ElSelect>
        </div>
        <div v-if="walletData">
          <p class="text-sm text-gray-500">{{ $t('info.walletBalance') }}</p>
          <p
            class="text-2xl font-bold mt-1"
            :class="walletData.balance >= 0 ? 'text-green-600' : 'text-red-500'"
          >
            {{ formatISK(walletData.balance) }} ISK
          </p>
        </div>
      </div>
    </ElCard>

    <!-- 流水表格 -->
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="card-title">{{ $t('info.walletJournal') }}</span>
          <ElButton :loading="loading" @click="loadData">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
      </template>

      <ElTable
        v-loading="loading"
        :data="walletData?.journals ?? []"
        stripe
        border
        style="width: 100%"
      >
        <ElTableColumn type="index" width="60" label="#" />
        <ElTableColumn prop="date" :label="$t('info.journalDate')" width="180" />
        <ElTableColumn prop="ref_type" :label="$t('info.journalType')" width="160" align="center">
          <template #default="{ row }">
            <ElTag size="small" effect="plain">{{ row.ref_type }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="amount" :label="$t('info.journalAmount')" width="160" align="right">
          <template #default="{ row }">
            <span :class="row.amount >= 0 ? 'text-green-600' : 'text-red-500'" class="font-medium">
              {{ row.amount >= 0 ? '+' : '' }}{{ formatISK(row.amount) }}
            </span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="balance" :label="$t('info.journalBalance')" width="160" align="right">
          <template #default="{ row }">{{ formatISK(row.balance) }}</template>
        </ElTableColumn>
        <ElTableColumn
          prop="description"
          :label="$t('info.journalDescription')"
          min-width="240"
          show-overflow-tooltip
        />
        <ElTableColumn
          prop="reason"
          :label="$t('info.journalReason')"
          min-width="160"
          show-overflow-tooltip
        />
      </ElTable>

      <ElEmpty
        v-if="!loading && (!walletData || walletData.journals.length === 0)"
        :description="$t('info.noJournalData')"
      />

      <div v-if="pagination.total > 0" class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElSelect,
    ElOption,
    ElAvatar,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElPagination,
    ElEmpty
  } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchInfoWallet } from '@/api/eve-info'

  defineOptions({ name: 'EveInfoWallet' })

  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterId = ref<number>()
  const walletData = ref<Api.EveInfo.WalletResponse | null>(null)
  const loading = ref(false)
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  const loadCharacters = async () => {
    try {
      characters.value = await fetchMyCharacters()
      if (characters.value.length > 0 && !selectedCharacterId.value) {
        selectedCharacterId.value = characters.value[0].character_id
        loadData()
      }
    } catch {
      characters.value = []
    }
  }

  const loadData = async () => {
    if (!selectedCharacterId.value) return
    loading.value = true
    try {
      const res = await fetchInfoWallet({
        character_id: selectedCharacterId.value,
        page: pagination.page,
        page_size: pagination.pageSize
      })
      walletData.value = res
      pagination.total = res.total
      pagination.page = res.page
      pagination.pageSize = res.page_size
    } catch {
      walletData.value = null
      pagination.total = 0
    } finally {
      loading.value = false
    }
  }

  const onCharacterChange = () => {
    pagination.page = 1
    loadData()
  }

  const handleSizeChange = () => {
    pagination.page = 1
    loadData()
  }

  const handleCurrentChange = () => {
    loadData()
  }

  onMounted(() => {
    loadCharacters()
  })
</script>

<style scoped>
  .card-title {
    font-size: 15px;
    font-weight: 500;
  }
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
