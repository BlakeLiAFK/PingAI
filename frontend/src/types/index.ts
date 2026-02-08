// 协议类型
export type ProtocolType = 'openai' | 'anthropic' | 'gemini'

// 供应商
export interface ProviderInfo {
  id: string
  name: string
  baseURL: string
  protocol: ProtocolType
  models: string[]
  isBuiltin: boolean
}

// 检测状态
export type CheckStatus = 'pending' | 'running' | 'success' | 'failed' | 'warning'
export type CheckItem = 'connectivity' | 'chat' | 'stream' | 'models' | 'multi_turn'

// 检测结果
export interface CheckResult {
  item: CheckItem
  status: CheckStatus
  latency: number
  ttft: number
  message: string
  detail: string
  tokenIn: number
  tokenOut: number
}

export interface FullCheckResult {
  providerID: string
  providerName: string
  baseURL: string
  model: string
  protocol: string
  results: CheckResult[]
  modelList: string[]
  startTime: string
  endTime: string
  totalLatency: number
}

// 配置
export interface CheckConfig {
  providerID: string
  providerName: string
  baseURL: string
  apiKey: string
  model: string
  protocol: ProtocolType
}

// 历史记录
export interface HistoryItem {
  id: number
  providerID: string
  providerName: string
  baseURL: string
  model: string
  protocol: string
  results: CheckResult[]
  modelList: string[]
  totalLatency: number
  status: string
  createdAt: string
}

export interface HistoryListResult {
  items: HistoryItem[]
  total: number
}

export const PROTOCOL_NAMES: Record<ProtocolType, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  gemini: 'Gemini',
}
