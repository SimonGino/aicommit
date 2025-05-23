package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url,omitempty"`
	Model    string `json:"model,omitempty"`
	Language string `json:"language"`
}

func LoadConfig() *Config {
	cfg := &Config{
		Model:    "gpt-4o",
		Language: "en",
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

func (c *Config) UpdateAPIKey(apiKey string) error {
	c.APIKey = apiKey
	return c.Save()
}

func (c *Config) UpdateBaseURL(baseURL string) error {
	c.BaseURL = baseURL
	return c.Save()
}

func (c *Config) UpdateModel(model string) error {
	c.Model = model
	return c.Save()
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
