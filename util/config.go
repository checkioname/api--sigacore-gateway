package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ConnStr              string        `mapstructure:"CONN_STR"`
	Addr                 string        `mapstructure:"ADDR"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
}

func LoadConfig() (config Config, err error) {
	viper.SetConfigName("app") // name of config file (without extension)
	viper.SetConfigType("env") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
