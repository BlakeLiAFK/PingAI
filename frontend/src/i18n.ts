import { ref, computed } from 'vue'

export type Locale = 'zh' | 'en'

const STORAGE_KEY = 'pingai_locale'

function getInitLocale(): Locale {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'en' || saved === 'zh') return saved
  return navigator.language.startsWith('zh') ? 'zh' : 'en'
}

export const locale = ref<Locale>(getInitLocale())

export function setLocale(l: Locale) {
  locale.value = l
  localStorage.setItem(STORAGE_KEY, l)
}

const messages: Record<Locale, Record<string, string>> = {
  zh: {
    // App
    'app.export': '导出',
    'app.batchCheck': '批量检测',
    'app.emptyHint': '选择供应商, 输入 API Key, 点击检测',
    'app.emptySubHint': '支持 OpenAI / Anthropic / Gemini 三种协议',

    // ConfigPanel
    'config.baseURL': 'Base URL',
    'config.apiKey': 'API Key',
    'config.protocol': '协议',
    'config.model': '模型',
    'config.reset': '重置',
    'config.batch': '批量',
    'config.check': '检测',

    // Sidebar
    'sidebar.providers': '供应商',
    'sidebar.history': '历史记录',

    // CheckCard
    'card.totalLatency': '总耗时',
    'card.availableModels': '可用模型',
    'card.collapse': '收起',

    // CheckItems
    'item.connectivity': '连通性',
    'item.chat': '对话测试',
    'item.stream': '流式输出',
    'item.models': '模型列表',
    'item.multi_turn': '多轮对话',

    // History
    'history.total': '共 {n} 条记录',
    'history.selectAll': '全选',
    'history.deselectAll': '取消全选',
    'history.deleteSelected': '删除选中 ({n})',
    'history.clearAll': '清空全部',
    'history.refresh': '刷新',
    'history.confirmDelete': '确定删除 {n} 条记录?',
    'history.confirmClear': '确定清空全部历史记录?',
    'history.empty': '暂无历史记录',

    // BatchKeyDialog
    'batchKey.title': '批量 Key 检测',
    'batchKey.label': 'API Keys（每行一个）',
    'batchKey.count': '共 {n} 个 Key',
    'batchKey.start': '开始检测',
    'batchKey.results': '检测结果',

    // AddProviderDialog
    'addProvider.title': '添加供应商',
    'addProvider.id': 'ID（唯一标识）',
    'addProvider.name': '名称',
    'addProvider.models': '模型列表（每行一个）',
    'addProvider.cancel': '取消',
    'addProvider.add': '添加',

    // Settings
    'settings.title': '设置',
    'settings.providerManage': '供应商管理',
    'settings.add': '添加',
    'settings.builtin': '内置',
    'settings.custom': '自定义',
    'settings.confirmDelete': '确定删除该供应商?',
    'settings.confirmReset': '确定重置? 将删除所有自定义供应商，清除全部配置和可见性设置。',
    'settings.resetDefault': '重置为默认',
    'settings.close': '关闭',
    'settings.language': '语言',
  },
  en: {
    'app.export': 'Export',
    'app.batchCheck': 'Batch Check',
    'app.emptyHint': 'Select a provider, enter API Key, click Check',
    'app.emptySubHint': 'Supports OpenAI / Anthropic / Gemini protocols',

    'config.baseURL': 'Base URL',
    'config.apiKey': 'API Key',
    'config.protocol': 'Protocol',
    'config.model': 'Model',
    'config.reset': 'Reset',
    'config.batch': 'Batch',
    'config.check': 'Check',

    'sidebar.providers': 'Providers',
    'sidebar.history': 'History',

    'card.totalLatency': 'Total',
    'card.availableModels': 'Available Models',
    'card.collapse': 'Collapse',

    'item.connectivity': 'Connectivity',
    'item.chat': 'Chat',
    'item.stream': 'Streaming',
    'item.models': 'Model List',
    'item.multi_turn': 'Multi-turn',

    'history.total': '{n} records',
    'history.selectAll': 'Select All',
    'history.deselectAll': 'Deselect All',
    'history.deleteSelected': 'Delete ({n})',
    'history.clearAll': 'Clear All',
    'history.refresh': 'Refresh',
    'history.confirmDelete': 'Delete {n} records?',
    'history.confirmClear': 'Clear all history?',
    'history.empty': 'No history yet',

    'batchKey.title': 'Batch Key Check',
    'batchKey.label': 'API Keys (one per line)',
    'batchKey.count': '{n} keys',
    'batchKey.start': 'Start',
    'batchKey.results': 'Results',

    'addProvider.title': 'Add Provider',
    'addProvider.id': 'ID (unique)',
    'addProvider.name': 'Name',
    'addProvider.models': 'Models (one per line)',
    'addProvider.cancel': 'Cancel',
    'addProvider.add': 'Add',

    'settings.title': 'Settings',
    'settings.providerManage': 'Provider Management',
    'settings.add': 'Add',
    'settings.builtin': 'Built-in',
    'settings.custom': 'Custom',
    'settings.confirmDelete': 'Delete this provider?',
    'settings.confirmReset': 'Reset all? This will remove custom providers, configs and visibility settings.',
    'settings.resetDefault': 'Reset to Default',
    'settings.close': 'Close',
    'settings.language': 'Language',
  },
}

// 翻译函数，支持 {n} 占位符
export function t(key: string, params?: Record<string, string | number>): string {
  let text = messages[locale.value]?.[key] || messages.zh[key] || key
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      text = text.replace(`{${k}}`, String(v))
    }
  }
  return text
}

// 检测项名称（响应式）
export function checkItemName(item: string): string {
  return t(`item.${item}`) || item
}
