package config

import "gin-admin/pkg/logger"

type Config struct {
	Global     Global
	Logger     logger.Options
	DataSource DataSource
	Util       Util
}

type Global struct {
	AppName            string `default:"ginadmin"`
	Version            string `default:"v1.0.0"`
	Debug              bool
	PprofAddr          string
	DisableSwagger     bool
	DisablePrintConfig bool
	MenuFile           string
	DenyDeleteMenu     bool
	Http               GlobalHttp
}

type GlobalHttp struct {
	Addr            string `default:":8080"`
	ShutdownTimeout int    `default:"10"`
	ReadTimeout     int    `default:"60"`
	WriteTimeout    int    `default:"60"`
	IdleTimeout     int    `default:"60"`
	SSLCertFile     string
	SSLKeyFile      string
}

type DataSource struct {
	DB DB
}

type DB struct {
	Debug        bool
	Type         string `default:"sqlite3"`
	DSN          string `default:"data/ginadmin.db"`
	MaxLifeTime  int    `default:"86400"`
	MaxIdleTime  int    `default:"3600"`
	MaxOpenConns int    `default:"100"`
	MaxIdleConns int    `default:"50"`
	TablePrefix  string `default:""`
	AutoMigrate  bool
	Resolver     []DBResolver
}

type DBResolver struct {
	DBType   string
	Sources  []string
	Replicas []string
	Tables   []string
}

type Util struct {
	Prometheus *Prometheus
}

type Prometheus struct {
	Enable         bool
	Port           int    `default:"9100"`
	BasicUsername  string `default:"admin"`
	BasicPassword  string `default:"admin"`
	LogApis        []string
	LogMethods     []string
	DefaultCollect bool
}

func (c *Config) IsDebug() bool {
	return c.Global.Debug
}
