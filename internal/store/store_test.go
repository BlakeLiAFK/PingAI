package store

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDB 创建临时数据库，返回清理函数
func setupTestDB(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	if err := InitWithPath(dbPath); err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	return func() {
		Close()
		os.RemoveAll(dir)
	}
}

// --- 供应商配置测试 ---

func TestProviderConfigCRUD(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	cfg := ProviderConfigRow{
		ProviderID: "test-provider",
		APIKey:     "sk-test-key",
		BaseURL:    "https://api.test.com/v1",
		Model:      "test-model",
		Protocol:   "openai",
	}

	// 保存配置
	if err := SaveProviderConfig(cfg); err != nil {
		t.Fatalf("SaveProviderConfig 失败: %v", err)
	}

	// 读取配置
	got, err := GetProviderConfig("test-provider")
	if err != nil {
		t.Fatalf("GetProviderConfig 失败: %v", err)
	}
	if got == nil {
		t.Fatal("GetProviderConfig 返回 nil")
	}
	if got.APIKey != "sk-test-key" {
		t.Errorf("APIKey = %q, 期望 %q", got.APIKey, "sk-test-key")
	}
	if got.BaseURL != "https://api.test.com/v1" {
		t.Errorf("BaseURL = %q, 期望 %q", got.BaseURL, "https://api.test.com/v1")
	}
	if got.Protocol != "openai" {
		t.Errorf("Protocol = %q, 期望 %q", got.Protocol, "openai")
	}

	// 更新配置（upsert）
	cfg.APIKey = "sk-updated"
	cfg.Protocol = "anthropic"
	if err := SaveProviderConfig(cfg); err != nil {
		t.Fatalf("更新 SaveProviderConfig 失败: %v", err)
	}
	got, _ = GetProviderConfig("test-provider")
	if got.APIKey != "sk-updated" {
		t.Errorf("更新后 APIKey = %q, 期望 %q", got.APIKey, "sk-updated")
	}
	if got.Protocol != "anthropic" {
		t.Errorf("更新后 Protocol = %q, 期望 %q", got.Protocol, "anthropic")
	}

	// 获取不存在的配置
	missing, err := GetProviderConfig("nonexistent")
	if err != nil {
		t.Fatalf("查询不存在配置报错: %v", err)
	}
	if missing != nil {
		t.Error("不存在的配置应返回 nil")
	}

	// 获取全部配置
	all, err := GetAllProviderConfigs()
	if err != nil {
		t.Fatalf("GetAllProviderConfigs 失败: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("配置总数 = %d, 期望 1", len(all))
	}
}

// --- 自定义供应商测试 ---

func TestCustomProviderCRUD(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	p := ProviderRow{
		ID:       "custom-1",
		Name:     "Custom Provider",
		BaseURL:  "https://custom.api.com/v1",
		Protocol: "openai",
		Models:   `["model-a","model-b"]`,
	}

	// 添加
	if err := AddCustomProvider(p); err != nil {
		t.Fatalf("AddCustomProvider 失败: %v", err)
	}

	// 查询
	rows, err := GetCustomProviders()
	if err != nil {
		t.Fatalf("GetCustomProviders 失败: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("自定义供应商数量 = %d, 期望 1", len(rows))
	}
	if rows[0].Name != "Custom Provider" {
		t.Errorf("Name = %q, 期望 %q", rows[0].Name, "Custom Provider")
	}
	if rows[0].IsBuiltin != 0 {
		t.Error("自定义供应商 IsBuiltin 应为 0")
	}

	// 更新（upsert 同 id）
	p.Name = "Updated Provider"
	if err := AddCustomProvider(p); err != nil {
		t.Fatalf("Upsert AddCustomProvider 失败: %v", err)
	}
	rows, _ = GetCustomProviders()
	if len(rows) != 1 {
		t.Fatalf("Upsert 后数量 = %d, 期望 1", len(rows))
	}
	if rows[0].Name != "Updated Provider" {
		t.Errorf("Upsert 后 Name = %q, 期望 %q", rows[0].Name, "Updated Provider")
	}

	// 删除
	if err := DeleteCustomProvider("custom-1"); err != nil {
		t.Fatalf("DeleteCustomProvider 失败: %v", err)
	}
	rows, _ = GetCustomProviders()
	if len(rows) != 0 {
		t.Errorf("删除后数量 = %d, 期望 0", len(rows))
	}
}

// --- 历史记录测试 ---

func TestHistoryCRUD(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	h := HistoryRow{
		ProviderID:   "openai",
		ProviderName: "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		Model:        "gpt-4o",
		Protocol:     "openai",
		ResultsJSON:  `[{"item":"chat","status":"success"}]`,
		ModelList:    `["gpt-4o"]`,
		TotalLatency: 500,
		Status:       "success",
	}

	// 保存
	id1, err := SaveHistory(h)
	if err != nil {
		t.Fatalf("SaveHistory 失败: %v", err)
	}
	if id1 <= 0 {
		t.Errorf("SaveHistory 返回 id = %d, 期望 > 0", id1)
	}

	// 再保存一条
	h.Model = "gpt-4"
	h.TotalLatency = 800
	id2, err := SaveHistory(h)
	if err != nil {
		t.Fatalf("SaveHistory(2) 失败: %v", err)
	}

	// 获取列表
	rows, err := GetHistory(10, 0)
	if err != nil {
		t.Fatalf("GetHistory 失败: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("历史记录数 = %d, 期望 2", len(rows))
	}

	// 获取总数
	count, err := GetHistoryCount()
	if err != nil {
		t.Fatalf("GetHistoryCount 失败: %v", err)
	}
	if count != 2 {
		t.Errorf("总数 = %d, 期望 2", count)
	}

	// 分页
	page, _ := GetHistory(1, 0)
	if len(page) != 1 {
		t.Errorf("分页 limit=1 返回 %d 条, 期望 1", len(page))
	}

	// 单条删除
	if err := DeleteHistoryByID(id1); err != nil {
		t.Fatalf("DeleteHistoryByID 失败: %v", err)
	}
	count, _ = GetHistoryCount()
	if count != 1 {
		t.Errorf("单删后总数 = %d, 期望 1", count)
	}

	// 再添加两条用于批量删除测试
	h.Model = "gpt-3.5-turbo"
	id3, _ := SaveHistory(h)
	h.Model = "o1"
	id4, _ := SaveHistory(h)

	// 批量删除
	if err := DeleteHistoryByIDs([]int64{id2, id3}); err != nil {
		t.Fatalf("DeleteHistoryByIDs 失败: %v", err)
	}
	count, _ = GetHistoryCount()
	if count != 1 {
		t.Errorf("批量删除后总数 = %d, 期望 1", count)
	}

	// 清空全部
	_ = id4 // 确保 id4 已使用
	if err := DeleteAllHistory(); err != nil {
		t.Fatalf("DeleteAllHistory 失败: %v", err)
	}
	count, _ = GetHistoryCount()
	if count != 0 {
		t.Errorf("清空后总数 = %d, 期望 0", count)
	}
}

// 空 ids 批量删除不应报错
func TestDeleteHistoryByIDsEmpty(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	if err := DeleteHistoryByIDs([]int64{}); err != nil {
		t.Errorf("空 ids 批量删除报错: %v", err)
	}
}

// 删除自定义供应商时同时清理配置
func TestDeleteCustomProviderCleansConfig(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// 添加供应商和配置
	AddCustomProvider(ProviderRow{
		ID: "cp-1", Name: "CP", BaseURL: "https://cp.com", Protocol: "openai", Models: "[]",
	})
	SaveProviderConfig(ProviderConfigRow{
		ProviderID: "cp-1", APIKey: "key", BaseURL: "https://cp.com", Model: "m", Protocol: "openai",
	})

	// 删除供应商
	DeleteCustomProvider("cp-1")

	// 配置也应被清除
	cfg, _ := GetProviderConfig("cp-1")
	if cfg != nil {
		t.Error("删除供应商后配置仍然存在")
	}
}
