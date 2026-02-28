<template>
  <ElRow :gutter="20" class="flex">
    <ElCol v-for="(item, index) in dataList" :key="index" :sm="12" :md="6" :lg="6">
      <div class="art-card relative flex flex-col justify-center h-35 px-5 mb-5 max-sm:mb-4">
        <span class="text-g-700 text-sm">{{ item.label }}</span>
        <ArtCountTo
          class="text-[26px] font-medium mt-2"
          :target="item.value"
          :duration="1300"
          :decimals="item.decimals"
          :suffix="item.suffix"
          separator=","
        />
        <div class="flex-c mt-1">
          <span class="text-xs text-g-600">{{ item.desc }}</span>
        </div>
        <div
          class="absolute top-0 bottom-0 right-5 m-auto size-12.5 rounded-xl flex-cc bg-theme/10"
        >
          <ArtSvgIcon :icon="item.icon" class="text-xl text-theme" />
        </div>
      </div>
    </ElCol>
  </ElRow>
</template>

<script setup lang="ts">
  import { humanizeNumber } from '@/utils/common/text'

  const props = defineProps<{
    cards?: Api.Dashboard.Cards
  }>()

  interface CardDataItem {
    label: string
    icon: string
    value: number
    decimals: number
    suffix: string
    desc: string
  }

  const dataList = computed<CardDataItem[]>(() => {
    const c = props.cards
    return [
      {
        label: 'EVE 钱包余额',
        icon: 'ri:wallet-3-line',
        value: c?.eve_wallet_balance ?? 0,
        decimals: 2,
        suffix: '',
        desc: humanizeNumber(c?.eve_wallet_balance ?? 0)
      },
      {
        label: 'EVE 技能点',
        icon: 'ri:brain-line',
        value: c?.eve_skill_points ?? 0,
        decimals: 0,
        suffix: '',
        desc: humanizeNumber(c?.eve_skill_points ?? 0)
      },
      {
        label: '系统钱包',
        icon: 'ri:money-cny-circle-line',
        value: c?.system_wallet_balance ?? 0,
        decimals: 2,
        suffix: '',
        desc: ''
      },
      {
        label: '当月联盟 PAP',
        icon: 'ri:shield-star-line',
        value: c?.alliance_pap ?? 0,
        decimals: 1,
        suffix: '',
        desc: ''
      }
    ]
  })
</script>
