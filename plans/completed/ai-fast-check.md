# AI Fast Check - AI API 可用性快速检测工具

> 创建: 2026-02-08
> 状态: 进行中 (第二轮迭代)

## 第二轮变更

- [x] 纯 Go SQLite 存储 (modernc.org/sqlite)
- [ ] 历史检测记录 + 手动清理 (逐条/批量/全部)
- [ ] 支持添加自定义供应商
- [ ] 多协议支持: OpenAI / Anthropic / Gemini，界面可切换

## 技术方案

### SQLite 表结构
- `providers`: 自定义供应商 (内置的保留在代码中)
- `provider_configs`: 每个供应商的 API Key 等配置
- `check_history`: 历史检测记录

### 协议适配
- OpenAI: /chat/completions (大多数厂商)
- Anthropic: /messages (x-api-key header, anthropic-version)
- Gemini: /models/{model}:generateContent (原生 Google API)

### 项目结构更新
```
internal/
  store/        -> SQLite 存储层
  checker/      -> 检测引擎 (多协议)
  protocol/     -> 协议适配器
  provider/     -> 厂商定义
```
