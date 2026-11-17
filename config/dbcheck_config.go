package config

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

// InitDBCheckConfig loads and parses the db-check TOML config.
func InitDBCheckConfig(configPath string) (DBCheckConfig, error) {
	var cfg DBCheckConfig
	if err := decodeConfigFile(configPath, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
