/*
 * Created: 2022-09-10 11:48:20
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Win-Man/ticheck/config"
	"github.com/Win-Man/ticheck/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newDBCheckCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "db-check",
		Short: "db-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.InitArgsCheckConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to db-check")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))
			cfgBytes, err := json.Marshal(cfg)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			log.Debug(fmt.Sprintf("Config:%v", string(cfgBytes)))

			executeArgsCheck(cfg)

			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "C", "", "config file path")
	cmd.Flags().StringVarP(&logLevel, "log-level", "L", "", "log level: info, debug, warn, error, fatal")
	cmd.Flags().StringVar(&logPath, "log-path", "", "The path of log file")
	return cmd
}

func executeDBCheck(cfg config.ArgsCheckConfig) {
	// 各个节点连接数情况

	//  tikv 磁盘使用率

	// 长时间运行连接情况 active time 

	// 自定义 SQL 检测
}
