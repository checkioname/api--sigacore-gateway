package util

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ConnStr                     string        `mapstructure:"DB_SOURCE"`
	AuthServerAddress           string        `mapstructure:"AUTH_SERVER_ADDRESS"`
	GatewayServerAddress        string        `mapstructure:"GATEWAY_SERVER_ADDRESS"`
	TokenSymmetricKey           string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration         time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration        time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	AllowedIPs                  []string      `mapstructure:"ALLOWED_IPS"`
	UserServiceAddress          string        `mapstructure:"USER_SERVICE_ADDRESS"`
	ReportServiceAddress        string        `mapstructure:"REPORT_SERVICE_ADDRESS"`
	NotificationServiceAddress  string        `mapstructure:"NOTIFICATION_SERVICE_ADDRESS"`
}

func LoadConfig() (config Config, err error) {
	// ... (código existente para carregar)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	// Corrigir o unmarshal de listas a partir de strings separadas por vírgula
	ips := viper.GetString("ALLOWED_IPS")
	config.AllowedIPs = strings.Split(ips, ",")

	return
}
