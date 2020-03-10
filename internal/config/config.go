package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Network struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"network"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Name string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`
}

func InitConfig() (Config, error) {
	var C Config
	viper.AddConfigPath("/etc/meetingbot")
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	if err := viper.ReadInConfig(); err != nil {
		return C, err
	}

	err := viper.Unmarshal(&C)
	if err != nil {
		return C, err
	}
	return C, err
}