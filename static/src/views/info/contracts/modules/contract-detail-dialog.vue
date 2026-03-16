<template>
  <ElDialog
    v-model="dialogVisible"
    :title="contractTitle || $t('info.contractDetailTitle')"
    width="700px"
    destroy-on-close
  >
    <div v-loading="loading">
      <!-- 物品列表 -->
      <template v-if="detail?.items?.length">
        <div class="mb-2 text-sm font-semibold text-gray-500">{{ $t('info.contractItems') }}</div>
        <ElTable :data="detail.items" size="small" border class="mb-4">
          <ElTableColumn width="40">
            <template #default="{ row }">
              <img
                :src="`https://images.evetech.net/types/${row.type_id}/icon?size=32`"
                class="w-7 h-7"
                loading="lazy"
              />
            </template>
          </ElTableColumn>
          <ElTableColumn
            prop="type_name"
            :label="$t('info.contractItemName')"
            min-width="160"
            show-overflow-tooltip
          />
          <ElTableColumn
            prop="group_name"
            :label="$t('info.contractItemGroup')"
            width="130"
            show-overflow-tooltip
          />
          <ElTableColumn
            prop="quantity"
            :label="$t('info.contractItemQty')"
            width="80"
            align="right"
          />
          <ElTableColumn :label="$t('info.contractItemIncluded')" width="90" align="center">
            <template #default="{ row }">
              <ElTag :type="row.is_included ? 'success' : 'info'" size="small">
                {{ row.is_included ? '✓' : '✗' }}
              </ElTag>
            </template>
          </ElTableColumn>
        </ElTable>
      </template>

      <!-- 竞价列表（仅拍卖合同） -->
      <template v-if="contractType === 'auction' && detail?.bids?.length">
        <div class="mb-2 text-sm font-semibold text-gray-500">{{ $t('info.contractBids') }}</div>
        <ElTable :data="sortedBids" size="small" border>
          <ElTableColumn
            prop="amount"
            :label="$t('info.contractBidAmount')"
            width="160"
            align="right"
          >
            <template #default="{ row }">
              <strong>{{ formatISK(row.amount) }}</strong>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="bidder_id" :label="$t('info.contractBidder')" />
          <ElTableColumn prop="date_bid" :label="$t('info.contractBidTime')" width="180">
            <template #default="{ row }">{{ new Date(row.date_bid).toLocaleString() }}</template>
          </ElTableColumn>
        </ElTable>
      </template>

      <ElEmpty
        v-if="
          !loading &&
          !detail?.items?.length &&
          !(contractType === 'auction' && detail?.bids?.length)
        "
        :image-size="60"
      />
    </div>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElDialog, ElTable, ElTableColumn, ElTag, ElEmpty } from 'element-plus'
  import { fetchInfoContractDetail } from '@/api/eve-info'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'ContractDetailDialog' })

  interface Props {
    visible: boolean
    characterId: number
    contractId: number
    contractType: string
    contractTitle?: string
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{ (e: 'update:visible', val: boolean): void }>()

  const userStore = useUserStore()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (val) => emit('update:visible', val)
  })

  const loading = ref(false)
  const detail = ref<Api.EveInfo.ContractDetailResponse | null>(null)

  const sortedBids = computed(() =>
    detail.value?.bids ? [...detail.value.bids].sort((a, b) => b.amount - a.amount) : []
  )

  const formatISK = (v: number) => {
    if (v >= 1_000_000_000) return `${(v / 1_000_000_000).toFixed(2)}B`
    if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(2)}M`
    if (v >= 1_000) return `${(v / 1_000).toFixed(2)}K`
    return v.toLocaleString()
  }

  const loadDetail = async () => {
    if (!props.characterId || !props.contractId) return
    loading.value = true
    detail.value = null
    try {
      const res = await fetchInfoContractDetail({
        character_id: props.characterId,
        contract_id: props.contractId,
        language: userStore.language
      })
      detail.value = res
    } catch {
      detail.value = null
    } finally {
      loading.value = false
    }
  }

  watch(
    () => props.visible,
    (v) => {
      if (v) loadDetail()
    }
  )
</script>
