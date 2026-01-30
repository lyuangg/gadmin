package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DBHost                  string `yaml:"db_host"`
	DBPort                  string `yaml:"db_port"`
	DBUser                  string `yaml:"db_user"`
	DBPassword              string `yaml:"db_password"`
	DBName                  string `yaml:"db_name"`
	JWTSecret               string `yaml:"jwt_secret"`
	Port                    string `yaml:"port"`
	LogType                 string `yaml:"log_type"`                   // 日志类型: json 或 text
	LogLevel                string `yaml:"log_level"`                  // 日志级别: debug, info, warn, error
	LogOutput               string `yaml:"log_output"`                 // 日志输出: stdout 或文件路径
	LogColorful             bool   `yaml:"log_colorful"`               // 是否开启日志颜色（level、请求日志状态码等，仅终端输出时有效）
	GinMode                 string `yaml:"gin_mode"`                   // Gin 模式: debug, release, test
	DBLogLevel              string `yaml:"db_log_level"`               // 数据库 SQL 日志级别: silent, error, warn, info
	DBSlowThresholdMs       int    `yaml:"db_slow_threshold_ms"`       // SQL 慢查询阈值（毫秒）
	DBLogColorful           bool   `yaml:"db_log_colorful"`            // SQL 日志是否带颜色（仅终端友好，文件建议关闭）
	DBTablePrefix           string `yaml:"db_table_prefix"`            // 数据库表前缀
	OperationLogRetainCount int    `yaml:"operation_log_retain_count"` // 操作日志保留条数，每日凌晨清理时保留最近 N 条，默认 10000
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{}
	if configPath != "" {
		if err := loadFromFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("加载配置文件失败: %w", err)
		}
	} else {
		loadFromEnv(cfg)
	}

	return cfg, nil
}

func loadFromFile(configPath string, cfg *Config) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	loadFromEnv(cfg)

	return nil
}

func loadFromEnv(cfg *Config) {
	if cfg.DBHost == "" {
		cfg.DBHost = getEnv("DB_HOST", "mysql")
	}
	if cfg.DBPort == "" {
		cfg.DBPort = getEnv("DB_PORT", "3306")
	}
	if cfg.DBUser == "" {
		cfg.DBUser = getEnv("DB_USER", "root")
	}
	if cfg.DBPassword == "" {
		cfg.DBPassword = getEnv("DB_PASSWORD", "root123456")
	}
	if cfg.DBName == "" {
		cfg.DBName = getEnv("DB_NAME", "gadmin")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	}
	if cfg.Port == "" {
		cfg.Port = getEnv("PORT", "8080")
	}
	if cfg.LogType == "" {
		cfg.LogType = getEnv("LOG_TYPE", "text")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = getEnv("LOG_LEVEL", "info")
	}
	if cfg.LogOutput == "" {
		cfg.LogOutput = getEnv("LOG_OUTPUT", "")
	}
	if v := os.Getenv("LOG_COLORFUL"); v != "" {
		cfg.LogColorful = v == "1" || strings.ToLower(v) == "true"
	}
	if cfg.GinMode == "" {
		cfg.GinMode = getEnv("GIN_MODE", "release")
	}
	if cfg.DBLogLevel == "" {
		cfg.DBLogLevel = getEnv("DB_LOG_LEVEL", "warn")
	}
	if cfg.DBSlowThresholdMs <= 0 {
		cfg.DBSlowThresholdMs = getEnvInt("DB_SLOW_THRESHOLD_MS", 200)
	}
	if !cfg.DBLogColorful {
		if v := os.Getenv("DB_LOG_COLORFUL"); v == "1" || strings.ToLower(v) == "true" {
			cfg.DBLogColorful = true
		}
	}
	if cfg.DBTablePrefix == "" {
		cfg.DBTablePrefix = getEnv("DB_TABLE_PREFIX", "")
	}
	if cfg.OperationLogRetainCount <= 0 {
		cfg.OperationLogRetainCount = getEnvInt("OPERATION_LOG_RETAIN_COUNT", 10000)
	}
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var n int
		if _, err := fmt.Sscanf(value, "%d", &n); err == nil && n > 0 {
			return n
		}
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) DSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}
