<!-- ESI 定时任务管理页面 -->
<template>
  <div class="esi-refresh-page">
    <!-- 任务定义列表 -->
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="card-title">任务列表</span>
          <ElButton type="primary" :loading="runAllLoading" @click="handleRunAll">
            <el-icon class="mr-1"><Refresh /></el-icon>
            全量刷新
          </ElButton>
        </div>
      </template>

      <div class="table-container">
        <ElTable v-loading="tasksLoading" :data="tasks" stripe border>
          <ElTableColumn prop="name" label="任务名称" width="200" />
          <ElTableColumn prop="description" label="描述" min-width="200" />
          <ElTableColumn prop="priority" label="优先级" width="100" align="center">
            <template #default="{ row }">
              <ElTag :type="priorityType(row.priority)" size="small">
                {{ priorityLabel(row.priority) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="active_interval" label="活跃间隔" width="120" align="center" />
          <ElTableColumn prop="inactive_interval" label="非活跃间隔" width="120" align="center" />
          <ElTableColumn prop="required_scopes" label="所需权限" min-width="260">
            <template #default="{ row }">
              <template v-if="row.required_scopes?.length">
                <ElTag
                  v-for="scope in row.required_scopes"
                  :key="scope"
                  size="small"
                  class="scope-tag"
                  effect="plain"
                >
                  {{ scope }}
                </ElTag>
              </template>
              <span v-else class="text-gray-400">无需权限</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="100" fixed="right" align="center">
            <template #default="{ row }">
              <ElButton
                type="primary"
                link
                size="small"
                :loading="runningByName.has(row.name)"
                @click="handleRunTaskByName(row)"
              >
                执行
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>
    </ElCard>

    <!-- 运行状态列表 -->
    <ElCard class="art-table-card mt-4" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="card-title">运行状态</span>
          <div class="status-toolbar">
            <ElSelect
              v-model="statusFilter.task_name"
              placeholder="全部任务"
              clearable
              style="width: 180px"
              @change="handleFilterChange"
            >
              <ElOption v-for="t in tasks" :key="t.name" :label="t.description" :value="t.name" />
            </ElSelect>
            <ElSelect
              v-model="statusFilter.status"
              placeholder="全部状态"
              clearable
              style="width: 120px; margin-left: 8px"
              @change="handleFilterChange"
            >
              <ElOption label="等待中" value="pending" />
              <ElOption label="运行中" value="running" />
              <ElOption label="成功" value="success" />
              <ElOption label="失败" value="failed" />
              <ElOption label="已跳过" value="skipped" />
            </ElSelect>
            <ElButton :loading="statusLoading" style="margin-left: 8px" @click="loadStatuses">
              <el-icon class="mr-1"><Refresh /></el-icon>
              刷新状态
            </ElButton>
          </div>
        </div>
      </template>

      <div class="table-container">
        <ElTable v-loading="statusLoading" :data="statuses" stripe border>
          <ElTableColumn prop="task_name" label="任务名称" width="200" />
          <ElTableColumn prop="description" label="描述" min-width="160" />
          <ElTableColumn prop="character_id" label="角色 ID" width="120" align="center" />
          <ElTableColumn prop="priority" label="优先级" width="100" align="center">
            <template #default="{ row }">
              <ElTag :type="priorityType(row.priority)" size="small">
                {{ priorityLabel(row.priority) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <ElTag :type="statusType(row.status)" size="small">
                {{ statusLabel(row.status) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="last_run" label="上次运行" width="180">
            <template #default="{ row }">
              {{ row.last_run || '-' }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="next_run" label="下次运行" width="180">
            <template #default="{ row }">
              {{ row.next_run || '-' }}
            </template>
          </ElTableColumn>
          <ElTableColumn prop="error" label="错误信息" min-width="200">
            <template #default="{ row }">
              <span v-if="row.error" class="text-red-500">{{ row.error }}</span>
              <span v-else class="text-gray-400">-</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="100" fixed="right" align="center">
            <template #default="{ row }">
              <ElButton
                type="primary"
                link
                size="small"
                :loading="runningTasks.has(`${row.task_name}_${row.character_id}`)"
                @click="handleRunTask(row)"
              >
                执行
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.current"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
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
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElMessageBox,
    ElPagination,
    ElSelect,
    ElOption
  } from 'element-plus'
  import {
    fetchESIRefreshTasks,
    fetchESIRefreshStatuses,
    runESIRefreshTask,
    runESIRefreshTaskByName,
    runESIRefreshAll
  } from '@/api/esi-refresh'

  defineOptions({ name: 'ESIRefresh' })

  // ---- 数据 ----
  const tasks = ref<Api.ESIRefresh.TaskInfo[]>([])
  const statuses = ref<Api.ESIRefresh.TaskStatus[]>([])
  const tasksLoading = ref(false)
  const statusLoading = ref(false)
  const runAllLoading = ref(false)
  const runningTasks = ref(new Set<string>())
  const runningByName = ref(new Set<string>())

  // ---- 分页 ----
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  // ---- 筛选 ----
  const statusFilter = reactive({
    task_name: '',
    status: ''
  })

  // ---- 优先级映射 ----
  const PRIORITY_MAP: Record<number, { type: string; text: string }> = {
    1: { type: 'danger', text: '极高' },
    10: { type: 'warning', text: '高' },
    50: { type: 'success', text: '普通' },
    90: { type: 'info', text: '低' }
  }

  const priorityType = (p: number) => (PRIORITY_MAP[p]?.type || 'info') as any
  const priorityLabel = (p: number) => PRIORITY_MAP[p]?.text ?? `P${p}`

  // ---- 状态映射 ----
  const STATUS_MAP: Record<string, { type: string; text: string }> = {
    pending: { type: 'info', text: '等待中' },
    running: { type: 'warning', text: '运行中' },
    success: { type: 'success', text: '成功' },
    failed: { type: 'danger', text: '失败' },
    skipped: { type: 'info', text: '已跳过' }
  }

  const statusType = (s: string) => (STATUS_MAP[s]?.type || 'info') as any
  const statusLabel = (s: string) => STATUS_MAP[s]?.text ?? s

  // ---- 加载数据 ----
  const loadTasks = async () => {
    tasksLoading.value = true
    try {
      const res = await fetchESIRefreshTasks()
      tasks.value = res ?? []
    } catch {
      tasks.value = []
    } finally {
      tasksLoading.value = false
    }
  }

  const loadStatuses = async () => {
    statusLoading.value = true
    try {
      const params: Api.ESIRefresh.TaskStatusSearchParams = {
        current: pagination.current,
        size: pagination.size
      }
      if (statusFilter.task_name) params.task_name = statusFilter.task_name
      if (statusFilter.status) params.status = statusFilter.status

      const res = await fetchESIRefreshStatuses(params)
      if (res) {
        statuses.value = res.records ?? []
        pagination.total = res.total ?? 0
        pagination.current = res.current ?? 1
        pagination.size = res.size ?? 20
      } else {
        statuses.value = []
        pagination.total = 0
      }
    } catch {
      statuses.value = []
      pagination.total = 0
    } finally {
      statusLoading.value = false
    }
  }

  // ---- 分页事件 ----
  const handleSizeChange = () => {
    pagination.current = 1
    loadStatuses()
  }

  const handleCurrentChange = () => {
    loadStatuses()
  }

  const handleFilterChange = () => {
    pagination.current = 1
    loadStatuses()
  }

  // ---- 手动触发（按任务名 — 所有角色） ----
  const handleRunTaskByName = async (row: Api.ESIRefresh.TaskInfo) => {
    try {
      await ElMessageBox.confirm(`确定要对所有角色执行「${row.description}」吗？`, '执行确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    runningByName.value.add(row.name)
    try {
      await runESIRefreshTaskByName({ task_name: row.name })
      ElMessage.success(`任务「${row.description}」已触发（所有角色）`)
      setTimeout(() => loadStatuses(), 2000)
    } catch {
      ElMessage.error(`任务「${row.description}」触发失败`)
    } finally {
      runningByName.value.delete(row.name)
    }
  }

  // ---- 手动触发（指定角色） ----
  const handleRunTask = async (row: Api.ESIRefresh.TaskStatus) => {
    const key = `${row.task_name}_${row.character_id}`
    runningTasks.value.add(key)
    try {
      await runESIRefreshTask({
        task_name: row.task_name,
        character_id: row.character_id
      })
      ElMessage.success(`任务 ${row.task_name} 已触发`)
      await loadStatuses()
    } catch {
      ElMessage.error(`任务 ${row.task_name} 触发失败`)
    } finally {
      runningTasks.value.delete(key)
    }
  }

  // ---- 全量刷新 ----
  const handleRunAll = async () => {
    try {
      await ElMessageBox.confirm(
        '确定要触发全量刷新吗？这将在后台异步执行所有数据刷新任务。',
        '全量刷新确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    runAllLoading.value = true
    try {
      await runESIRefreshAll()
      ElMessage.success('全量刷新已触发，任务将在后台执行')
      setTimeout(() => loadStatuses(), 2000)
    } catch {
      ElMessage.error('全量刷新触发失败')
    } finally {
      runAllLoading.value = false
    }
  }

  // ---- 初始化 ----
  onMounted(() => {
    loadTasks()
    loadStatuses()
  })
</script>

<style scoped lang="scss">
  .esi-refresh-page {
    padding: 16px;
    height: 100%;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }

  .art-table-card {
    flex-shrink: 0;
  }

  .table-container {
    width: 100%;
    overflow-x: auto;
  }

  .table-container :deep(.el-table) {
    width: 100%;
    min-width: 1200px;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .card-title {
    font-size: 16px;
    font-weight: 600;
  }

  .status-toolbar {
    display: flex;
    align-items: center;
  }

  .scope-tag {
    margin: 2px 4px 2px 0;
  }

  .mt-4 {
    margin-top: 16px;
  }

  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    padding: 16px 0 4px;
  }

  .mr-1 {
    margin-right: 4px;
  }
</style>
