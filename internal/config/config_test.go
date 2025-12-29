package config

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestConfig 创建测试用的临时配置目录
func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "aicommit-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	// 设置临时目录 (HOME for Unix, USERPROFILE for Windows)
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestLoadConfig_Default(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	// 验证默认值
	if cfg.Model != "gpt-4o" {
		t.Errorf("期望 Model='gpt-4o', 实际='%s'", cfg.Model)
	}
	if cfg.Language != "en" {
		t.Errorf("期望 Language='en', 实际='%s'", cfg.Language)
	}
	if cfg.Provider != "openai" {
		t.Errorf("期望 Provider='openai', 实际='%s'", cfg.Provider)
	}
	if cfg.AzureAPIVersion != "2024-02-15-preview" {
		t.Errorf("期望 AzureAPIVersion='2024-02-15-preview', 实际='%s'", cfg.AzureAPIVersion)
	}
}

func TestConfigSave(t *testing.T) {
	tmpDir, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()
	cfg.APIKey = "test-api-key"
	cfg.Model = "gpt-4"

	err := cfg.Save()
	if err != nil {
		t.Fatalf("保存配置失败: %v", err)
	}

	// 验证文件是否创建
	configPath := filepath.Join(tmpDir, ".config", "aicommit", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("配置文件未创建: %s", configPath)
	}

	// 重新加载验证
	cfg2 := LoadConfig()
	if cfg2.APIKey != "test-api-key" {
		t.Errorf("期望 APIKey='test-api-key', 实际='%s'", cfg2.APIKey)
	}
	if cfg2.Model != "gpt-4" {
		t.Errorf("期望 Model='gpt-4', 实际='%s'", cfg2.Model)
	}
}

func TestUpdateLanguage_Valid(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	testCases := []string{"en", "zh-CN", "zh-TW"}
	for _, lang := range testCases {
		err := cfg.UpdateLanguage(lang)
		if err != nil {
			t.Errorf("更新语言 '%s' 失败: %v", lang, err)
		}
		if cfg.Language != lang {
			t.Errorf("期望 Language='%s', 实际='%s'", lang, cfg.Language)
		}
	}
}

func TestUpdateLanguage_Invalid(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	err := cfg.UpdateLanguage("invalid-lang")
	if err == nil {
		t.Error("期望无效语言返回错误")
	}
}

func TestUpdateProvider_Valid(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	testCases := []string{"openai", "azure"}
	for _, provider := range testCases {
		err := cfg.UpdateProvider(provider)
		if err != nil {
			t.Errorf("更新提供商 '%s' 失败: %v", provider, err)
		}
		if cfg.Provider != provider {
			t.Errorf("期望 Provider='%s', 实际='%s'", provider, cfg.Provider)
		}
	}
}

func TestUpdateProvider_Invalid(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	err := cfg.UpdateProvider("invalid-provider")
	if err == nil {
		t.Error("期望无效提供商返回错误")
	}
}

func TestUpdateAPIKey(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	err := cfg.UpdateAPIKey("sk-test-key-123")
	if err != nil {
		t.Errorf("更新 API Key 失败: %v", err)
	}
	if cfg.APIKey != "sk-test-key-123" {
		t.Errorf("期望 APIKey='sk-test-key-123', 实际='%s'", cfg.APIKey)
	}
}

func TestUpdateBaseURL(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	testURL := "https://custom-api.example.com/v1"
	err := cfg.UpdateBaseURL(testURL)
	if err != nil {
		t.Errorf("更新 BaseURL 失败: %v", err)
	}
	if cfg.BaseURL != testURL {
		t.Errorf("期望 BaseURL='%s', 实际='%s'", testURL, cfg.BaseURL)
	}
}

func TestUpdateModel(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	err := cfg.UpdateModel("gpt-4-turbo")
	if err != nil {
		t.Errorf("更新 Model 失败: %v", err)
	}
	if cfg.Model != "gpt-4-turbo" {
		t.Errorf("期望 Model='gpt-4-turbo', 实际='%s'", cfg.Model)
	}
}

func TestUpdateAzureAPIVersion(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg := LoadConfig()

	testVersion := "2024-06-01"
	err := cfg.UpdateAzureAPIVersion(testVersion)
	if err != nil {
		t.Errorf("更新 AzureAPIVersion 失败: %v", err)
	}
	if cfg.AzureAPIVersion != testVersion {
		t.Errorf("期望 AzureAPIVersion='%s', 实际='%s'", testVersion, cfg.AzureAPIVersion)
	}
}
