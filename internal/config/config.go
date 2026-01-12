package config

// config.go 负责加载与校验 YAML 配置，并提供默认值填充。
import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Log      LogConfig      `yaml:"log"`
	Server   ServerConfig   `yaml:"server"`
	WeCom    WeComConfig    `yaml:"wecom"`
	Unraid   UnraidConfig   `yaml:"unraid"`
	Qinglong QinglongConfig `yaml:"qinglong"`
	Auth     AuthConfig     `yaml:"auth"`
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

type QinglongConfig struct {
	Instances []QinglongInstance `yaml:"instances"`
}

type QinglongInstance struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	BaseURL      string `yaml:"base_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type AuthConfig struct {
	AllowedUserIDs []string `yaml:"allowed_userids"`
}

var qinglongInstanceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,31}$`)

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

	hasUnraid := strings.TrimSpace(cfg.Unraid.Endpoint) != "" || strings.TrimSpace(cfg.Unraid.APIKey) != ""
	if hasUnraid {
		if cfg.Unraid.Endpoint == "" {
			problems = append(problems, "unraid.endpoint 不能为空")
		}
		if cfg.Unraid.APIKey == "" {
			problems = append(problems, "unraid.api_key 不能为空")
		}
	}

	if len(cfg.Qinglong.Instances) > 0 {
		seen := make(map[string]struct{})
		for i, ins := range cfg.Qinglong.Instances {
			prefix := fmt.Sprintf("qinglong.instances[%d].", i)
			if strings.TrimSpace(ins.ID) == "" {
				problems = append(problems, prefix+"id 不能为空")
			} else {
				if !qinglongInstanceIDPattern.MatchString(ins.ID) {
					problems = append(problems, prefix+"id 不合法（仅允许字母数字及 _ -，长度≤32，且首字符为字母数字）")
				}
				if _, ok := seen[ins.ID]; ok {
					problems = append(problems, prefix+"id 重复")
				}
				seen[ins.ID] = struct{}{}
			}
			if strings.TrimSpace(ins.Name) == "" {
				problems = append(problems, prefix+"name 不能为空")
			}
			if strings.TrimSpace(ins.BaseURL) == "" {
				problems = append(problems, prefix+"base_url 不能为空")
			} else {
				u, err := url.Parse(ins.BaseURL)
				if err != nil || u.Scheme == "" || u.Host == "" {
					problems = append(problems, prefix+"base_url 不合法")
				}
			}
			if strings.TrimSpace(ins.ClientID) == "" {
				problems = append(problems, prefix+"client_id 不能为空")
			}
			if strings.TrimSpace(ins.ClientSecret) == "" {
				problems = append(problems, prefix+"client_secret 不能为空")
			}
		}
	}

	if len(cfg.Auth.AllowedUserIDs) == 0 {
		problems = append(problems, "auth.allowed_userids 不能为空（MVP 仅支持白名单）")
	}

	hasQinglong := len(cfg.Qinglong.Instances) > 0
	if !hasUnraid && !hasQinglong {
		problems = append(problems, "至少配置一个后端服务：unraid 或 qinglong.instances")
	}

	if len(problems) > 0 {
		return errors.New(fmt.Sprintf("配置校验失败: %s", strings.Join(problems, "; ")))
	}
	return nil
}
