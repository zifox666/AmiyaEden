<!-- 舰队配置管理页面 -->
<template>
  <div class="fleet-config-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="getData">
        <template #left>
          <ElButton v-if="canManage" type="primary" :icon="Plus" @click="openCreateDialog">
            {{ $t('fleetConfig.create') }}
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

    <!-- 创建 / 编辑 / 查看弹窗 -->
    <FleetConfigDialog
      v-model:visible="dialogVisible"
      :editing="editingConfig"
      :readonly="dialogReadonly"
      @success="handleDialogSuccess"
      @created="handleDialogCreated"
    />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import FleetConfigDialog from './modules/fleet-config-dialog.vue'
  import { fetchFleetConfigList, deleteFleetConfig } from '@/api/fleet-config'
  import { ElButton, ElTag, ElMessageBox } from 'element-plus'
  import { Plus } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'FleetConfigs' })

  type FleetConfigItem = Api.FleetConfig.FleetConfigItem

  const { t } = useI18n()

  const userStore = useUserStore()
  const canManage = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((r) => ['super_admin', 'admin', 'fc', 'srp'].includes(r))
  })

  const formatTime = (v: string) => {
    if (!v) return '-'
    return new Date(v).toLocaleString()
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    handleSizeChange,
    handleCurrentChange
  } = useTable({
    core: {
      apiFn: fetchFleetConfigList,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        {
          prop: 'name',
          label: t('fleetConfig.fields.name'),
          minWidth: 180,
          showOverflowTooltip: true,
          formatter: (row: FleetConfigItem) =>
            h(
              ElButton,
              { type: 'primary', link: true, onClick: () => openEditDialog(row) },
              () => row.name
            )
        },
        {
          prop: 'description',
          label: t('fleetConfig.fields.description'),
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'fittings',
          label: t('fleetConfig.fields.fittings'),
          width: 120,
          formatter: (row: FleetConfigItem) =>
            h(
              ElTag,
              { type: 'info', size: 'small' },
              () => `${row.fittings?.length ?? 0} ${t('fleetConfig.fittingUnit')}`
            )
        },
        {
          prop: 'created_at',
          label: t('common.createdAt'),
          width: 180,
          formatter: (row: FleetConfigItem) => h('span', {}, formatTime(row.created_at))
        },
        ...(canManage.value
          ? [
              {
                prop: 'actions',
                label: t('common.operation'),
                width: 160,
                fixed: 'right' as const,
                formatter: (row: FleetConfigItem) =>
                  h('div', { class: 'flex gap-1' }, [
                    h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
                    h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) }),
                    h(ArtButtonTable, {
                      type: 'view',
                      onClick: () => openEditDialog(row),
                      disabled: !canManage.value
                    })
                  ])
              }
            ]
          : [])
      ]
    }
  })

  // ─── 创建 / 编辑 / 查看 ───
  const dialogVisible = ref(false)
  const editingConfig = ref<FleetConfigItem | null>(null)
  const dialogReadonly = ref(false)

  function openCreateDialog() {
    editingConfig.value = null
    dialogReadonly.value = false
    dialogVisible.value = true
  }

  function openEditDialog(row: FleetConfigItem) {
    editingConfig.value = row
    dialogReadonly.value = !canManage.value
    dialogVisible.value = true
  }

  function handleDialogSuccess() {
    dialogVisible.value = false
    getData()
  }

  function handleDialogCreated(config: FleetConfigItem) {
    dialogVisible.value = false
    getData()
    nextTick(() => {
      openEditDialog(config)
    })
  }

  // ─── 删除 ───
  async function handleDelete(row: FleetConfigItem) {
    await ElMessageBox.confirm(
      t('fleetConfig.deleteConfirm', { name: row.name }),
      t('fleetConfig.delete'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    try {
      await deleteFleetConfig(row.id)
      ElMessage.success(t('fleetConfig.deleteSuccess'))
      getData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    }
  }
</script>
