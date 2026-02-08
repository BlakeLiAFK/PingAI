package checker

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Report 检测报告
type Report struct {
	GeneratedAt string            `json:"generatedAt"`
	Results     []FullCheckResult `json:"results"`
	Summary     ReportSummary     `json:"summary"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
	Warning int `json:"warning"`
}

// GenerateReport 生成 JSON 报告
func GenerateReport(results []FullCheckResult) string {
	summary := ReportSummary{Total: len(results)}
	for _, r := range results {
		allSuccess := true
		hasFailed := false
		for _, item := range r.Results {
			if item.Status == StatusFailed {
				hasFailed = true
				allSuccess = false
			} else if item.Status == StatusWarning {
				allSuccess = false
			}
		}
		if hasFailed {
			summary.Failed++
		} else if allSuccess {
			summary.Success++
		} else {
			summary.Warning++
		}
	}

	report := Report{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Results:     results,
		Summary:     summary,
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	return string(data)
}

// GenerateTextSummary 生成文本摘要
func GenerateTextSummary(results []FullCheckResult) string {
	var sb strings.Builder
	sb.WriteString("=== AI API Check Report ===\n")
	sb.WriteString(fmt.Sprintf("Time: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s] %s (%s)\n", r.ProviderName, r.Model, r.BaseURL))
		for _, item := range r.Results {
			icon := "?"
			switch item.Status {
			case StatusSuccess:
				icon = "OK"
			case StatusFailed:
				icon = "FAIL"
			case StatusWarning:
				icon = "WARN"
			}
			sb.WriteString(fmt.Sprintf("  %-15s [%s] %s (%dms)\n",
				string(item.Item), icon, item.Message, item.Latency))
		}
		sb.WriteString(fmt.Sprintf("  Total: %dms\n\n", r.TotalLatency))
	}

	return sb.String()
}
