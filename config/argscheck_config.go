package config

import (
	"github.com/BurntSushi/toml"
)

type ArgsCheckConfig struct {
	ResultFilePath string        `toml:"result-file-path" json:"result-file-path"`
	Log            Log           `toml:"log" json:"log"`
	DBConfig       DBConfig      `toml:"db-config" json:"db-config"`
	CheckTemp      CheckTemplate `toml:"check-template"`
}

type CheckTemplate struct {
	TiDBConfig []ConfigKV `toml:"tidb-config"`
	PDConfig   []ConfigKV `toml:"pd-config"`
	TiKVConfig []ConfigKV `toml:"tikv-config"`
	TiDBVars   []ConfigKV `toml:"tidb-variables"`
}



type ConfigKV struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

// InitConfig Func
func InitArgsCheckConfig(configPath string) (cfg ArgsCheckConfig) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
