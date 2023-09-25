package config

import (
	"github.com/spf13/viper"
)

func InitializeConfig() error {
	viper.SetConfigName("config")  // 配置文件的前缀名
	viper.AddConfigPath("config/") // 配置文件的查找路径

	// 设置环境变量前缀，以便可以通过环境变量来覆盖配置文件中的值
	viper.SetEnvPrefix("MYAPP")

	// 读取环境变量，例如 MYAPP_ENV 可以是 "local" 或 "prod"
	viper.AutomaticEnv()

	// 根据环境变量来确定加载哪个配置文件
	env := viper.GetString("ENV")
	if env == "" {
		env = "prod"
	}
	viper.SetConfigName("config." + env + ".yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

func GetString(key string) string {
	return viper.GetString(key)
}
func GetBool(key string) bool {
	return viper.GetBool(key)
}
func GetInt(key string) int {
	return viper.GetInt(key)
}
