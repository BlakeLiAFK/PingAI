package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Protocol 协议类型
type Protocol string

const (
	ProtocolOpenAI    Protocol = "openai"
	ProtocolAnthropic Protocol = "anthropic"
	ProtocolGemini    Protocol = "gemini"
)

// ChatRequest 统一请求
type ChatRequest struct {
	BaseURL  string
	APIKey   string
	Model    string
	Messages []Message
	Stream   bool
}

// Message 统一消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 统一响应
type ChatResponse struct {
	Content      string
	PromptTokens int
	CompTokens   int
	StatusCode   int
	RawBody      string
	Error        string
}

// StreamCallback 流式回调
type StreamCallback func(chunk string, isFirst bool)

// Adapter 协议适配器接口
type Adapter interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest, cb StreamCallback) (*ChatResponse, error)
	ListModels(ctx context.Context, baseURL, apiKey string) ([]string, error)
	CheckConnectivity(ctx context.Context, baseURL, apiKey string) (int, error)
}

// GetAdapter 根据协议类型获取适配器
func GetAdapter(p Protocol) Adapter {
	switch p {
	case ProtocolAnthropic:
		return &AnthropicAdapter{}
	case ProtocolGemini:
		return &GeminiAdapter{}
	default:
		return &OpenAIAdapter{}
	}
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// --- OpenAI 适配器 ---

type OpenAIAdapter struct{}

func (a *OpenAIAdapter) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	msgs := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}
	body, _ := json.Marshal(map[string]any{
		"model":    req.Model,
		"messages": msgs,
	})

	url := strings.TrimSuffix(req.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return &ChatResponse{StatusCode: resp.StatusCode, RawBody: truncate(string(respBody), 300),
			Error: fmt.Sprintf("HTTP %d", resp.StatusCode)}, nil
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: "JSON parse error", RawBody: truncate(string(respBody), 300)}, nil
	}
	if result.Error != nil {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: result.Error.Message}, nil
	}
	if len(result.Choices) == 0 {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: "empty choices"}, nil
	}

	cr := &ChatResponse{
		Content:    result.Choices[0].Message.Content,
		StatusCode: 200,
	}
	if result.Usage != nil {
		cr.PromptTokens = result.Usage.PromptTokens
		cr.CompTokens = result.Usage.CompletionTokens
	}
	return cr, nil
}

func (a *OpenAIAdapter) ChatStream(ctx context.Context, req ChatRequest, cb StreamCallback) (*ChatResponse, error) {
	msgs := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}
	body, _ := json.Marshal(map[string]any{
		"model":    req.Model,
		"messages": msgs,
		"stream":   true,
	})

	url := strings.TrimSuffix(req.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return &ChatResponse{StatusCode: resp.StatusCode, Error: fmt.Sprintf("HTTP %d", resp.StatusCode),
			RawBody: truncate(string(respBody), 300)}, nil
	}

	return readSSE(resp.Body, cb)
}

func (a *OpenAIAdapter) ListModels(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	url := strings.TrimSuffix(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}
	return models, nil
}

func (a *OpenAIAdapter) CheckConnectivity(ctx context.Context, baseURL, apiKey string) (int, error) {
	url := strings.TrimSuffix(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// --- Anthropic 适配器 ---

type AnthropicAdapter struct{}

func (a *AnthropicAdapter) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	msgs := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}
	body, _ := json.Marshal(map[string]any{
		"model":      req.Model,
		"messages":   msgs,
		"max_tokens": 256,
	})

	url := strings.TrimSuffix(req.BaseURL, "/") + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: fmt.Sprintf("HTTP %d", resp.StatusCode),
			RawBody: truncate(string(respBody), 300)}, nil
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: "JSON parse error", RawBody: truncate(string(respBody), 300)}, nil
	}
	if result.Error != nil {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: result.Error.Message}, nil
	}
	if len(result.Content) == 0 {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: "empty content"}, nil
	}

	cr := &ChatResponse{Content: result.Content[0].Text, StatusCode: 200}
	if result.Usage != nil {
		cr.PromptTokens = result.Usage.InputTokens
		cr.CompTokens = result.Usage.OutputTokens
	}
	return cr, nil
}

func (a *AnthropicAdapter) ChatStream(ctx context.Context, req ChatRequest, cb StreamCallback) (*ChatResponse, error) {
	msgs := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}
	body, _ := json.Marshal(map[string]any{
		"model":      req.Model,
		"messages":   msgs,
		"max_tokens": 256,
		"stream":     true,
	})

	url := strings.TrimSuffix(req.BaseURL, "/") + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return &ChatResponse{StatusCode: resp.StatusCode, Error: fmt.Sprintf("HTTP %d", resp.StatusCode),
			RawBody: truncate(string(respBody), 300)}, nil
	}

	return readAnthropicSSE(resp.Body, cb)
}

func (a *AnthropicAdapter) ListModels(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	url := strings.TrimSuffix(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}
	return models, nil
}

func (a *AnthropicAdapter) CheckConnectivity(ctx context.Context, baseURL, apiKey string) (int, error) {
	url := strings.TrimSuffix(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// --- Gemini 适配器 ---

type GeminiAdapter struct{}

func (a *GeminiAdapter) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	contents := make([]map[string]any, len(req.Messages))
	for i, m := range req.Messages {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents[i] = map[string]any{
			"role":  role,
			"parts": []map[string]string{{"text": m.Content}},
		}
	}
	body, _ := json.Marshal(map[string]any{
		"contents": contents,
	})

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		strings.TrimSuffix(req.BaseURL, "/"), req.Model, req.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: fmt.Sprintf("HTTP %d", resp.StatusCode),
			RawBody: truncate(string(respBody), 300)}, nil
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata *struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: "JSON parse error", RawBody: truncate(string(respBody), 300)}, nil
	}
	if result.Error != nil {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: result.Error.Message}, nil
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return &ChatResponse{StatusCode: resp.StatusCode, Error: "empty response"}, nil
	}

	cr := &ChatResponse{Content: result.Candidates[0].Content.Parts[0].Text, StatusCode: 200}
	if result.UsageMetadata != nil {
		cr.PromptTokens = result.UsageMetadata.PromptTokenCount
		cr.CompTokens = result.UsageMetadata.CandidatesTokenCount
	}
	return cr, nil
}

func (a *GeminiAdapter) ChatStream(ctx context.Context, req ChatRequest, cb StreamCallback) (*ChatResponse, error) {
	contents := make([]map[string]any, len(req.Messages))
	for i, m := range req.Messages {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents[i] = map[string]any{
			"role":  role,
			"parts": []map[string]string{{"text": m.Content}},
		}
	}
	body, _ := json.Marshal(map[string]any{
		"contents": contents,
	})

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s",
		strings.TrimSuffix(req.BaseURL, "/"), req.Model, req.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return &ChatResponse{StatusCode: resp.StatusCode, Error: fmt.Sprintf("HTTP %d", resp.StatusCode),
			RawBody: truncate(string(respBody), 300)}, nil
	}

	return readGeminiSSE(resp.Body, cb)
}

func (a *GeminiAdapter) ListModels(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	url := fmt.Sprintf("%s/models?key=%s", strings.TrimSuffix(baseURL, "/"), apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		// "models/gemini-1.5-pro" -> "gemini-1.5-pro"
		name := m.Name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		models[i] = name
	}
	return models, nil
}

func (a *GeminiAdapter) CheckConnectivity(ctx context.Context, baseURL, apiKey string) (int, error) {
	url := fmt.Sprintf("%s/models?key=%s", strings.TrimSuffix(baseURL, "/"), apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// --- SSE 读取工具 ---

func readSSE(reader io.Reader, cb StreamCallback) (*ChatResponse, error) {
	buf := make([]byte, 4096)
	var fullContent strings.Builder
	chunkCount := 0
	isFirst := true

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					continue
				}
				var chunk struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}
				if json.Unmarshal([]byte(data), &chunk) == nil && len(chunk.Choices) > 0 {
					text := chunk.Choices[0].Delta.Content
					if text != "" {
						fullContent.WriteString(text)
						chunkCount++
						if cb != nil {
							cb(text, isFirst)
							isFirst = false
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}

	return &ChatResponse{
		Content:    fullContent.String(),
		StatusCode: 200,
	}, nil
}

func readAnthropicSSE(reader io.Reader, cb StreamCallback) (*ChatResponse, error) {
	buf := make([]byte, 4096)
	var fullContent strings.Builder
	isFirst := true

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				data := strings.TrimPrefix(line, "data: ")
				var event struct {
					Type  string `json:"type"`
					Delta *struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}
				if json.Unmarshal([]byte(data), &event) == nil && event.Delta != nil && event.Delta.Text != "" {
					fullContent.WriteString(event.Delta.Text)
					if cb != nil {
						cb(event.Delta.Text, isFirst)
						isFirst = false
					}
				}
			}
		}
		if err != nil {
			break
		}
	}

	return &ChatResponse{Content: fullContent.String(), StatusCode: 200}, nil
}

func readGeminiSSE(reader io.Reader, cb StreamCallback) (*ChatResponse, error) {
	buf := make([]byte, 4096)
	var fullContent strings.Builder
	isFirst := true

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				data := strings.TrimPrefix(line, "data: ")
				var chunk struct {
					Candidates []struct {
						Content struct {
							Parts []struct {
								Text string `json:"text"`
							} `json:"parts"`
						} `json:"content"`
					} `json:"candidates"`
				}
				if json.Unmarshal([]byte(data), &chunk) == nil && len(chunk.Candidates) > 0 {
					for _, part := range chunk.Candidates[0].Content.Parts {
						if part.Text != "" {
							fullContent.WriteString(part.Text)
							if cb != nil {
								cb(part.Text, isFirst)
								isFirst = false
							}
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}

	return &ChatResponse{Content: fullContent.String(), StatusCode: 200}, nil
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
