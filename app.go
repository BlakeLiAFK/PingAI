package main

import (
	"pingai/internal/checker"
	"pingai/internal/provider"
	"pingai/internal/store"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 应用核心
type App struct {
	ctx     context.Context
	checker *checker.Checker
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{
		checker: checker.NewChecker(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := store.Init(); err != nil {
		runtime.LogErrorf(ctx, "数据库初始化失败: %v", err)
	}
}

func (a *App) shutdown(_ context.Context) {
	store.Close()
}

// --- 供应商 ---

// ProviderInfo 返回给前端的供应商信息
type ProviderInfo struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	BaseURL   string   `json:"baseURL"`
	Protocol  string   `json:"protocol"`
	Models    []string `json:"models"`
	IsBuiltin bool     `json:"isBuiltin"`
}

// GetProviders 获取全部供应商 (内置 + 自定义)
func (a *App) GetProviders() []ProviderInfo {
	var result []ProviderInfo

	// 内置
	for _, p := range provider.GetPresets() {
		result = append(result, ProviderInfo{
			ID: p.ID, Name: p.Name, BaseURL: p.BaseURL,
			Protocol: p.Protocol, Models: p.Models, IsBuiltin: true,
		})
	}

	// 自定义
	customs, _ := store.GetCustomProviders()
	for _, c := range customs {
		var models []string
		json.Unmarshal([]byte(c.Models), &models)
		result = append(result, ProviderInfo{
			ID: c.ID, Name: c.Name, BaseURL: c.BaseURL,
			Protocol: c.Protocol, Models: models, IsBuiltin: false,
		})
	}

	return result
}

// AddProviderReq 添加供应商请求
type AddProviderReq struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	BaseURL  string   `json:"baseURL"`
	Protocol string   `json:"protocol"`
	Models   []string `json:"models"`
}

// AddProvider 添加自定义供应商
func (a *App) AddProvider(req AddProviderReq) error {
	modelsJSON, _ := json.Marshal(req.Models)
	return store.AddCustomProvider(store.ProviderRow{
		ID:       req.ID,
		Name:     req.Name,
		BaseURL:  req.BaseURL,
		Protocol: req.Protocol,
		Models:   string(modelsJSON),
	})
}

// DeleteProvider 删除自定义供应商
func (a *App) DeleteProvider(id string) error {
	return store.DeleteCustomProvider(id)
}

// --- 可见性 ---

// SetProviderVisibility 设置供应商是否显示在侧边栏
func (a *App) SetProviderVisibility(providerID string, visible bool) error {
	return store.SetProviderVisibility(providerID, visible)
}

// GetHiddenProviderIDs 获取被隐藏的供应商 ID 列表
func (a *App) GetHiddenProviderIDs() []string {
	ids, _ := store.GetHiddenProviderIDs()
	if ids == nil {
		return []string{}
	}
	return ids
}

// ResetAllProviders 重置全部供应商设置（删除自定义、清除配置和可见性）
func (a *App) ResetAllProviders() error {
	return store.ResetAll()
}

// --- 配置 ---

// SaveProviderConfig 保存供应商配置 (API Key 等)
func (a *App) SaveProviderConfig(providerID, apiKey, baseURL, model, protocol string) error {
	return store.SaveProviderConfig(store.ProviderConfigRow{
		ProviderID: providerID,
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Model:      model,
		Protocol:   protocol,
	})
}

// GetAllConfigs 获取所有保存的配置
func (a *App) GetAllConfigs() []store.ProviderConfigRow {
	configs, _ := store.GetAllProviderConfigs()
	if configs == nil {
		return []store.ProviderConfigRow{}
	}
	return configs
}

// GetProviderDefaults 获取内置供应商的默认配置
func (a *App) GetProviderDefaults(providerID string) *ProviderInfo {
	for _, p := range provider.GetPresets() {
		if p.ID == providerID {
			return &ProviderInfo{
				ID: p.ID, Name: p.Name, BaseURL: p.BaseURL,
				Protocol: p.Protocol, Models: p.Models, IsBuiltin: true,
			}
		}
	}
	return nil
}

// --- 检测 ---

// RunCheck 执行单个检测
func (a *App) RunCheck(baseURL, apiKey, model, providerID, providerName, protocol string) checker.FullCheckResult {
	result := a.checker.RunFullCheck(baseURL, apiKey, model, providerID, providerName, protocol)
	a.saveHistory(result)
	return result
}

// BatchCheckItem 批量检测项
type BatchCheckItem struct {
	BaseURL      string `json:"baseURL"`
	APIKey       string `json:"apiKey"`
	Model        string `json:"model"`
	ProviderID   string `json:"providerID"`
	ProviderName string `json:"providerName"`
	Protocol     string `json:"protocol"`
}

// RunBatchCheck 批量检测
func (a *App) RunBatchCheck(items []BatchCheckItem) []checker.FullCheckResult {
	type indexed struct {
		idx    int
		result checker.FullCheckResult
	}

	results := make([]checker.FullCheckResult, len(items))
	ch := make(chan indexed, len(items))

	for i, item := range items {
		go func(idx int, it BatchCheckItem) {
			r := a.checker.RunFullCheck(it.BaseURL, it.APIKey, it.Model, it.ProviderID, it.ProviderName, it.Protocol)
			ch <- indexed{idx: idx, result: r}
		}(i, item)
	}

	for range items {
		ir := <-ch
		results[ir.idx] = ir.result
	}

	// 保存历史
	for _, r := range results {
		a.saveHistory(r)
	}
	return results
}

// RunBatchKeyCheck 批量 Key 检测：同一供应商配置，多个 API Key
func (a *App) RunBatchKeyCheck(baseURL, model, providerID, providerName, protocol string, apiKeys []string) []checker.FullCheckResult {
	type indexed struct {
		idx    int
		result checker.FullCheckResult
	}

	results := make([]checker.FullCheckResult, len(apiKeys))
	ch := make(chan indexed, len(apiKeys))

	for i, key := range apiKeys {
		go func(idx int, k string) {
			// 供应商名称附加脱敏 Key 标识
			masked := maskKey(k)
			name := providerName + " (" + masked + ")"
			r := a.checker.RunFullCheck(baseURL, k, model, providerID, name, protocol)
			ch <- indexed{idx: idx, result: r}
		}(i, key)
	}

	for range apiKeys {
		ir := <-ch
		results[ir.idx] = ir.result
	}

	for _, r := range results {
		a.saveHistory(r)
	}
	return results
}

// maskKey 脱敏 API Key，保留前3后4位
func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:3] + "..." + key[len(key)-4:]
}

func (a *App) saveHistory(r checker.FullCheckResult) {
	resultsJSON, _ := json.Marshal(r.Results)
	modelListJSON, _ := json.Marshal(r.ModelList)

	status := "success"
	for _, item := range r.Results {
		if item.Status == "failed" {
			status = "failed"
			break
		}
		if item.Status == "warning" {
			status = "warning"
		}
	}

	store.SaveHistory(store.HistoryRow{
		ProviderID:   r.ProviderID,
		ProviderName: r.ProviderName,
		BaseURL:      r.BaseURL,
		Model:        r.Model,
		Protocol:     r.Protocol,
		ResultsJSON:  string(resultsJSON),
		ModelList:    string(modelListJSON),
		TotalLatency: r.TotalLatency,
		Status:       status,
	})
}

// --- 历史记录 ---

// HistoryItem 前端展示的历史项
type HistoryItem struct {
	ID           int64                  `json:"id"`
	ProviderID   string                 `json:"providerID"`
	ProviderName string                 `json:"providerName"`
	BaseURL      string                 `json:"baseURL"`
	Model        string                 `json:"model"`
	Protocol     string                 `json:"protocol"`
	Results      []checker.CheckResult  `json:"results"`
	ModelList    []string               `json:"modelList"`
	TotalLatency int64                  `json:"totalLatency"`
	Status       string                 `json:"status"`
	CreatedAt    string                 `json:"createdAt"`
}

// HistoryListResult 历史列表返回
type HistoryListResult struct {
	Items []HistoryItem `json:"items"`
	Total int           `json:"total"`
}

// GetHistory 获取历史记录
func (a *App) GetHistory(limit, offset int) HistoryListResult {
	rows, _ := store.GetHistory(limit, offset)
	total, _ := store.GetHistoryCount()

	items := make([]HistoryItem, 0, len(rows))
	for _, row := range rows {
		item := HistoryItem{
			ID:           row.ID,
			ProviderID:   row.ProviderID,
			ProviderName: row.ProviderName,
			BaseURL:      row.BaseURL,
			Model:        row.Model,
			Protocol:     row.Protocol,
			TotalLatency: row.TotalLatency,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt,
		}
		json.Unmarshal([]byte(row.ResultsJSON), &item.Results)
		json.Unmarshal([]byte(row.ModelList), &item.ModelList)
		items = append(items, item)
	}

	return HistoryListResult{Items: items, Total: total}
}

// DeleteHistory 删除单条历史
func (a *App) DeleteHistory(id int64) error {
	return store.DeleteHistoryByID(id)
}

// DeleteHistoryBatch 批量删除
func (a *App) DeleteHistoryBatch(ids []int64) error {
	return store.DeleteHistoryByIDs(ids)
}

// DeleteAllHistory 清空全部历史
func (a *App) DeleteAllHistory() error {
	return store.DeleteAllHistory()
}

// --- 导出 ---

// ExportReport 导出报告到文件
func (a *App) ExportReport(results []checker.FullCheckResult) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Report",
		DefaultFilename: "ai_check_report_" + time.Now().Format("20060102_150405") + ".json",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files", Pattern: "*.json"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}

	report := checker.GenerateReport(results)
	if err := os.WriteFile(path, []byte(report), 0644); err != nil {
		return "", err
	}
	return path, nil
}
