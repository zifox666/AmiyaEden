<!-- 导入 Excel 文件 -->
<template>
  <div class="inline-block">
    <ElUpload
      :auto-upload="false"
      accept=".xlsx, .xls"
      :show-file-list="false"
      @change="handleFileChange"
    >
      <ElButton type="primary" v-ripple>
        <slot>导入 Excel</slot>
      </ElButton>
    </ElUpload>
  </div>
</template>

<script setup lang="ts">
  import ExcelJS from 'exceljs'
  import type { UploadFile } from 'element-plus'

  defineOptions({ name: 'ArtExcelImport' })

  // Excel 导入工具函数
  async function importExcel(file: File): Promise<Array<Record<string, unknown>>> {
    const arrayBuffer = await file.arrayBuffer()
    const workbook = new ExcelJS.Workbook()
    await workbook.xlsx.load(arrayBuffer)

    const worksheet = workbook.worksheets[0]
    const headers: string[] = []
    worksheet.getRow(1).eachCell((cell) => {
      headers.push(String(cell.value ?? ''))
    })

    const results: Array<Record<string, unknown>> = []
    worksheet.eachRow((row, rowNumber) => {
      if (rowNumber === 1) return
      const rowData: Record<string, unknown> = {}
      row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
        const header = headers[colNumber - 1]
        if (header) rowData[header] = cell.value
      })
      results.push(rowData)
    })

    return results
  }

  // 定义 emits
  const emit = defineEmits<{
    'import-success': [data: Array<Record<string, unknown>>]
    'import-error': [error: Error]
  }>()

  // 处理文件导入
  const handleFileChange = async (uploadFile: UploadFile) => {
    try {
      if (!uploadFile.raw) return
      const results = await importExcel(uploadFile.raw)
      emit('import-success', results)
    } catch (error) {
      emit('import-error', error as Error)
    }
  }
</script>
