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

import "github.com/BurntSushi/toml"

type Config struct {
	Log           Log            `toml:"log" json:"log"`
	MySQLConfig   DBConfig       `toml:"mysql-config" json:"mysql-config"`
	TiDBConfig    DBConfig       `toml:"tidb-config" json:"tidb-config"`
}

type Log struct {
	Level   string `toml:"log-level" json:"log-level"`
	LogPath string `toml:"log-path" json:"log-path"`
	LogDir  string `toml:"log-dir" json:"log-dir"`
}









// InitConfig Func
func InitConfig(configPath string) (cfg Config) {

	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
