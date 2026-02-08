package checker

import (
	"pingai/internal/protocol"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// CheckItem 单项检测类型
type CheckItem string

const (
	CheckConnectivity CheckItem = "connectivity"
	CheckChat         CheckItem = "chat"
	CheckStream       CheckItem = "stream"
	CheckModels       CheckItem = "models"
	CheckMultiTurn    CheckItem = "multi_turn"
)

// CheckStatus 检测状态
type CheckStatus string

const (
	StatusPending CheckStatus = "pending"
	StatusRunning CheckStatus = "running"
	StatusSuccess CheckStatus = "success"
	StatusFailed  CheckStatus = "failed"
	StatusWarning CheckStatus = "warning"
)

// CheckResult 单项检测结果
type CheckResult struct {
	Item     CheckItem   `json:"item"`
	Status   CheckStatus `json:"status"`
	Latency  int64       `json:"latency"`
	TTFT     int64       `json:"ttft"`
	Message  string      `json:"message"`
	Detail   string      `json:"detail"`
	TokenIn  int         `json:"tokenIn"`
	TokenOut int         `json:"tokenOut"`
}

// FullCheckResult 完整检测结果
type FullCheckResult struct {
	ProviderID   string        `json:"providerID"`
	ProviderName string        `json:"providerName"`
	BaseURL      string        `json:"baseURL"`
	Model        string        `json:"model"`
	Protocol     string        `json:"protocol"`
	Results      []CheckResult `json:"results"`
	ModelList    []string      `json:"modelList"`
	StartTime    string        `json:"startTime"`
	EndTime      string        `json:"endTime"`
	TotalLatency int64         `json:"totalLatency"`
}

// Checker 检测引擎
type Checker struct{}

// NewChecker 创建检测引擎
func NewChecker() *Checker {
	return &Checker{}
}

const timeFmt = "2006-01-02 15:04:05"

// RunFullCheck 执行全量检测
func (c *Checker) RunFullCheck(baseURL, apiKey, model, providerID, providerName, proto string) FullCheckResult {
	startTime := time.Now()
	adapter := protocol.GetAdapter(protocol.Protocol(proto))

	result := FullCheckResult{
		ProviderID:   providerID,
		ProviderName: providerName,
		BaseURL:      baseURL,
		Model:        model,
		Protocol:     proto,
		StartTime:    startTime.Format(timeFmt),
	}

	// 连通性检测
	connResult := c.checkConnectivity(adapter, baseURL, apiKey)
	result.Results = append(result.Results, connResult)
	if connResult.Status == StatusFailed {
		result.EndTime = time.Now().Format(timeFmt)
		result.TotalLatency = time.Since(startTime).Milliseconds()
		return result
	}

	// 并行执行其余检测
	var wg sync.WaitGroup
	var chatResult, streamResult, modelResult, multiTurnResult CheckResult
	var modelList []string

	wg.Add(4)
	go func() { defer wg.Done(); chatResult = c.checkChat(adapter, baseURL, apiKey, model) }()
	go func() { defer wg.Done(); streamResult = c.checkStream(adapter, baseURL, apiKey, model) }()
	go func() { defer wg.Done(); modelResult, modelList = c.checkModels(adapter, baseURL, apiKey) }()
	go func() { defer wg.Done(); multiTurnResult = c.checkMultiTurn(adapter, baseURL, apiKey, model) }()
	wg.Wait()

	result.Results = append(result.Results, chatResult, streamResult, modelResult, multiTurnResult)
	result.ModelList = modelList
	result.EndTime = time.Now().Format(timeFmt)
	result.TotalLatency = time.Since(startTime).Milliseconds()
	return result
}

// checkConnectivity 连通性检测
func (c *Checker) checkConnectivity(adapter protocol.Adapter, baseURL, apiKey string) CheckResult {
	start := time.Now()
	r := CheckResult{Item: CheckConnectivity}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	code, err := adapter.CheckConnectivity(ctx, baseURL, apiKey)
	r.Latency = time.Since(start).Milliseconds()

	if err != nil {
		r.Status = StatusFailed
		r.Message = "网络不可达"
		r.Detail = err.Error()
		return r
	}
	if code == 401 || code == 403 {
		r.Status = StatusWarning
		r.Message = fmt.Sprintf("网络可达, 认证失败 (HTTP %d)", code)
		r.Detail = "请检查 API Key"
		return r
	}
	if code >= 200 && code < 500 {
		r.Status = StatusSuccess
		r.Message = fmt.Sprintf("网络可达 (HTTP %d)", code)
		return r
	}
	r.Status = StatusFailed
	r.Message = fmt.Sprintf("服务异常 (HTTP %d)", code)
	return r
}

// checkChat 对话测试
func (c *Checker) checkChat(adapter protocol.Adapter, baseURL, apiKey, model string) CheckResult {
	start := time.Now()
	r := CheckResult{Item: CheckChat}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := adapter.Chat(ctx, protocol.ChatRequest{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Messages: []protocol.Message{{Role: "user", Content: "Hi, reply with exactly: OK"}},
	})
	r.Latency = time.Since(start).Milliseconds()

	if err != nil {
		r.Status = StatusFailed
		r.Message = "请求失败"
		r.Detail = err.Error()
		return r
	}
	if resp.Error != "" {
		r.Status = StatusFailed
		r.Message = resp.Error
		r.Detail = resp.RawBody
		return r
	}

	r.Status = StatusSuccess
	r.Message = "对话正常"
	r.Detail = truncate(resp.Content, 100)
	r.TokenIn = resp.PromptTokens
	r.TokenOut = resp.CompTokens
	return r
}

// checkStream 流式输出测试
func (c *Checker) checkStream(adapter protocol.Adapter, baseURL, apiKey, model string) CheckResult {
	start := time.Now()
	r := CheckResult{Item: CheckStream}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chunkCount := 0
	resp, err := adapter.ChatStream(ctx, protocol.ChatRequest{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Messages: []protocol.Message{{Role: "user", Content: "Count from 1 to 5"}},
		Stream:   true,
	}, func(chunk string, isFirst bool) {
		chunkCount++
		if isFirst {
			r.TTFT = time.Since(start).Milliseconds()
		}
	})
	r.Latency = time.Since(start).Milliseconds()

	if err != nil {
		r.Status = StatusFailed
		r.Message = "请求失败"
		r.Detail = err.Error()
		return r
	}
	if resp.Error != "" {
		r.Status = StatusFailed
		r.Message = resp.Error
		r.Detail = resp.RawBody
		return r
	}
	if chunkCount == 0 {
		r.Status = StatusFailed
		r.Message = "未收到流式数据"
		return r
	}

	r.Status = StatusSuccess
	r.Message = fmt.Sprintf("流式正常, %d chunks, TTFT %dms", chunkCount, r.TTFT)
	r.Detail = truncate(resp.Content, 100)
	return r
}

// checkModels 模型列表获取
func (c *Checker) checkModels(adapter protocol.Adapter, baseURL, apiKey string) (CheckResult, []string) {
	start := time.Now()
	r := CheckResult{Item: CheckModels}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	models, err := adapter.ListModels(ctx, baseURL, apiKey)
	r.Latency = time.Since(start).Milliseconds()

	if err != nil {
		r.Status = StatusWarning
		r.Message = "模型列表获取失败"
		r.Detail = err.Error()
		return r, nil
	}

	r.Status = StatusSuccess
	r.Message = fmt.Sprintf("获取到 %d 个模型", len(models))
	if len(models) > 5 {
		r.Detail = strings.Join(models[:5], ", ") + "..."
	} else {
		r.Detail = strings.Join(models, ", ")
	}
	return r, models
}

// checkMultiTurn 多轮对话测试
func (c *Checker) checkMultiTurn(adapter protocol.Adapter, baseURL, apiKey, model string) CheckResult {
	start := time.Now()
	r := CheckResult{Item: CheckMultiTurn}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 第一轮
	resp1, err := adapter.Chat(ctx, protocol.ChatRequest{
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
		Messages: []protocol.Message{{Role: "user", Content: "Remember this number: 42. Just reply OK."}},
	})
	if err != nil || resp1.Error != "" {
		r.Status = StatusFailed
		r.Message = "第一轮对话失败"
		if err != nil {
			r.Detail = err.Error()
		} else {
			r.Detail = resp1.Error
		}
		r.Latency = time.Since(start).Milliseconds()
		return r
	}

	// 第二轮
	resp2, err := adapter.Chat(ctx, protocol.ChatRequest{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		Messages: []protocol.Message{
			{Role: "user", Content: "Remember this number: 42. Just reply OK."},
			{Role: "assistant", Content: resp1.Content},
			{Role: "user", Content: "What number did I ask you to remember?"},
		},
	})
	r.Latency = time.Since(start).Milliseconds()

	if err != nil || resp2.Error != "" {
		r.Status = StatusFailed
		r.Message = "第二轮对话失败"
		if err != nil {
			r.Detail = err.Error()
		} else {
			r.Detail = resp2.Error
		}
		return r
	}

	if strings.Contains(resp2.Content, "42") {
		r.Status = StatusSuccess
		r.Message = "多轮对话正常, 上下文保持良好"
	} else {
		r.Status = StatusWarning
		r.Message = "多轮完成, 上下文可能丢失"
	}
	r.Detail = fmt.Sprintf("R1: %s | R2: %s", truncate(resp1.Content, 50), truncate(resp2.Content, 50))
	r.TokenIn = resp1.PromptTokens + resp2.PromptTokens
	r.TokenOut = resp1.CompTokens + resp2.CompTokens
	return r
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
