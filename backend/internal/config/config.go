// Package config loads and validates process configuration from the
// environment. Required values fail fast with actionable messages; no real
// secrets may be committed — see .env.example for placeholders.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env          string // development | staging | production
	AppName      string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	// BodyLimitBytes caps ordinary JSON/API request bodies. Multipart
	// evidence uploads use the larger EvidenceUploadLimitBytes instead.
	BodyLimitBytes           int64
	EvidenceUploadLimitBytes int64

	DB   DBConfig
	S3   S3Config
	Auth AuthConfig
	Log  LogConfig
}

type DBConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string // require | verify-full in production
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		d.Host, d.Port, d.Name, d.User, d.Password, d.SSLMode,
	)
}

type S3Config struct {
	Endpoint  string // host:port, no scheme — server-side operations
	UseSSL    bool
	AccessKey string
	SecretKey string
	Bucket    string
	// PublicEndpoint/PublicSSL shape browser-facing presigned URLs. Set when
	// the internal Endpoint is not reachable by clients (e.g. `minio:9000`
	// inside a container network vs `ojt.example.edu` proxied at the edge).
	// The signature is computed over this host, so the edge proxy MUST pass
	// the client Host header through to object storage.
	PublicEndpoint string
	PublicSSL      bool
	// PresignTTL bounds how long generated evidence read URLs stay valid.
	PresignTTL time.Duration
}

type AuthConfig struct {
	// SessionTTL bounds an authenticated session's absolute lifetime.
	SessionTTL           time.Duration
	CookieName           string
	CookieSecure         bool
	CookieDomain         string // empty = host-only cookie
	AllowedOrigins       []string
	LoginRateLimit       int // attempts per window per key
	LoginRateWindow      time.Duration
	AttendanceRateLimit  int
	AttendanceRateWindow time.Duration
}

type LogConfig struct {
	Level  string
	Format string // console | json
}

// Load reads configuration from the environment (a local .env file in the
// working directory is honored for development). It returns an error when a
// required value is missing or malformed.
func Load() (*Config, error) {
	loadDotEnv(".env")

	c := &Config{
		Env:                      get("APP_ENV", "development"),
		AppName:                  get("APP_NAME", "OJT Attendance"),
		Port:                     getInt("APP_PORT", 8080),
		ReadTimeout:              getDur("APP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:             getDur("APP_WRITE_TIMEOUT", 30*time.Second),
		BodyLimitBytes:           getSize("APP_BODY_LIMIT", 1<<20),
		EvidenceUploadLimitBytes: getSize("EVIDENCE_UPLOAD_LIMIT", 6<<20),

		DB: DBConfig{
			Host:            get("DB_HOST", "127.0.0.1"),
			Port:            getInt("DB_PORT", 5432),
			Name:            get("DB_NAME", "ojt_management"),
			User:            get("DB_USER", "postgres"),
			Password:        get("DB_PASSWORD", ""),
			SSLMode:         get("DB_SSLMODE", "disable"),
			MaxConns:        int32(getInt("DB_MAX_CONNS", 20)),
			MinConns:        int32(getInt("DB_MIN_CONNS", 2)),
			MaxConnLifetime: getDur("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: getDur("DB_MAX_CONN_IDLE_TIME", 15*time.Minute),
		},
		S3: S3Config{
			Endpoint:       get("S3_ENDPOINT", "127.0.0.1:9000"),
			UseSSL:         getBool("S3_USE_SSL", false),
			AccessKey:      get("S3_ACCESS_KEY", ""),
			SecretKey:      get("S3_SECRET_KEY", ""),
			Bucket:         get("S3_BUCKET", "ojt-evidence"),
			PublicEndpoint: get("S3_PUBLIC_ENDPOINT", ""),
			PublicSSL:      getBool("S3_PUBLIC_SSL", getBool("S3_USE_SSL", false)),
			PresignTTL:     getDur("S3_PRESIGN_TTL", 5*time.Minute),
		},
		Auth: AuthConfig{
			SessionTTL:   getDur("SESSION_TTL", 12*time.Hour),
			CookieName:   get("SESSION_COOKIE_NAME", "ojt_session"),
			CookieSecure: getBool("SESSION_COOKIE_SECURE", false),
			CookieDomain: get("SESSION_COOKIE_DOMAIN", ""),
			AllowedOrigins: getList("ALLOWED_ORIGINS", []string{
				"http://localhost:5173", "http://localhost:4173",
			}),
			LoginRateLimit:       getInt("LOGIN_RATE_LIMIT_MAX", 10),
			LoginRateWindow:      getDur("LOGIN_RATE_LIMIT_WINDOW", time.Minute),
			AttendanceRateLimit:  getInt("ATTENDANCE_RATE_LIMIT_MAX", 20),
			AttendanceRateWindow: getDur("ATTENDANCE_RATE_LIMIT_WINDOW", time.Minute),
		},
		Log: LogConfig{
			Level:  get("LOG_LEVEL", "info"),
			Format: get("LOG_FORMAT", "console"),
		},
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

// validate checks required configuration. In production it refuses to start
// with insecure defaults rather than silently accepting them.
func (c *Config) validate() error {
	var problems []string
	bad := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	switch c.Env {
	case "development", "staging", "production":
	default:
		bad("APP_ENV must be development, staging, or production (got %q)", c.Env)
	}
	if c.DB.Name == "" {
		bad("DB_NAME is required")
	}
	if c.DB.User == "" {
		bad("DB_USER is required")
	}
	if c.DB.SSLMode == "" {
		bad("DB_SSLMODE is required (disable|require|verify-full)")
	}
	if c.S3.Endpoint == "" || c.S3.Bucket == "" {
		bad("S3_ENDPOINT and S3_BUCKET are required")
	}
	if c.S3.AccessKey == "" || c.S3.SecretKey == "" {
		bad("S3_ACCESS_KEY and S3_SECRET_KEY are required (see .env.example)")
	}
	if c.SessionTTL() <= 0 {
		bad("SESSION_TTL must be positive")
	}

	if c.IsProduction() {
		if c.DB.Password == "" {
			bad("DB_PASSWORD is required in production")
		}
		if c.DB.SSLMode == "disable" {
			bad("DB_SSLMODE=disable is not allowed in production; use require or verify-full")
		}
		if len(c.Auth.AllowedOrigins) == 0 {
			bad("ALLOWED_ORIGINS must list explicit origins in production (empty blocks all cross-origin requests)")
		}
		for _, o := range c.Auth.AllowedOrigins {
			if o == "*" {
				bad("ALLOWED_ORIGINS must not contain * in production")
			}
		}
		if !c.Auth.CookieSecure {
			bad("SESSION_COOKIE_SECURE=false is not allowed in production (session cookie must be Secure)")
		}
		if !c.S3.UseSSL {
			bad("S3_USE_SSL=false is not allowed in production")
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

// SessionTTL is a small helper so call sites read consistently.
func (c *Config) SessionTTL() time.Duration { return c.Auth.SessionTTL }

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDur(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getList(key string, def []string) []string {
	if v, ok := os.LookupEnv(key); ok {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return def
}

// getSize parses byte sizes like "1MB", "512KB", or a raw integer byte count.
func getSize(key string, def int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	v = strings.TrimSpace(strings.ToUpper(v))
	mult := int64(1)
	for suffix, m := range map[string]int64{"KB": 1 << 10, "MB": 1 << 20, "GB": 1 << 30, "B": 1} {
		if strings.HasSuffix(v, suffix) {
			mult = m
			v = strings.TrimSpace(strings.TrimSuffix(v, suffix))
			break
		}
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return def
	}
	return n * mult
}

// loadDotEnv populates unset environment variables from a .env file so local
// development works without exporting variables manually.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			_ = os.Setenv(k, v)
		}
	}
}
