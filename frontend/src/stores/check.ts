import { computed, reactive, ref } from 'vue'
import type { ProviderInfo, CheckConfig, FullCheckResult, ProtocolType, HistoryItem } from '../types'

// 全局状态
export const providers = ref<ProviderInfo[]>([])
export const selectedProviderID = ref<string>('')
export const checkConfigs = reactive<Map<string, CheckConfig>>(new Map())
export const checkResults = reactive<Map<string, FullCheckResult>>(new Map())
export const isRunning = ref(false)
export const allResults = ref<FullCheckResult[]>([])

// 视图切换: 'check' | 'history'
export const activeView = ref<'check' | 'history'>('check')

// 历史记录
export const historyItems = ref<HistoryItem[]>([])
export const historyTotal = ref(0)

const wails = () => (window as any).go.main.App

// 防抖保存定时器
const saveTimers = new Map<string, ReturnType<typeof setTimeout>>()

// 自动保存配置（防抖 500ms）
export function autoSaveConfig(providerID: string) {
  if (saveTimers.has(providerID)) {
    clearTimeout(saveTimers.get(providerID)!)
  }
  saveTimers.set(providerID, setTimeout(async () => {
    saveTimers.delete(providerID)
    const cfg = checkConfigs.get(providerID)
    if (!cfg) return
    try {
      await wails().SaveProviderConfig(cfg.providerID, cfg.apiKey, cfg.baseURL, cfg.model, cfg.protocol)
    } catch (e) {
      console.error('Auto save config failed:', e)
    }
  }, 500))
}

// 重置内置供应商配置为默认值
export async function resetProviderConfig(providerID: string) {
  try {
    const defaults = await wails().GetProviderDefaults(providerID)
    if (!defaults) return
    const cfg = checkConfigs.get(providerID)
    if (!cfg) return
    cfg.baseURL = defaults.baseURL
    cfg.protocol = defaults.protocol as ProtocolType
    cfg.model = defaults.models?.length > 0 ? defaults.models[0] : ''
    cfg.apiKey = ''
    await wails().SaveProviderConfig(cfg.providerID, cfg.apiKey, cfg.baseURL, cfg.model, cfg.protocol)
  } catch (e) {
    console.error('Reset config failed:', e)
  }
}

// 初始化
export async function initProviders() {
  try {
    const list: ProviderInfo[] = await wails().GetProviders()
    providers.value = list

    // 加载已保存的配置
    const configs = await wails().GetAllConfigs()
    const savedMap = new Map<string, any>()
    if (configs) {
      for (const c of configs) {
        savedMap.set(c.providerID, c)
      }
    }

    // 初始化每个供应商的配置
    for (const p of list) {
      const saved = savedMap.get(p.id)
      checkConfigs.set(p.id, {
        providerID: p.id,
        providerName: p.name,
        baseURL: saved?.baseURL || p.baseURL,
        apiKey: saved?.apiKey || '',
        model: saved?.model || (p.models?.length > 0 ? p.models[0] : ''),
        protocol: (saved?.protocol || p.protocol) as ProtocolType,
      })
    }

    // 加载可见性
    await loadHiddenProviders()

    if (list.length > 0) {
      // 默认选中第一个可见供应商
      const firstVisible = list.find(p => !hiddenProviderIDs.value.has(p.id))
      selectedProviderID.value = firstVisible?.id || list[0].id
    }
  } catch (e) {
    console.error('Init failed:', e)
  }
}

// 执行单个检测
export async function runSingleCheck(providerID: string) {
  const cfg = checkConfigs.get(providerID)
  if (!cfg || !cfg.model) return

  isRunning.value = true
  try {
    await wails().SaveProviderConfig(cfg.providerID, cfg.apiKey, cfg.baseURL, cfg.model, cfg.protocol)
    const result: FullCheckResult = await wails().RunCheck(
      cfg.baseURL, cfg.apiKey, cfg.model, cfg.providerID, cfg.providerName, cfg.protocol
    )
    checkResults.set(providerID, result)
    updateAllResults()
  } catch (e) {
    console.error('Check failed:', e)
  } finally {
    isRunning.value = false
  }
}

// 批量检测
export async function runAllChecks() {
  const items: any[] = []
  checkConfigs.forEach((cfg) => {
    if (cfg.model && cfg.baseURL) {
      items.push({
        baseURL: cfg.baseURL,
        apiKey: cfg.apiKey,
        model: cfg.model,
        providerID: cfg.providerID,
        providerName: cfg.providerName,
        protocol: cfg.protocol,
      })
    }
  })
  if (items.length === 0) return

  isRunning.value = true
  try {
    for (const item of items) {
      await wails().SaveProviderConfig(item.providerID, item.apiKey, item.baseURL, item.model, item.protocol)
    }
    const results: FullCheckResult[] = await wails().RunBatchCheck(items)
    for (const r of results) {
      checkResults.set(r.providerID, r)
    }
    updateAllResults()
  } catch (e) {
    console.error('Batch check failed:', e)
  } finally {
    isRunning.value = false
  }
}

// 导出报告
export async function exportReport() {
  const results = Array.from(checkResults.values())
  if (results.length === 0) return
  try {
    await wails().ExportReport(results)
  } catch (e) {
    console.error('Export failed:', e)
  }
}

// --- 供应商管理 ---

export async function addProvider(data: { id: string; name: string; baseURL: string; protocol: ProtocolType; models: string[] }) {
  await wails().AddProvider(data)
  await initProviders()
}

export async function deleteProvider(id: string) {
  await wails().DeleteProvider(id)
  providers.value = providers.value.filter(p => p.id !== id)
  checkConfigs.delete(id)
  checkResults.delete(id)
  updateAllResults()
  if (selectedProviderID.value === id && providers.value.length > 0) {
    selectedProviderID.value = providers.value[0].id
  }
}

// --- 可见性管理 ---

export const hiddenProviderIDs = ref<Set<string>>(new Set())

// 可见的供应商（侧边栏用）
export const visibleProviders = computed(() =>
  providers.value.filter(p => !hiddenProviderIDs.value.has(p.id))
)

export async function loadHiddenProviders() {
  try {
    const ids: string[] = await wails().GetHiddenProviderIDs()
    hiddenProviderIDs.value = new Set(ids || [])
  } catch (e) {
    console.error('Load hidden providers failed:', e)
  }
}

export async function setProviderVisibility(providerID: string, visible: boolean) {
  try {
    await wails().SetProviderVisibility(providerID, visible)
    if (visible) {
      hiddenProviderIDs.value.delete(providerID)
    } else {
      hiddenProviderIDs.value.add(providerID)
    }
    // 触发响应式更新
    hiddenProviderIDs.value = new Set(hiddenProviderIDs.value)
  } catch (e) {
    console.error('Set visibility failed:', e)
  }
}

export async function resetAllProviders() {
  try {
    await wails().ResetAllProviders()
    hiddenProviderIDs.value = new Set()
    checkConfigs.clear()
    checkResults.clear()
    allResults.value = []
    await initProviders()
  } catch (e) {
    console.error('Reset all failed:', e)
  }
}

// --- 历史记录 ---

export async function loadHistory(limit = 50, offset = 0) {
  try {
    const result = await wails().GetHistory(limit, offset)
    historyItems.value = result.items || []
    historyTotal.value = result.total
  } catch (e) {
    console.error('Load history failed:', e)
  }
}

export async function deleteHistoryItem(id: number) {
  await wails().DeleteHistory(id)
  await loadHistory()
}

export async function deleteHistoryBatch(ids: number[]) {
  await wails().DeleteHistoryBatch(ids)
  await loadHistory()
}

export async function clearAllHistory() {
  await wails().DeleteAllHistory()
  historyItems.value = []
  historyTotal.value = 0
}

// --- 批量 Key 检测 ---

export const isBatchRunning = ref(false)
export const batchKeyResults = ref<FullCheckResult[]>([])

export async function runBatchKeyCheck(apiKeys: string[]): Promise<FullCheckResult[]> {
  const cfg = checkConfigs.get(selectedProviderID.value)
  if (!cfg || !cfg.model || !cfg.baseURL || apiKeys.length === 0) return []

  isBatchRunning.value = true
  batchKeyResults.value = []
  try {
    const results: FullCheckResult[] = await wails().RunBatchKeyCheck(
      cfg.baseURL, cfg.model, cfg.providerID, cfg.providerName, cfg.protocol, apiKeys
    )
    batchKeyResults.value = results
    return results
  } catch (e) {
    console.error('Batch key check failed:', e)
    return []
  } finally {
    isBatchRunning.value = false
  }
}

function updateAllResults() {
  allResults.value = Array.from(checkResults.values())
}
