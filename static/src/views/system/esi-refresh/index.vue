<!-- ESI 定时任务管理页面 -->
<template>
  <div class="esi-refresh-page art-full-height">
    <!-- 任务定义列表 -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader
        v-model:columns="taskColumnChecks"
        :loading="tasksLoading"
        @refresh="loadTasks"
      >
        <template #left>
          <ElButton type="primary" :loading="runAllLoading" @click="handleRunAll">
            全量刷新
          </ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable :loading="tasksLoading" :data="tasks" :columns="taskColumns" />
    </ElCard>

    <!-- 运行状态列表 -->
    <ElCard class="art-table-card mt-4" shadow="never">
      <ArtTableHeader
        v-model:columns="statusColumnChecks"
        :loading="statusLoading"
        @refresh="refreshData"
      >
        <template #left>
          <ElSelect
            v-model="filterForm.task_name"
            placeholder="全部任务"
            clearable
            style="width: 180px"
            @change="handleSearch"
          >
            <ElOption v-for="t in tasks" :key="t.name" :label="t.description" :value="t.name" />
          </ElSelect>
          <ElSelect
            v-model="filterForm.status"
            placeholder="全部状态"
            clearable
            style="width: 130px"
            @change="handleSearch"
          >
            <ElOption label="等待中" value="pending" />
            <ElOption label="运行中" value="running" />
            <ElOption label="成功" value="success" />
            <ElOption label="失败" value="failed" />
            <ElOption label="已跳过" value="skipped" />
          </ElSelect>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="statusLoading"
        :data="statusData"
        :columns="statusColumns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElCard, ElTag, ElButton, ElMessageBox, ElSelect, ElOption } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    fetchESIRefreshTasks,
    fetchESIRefreshStatuses,
    runESIRefreshTask,
    runESIRefreshTaskByName,
    runESIRefreshAll
  } from '@/api/esi-refresh'

  defineOptions({ name: 'ESIRefresh' })

  type TaskInfo = Api.ESIRefresh.TaskInfo
  type TaskStatus = Api.ESIRefresh.TaskStatus

  // ─── 优先级 / 状态映射 ───
  const PRIORITY_MAP: Record<number, { type: string; text: string }> = {
    1: { type: 'danger', text: '极高' },
    10: { type: 'warning', text: '高' },
    50: { type: 'success', text: '普通' },
    90: { type: 'info', text: '低' }
  }
  const priorityType = (p: number) => (PRIORITY_MAP[p]?.type ?? 'info') as any
  const priorityLabel = (p: number) => PRIORITY_MAP[p]?.text ?? `P${p}`

  const STATUS_MAP: Record<string, { type: string; text: string }> = {
    pending: { type: 'info', text: '等待中' },
    running: { type: 'warning', text: '运行中' },
    success: { type: 'success', text: '成功' },
    failed: { type: 'danger', text: '失败' },
    skipped: { type: 'info', text: '已跳过' }
  }
  const statusType = (s: string) => (STATUS_MAP[s]?.type ?? 'info') as any
  const statusLabel = (s: string) => STATUS_MAP[s]?.text ?? s

  // ─── 触发状态（loading per row） ───
  const runningByName = ref(new Set<string>())
  const runningTasks = ref(new Set<string>())

  // ══════════════════════════════════════════
  // 表 1：任务定义（无分页）
  // ══════════════════════════════════════════
  const { columns: taskColumns, columnChecks: taskColumnChecks } = useTableColumns<TaskInfo>(() => [
    { type: 'index', width: 60, label: '#' },
    { prop: 'name', label: '任务名称', width: 200, showOverflowTooltip: true },
    { prop: 'description', label: '描述', minWidth: 180, showOverflowTooltip: true },
    {
      prop: 'priority',
      label: '优先级',
      width: 100,
      formatter: (row: TaskInfo) =>
        h(ElTag, { type: priorityType(row.priority), size: 'small' }, () =>
          priorityLabel(row.priority)
        )
    },
    { prop: 'active_interval', label: '活跃间隔', width: 120 },
    { prop: 'inactive_interval', label: '非活跃间隔', width: 120 },
    {
      prop: 'required_scopes',
      label: '所需权限',
      minWidth: 260,
      formatter: (row: TaskInfo) =>
        row.required_scopes?.length
          ? h(
              'div',
              { class: 'flex flex-wrap gap-1 py-1' },
              row.required_scopes.map((scope) =>
                h(ElTag, { size: 'small', effect: 'plain', key: scope }, () => scope)
              )
            )
          : h('span', { class: 'text-gray-400' }, '无需权限')
    },
    {
      prop: 'actions',
      label: '操作',
      width: 100,
      fixed: 'right',
      formatter: (row: TaskInfo) =>
        h(
          ElButton,
          {
            size: 'small',
            type: 'primary',
            loading: runningByName.value.has(row.name),
            onClick: () => handleRunTaskByName(row)
          },
          () => '执行'
        )
    }
  ])

  const tasks = ref<TaskInfo[]>([])
  const tasksLoading = ref(false)
  const runAllLoading = ref(false)

  const loadTasks = async () => {
    tasksLoading.value = true
    try {
      tasks.value = (await fetchESIRefreshTasks()) ?? []
    } catch {
      tasks.value = []
    } finally {
      tasksLoading.value = false
    }
  }

  // ══════════════════════════════════════════
  // 表 2：运行状态（分页 + 筛选）
  // ══════════════════════════════════════════
  const filterForm = reactive({ task_name: '', status: '' })

  const {
    columns: statusColumns,
    columnChecks: statusColumnChecks,
    data: statusData,
    loading: statusLoading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    getData,
    searchParams
  } = useTable({
    core: {
      apiFn: fetchESIRefreshStatuses,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        { prop: 'task_name', label: '任务名称', width: 200, showOverflowTooltip: true },
        { prop: 'description', label: '描述', minWidth: 160, showOverflowTooltip: true },
        { prop: 'character_id', label: '人物 ID', width: 120 },
        {
          prop: 'priority',
          label: '优先级',
          width: 100,
          formatter: (row: TaskStatus) =>
            h(ElTag, { type: priorityType(row.priority), size: 'small' }, () =>
              priorityLabel(row.priority)
            )
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: TaskStatus) =>
            h(ElTag, { type: statusType(row.status), size: 'small' }, () => statusLabel(row.status))
        },
        {
          prop: 'last_run',
          label: '上次运行',
          width: 180,
          formatter: (row: TaskStatus) => h('span', {}, row.last_run ?? '-')
        },
        {
          prop: 'next_run',
          label: '下次运行',
          width: 180,
          formatter: (row: TaskStatus) => h('span', {}, row.next_run ?? '-')
        },
        {
          prop: 'error',
          label: '错误信息',
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: TaskStatus) =>
            row.error
              ? h('span', { class: 'text-red-500' }, row.error)
              : h('span', { class: 'text-gray-400' }, '-')
        },
        {
          prop: 'actions',
          label: '操作',
          width: 100,
          fixed: 'right',
          formatter: (row: TaskStatus) =>
            h(
              ElButton,
              {
                size: 'small',
                type: 'primary',
                loading: runningTasks.value.has(`${row.task_name}_${row.character_id}`),
                onClick: () => handleRunTask(row)
              },
              () => '执行'
            )
        }
      ]
    }
  })

  const handleSearch = () => {
    Object.assign(searchParams, {
      task_name: filterForm.task_name || undefined,
      status: filterForm.status || undefined
    })
    getData()
  }

  // ─── 手动触发（按任务名 — 所有人物） ───
  const handleRunTaskByName = async (row: TaskInfo) => {
    await ElMessageBox.confirm(`确定要对所有人物执行「${row.description}」吗？`, '执行确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    runningByName.value.add(row.name)
    try {
      await runESIRefreshTaskByName({ task_name: row.name })
      ElMessage.success(`任务「${row.description}」已触发（所有人物）`)
      setTimeout(() => refreshData(), 2000)
    } catch {
      ElMessage.error(`任务「${row.description}」触发失败`)
    } finally {
      runningByName.value.delete(row.name)
    }
  }

  // ─── 手动触发（指定人物） ───
  const handleRunTask = async (row: TaskStatus) => {
    const key = `${row.task_name}_${row.character_id}`
    runningTasks.value.add(key)
    try {
      await runESIRefreshTask({ task_name: row.task_name, character_id: row.character_id })
      ElMessage.success(`任务 ${row.task_name} 已触发`)
      refreshData()
    } catch {
      ElMessage.error(`任务 ${row.task_name} 触发失败`)
    } finally {
      runningTasks.value.delete(key)
    }
  }

  // ─── 全量刷新 ───
  const handleRunAll = async () => {
    await ElMessageBox.confirm(
      '确定要触发全量刷新吗？这将在后台异步执行所有数据刷新任务。',
      '全量刷新确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    runAllLoading.value = true
    try {
      await runESIRefreshAll()
      ElMessage.success('全量刷新已触发，任务将在后台执行')
      setTimeout(() => refreshData(), 2000)
    } catch {
      ElMessage.error('全量刷新触发失败')
    } finally {
      runAllLoading.value = false
    }
  }

  // ─── 初始化 ───
  onMounted(loadTasks)
</script>
