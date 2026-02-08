# i18n 国际化 + GitHub Actions CI

> 创建: 2026-02-08
> 状态: 已完成

## 目标

完成中英文国际化支持，创建三端 CI 构建，初始化 Git 推送到 GitHub

## 步骤

- [x] 创建 i18n.ts 翻译模块
- [x] 更新所有 Vue 组件使用 t() 函数
- [x] 设置页面添加语言切换
- [x] 移除 types/index.ts 中的硬编码 CHECK_ITEM_NAMES
- [x] 前端编译验证通过
- [x] Go 编译和测试通过
- [x] 创建 .github/workflows/build.yml (Win/Mac/Linux)
- [x] 创建 .gitignore
- [x] Git 初始化并推送到 github.com:BlakeLiAFK/PingAI.git

## 完成标准

- [x] 切换语言后所有界面文字即时切换
- [x] GitHub Actions 能构建三端产物
- [x] 代码已推送到远程仓库
