package util

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment    string   `mapstructure:"ENVIRONMENT"`
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`
	DBSource       string   `mapstructure:"DB_SOURCE"`
	DBDriver       string   `mapstructure:"DB_DRIVER"`
	KafkaBroker    string   `mapstructure:"KAFKA_BROKER"`
	KafkaTopic     string   `mapstructure:"KAFKA_Topic"`
	RestPort       string   `mapstructure:"REST_PORT"`
	WsPort         string   `mapstructure:"WS_PORT"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
