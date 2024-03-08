package gormx

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdmysql "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

type ResolverConfig struct {
	DBType   string
	Sources  []string
	Replicas []string
	Tables   []string
}

type Config struct {
	Debug        bool
	DBType       string
	DSN          string
	MaxLifeTime  int
	MaxIdleTime  int
	MaxOpenConns int
	MaxIdleConns int
	TablePrefix  string
	Resolver     []ResolverConfig
}

func New(cfg Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch strings.ToLower(cfg.DBType) {
	case "mysql":
		if err := createDatabaseWithMysql(cfg.DSN); err != nil {
			return nil, err
		}
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "sqlite3":
		_ = os.MkdirAll(filepath.Dir(cfg.DSN), os.ModePerm)
		dialector = sqlite.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	gormCfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.TablePrefix,
			SingularTable: true,
		},
		Logger: logger.Discard,
	}
	if cfg.Debug {
		gormCfg.Logger = logger.Default
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.Resolver) > 0 {
		resolver := &dbresolver.DBResolver{}

		for _, r := range cfg.Resolver {
			resolverCfg := dbresolver.Config{}
			dbType := strings.ToLower(r.DBType)
			var open func(dsn string) gorm.Dialector

			switch dbType {
			case "mysql":
				open = mysql.Open
			case "postgres":
				open = postgres.Open
			case "sqlite3":
				open = sqlite.Open
			default:
				continue
			}

			for _, replica := range r.Replicas {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(cfg.DSN), os.ModePerm)
				}
				resolverCfg.Replicas = append(resolverCfg.Replicas, open(replica))
			}
			for _, source := range r.Sources {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(cfg.DSN), os.ModePerm)
				}
				resolverCfg.Sources = append(resolverCfg.Sources, open(source))
			}
			tables := stringSliceToInterfaceSlice(r.Tables)
			resolver.Register(resolverCfg, tables...)
			zap.L().Info(fmt.Sprintf("use resolver, #tables: %v, #replicas: %v, #sources: %v \n", tables, r.Replicas, r.Sources))
		}

		resolver.
			SetMaxIdleConns(cfg.MaxIdleConns).
			SetMaxOpenConns(cfg.MaxOpenConns).
			SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Second).
			SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)
		err = db.Use(resolver)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Debug {
		db = db.Debug()
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDb.SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Second)
	sqlDb.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)
	return db, nil
}

func stringSliceToInterfaceSlice(s []string) []interface{} {
	r := make([]interface{}, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r
}

func createDatabaseWithMysql(dsn string) error {
	cfg, err := sdmysql.ParseDSN(dsn)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET = `utf8mb4`", cfg.DBName))
	return err
}
