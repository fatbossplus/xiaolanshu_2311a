package config

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
)

// Config 全局配置结构体
type Config struct {
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Logger LoggerConfig `mapstructure:"logger"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level   string `mapstructure:"level"`   // debug/info/warn/error
	File    string `mapstructure:"file"`    // 日志根目录，空则只输出到控制台
	Service string `mapstructure:"service"` // 服务名，注入每条日志
}

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

var (
	globalConfig *Config
	once         sync.Once
)

func LoadConfig(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		// 支持环境变量覆盖配置
		viper.AutomaticEnv() // 读取环境变量，覆盖配置文件中的相同配置

		// 读取配置文件
		if err = viper.ReadInConfig(); err != nil {
			err = fmt.Errorf("读取配置文件失败: %w", err)
			return
		}

		// 解析配置到结构体
		globalConfig = &Config{}
		if err = viper.Unmarshal(globalConfig); err != nil {
			err = fmt.Errorf("解析配置文件失败: %w", err)
			return
		}
	})

	return globalConfig, err
}
