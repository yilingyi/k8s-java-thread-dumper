package global

import (
	"log"

	"github.com/spf13/viper"
)

var NOTIFY_VIPER *viper.Viper

func init() {
	NOTIFY_VIPER = viper.New()
	NOTIFY_VIPER.SetConfigName("config")   // 配置文件名 (不带扩展名)
	NOTIFY_VIPER.SetConfigType("yaml")     // 配置文件的扩展名
	NOTIFY_VIPER.AddConfigPath("./config") // 配置文件所在的目录

	err := NOTIFY_VIPER.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}
