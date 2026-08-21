package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds every runtime knob of the application. It is populated from
// environment variables (and optionally a YAML file) so the same binary runs
// in dev and inside Docker Compose.
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	MySQL    MySQLConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Password PasswordConfig
	Rate     RateConfig
	Log      LogConfig
}

type AppConfig struct {
	Name         string
	Env          string // dev | prod
	DefaultAdmin DefaultAdmin
}

type DefaultAdmin struct {
	Username string
	Password string
	Email    string
	FullName string
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
}

type MySQLConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Params          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN assembles a MySQL data-source name with sensible defaults
// (parseTime for DATETIME <-> time.Time, utf8mb4 charset, timezone=Local).
// Only charset is set: the collation DSN param trips up MySQL 8.4's handshake
// and silently drops the connection back to latin1, corrupting Chinese text.
func (m MySQLConfig) DSN() string {
	params := m.Params
	if params == "" {
		params = "parseTime=true&loc=Local&charset=utf8mb4&allowNativePasswords=true"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", m.User, m.Password, m.Host, m.Port, m.Database, params)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns host:port.
func (r RedisConfig) Addr() string { return fmt.Sprintf("%s:%d", r.Host, r.Port) }

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

type PasswordConfig struct {
	BcryptCost int
}

type RateConfig struct {
	LoginPerIPPerMin   int
	LoginPerUserPerMin int
	WritePerMin        int // generic idempotent-write limiter
}

type LogConfig struct {
	Level  string // debug | info | warn | error
	Format string // json | console
}

// Load reads configuration. Precedence: defaults < env vars. Environment keys
// are uppercased and dot-segmented with underscores, e.g. MYSQL_HOST.
func Load() (*Config, error) {
	v := viper.New()
	setDefaults(v)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Optional config file (./configs/config.yaml) for local dev.
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	_ = v.ReadInConfig() // ignore error: env vars are sufficient

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "go-farm-production")
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.defaultadmin.username", "admin")
	v.SetDefault("app.defaultadmin.password", "Admin123!")
	v.SetDefault("app.defaultadmin.email", "admin@farm.local")
	v.SetDefault("app.defaultadmin.fullname", "系统管理员")

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.readtimeout", "15s")
	v.SetDefault("http.writetimeout", "30s")
	v.SetDefault("http.idletimeout", "60s")
	v.SetDefault("http.shutdowntimeout", "30s")
	v.SetDefault("http.maxbodybytes", 10*1024*1024)

	v.SetDefault("mysql.host", "127.0.0.1")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.user", "farm")
	v.SetDefault("mysql.password", "farmsecret")
	v.SetDefault("mysql.database", "farm_production")
	v.SetDefault("mysql.maxopenconns", 50)
	v.SetDefault("mysql.maxidleconns", 10)
	v.SetDefault("mysql.connmaxlifetime", "5m")

	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("jwt.secret", "change-me-in-production-please-32bytes-min")
	v.SetDefault("jwt.accesstokenttl", "15m")
	v.SetDefault("jwt.refreshtokenttl", "168h") // 7d
	v.SetDefault("jwt.issuer", "go-farm-production")

	v.SetDefault("password.bcryptcost", 10)

	v.SetDefault("rate.loginperippermin", 5)
	v.SetDefault("rate.loginperuserpermin", 10)
	v.SetDefault("rate.writepermin", 60)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}

func (c *Config) validate() error {
	if c.JWT.Secret == "" || len(c.JWT.Secret) < 32 {
		return fmt.Errorf("config: jwt.secret must be at least 32 bytes")
	}
	if c.Password.BcryptCost < 4 || c.Password.BcryptCost > 31 {
		return fmt.Errorf("config: password.bcryptcost must be between 4 and 31")
	}
	return nil
}

// IsProd reports whether the app runs in production mode.
func (c *Config) IsProd() bool { return strings.EqualFold(c.App.Env, "prod") }
