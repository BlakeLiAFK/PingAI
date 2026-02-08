package provider

// Provider 厂商定义
type Provider struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	BaseURL  string   `json:"baseURL"`
	Models   []string `json:"models"`
	Protocol string   `json:"protocol"` // "openai" 或 "anthropic"
}
