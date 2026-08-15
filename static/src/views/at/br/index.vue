<template>
  <div class="p-4 md:p-6">
    <div class="art-card p-5">
      <div class="flex-b mb-5"
        ><h1 class="text-xl">{{ t('at.br.title') }}</h1
        ><ElButton v-if="canManage" type="primary" @click="build">{{
          t('at.br.build')
        }}</ElButton></div
      >
      <template v-if="br">
        <div class="grid grid-cols-2 gap-4 mb-5"
          ><div v-for="team in br.teams" :key="team.side" class="border border-g-300 rounded p-4"
            ><h2 class="text-lg">{{ t('at.br.team') }} {{ team.side }}</h2
            ><div class="text-3xl mt-2">{{ team.destroyed_points }}</div
            ><div class="text-sm text-g-500 mt-2"
              >{{ t('at.br.hostileDamage') }} {{ team.hostile_damage }} ·
              {{ t('at.br.friendlyDamage') }} {{ team.friendly_damage }}</div
            ></div
          ></div
        >
        <div class="grid grid-cols-1 xl:grid-cols-2 gap-5"
          ><div v-for="side in ['A', 'B']" :key="side"
            ><h3 class="mb-2">{{ t('at.br.team') }} {{ side }}</h3
            ><ElTable :data="participants(side)"
              ><ElTableColumn prop="ship_name" :label="t('at.km.ship')" /><ElTableColumn
                prop="character_name"
                :label="t('at.br.character')"
                ><template #default="{ row }"
                  ><span :class="row.dead ? 'bg-red-100 dark:bg-red-950 px-2 py-1 rounded' : ''">{{
                    row.character_name
                  }}</span></template
                ></ElTableColumn
              ><ElTableColumn :label="t('at.br.damage')"
                ><template #default="{ row }"
                  ><div>{{ row.hostile_damage }} ({{ row.damage_percentage.toFixed(1) }}%)</div
                  ><div v-if="row.friendly_damage" class="text-orange-500"
                    >+{{ row.friendly_damage }}</div
                  ></template
                ></ElTableColumn
              ><ElTableColumn :label="t('at.br.signatures')"
                ><template #default="{ row }"
                  >{{ row.signatures }} / {{ row.final_blows }}</template
                ></ElTableColumn
              ></ElTable
            ></div
          ></div
        >
      </template>
      <ElEmpty v-else :description="t('at.br.empty')" />
    </div>
  </div>
</template>
<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { atApi } from '@/api/at'
  import { useUserStore } from '@/store/modules/user'
  const { t } = useI18n(),
    route = useRoute(),
    user = useUserStore(),
    br = ref<Api.AT.BRDetail>()
  const roundId = computed(() => Number(route.params.roundId)),
    canManage = computed(() =>
      (user.info.roles ?? []).some((r) => ['fc', 'admin', 'super_admin'].includes(r))
    )
  const participants = (side: string) =>
    br.value?.participants.filter((item) => item.side === side) ?? []
  async function load() {
    try {
      br.value = await atApi.latestBR(roundId.value)
    } catch {
      br.value = undefined
    }
  }
  async function build() {
    br.value = await atApi.buildBR(roundId.value)
  }
  onMounted(load)
</script>
