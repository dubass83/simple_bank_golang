package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config store all configuration of the application
// the values read by viper from config file or enviroment variables
type Config struct {
	Enviroment           string        `mapstructure:"ENVIROMENT"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	HTTPAddressString    string        `mapstructure:"HTTP_ADDRESS_STRING"`
	GRPCAddressString    string        `mapstructure:"GRPC_ADDRESS_STRING"`
	TokenString          string        `mapstructure:"TOKEN_STRING"`
	TokenDuration        time.Duration `mapstructure:"TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

// LoadConfig read configuration from config file or enviroment variables
func LoadConfig(configPath string) (config Config, err error) {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
