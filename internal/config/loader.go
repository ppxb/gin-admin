package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gin-admin/pkg/encoding/toml"
	"github.com/creasty/defaults"
	"github.com/pkg/errors"
)

var (
	once        sync.Once
	C           = new(Config)
	configTypes = []string{".toml"}
)

func MustLoad(dir string) {
	once.Do(func() {
		if err := Load(dir); err != nil {
			panic(err)
		}
	})
}

func Load(dir string) error {
	if err := defaults.Set(C); err != nil {
		return err
	}

	configs, err := os.ReadDir(dir)
	if err != nil {
		return errors.Wrapf(err, "failed to read config dir %s", dir)
	}

	for _, config := range configs {
		fullname := filepath.Join(dir, config.Name())
		return parseFile(fullname)
	}
	return nil
}

func parseFile(name string) error {
	ext := filepath.Ext(name)

	if ext == "" || !strings.Contains(strings.Join(configTypes, ""), ext) {
		return nil
	}

	buf, err := os.ReadFile(name)
	if err != nil {
		return errors.Wrapf(err, "failed to read config file %s", name)
	}

	switch ext {
	case ".toml":
		err = toml.Unmarshal(buf, C)
	}
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal config file %s", name)

	}
	return nil
}
