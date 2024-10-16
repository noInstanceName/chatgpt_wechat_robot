package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/qingconglaixueit/wechatbot/pkg/logger"
)

// Configuration 项目配置
type Configuration struct {
	// 扫描二维码方式, url or console
	QRCallback string `json:"qr_callback" yaml:"qr-callback"`
	// 自动通过好友
	AutoPass bool `json:"auto_pass" yaml:"auto-pass"`
	// gpt url
	GptUrl string `json:"gpt_url" yaml:"gpt-url"`
	// gpt apikey
	ApiKey string `json:"api_key" yaml:"api-key"`
	// 会话超时时间
	SessionTimeout time.Duration `json:"session_timeout" yaml:"session-timeout"`
	// GPT请求最大字符数
	MaxTokens uint `json:"max_tokens" yaml:"max-tokens"`
	// 记录的最大历史轮数
	MaxHistoryRound int `json:"max_history_round" yaml:"max-history-round"`
	// GPT模型
	Model string `json:"model" yaml:"model"`
	// 热度
	Temperature float64 `json:"temperature" yaml:"temperature"`
	// 回复前缀
	ReplyPrefix string `json:"reply_prefix" yaml:"reply-prefix"`
	// 清空会话口令
	SessionClearToken string `json:"session_clear_token" yaml:"session-clear-token"`
}

var config *Configuration
var once sync.Once

// LoadConfig 加载配置
func LoadConfig() *Configuration {
	once.Do(func() {
		// 给配置赋默认值
		config = &Configuration{
			AutoPass:          false,
			SessionTimeout:    60,
			MaxTokens:         512,
			Model:             "o1-mini",
			Temperature:       0.9,
			SessionClearToken: "下个问题",
		}

		// 判断配置文件是否存在，存在直接JSON读取
		_, err := os.Stat("config.json")
		if err == nil {
			f, err := os.Open("config.json")
			if err != nil {
				logger.Danger(fmt.Sprintf("open config error: %v", err))
				return
			}
			defer f.Close()
			encoder := json.NewDecoder(f)
			err = encoder.Decode(config)
			if err != nil {
				logger.Danger(fmt.Sprintf("decode config error: %v", err))
				return
			}
		}
		// 有环境变量使用环境变量
		ApiKey := os.Getenv("APIKEY")
		AutoPass := os.Getenv("AUTO_PASS")
		SessionTimeout := os.Getenv("SESSION_TIMEOUT")
		Model := os.Getenv("MODEL")
		MaxTokens := os.Getenv("MAX_TOKENS")
		Temperature := os.Getenv("TEMPREATURE")
		ReplyPrefix := os.Getenv("REPLY_PREFIX")
		SessionClearToken := os.Getenv("SESSION_CLEAR_TOKEN")
		if ApiKey != "" {
			config.ApiKey = ApiKey
		}
		if AutoPass == "true" {
			config.AutoPass = true
		}
		if SessionTimeout != "" {
			duration, err := time.ParseDuration(SessionTimeout)
			if err != nil {
				logger.Danger(fmt.Sprintf("config session timeout error: %v, get is %v", err, SessionTimeout))
				return
			}
			config.SessionTimeout = duration
		}
		if Model != "" {
			config.Model = Model
		}
		if MaxTokens != "" {
			max, err := strconv.Atoi(MaxTokens)
			if err != nil {
				logger.Danger(fmt.Sprintf("config max tokens error: %v ,get is %v", err, MaxTokens))
				return
			}
			config.MaxTokens = uint(max)
		}
		if Temperature != "" {
			temp, err := strconv.ParseFloat(Temperature, 64)
			if err != nil {
				logger.Danger(fmt.Sprintf("config temperature error: %v, get is %v", err, Temperature))
				return
			}
			config.Temperature = temp
		}
		if ReplyPrefix != "" {
			config.ReplyPrefix = ReplyPrefix
		}
		if SessionClearToken != "" {
			config.SessionClearToken = SessionClearToken
		}

	})
	if config.ApiKey == "" {
		logger.Danger("config error: api key required")
	}

	return config
}
