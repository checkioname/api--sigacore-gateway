package util

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Environment                string        `mapstructure:"ENVIRONMENT"`
	ConnStr                    string        `mapstructure:"DB_SOURCE"`
	AuthServerAddress          string        `mapstructure:"AUTH_SERVER_ADDRESS"`
	GatewayServerAddress       string        `mapstructure:"GATEWAY_SERVER_ADDRESS"`
	TokenSymmetricKey          string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration        time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration       time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	AllowedIPs                 []string      `mapstructure:"ALLOWED_IPS"`
	UserServiceAddress         string        `mapstructure:"USER_SERVICE_ADDRESS"`
	DocServiceAddress          string        `mapstructure:"DOC_SERVICE_ADDRESS"`
	NotificationServiceAddress string        `mapstructure:"NOTIFICATION_SERVICE_ADDRESS"`
}

// Constantes para ambientes
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTesting     = "testing"
)

// Chaves inseguras que não devem ser usadas em produção
var _unsafeKeys = map[string]bool{
	"12345678901234567890123456789012": true,
	"abcdefghijklmnopqrstuvwxyz123456": true,
	"00000000000000000000000000000000": true,
	"11111111111111111111111111111111": true,
	"testtesttesttesttesttesttest1234": true,
}

// LoadConfig loads configuration from environment variables and config files.
func LoadConfig() (config Config, err error) {
	// Configurar Viper
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Definir valores padrão
	setDefaults()

	// Tentar ler arquivo de configuração (opcional)
	if err := viper.ReadInConfig(); err != nil {
		// Em produção, não deve depender de arquivo .env
		if isProduction() {
			return config, fmt.Errorf("configuration file required in production: %w", err)
		}
		// Em desenvolvimento, é opcional
		fmt.Printf("Warning: No config file found, using environment variables only\n")
	}

	// Unmarshal para struct
	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Processar AllowedIPs (separados por vírgula)
	if ipsStr := viper.GetString("ALLOWED_IPS"); ipsStr != "" {
		config.AllowedIPs = strings.Split(ipsStr, ",")
		for i, ip := range config.AllowedIPs {
			config.AllowedIPs[i] = strings.TrimSpace(ip)
		}
	}

	// Validar configuração
	if err := validateConfig(&config); err != nil {
		return config, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// setDefaults define valores padrão para configuração
func setDefaults() {
	viper.SetDefault("ENVIRONMENT", EnvDevelopment)
	viper.SetDefault("AUTH_SERVER_ADDRESS", "localhost:8080")
	viper.SetDefault("GATEWAY_SERVER_ADDRESS", "localhost:8081")
	viper.SetDefault("ACCESS_TOKEN_DURATION", "15m")
	viper.SetDefault("REFRESH_TOKEN_DURATION", "24h")
	viper.SetDefault("ALLOWED_IPS", "127.0.0.1")
}

// validateConfig valida toda a configuração
func validateConfig(config *Config) error {
	// Validar ambiente
	if err := validateEnvironment(config.Environment); err != nil {
		return err
	}

	// Validar chave simétrica
	if err := validateSymmetricKey(config.TokenSymmetricKey, config.Environment); err != nil {
		return err
	}

	// Validar string de conexão do banco
	if err := validateDatabaseConfig(config.ConnStr, config.Environment); err != nil {
		return err
	}

	// Validar endereços dos serviços
	if err := validateServiceAddresses(config); err != nil {
		return err
	}

	// Validar IPs permitidos
	if err := validateAllowedIPs(config.AllowedIPs, config.Environment); err != nil {
		return err
	}

	return nil
}

// validateEnvironment valida se o ambiente é válido
func validateEnvironment(env string) error {
	switch env {
	case EnvDevelopment, EnvProduction, EnvTesting:
		return nil
	default:
		return fmt.Errorf("invalid environment '%s', must be one of: %s, %s, %s",
			env, EnvDevelopment, EnvProduction, EnvTesting)
	}
}

// validateSymmetricKey valida a chave simétrica
func validateSymmetricKey(key, environment string) error {
	if key == "" {
		return fmt.Errorf("TOKEN_SYMMETRIC_KEY is required")
	}

	if len(key) != 32 {
		return fmt.Errorf("TOKEN_SYMMETRIC_KEY must be exactly 32 characters, got %d", len(key))
	}

	// Em produção, verificar se não é uma chave insegura
	if environment == EnvProduction {
		if _unsafeKeys[key] {
			return fmt.Errorf("unsafe symmetric key detected in production environment")
		}

		// Verificar se tem entropia suficiente (não pode ser muito repetitiva)
		if !hasGoodEntropy(key) {
			return fmt.Errorf("symmetric key has low entropy, use a cryptographically secure key")
		}
	}

	return nil
}

// validateDatabaseConfig valida a configuração do banco
func validateDatabaseConfig(connStr, environment string) error {
	if connStr == "" {
		return fmt.Errorf("DB_SOURCE is required")
	}

	// Em produção, garantir que não usa credenciais padrão
	if environment == EnvProduction {
		if strings.Contains(connStr, "admin:admin") {
			return fmt.Errorf("default database credentials detected in production")
		}
		if !strings.Contains(connStr, "sslmode=require") && !strings.Contains(connStr, "sslmode=verify-full") {
			return fmt.Errorf("SSL required for database connection in production")
		}
	}

	return nil
}

// validateServiceAddresses valida os endereços dos serviços
func validateServiceAddresses(config *Config) error {
	addresses := map[string]string{
		"AUTH_SERVER_ADDRESS":          config.AuthServerAddress,
		"GATEWAY_SERVER_ADDRESS":       config.GatewayServerAddress,
		"USER_SERVICE_ADDRESS":         config.UserServiceAddress,
		"DOC_SERVICE_ADDRESS":          config.DocServiceAddress,
		"NOTIFICATION_SERVICE_ADDRESS": config.NotificationServiceAddress,
	}

	for name, addr := range addresses {
		if addr == "" {
			return fmt.Errorf("%s is required", name)
		}
	}

	return nil
}

// validateAllowedIPs valida a lista de IPs permitidos
func validateAllowedIPs(ips []string, environment string) error {
	if len(ips) == 0 {
		return fmt.Errorf("ALLOWED_IPS must contain at least one IP address")
	}

	// Em produção, avisar se localhost está na lista
	if environment == EnvProduction {
		for _, ip := range ips {
			if ip == "127.0.0.1" || ip == "localhost" {
				return fmt.Errorf("localhost IPs detected in production environment")
			}
		}
	}

	return nil
}

// hasGoodEntropy verifica se a string tem entropia suficiente
func hasGoodEntropy(s string) bool {
	// Contar caracteres únicos
	unique := make(map[rune]bool)
	for _, r := range s {
		unique[r] = true
	}

	// Deve ter pelo menos 16 caracteres únicos em uma chave de 32 chars
	return len(unique) >= 16
}

// isProduction verifica se está em ambiente de produção
func isProduction() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == EnvProduction
}

// IsDevelopment retorna true se estiver em ambiente de desenvolvimento
func (c *Config) IsDevelopment() bool {
	return c.Environment == EnvDevelopment
}

// IsProduction retorna true se estiver em ambiente de produção
func (c *Config) IsProduction() bool {
	return c.Environment == EnvProduction
}

// IsTesting retorna true se estiver em ambiente de teste
func (c *Config) IsTesting() bool {
	return c.Environment == EnvTesting
}
