/*
 * Created: 2021-03-25 14:34:55
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Log         Log      `toml:"log" json:"log"`
	MySQLConfig DBConfig `toml:"mysql-config" json:"mysql-config"`
	TiDBConfig  DBConfig `toml:"tidb-config" json:"tidb-config"`
}

type Log struct {
	Level   string `toml:"log-level" json:"log-level"`
	LogPath string `toml:"log-path" json:"log-path"`
	LogDir  string `toml:"log-dir" json:"log-dir"`
}

type DBConfig struct {
	Host     string `toml:"host" json:"host"`
	Port     int    `toml:"port" json:"port"`
	User     string `toml:"user" json:"user"`
	Password string `toml:"password" json:"password"`
	Database string `toml:"database" json:"database"`
}

// decodeConfigFile loads a TOML config file into cfg.
// A missing or unreadable config file must surface as an error instead of a
// panic, so the CLI can exit with a readable message and a non-zero status.
func decodeConfigFile(configPath string, cfg interface{}) error {
	if strings.TrimSpace(configPath) == "" {
		return fmt.Errorf("config file path is required, please set --config/-C (see config/*.toml.example)")
	}
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return fmt.Errorf("load config file %q failed: %w", configPath, err)
	}
	return nil
}

// InitConfig Func
func InitConfig(configPath string) (Config, error) {
	var cfg Config
	if err := decodeConfigFile(configPath, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
