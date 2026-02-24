package bootstrap

import (
	"amiya-eden/config"
	"amiya-eden/global"
	"fmt"

	"github.com/spf13/viper"
)

// InitConfig 读取并解析配置文件
func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 支持从环境变量覆盖配置
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("读取配置文件失败: %v", err))
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}

	global.Config = &cfg

	// 监听配置文件变化（热重载）
	viper.WatchConfig()
}
