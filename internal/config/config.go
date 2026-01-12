package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Log    LogConfig    `yaml:"log"`
	Server ServerConfig `yaml:"server"`
	WeCom  WeComConfig  `yaml:"wecom"`
	Unraid UnraidConfig `yaml:"unraid"`
	Auth   AuthConfig   `yaml:"auth"`
}

type LogConfig struct {
	Level LogLevel `yaml:"level"`
}

type LogLevel string

func (l LogLevel) ToSlogLevel() slog.Level {
	switch strings.ToLower(string(l)) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type ServerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	BaseURL    string `yaml:"base_url"`
}

type WeComConfig struct {
	CorpID         string `yaml:"corpid"`
	AgentID        int    `yaml:"agentid"`
	Secret         string `yaml:"secret"`
	Token          string `yaml:"token"`
	EncodingAESKey string `yaml:"encoding_aes_key"`
	APIBaseURL     string `yaml:"api_base_url"`
}

type UnraidConfig struct {
	Endpoint            string `yaml:"endpoint"`
	APIKey              string `yaml:"api_key"`
	Origin              string `yaml:"origin"`
	ForceUpdateMutation string `yaml:"force_update_mutation"`
}

type AuthConfig struct {
	AllowedUserIDs []string `yaml:"allowed_userids"`
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}

	applyDefaults(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}
	if cfg.WeCom.APIBaseURL == "" {
		cfg.WeCom.APIBaseURL = "https://qyapi.weixin.qq.com/cgi-bin"
	}
	if cfg.Unraid.Origin == "" {
		cfg.Unraid.Origin = "daily-help"
	}
}

func validate(cfg Config) error {
	var problems []string

	if cfg.Server.ListenAddr == "" {
		problems = append(problems, "server.listen_addr 不能为空")
	}

	if cfg.WeCom.CorpID == "" {
		problems = append(problems, "wecom.corpid 不能为空")
	}
	if cfg.WeCom.AgentID == 0 {
		problems = append(problems, "wecom.agentid 不能为空")
	}
	if cfg.WeCom.Secret == "" {
		problems = append(problems, "wecom.secret 不能为空")
	}
	if cfg.WeCom.Token == "" {
		problems = append(problems, "wecom.token 不能为空")
	}
	if cfg.WeCom.EncodingAESKey == "" {
		problems = append(problems, "wecom.encoding_aes_key 不能为空")
	}
	if cfg.WeCom.APIBaseURL == "" {
		problems = append(problems, "wecom.api_base_url 不能为空")
	}

	if cfg.Unraid.Endpoint == "" {
		problems = append(problems, "unraid.endpoint 不能为空")
	}
	if cfg.Unraid.APIKey == "" {
		problems = append(problems, "unraid.api_key 不能为空")
	}

	if len(cfg.Auth.AllowedUserIDs) == 0 {
		problems = append(problems, "auth.allowed_userids 不能为空（MVP 仅支持白名单）")
	}

	if len(problems) > 0 {
		return errors.New(fmt.Sprintf("配置校验失败: %s", strings.Join(problems, "; ")))
	}
	return nil
}
