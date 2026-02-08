package store

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// DB 全局数据库实例
var DB *sqlx.DB

// Init 初始化数据库
func Init() error {
	dbPath := getDBPath()
	var err error
	DB, err = sqlx.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)")
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(1)
	return migrate()
}

// InitWithPath 使用指定路径初始化数据库，供测试使用
func InitWithPath(dbPath string) error {
	dir := filepath.Dir(dbPath)
	os.MkdirAll(dir, 0755)
	var err error
	DB, err = sqlx.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)")
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(1)
	return migrate()
}

// Close 关闭数据库
func Close() {
	if DB != nil {
		DB.Close()
	}
}

func getDBPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".pingai")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "data.db")
}

func migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS providers (
		id          TEXT PRIMARY KEY,
		name        TEXT NOT NULL,
		base_url    TEXT NOT NULL,
		protocol    TEXT NOT NULL DEFAULT 'openai',
		models      TEXT NOT NULL DEFAULT '[]',
		is_builtin  INTEGER NOT NULL DEFAULT 0,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS provider_configs (
		provider_id TEXT PRIMARY KEY,
		api_key     TEXT NOT NULL DEFAULT '',
		base_url    TEXT NOT NULL DEFAULT '',
		model       TEXT NOT NULL DEFAULT '',
		protocol    TEXT NOT NULL DEFAULT 'openai',
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS check_history (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		provider_id   TEXT NOT NULL,
		provider_name TEXT NOT NULL,
		base_url      TEXT NOT NULL,
		model         TEXT NOT NULL,
		protocol      TEXT NOT NULL DEFAULT 'openai',
		results_json  TEXT NOT NULL,
		model_list    TEXT NOT NULL DEFAULT '[]',
		total_latency INTEGER NOT NULL DEFAULT 0,
		status        TEXT NOT NULL DEFAULT 'mixed',
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_history_provider ON check_history(provider_id);
	CREATE INDEX IF NOT EXISTS idx_history_created ON check_history(created_at);

	CREATE TABLE IF NOT EXISTS provider_visibility (
		provider_id TEXT PRIMARY KEY,
		visible     INTEGER NOT NULL DEFAULT 1
	);
	`
	_, err := DB.Exec(schema)
	return err
}

// ProviderRow 供应商数据行
type ProviderRow struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	BaseURL   string `db:"base_url" json:"baseURL"`
	Protocol  string `db:"protocol" json:"protocol"`
	Models    string `db:"models" json:"models"`
	IsBuiltin int    `db:"is_builtin" json:"isBuiltin"`
	CreatedAt string `db:"created_at" json:"createdAt"`
}

// ProviderConfigRow 供应商配置行
type ProviderConfigRow struct {
	ProviderID string `db:"provider_id" json:"providerID"`
	APIKey     string `db:"api_key" json:"apiKey"`
	BaseURL    string `db:"base_url" json:"baseURL"`
	Model      string `db:"model" json:"model"`
	Protocol   string `db:"protocol" json:"protocol"`
	UpdatedAt  string `db:"updated_at" json:"updatedAt"`
}

// HistoryRow 历史记录行
type HistoryRow struct {
	ID           int64  `db:"id" json:"id"`
	ProviderID   string `db:"provider_id" json:"providerID"`
	ProviderName string `db:"provider_name" json:"providerName"`
	BaseURL      string `db:"base_url" json:"baseURL"`
	Model        string `db:"model" json:"model"`
	Protocol     string `db:"protocol" json:"protocol"`
	ResultsJSON  string `db:"results_json" json:"resultsJSON"`
	ModelList    string `db:"model_list" json:"modelList"`
	TotalLatency int64  `db:"total_latency" json:"totalLatency"`
	Status       string `db:"status" json:"status"`
	CreatedAt    string `db:"created_at" json:"createdAt"`
}

// --- 供应商配置 CRUD ---

// SaveProviderConfig 保存供应商配置
func SaveProviderConfig(cfg ProviderConfigRow) error {
	_, err := DB.Exec(`
		INSERT INTO provider_configs (provider_id, api_key, base_url, model, protocol, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(provider_id) DO UPDATE SET
			api_key = excluded.api_key,
			base_url = excluded.base_url,
			model = excluded.model,
			protocol = excluded.protocol,
			updated_at = CURRENT_TIMESTAMP
	`, cfg.ProviderID, cfg.APIKey, cfg.BaseURL, cfg.Model, cfg.Protocol)
	return err
}

// GetProviderConfig 获取供应商配置
func GetProviderConfig(providerID string) (*ProviderConfigRow, error) {
	var row ProviderConfigRow
	err := DB.Get(&row, "SELECT * FROM provider_configs WHERE provider_id = ?", providerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &row, err
}

// GetAllProviderConfigs 获取所有供应商配置
func GetAllProviderConfigs() ([]ProviderConfigRow, error) {
	var rows []ProviderConfigRow
	err := DB.Select(&rows, "SELECT * FROM provider_configs ORDER BY updated_at DESC")
	return rows, err
}

// --- 自定义供应商 CRUD ---

// AddCustomProvider 添加自定义供应商
func AddCustomProvider(p ProviderRow) error {
	_, err := DB.Exec(`
		INSERT INTO providers (id, name, base_url, protocol, models, is_builtin)
		VALUES (?, ?, ?, ?, ?, 0)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			base_url = excluded.base_url,
			protocol = excluded.protocol,
			models = excluded.models
	`, p.ID, p.Name, p.BaseURL, p.Protocol, p.Models)
	return err
}

// GetCustomProviders 获取所有自定义供应商
func GetCustomProviders() ([]ProviderRow, error) {
	var rows []ProviderRow
	err := DB.Select(&rows, "SELECT * FROM providers WHERE is_builtin = 0 ORDER BY created_at DESC")
	return rows, err
}

// DeleteCustomProvider 删除自定义供应商
func DeleteCustomProvider(id string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM providers WHERE id = ? AND is_builtin = 0", id)
	tx.Exec("DELETE FROM provider_configs WHERE provider_id = ?", id)
	return tx.Commit()
}

// --- 历史记录 CRUD ---

// SaveHistory 保存检测历史
func SaveHistory(h HistoryRow) (int64, error) {
	result, err := DB.Exec(`
		INSERT INTO check_history (provider_id, provider_name, base_url, model, protocol, results_json, model_list, total_latency, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, h.ProviderID, h.ProviderName, h.BaseURL, h.Model, h.Protocol, h.ResultsJSON, h.ModelList, h.TotalLatency, h.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetHistory 获取历史记录列表
func GetHistory(limit, offset int) ([]HistoryRow, error) {
	var rows []HistoryRow
	err := DB.Select(&rows, "SELECT * FROM check_history ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	return rows, err
}

// GetHistoryCount 获取历史总数
func GetHistoryCount() (int, error) {
	var count int
	err := DB.Get(&count, "SELECT COUNT(*) FROM check_history")
	return count, err
}

// DeleteHistoryByID 删除单条历史
func DeleteHistoryByID(id int64) error {
	_, err := DB.Exec("DELETE FROM check_history WHERE id = ?", id)
	return err
}

// DeleteHistoryByIDs 批量删除历史
func DeleteHistoryByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query, args, err := sqlx.In("DELETE FROM check_history WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	_, err = DB.Exec(query, args...)
	return err
}

// DeleteAllHistory 清空所有历史
func DeleteAllHistory() error {
	_, err := DB.Exec("DELETE FROM check_history")
	return err
}

// --- 供应商可见性 ---

// SetProviderVisibility 设置供应商是否在侧边栏显示
func SetProviderVisibility(providerID string, visible bool) error {
	v := 0
	if visible {
		v = 1
	}
	_, err := DB.Exec(`
		INSERT INTO provider_visibility (provider_id, visible) VALUES (?, ?)
		ON CONFLICT(provider_id) DO UPDATE SET visible = excluded.visible
	`, providerID, v)
	return err
}

// GetHiddenProviderIDs 获取所有被隐藏的供应商 ID
func GetHiddenProviderIDs() ([]string, error) {
	var ids []string
	err := DB.Select(&ids, "SELECT provider_id FROM provider_visibility WHERE visible = 0")
	return ids, err
}

// ResetProviderVisibility 重置所有可见性设置
func ResetProviderVisibility() error {
	_, err := DB.Exec("DELETE FROM provider_visibility")
	return err
}

// ResetAll 重置全部数据：删除自定义供应商、配置、可见性
func ResetAll() error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec("DELETE FROM providers WHERE is_builtin = 0")
	tx.Exec("DELETE FROM provider_configs")
	tx.Exec("DELETE FROM provider_visibility")
	return tx.Commit()
}
