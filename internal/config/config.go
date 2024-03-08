package config

import "gin-admin/pkg/logger"

type Config struct {
	Global Global
	Logger logger.Options
	Util   Util
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
