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
	"fmt"
	"strings"

	"github.com/Win-Man/ticheck/config"
	"github.com/Win-Man/ticheck/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newDRCheckCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "dr-check",
		Short: "dr-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.InitConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to dr-check")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))

			executeDRCheck(cfg)

			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "C", "", "config file path")
	cmd.Flags().StringVarP(&logLevel, "log-level", "L", "", "log level: info, debug, warn, error, fatal")
	cmd.Flags().StringVar(&logPath, "log-path", "", "The path of log file")

	return cmd
}

func executeDRCheck(cfg config.Config) {

}
