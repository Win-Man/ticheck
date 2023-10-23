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
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Win-Man/ticheck/config"
	"github.com/Win-Man/ticheck/database"
	"github.com/Win-Man/ticheck/pkg"
	"github.com/Win-Man/ticheck/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newDBCheckCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "db-check",
		Short: "db-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.InitDBCheckConfig(configPath)
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

			executeDBCheck(cfg)

			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "C", "", "config file path")
	cmd.Flags().StringVarP(&logLevel, "log-level", "L", "", "log level: info, debug, warn, error, fatal")
	cmd.Flags().StringVar(&logPath, "log-path", "", "The path of log file")
	return cmd
}

func executeDBCheck(cfg config.DBCheckConfig) {
	var db *sql.DB
	var err error

	db, err = database.OpenMySQLDB(&cfg.DBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		os.Exit(1)
	}

	for _, kv := range cfg.DBCheckItems.UDSQLs {
		log.Debug(fmt.Sprintf("Execute item %v", kv))
		var sqlTable = table.Table{}
		
		sqlTable, err = getTableBySQL(db, kv.Sql)
		if err != nil {
			log.Error(fmt.Sprintf("Query Table failed! SQL is :%s Get error:%v", kv.Sql, err))
		} else {
			sqlTable.SetTitle(kv.Name)
			fmt.Println(sqlTable.Render())
		}
	}

	//  tikv 磁盘使用率

	// 长时间运行连接情况 active time
}

func getTableBySQL(db *sql.DB, sql string) (table.Table, error) {
	var resTable table.Table
	var err error
	tmpTable, err := pkg.QueryTable(db, sql)
	if err != nil {
		log.Error("Query error:%v", err)
		return resTable, err
	}
	var headerRow = table.Row{}
	for _, v := range tmpTable.ColumnHeader {
		headerRow = append(headerRow, v)
	}
	resTable.AppendHeader(headerRow)

	for _, row := range tmpTable.RecordList {
		var contentRow = table.Row{}
		for _, v := range row {
			contentRow = append(contentRow, v)
		}

		resTable.AppendRow(contentRow)
	}
	resTable.Style().Options.SeparateRows = true

	return resTable, err
}
