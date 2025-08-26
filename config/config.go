package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Bind9 struct {
		ZoneFileDir string `mapstructure:"zone_file_dir"`
		ConfigFile  string `mapstructure:"config_file"`
		DefaultTTL  uint32 `mapstructure:"default_ttl"`
	} `mapstructure:"bind9"`
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	ApiAccess []struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"api_access"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	err = viper.Unmarshal(&config)
	return &config, err
}
