<!-- 军团建筑管理页面 -->
<template>
  <div class="corp-structures art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="getData">
        <template #left>
          <!-- 军团选择器（多军团时显示） -->
          <ElSelect
            v-if="corpIDs.length > 1"
            v-model="selectedCorpID"
            :placeholder="$t('corpStructure.selectCorp')"
            style="width: 220px; margin-right: 12px"
            @change="handleCorpChange"
          >
            <ElOption v-for="id in corpIDs" :key="id" :label="getCorpLabel(id)" :value="id" />
          </ElSelect>

          <!-- 状态筛选 -->
          <ElSelect
            v-model="stateFilter"
            :placeholder="$t('corpStructure.stateFilter')"
            style="width: 180px; margin-right: 12px"
            clearable
            @change="handleFilterChange"
          >
            <ElOption
              v-for="opt in stateOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </ElSelect>

          <!-- 低油量开关 -->
          <span style="display: inline-flex; align-items: center; gap: 6px">
            <ElSwitch v-model="fuelExpiresSoon" @change="handleFilterChange" />
            <span style="font-size: 13px; color: var(--el-text-color-regular)">
              {{ $t('corpStructure.fuelExpiresFilter') }}
            </span>
          </span>
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
  </div>
</template>

<script setup lang="ts">
  import { ref, h, onMounted, computed } from 'vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchCorpStructureList, fetchCorpIDs } from '@/api/corp-structure'
  import { fetchNames } from '@/api/sde'
  import { ElTag, ElSelect, ElOption, ElSwitch } from 'element-plus'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'CorpStructures' })

  const { t } = useI18n()

  const corpIDs = ref<number[]>([])
  const selectedCorpID = ref<number | undefined>(undefined)
  const corpNames = ref<Map<number, string>>(new Map())
  const stateFilter = ref<string>('')
  const fuelExpiresSoon = ref<boolean>(false)

  function getCorpLabel(id: number): string {
    return corpNames.value.get(id) ?? t('corpStructure.corpPrefix', { id })
  }

  const stateOptions = computed(() => [
    { value: 'shield_vulnerable', label: t('corpStructure.stateLabels.shield_vulnerable') },
    { value: 'armor_reinforce', label: t('corpStructure.stateLabels.armor_reinforce') },
    { value: 'armor_vulnerable', label: t('corpStructure.stateLabels.armor_vulnerable') },
    { value: 'hull_reinforce', label: t('corpStructure.stateLabels.hull_reinforce') },
    { value: 'hull_vulnerable', label: t('corpStructure.stateLabels.hull_vulnerable') },
    { value: 'anchoring', label: t('corpStructure.stateLabels.anchoring') },
    { value: 'unanchored', label: t('corpStructure.stateLabels.unanchored') },
    { value: 'fitting_invulnerable', label: t('corpStructure.stateLabels.fitting_invulnerable') },
    { value: 'onlining_vulnerable', label: t('corpStructure.stateLabels.onlining_vulnerable') }
  ])

  // 建筑状态颜色映射
  const STATE_MAP: Record<string, string> = {
    shield_vulnerable: 'success',
    armor_reinforce: 'warning',
    armor_vulnerable: 'warning',
    hull_reinforce: 'danger',
    hull_vulnerable: 'danger',
    anchoring: 'info',
    unanchored: 'info',
    fitting_invulnerable: 'success',
    onlining_vulnerable: 'info',
    online_deprecated: 'info',
    anchor_vulnerable: 'warning',
    deploy_vulnerable: 'warning',
    unknown: 'info'
  }

  // 服务状态颜色映射
  const SERVICE_STATE_MAP: Record<string, string> = {
    online: 'success',
    offline: 'danger',
    cleanup: 'warning'
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    handleSizeChange,
    handleCurrentChange
  } = useTable({
    core: {
      apiFn: fetchCorpStructureList,
      apiParams: { current: 1, size: 20 },
      immediate: false,
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'name',
          label: t('corpStructure.fields.name'),
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'state',
          label: t('corpStructure.fields.state'),
          width: 160,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            const label = t(`corpStructure.stateLabels.${row.state}`, row.state)
            return h(
              ElTag,
              {
                type: (STATE_MAP[row.state] ?? 'info') as any,
                effect: 'dark',
                size: 'small'
              },
              () => label
            )
          }
        },
        {
          prop: 'system_id',
          label: t('corpStructure.fields.systemId'),
          width: 120
        },
        {
          prop: 'type_id',
          label: t('corpStructure.fields.typeId'),
          width: 120
        },
        {
          prop: 'fuel_expires',
          label: t('corpStructure.fields.fuelExpires'),
          width: 180,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            if (!row.fuel_expires) return '-'
            return new Date(row.fuel_expires).toLocaleString()
          }
        },
        {
          prop: 'reinforce_hour',
          label: t('corpStructure.fields.reinforceHour'),
          width: 100
        },
        {
          prop: 'services',
          label: t('corpStructure.fields.services'),
          minWidth: 200,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            if (!row.services || row.services.length === 0) return '-'
            return h(
              'div',
              { class: 'flex flex-wrap gap-1' },
              row.services.map((svc) =>
                h(
                  ElTag,
                  {
                    type: (SERVICE_STATE_MAP[svc.state] ?? 'info') as any,
                    size: 'small'
                  },
                  () => svc.name
                )
              )
            )
          }
        },
        {
          prop: 'state_timer_end',
          label: t('corpStructure.fields.stateTimerEnd'),
          width: 180,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            if (!row.state_timer_end) return '-'
            return new Date(row.state_timer_end).toLocaleString()
          }
        },
        {
          prop: 'unanchors_at',
          label: t('corpStructure.fields.unanchorsAt'),
          width: 180,
          formatter: (row: Api.CorpStructure.StructureItem) => {
            if (!row.unanchors_at) return '-'
            return new Date(row.unanchors_at).toLocaleString()
          }
        }
      ]
    }
  })

  function handleCorpChange(corpID: number) {
    Object.assign(searchParams, { corp_id: corpID })
    getData()
  }

  function handleFilterChange() {
    Object.assign(searchParams, {
      state: stateFilter.value || undefined,
      fuel_expires_soon: fuelExpiresSoon.value || undefined
    })
    getData()
  }

  async function loadCorpIDs() {
    try {
      const res = await fetchCorpIDs()
      corpIDs.value = res ?? []
      if (corpIDs.value.length > 0) {
        // 通过 ESI /universe/names 解析军团名称
        try {
          const nameMap = await fetchNames({ esi: corpIDs.value })
          const map = new Map<number, string>()
          corpIDs.value.forEach((id) => {
            const name = (nameMap as Record<string, string>)[String(id)]
            if (name) map.set(id, name)
          })
          corpNames.value = map
        } catch {
          // 解析名称失败，降级显示 ID
        }
        selectedCorpID.value = corpIDs.value[0]
        Object.assign(searchParams, { corp_id: corpIDs.value[0] })
        getData()
      }
    } catch {
      corpIDs.value = []
    }
  }

  onMounted(() => {
    loadCorpIDs()
  })
</script>
