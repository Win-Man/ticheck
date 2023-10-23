package config

import (
	"github.com/BurntSushi/toml"
)

type CDCCheckConfig struct {
	Log            Log       `toml:"log" json:"log"`
	CDCCfg         CDCConfig `toml:"cdc-config" json:"cdc-config"`
}

type CDCConfig struct {
	CDCAddr string `toml:"cdc-address" json:"cdc-address"`
}

// InitConfig Func
func InitCDCCheckConfig(configPath string) (cfg CDCCheckConfig) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
