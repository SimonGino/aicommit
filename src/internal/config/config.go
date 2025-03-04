package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	QwenAPIKey      string `json:"qwen_api_key,omitempty"`
	OpenAIAPIKey    string `json:"openai_api_key,omitempty"`
	DeepseekAPIKey  string `json:"deepseek_api_key,omitempty"`
	DefaultProvider string `json:"default_provider"`
	Language        string `json:"language"`
}

func LoadConfig() *Config {
	cfg := &Config{
		DefaultProvider: "qwen",
		Language:        "en",
	}

	configFile := cfg.ConfigFile()
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return cfg
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg
	}

	return cfg
}

func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.ConfigFile()), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(c.ConfigFile(), data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

func (c *Config) UpdateAPIKey(provider, apiKey string) error {
	switch provider {
	case "qwen":
		c.QwenAPIKey = apiKey
	case "openai":
		c.OpenAIAPIKey = apiKey
	case "deepseek":
		c.DeepseekAPIKey = apiKey
	default:
		return fmt.Errorf("不支持的AI提供商: %s", provider)
	}

	return c.Save()
}

func (c *Config) GetAPIKey(provider string) string {
	switch provider {
	case "qwen":
		return c.QwenAPIKey
	case "openai":
		return c.OpenAIAPIKey
	case "deepseek":
		return c.DeepseekAPIKey
	default:
		return ""
	}
}

func (c *Config) UpdateLanguage(language string) error {
	switch language {
	case "en", "zh-CN", "zh-TW":
		c.Language = language
	default:
		return fmt.Errorf("不支持的语言: %s", language)
	}

	return c.Save()
}

func (c *Config) ConfigFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".config", "aicommit", "config.json")
}
