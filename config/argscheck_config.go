package config

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

// InitArgsCheckConfig loads and parses the args-check TOML config.
func InitArgsCheckConfig(configPath string) (ArgsCheckConfig, error) {
	var cfg ArgsCheckConfig
	if err := decodeConfigFile(configPath, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
