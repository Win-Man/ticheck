package config

import (
	"github.com/BurntSushi/toml"
)

type DBCheckConfig struct {
	Log          Log          `toml:"log" json:"log"`
	DBConfig     DBConfig     `toml:"db-config" json:"db-config"`
	DBCheckItems DBCheckItems `toml:"db-check-items" json:"db-check-items"`
}

type DBCheckItems struct {
	UDSQLs []UDSQL `toml:"user-defined-sqls"`
}

type UDSQL struct {
	Name string `toml:"name" json:"name"`
	Sql  string `toml:"sql" json:"sql"`
}

// InitConfig Func
func InitDBCheckConfig(configPath string) (cfg DBCheckConfig) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
