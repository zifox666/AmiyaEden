// 从 ESLint 插件中导入推荐配置
import pluginJs from '@eslint/js'
import eslintPluginPrettierRecommended from 'eslint-plugin-prettier/recommended'
import pluginVue from 'eslint-plugin-vue'
import globals from 'globals'
import tseslint from 'typescript-eslint'

export default [
  // 指定文件匹配规则
  {
    files: ['**/*.{js,mjs,cjs,ts,vue}']
  },
  // 指定全局变量和环境
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      }
    }
  },
  // 扩展配置
  pluginJs.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs['flat/essential'],
  // 自定义规则
  {
    // 针对所有 JavaScript、TypeScript 和 Vue 文件应用以下配置
    files: ['**/*.{js,mjs,cjs,ts,vue}'],

    languageOptions: {
      globals: {
        // TypeScript 全局命名空间
        Api: 'readonly'
      }
    },
    rules: {
      quotes: ['error', 'single'], // 使用单引号
      semi: ['error', 'never'], // 语句末尾不加分号
      'no-var': 'error', // 要求使用 let 或 const 而不是 var
      '@typescript-eslint/no-explicit-any': 'off', // 禁用 any 检查
      'vue/multi-word-component-names': 'off', // 禁用对 Vue 组件名称的多词要求检查
      'no-multiple-empty-lines': ['warn', { max: 1 }], // 不允许多个空行
      'no-unexpected-multiline': 'error' // 禁止空余的多行
    }
  },
  {
    files: ['**/*.{ts,vue}'],
    rules: {
      // TypeScript 负责标识符解析，避免与自动导入声明重复报错。
      'no-undef': 'off'
    }
  },
  // vue 规则
  {
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: { parser: tseslint.parser }
    }
  },
  // 忽略文件
  {
    ignores: [
      'node_modules',
      'dist',
      'public',
      '.vscode/**',
      'src/assets/**',
      'src/utils/console.ts'
    ]
  },
  // prettier 配置
  eslintPluginPrettierRecommended
]
