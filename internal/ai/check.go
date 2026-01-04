package ai

import (
	"context"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// CheckResult 配置检测结果
type CheckResult struct {
	ConfigExists     bool
	APIKeyConfigured bool
	APIKeyMasked     string
	Provider         string
	Model            string
	BaseURL          string
	APIConnected     bool
	ResponseTime     time.Duration
	Error            error
}

// Check 测试 API 连通性
// 发送一个简单的请求来验证 API Key 和网络连接
func (p *OpenAIProvider) Check(ctx context.Context) *CheckResult {
	result := &CheckResult{
		ConfigExists:     true, // 如果能到这里，配置已存在
		APIKeyConfigured: p.apiKey != "",
		Provider:         p.provider,
		Model:            p.model,
		BaseURL:          p.baseURL,
	}

	// 掩码 API Key
	if len(p.apiKey) > 8 {
		result.APIKeyMasked = p.apiKey[:4] + "..." + p.apiKey[len(p.apiKey)-4:]
	} else if p.apiKey != "" {
		result.APIKeyMasked = "***"
	}

	if !result.APIKeyConfigured {
		result.Error = fmt.Errorf("API Key 未配置")
		return result
	}

	// 发送测试请求
	start := time.Now()
	_, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hi",
				},
			},
			MaxTokens: 5,
		},
	)

	result.ResponseTime = time.Since(start)

	if err != nil {
		result.APIConnected = false
		result.Error = fmt.Errorf("API 连接失败: %w", err)
	} else {
		result.APIConnected = true
	}

	return result
}

// PrintCheckResult 打印检测结果
func PrintCheckResult(result *CheckResult) {
	fmt.Println("\n检查 AICommit 配置...")

	// 配置状态
	printStatus("配置文件", result.ConfigExists, "")
	printStatus("API Key", result.APIKeyConfigured, result.APIKeyMasked)
	printStatus("Provider", true, result.Provider)
	printStatus("Model", true, result.Model)
	if result.BaseURL != "" {
		printStatus("Base URL", true, result.BaseURL)
	}

	fmt.Println()

	// API 连通性
	if result.APIConnected {
		fmt.Printf("✓ API 连接: \033[32m成功\033[0m (%dms)\n", result.ResponseTime.Milliseconds())
	} else if result.Error != nil {
		fmt.Printf("✗ API 连接: \033[31m失败\033[0m\n")
		fmt.Printf("  错误: %v\n", result.Error)
	}

	fmt.Println()

	if result.APIConnected {
		fmt.Println("\033[32m所有检查通过 ✅\033[0m")
	} else {
		fmt.Println("\033[31m检查未通过 ❌\033[0m")
	}
}

func printStatus(name string, ok bool, value string) {
	status := "\033[32m✓\033[0m"
	if !ok {
		status = "\033[31m✗\033[0m"
	}

	if value != "" {
		fmt.Printf("%s %s: %s\n", status, name, value)
	} else {
		fmt.Printf("%s %s\n", status, name)
	}
}
