package provider

import (
	"strings"
	"testing"
)

func TestPresetsNotEmpty(t *testing.T) {
	presets := GetPresets()
	if len(presets) == 0 {
		t.Fatal("预设列表为空")
	}
}

func TestPresetsFieldsValid(t *testing.T) {
	validProtocols := map[string]bool{"openai": true, "anthropic": true, "gemini": true}
	seen := make(map[string]bool)

	for _, p := range GetPresets() {
		// ID 唯一且非空
		if p.ID == "" {
			t.Errorf("预设存在空 ID, Name=%q", p.Name)
		}
		if seen[p.ID] {
			t.Errorf("预设 ID 重复: %q", p.ID)
		}
		seen[p.ID] = true

		// 名称非空
		if p.Name == "" {
			t.Errorf("预设 %q 名称为空", p.ID)
		}

		// BaseURL 格式校验
		if !strings.HasPrefix(p.BaseURL, "https://") && !strings.HasPrefix(p.BaseURL, "http://") {
			t.Errorf("预设 %q BaseURL 格式无效: %q", p.ID, p.BaseURL)
		}

		// 协议校验
		if !validProtocols[p.Protocol] {
			t.Errorf("预设 %q 协议无效: %q", p.ID, p.Protocol)
		}

		// 模型列表非空
		if len(p.Models) == 0 {
			t.Errorf("预设 %q 模型列表为空", p.ID)
		}

		// 检查模型名称不含空白
		for _, m := range p.Models {
			if strings.TrimSpace(m) != m || m == "" {
				t.Errorf("预设 %q 模型名称无效: %q", p.ID, m)
			}
		}
	}
}

func TestPresetsProtocolDistribution(t *testing.T) {
	protocolCount := make(map[string]int)
	for _, p := range GetPresets() {
		protocolCount[p.Protocol]++
	}

	// 至少有三种协议的供应商
	if protocolCount["openai"] == 0 {
		t.Error("缺少 openai 协议供应商")
	}
	if protocolCount["anthropic"] == 0 {
		t.Error("缺少 anthropic 协议供应商")
	}
	if protocolCount["gemini"] == 0 {
		t.Error("缺少 gemini 协议供应商")
	}
}
