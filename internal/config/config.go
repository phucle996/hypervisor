package config

import "time"

// Config is the root typed config object, loaded once at startup.
type Config struct {
	App   AppCfg
	Psql  PsqlCfg
	Redis RedisCfg
	IAM   IAMCfg
	GRPC  GRPCCfg
	Agent AgentCfg
}

// AppCfg holds application-level settings.
type AppCfg struct {
	TimeZone          string
	HTTPPort          int
	LogLV             string
	PublicURL         string
	TrustedProxies    []string
	AllowedOrigins    []string
	AdminAllowedCIDRs []string
}

// PsqlCfg holds PostgreSQL connection parameters.
type PsqlCfg struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Schema   string
	SSLMode  string

	TLSEnabled bool
	CACertPath string
	CertPath   string
	KeyPath    string

	MaxConns    int
	MinConns    int
	MaxConnLife time.Duration
	MaxConnIdle time.Duration

	PingTimeout   time.Duration
	MaxRetries    int
	RetryInterval time.Duration
}

// RedisCfg holds Redis connection parameters for cache and rate-limit usage.
type RedisCfg struct {
	Addr     string
	Password string
	DB       int

	TLSEnabled bool
	CACertPath string
	CertPath   string
	KeyPath    string

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PoolSize     int
	MinIdleConns int

	PingTimeout   time.Duration
	MaxRetries    int
	RetryInterval time.Duration
}

// IAMCfg holds IAM admin-session gRPC client settings.
type IAMCfg struct {
	GRPCTarget        string
	GRPCTLSCACertPath string
	GRPCTLSCertPath   string
	GRPCTLSKeyPath    string
	GRPCTLSServerName string
	RequestTimeout    time.Duration
}

// GRPCCfg holds Hypervisor gRPC server settings for kvm-agent.
type GRPCCfg struct {
	Enabled           bool
	ServerPort        string
	ServerPublicAddr  string
	ServerTLSCertPath string
	ServerTLSKeyPath  string
	ClientCACertPath  string
}

// AgentCfg holds kvm-agent bootstrap packaging and certificate settings.
type AgentCfg struct {
	CACertPath string
	CAKeyPath  string
	CertTTL    time.Duration
}

// LoadConfig reads environment variables and returns the root typed config.
func LoadConfig() *Config {
	return &Config{
		App: AppCfg{
			TimeZone:          getEnv("APP_TIMEZONE", "UTC"),
			HTTPPort:          getEnvAsInt("APP_HTTP_PORT", 8080),
			PublicURL:         getEnv("APP_PUBLIC_URL", "http://localhost:8080"),
			TrustedProxies:    getEnvAsCSV("APP_TRUSTED_PROXIES", nil),
			AllowedOrigins:    getEnvAsCSV("APP_ALLOWED_ORIGINS", nil),
			AdminAllowedCIDRs: getEnvAsCSV("ADMIN_ALLOWED_CIDRS", []string{"0.0.0.0/0", "::/0"}),
		},
		Psql: PsqlCfg{
			Host:          getEnv("PSQL_HOST", "localhost"),
			Port:          getEnvAsInt("PSQL_PORT", 5432),
			User:          getEnv("PSQL_USER", "postgres"),
			Password:      getEnv("PSQL_PASSWORD", ""),
			DBName:        getEnv("PSQL_DBNAME", "controlplane"),
			Schema:        getEnv("PSQL_SCHEMA", "hypervisor"),
			SSLMode:       getEnv("PSQL_SSLMODE", "disable"),
			TLSEnabled:    getEnvAsBool("PSQL_TLS_ENABLED", false),
			CACertPath:    getEnv("PSQL_TLS_CA", ""),
			CertPath:      getEnv("PSQL_TLS_CERT", ""),
			KeyPath:       getEnv("PSQL_TLS_KEY", ""),
			MaxConns:      getEnvAsInt("PSQL_MAX_CONNS", 20),
			MinConns:      getEnvAsInt("PSQL_MIN_CONNS", 5),
			MaxConnLife:   getEnvAsDuration("PSQL_MAX_CONN_LIFE", 30*time.Minute),
			MaxConnIdle:   getEnvAsDuration("PSQL_MAX_CONN_IDLE", 5*time.Minute),
			PingTimeout:   getEnvAsDuration("PSQL_PING_TIMEOUT", 5*time.Second),
			MaxRetries:    getEnvAsInt("PSQL_MAX_RETRIES", 5),
			RetryInterval: getEnvAsDuration("PSQL_RETRY_INTERVAL", 3*time.Second),
		},
		Redis: RedisCfg{
			Addr:          getEnv("REDIS_ADDR", "localhost:6379"),
			Password:      getEnv("REDIS_PASSWORD", ""),
			DB:            getEnvAsInt("REDIS_DB", 0),
			TLSEnabled:    getEnvAsBool("REDIS_TLS_ENABLED", false),
			CACertPath:    getEnv("REDIS_TLS_CA", ""),
			CertPath:      getEnv("REDIS_TLS_CERT", ""),
			KeyPath:       getEnv("REDIS_TLS_KEY", ""),
			DialTimeout:   getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:   getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:  getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolSize:      getEnvAsInt("REDIS_POOL_SIZE", 20),
			MinIdleConns:  getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			PingTimeout:   getEnvAsDuration("REDIS_PING_TIMEOUT", 5*time.Second),
			MaxRetries:    getEnvAsInt("REDIS_MAX_RETRIES", 5),
			RetryInterval: getEnvAsDuration("REDIS_RETRY_INTERVAL", 3*time.Second),
		},
		IAM: IAMCfg{
			GRPCTarget:        getEnv("IAM_GRPC_TARGET", "127.0.0.1:9090"),
			GRPCTLSCACertPath: getEnv("IAM_GRPC_TLS_CA", ""),
			GRPCTLSCertPath:   getEnv("IAM_GRPC_TLS_CERT", ""),
			GRPCTLSKeyPath:    getEnv("IAM_GRPC_TLS_KEY", ""),
			GRPCTLSServerName: getEnv("IAM_GRPC_TLS_SERVER_NAME", ""),
			RequestTimeout:    getEnvAsDuration("IAM_ADMIN_SESSION_REQUEST_TIMEOUT", 3*time.Second),
		},
		GRPC: GRPCCfg{
			Enabled:           getEnvAsBool("GRPC_SERVER_ENABLED", true),
			ServerPort:        getEnv("GRPC_SERVER_PORT", "9443"),
			ServerPublicAddr:  getEnv("GRPC_SERVER_PUBLIC_ADDR", ""),
			ServerTLSCertPath: getEnv("GRPC_SERVER_TLS_CERT", ""),
			ServerTLSKeyPath:  getEnv("GRPC_SERVER_TLS_KEY", ""),
			ClientCACertPath:  getEnv("GRPC_SERVER_CLIENT_CA", ""),
		},
		Agent: AgentCfg{
			CACertPath: getEnv("AGENT_CA_CERT", ""),
			CAKeyPath:  getEnv("AGENT_CA_KEY", ""),
			CertTTL:    getEnvAsDuration("AGENT_CERT_TTL", 720*time.Hour),
		},
	}
}
