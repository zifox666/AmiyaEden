<!-- 技能规划管理页面（管理员/FC） -->
<template>
  <div class="skill-plan-admin art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="getData">
        <template #left>
          <ElButton type="primary" :icon="Plus" @click="openCreateDialog">
            {{ $t('skillPlan.create') }}
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

    <!-- 创建/编辑对话框 -->
    <ElDialog
      v-model="dialogVisible"
      :title="editingPlan ? $t('skillPlan.edit') : $t('skillPlan.create')"
      width="640px"
      destroy-on-close
    >
      <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <ElFormItem :label="$t('skillPlan.fields.name')" prop="name">
          <ElInput v-model="formData.name" :placeholder="$t('skillPlan.fields.namePlaceholder')" />
        </ElFormItem>
        <ElFormItem :label="$t('skillPlan.fields.description')" prop="description">
          <ElInput
            v-model="formData.description"
            type="textarea"
            :rows="2"
            :placeholder="$t('skillPlan.fields.descriptionPlaceholder')"
          />
        </ElFormItem>
        <ElFormItem :label="$t('skillPlan.fields.skillText')" prop="skill_text">
          <ElInput
            v-model="formData.skill_text"
            type="textarea"
            :rows="10"
            :placeholder="$t('skillPlan.fields.skillTextPlaceholder')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="saving" @click="handleSave">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, h } from 'vue'
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import {
    fetchSkillPlanList,
    fetchSkillPlanDetail,
    createSkillPlan,
    updateSkillPlan,
    deleteSkillPlan
  } from '@/api/skill-plan'
  import {
    ElButton,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInput,
    ElMessageBox,
    ElMessage,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { Plus } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'SkillPlanAdmin' })

  const { t } = useI18n()

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
      apiFn: fetchSkillPlanList,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'name',
          label: t('skillPlan.fields.name'),
          minWidth: 180,
          showOverflowTooltip: true
        },
        {
          prop: 'description',
          label: t('skillPlan.fields.description'),
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'items',
          label: t('skillPlan.fields.skillCount'),
          width: 120,
          formatter: (row: Api.SkillPlan.SkillPlanDTO) => row.items?.length ?? 0
        },
        {
          prop: 'created_at',
          label: t('skillPlan.fields.createdAt'),
          width: 180,
          formatter: (row: Api.SkillPlan.SkillPlanDTO) => {
            if (!row.created_at) return '-'
            return new Date(row.created_at).toLocaleString()
          }
        },
        {
          prop: 'actions',
          label: t('common.operation'),
          width: 160,
          fixed: 'right',
          formatter: (row: Api.SkillPlan.SkillPlanDTO) =>
            h('div', { class: 'flex gap-1' }, [
              h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
              h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
            ])
        }
      ]
    }
  })

  // ── 对话框 ──
  const dialogVisible = ref(false)
  const saving = ref(false)
  const editingPlan = ref<Api.SkillPlan.SkillPlanDTO | null>(null)
  const formRef = ref<FormInstance>()

  const formData = reactive({
    name: '',
    description: '',
    skill_text: ''
  })

  const formRules: FormRules = {
    name: [{ required: true, message: t('skillPlan.validation.nameRequired'), trigger: 'blur' }],
    skill_text: [
      { required: true, message: t('skillPlan.validation.skillTextRequired'), trigger: 'blur' }
    ]
  }

  function openCreateDialog() {
    editingPlan.value = null
    formData.name = ''
    formData.description = ''
    formData.skill_text = ''
    dialogVisible.value = true
  }

  async function openEditDialog(row: Api.SkillPlan.SkillPlanDTO) {
    editingPlan.value = row
    formData.name = row.name
    formData.description = row.description
    formData.skill_text = ''
    dialogVisible.value = true
    try {
      const detail = await fetchSkillPlanDetail(row.id, 'en')
      formData.skill_text = (detail.items ?? [])
        .map((item) => `${item.skill_name} ${item.required_level}`)
        .join('\n')
    } catch {
      // 获取失败静默处理，用户可手动重新输入
    }
  }

  async function handleSave() {
    if (!formRef.value) return
    await formRef.value.validate()
    saving.value = true
    try {
      if (editingPlan.value) {
        await updateSkillPlan(editingPlan.value.id, {
          name: formData.name,
          description: formData.description,
          skill_text: formData.skill_text
        })
        ElMessage.success(t('common.updateSuccess'))
      } else {
        await createSkillPlan({
          name: formData.name,
          description: formData.description,
          skill_text: formData.skill_text
        })
        ElMessage.success(t('common.createSuccess'))
      }
      dialogVisible.value = false
      getData()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e)
      ElMessage.error(msg)
    } finally {
      saving.value = false
    }
  }

  function handleDelete(row: Api.SkillPlan.SkillPlanDTO) {
    ElMessageBox.confirm(t('skillPlan.deleteConfirm'), t('common.tips'), {
      type: 'warning'
    }).then(async () => {
      try {
        await deleteSkillPlan(row.id)
        ElMessage.success(t('common.deleteSuccess'))
        getData()
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : String(e)
        ElMessage.error(msg)
      }
    })
  }
</script>
