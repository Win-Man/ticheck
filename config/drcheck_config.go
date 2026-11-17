package config

type DRCheckConfig struct {
	Log   Log      `toml:"log" json:"log"`
	DRCfg DRConfig `toml:"dr-config" json:"dr-config"`
}

type DRConfig struct {
	PDAddr string `toml:"pd-address" json:"pd-address"`
}

// InitDRCheckConfig loads and parses the dr-check TOML config.
func InitDRCheckConfig(configPath string) (DRCheckConfig, error) {
	var cfg DRCheckConfig
	if err := decodeConfigFile(configPath, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
