package config

import (
	"github.com/BurntSushi/toml"
)

type DRCheckConfig struct {
	ResultFilePath string   `toml:"result-file-path" json:"result-file-path"`
	Log            Log      `toml:"log" json:"log"`
	DBConfig       DBConfig `toml:"db-config" json:"db-config"`
	DRCfg          DRConfig `toml:"dr-config" json:"dr-config"`
}

type DRConfig struct {
	PDAddr string `toml:"pd-address" json:"pd-address"`
}

// InitConfig Func
func InitDRCheckConfig(configPath string) (cfg DRCheckConfig) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
