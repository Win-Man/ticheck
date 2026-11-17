package config

type CDCCheckConfig struct {
	Log    Log       `toml:"log" json:"log"`
	CDCCfg CDCConfig `toml:"cdc-config" json:"cdc-config"`
}

type CDCConfig struct {
	CDCAddr []string `toml:"cdc-address" json:"cdc-address"`
}

// InitCDCCheckConfig loads and parses the cdc-check TOML config.
func InitCDCCheckConfig(configPath string) (CDCCheckConfig, error) {
	var cfg CDCCheckConfig
	if err := decodeConfigFile(configPath, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
