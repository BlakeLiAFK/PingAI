# 单元测试编写

> 创建: 2026-02-08
> 状态: 进行中

## 目标

为 store 和 provider/presets 模块编写单元测试

## 步骤

- [x] 修改 store.go，新增 InitWithPath 函数支持自定义数据库路径
- [x] 编写 store_test.go，覆盖 ProviderConfig / CustomProvider / History CRUD
- [x] 编写 presets_test.go，验证预设数据完整性
- [x] 运行 go test ./... -v 通过
- [x] 添加 Antigravity Tools 和 Ollama 到预设列表

## 完成标准

- [x] store 测试全部通过 (5 tests)
- [x] presets 测试全部通过 (3 tests)
- [x] 测试文件不超过 200 行
- [x] 注释使用中文
