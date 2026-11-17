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
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Win-Man/ticheck/config"
	"github.com/Win-Man/ticheck/database"
	"github.com/Win-Man/ticheck/pkg"
	"github.com/Win-Man/ticheck/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newArgsCheckCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "args-check",
		Short: "args-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.InitArgsCheckConfig(configPath)
			if err != nil {
				return err
			}
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to args-check")
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
	cmd.Flags().StringVar(&output, "output", "print", "print|file")
	return cmd
}

func executeArgsCheck(cfg config.ArgsCheckConfig) {
	var dbconn *sql.DB
	var err error

	dbconn, err = database.OpenMySQLDB(&cfg.DBConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Connect source database error:%v", err))
		os.Exit(1)
	}

	// check tidb config
	cTable, eTable, err := checkConfigbyComponent(dbconn, cfg.CheckTemp.TiDBConfig, COMPONENT_TIDB)
	if err != nil {
		log.Error(fmt.Sprintf("Check TiDB Config Error:%v", err))
	}
	if output == "print" {
		println(cTable.Render())
		println(eTable.Render())
	} else {
		WriteFile(cfg.ResultFilePath, cTable.Render())
		WriteFile(cfg.ResultFilePath, eTable.Render())
	}
	// check pd config
	cTable, eTable, err = checkConfigbyComponent(dbconn, cfg.CheckTemp.PDConfig, COMPONENT_PD)
	if err != nil {
		log.Error(fmt.Sprintf("Check PD Config Error:%v", err))
	}
	if output == "print" {
		println(cTable.Render())
		println(eTable.Render())
	} else {
		WriteFile(cfg.ResultFilePath, cTable.Render())
		WriteFile(cfg.ResultFilePath, eTable.Render())
	}
	// check tikv config
	cTable, eTable, err = checkConfigbyComponent(dbconn, cfg.CheckTemp.TiKVConfig, COMPONENT_TIKV)
	if err != nil {
		log.Error(fmt.Sprintf("Check TiKV Config Error:%v", err))
	}
	if output == "print" {
		println(cTable.Render())
		println(eTable.Render())
	} else {
		WriteFile(cfg.ResultFilePath, cTable.Render())
		WriteFile(cfg.ResultFilePath, eTable.Render())
	}
	// check tidb variables
	cTable, eTable, err = checkVariables(dbconn, cfg.DBConfig, cfg.CheckTemp.TiDBVars)
	if err != nil {
		log.Error(fmt.Sprintf("Check TiDB Variables Error:%v", err))
	}
	if output == "print" {
		println(cTable.Render())
		println(eTable.Render())
	} else {
		WriteFile(cfg.ResultFilePath, cTable.Render())
		WriteFile(cfg.ResultFilePath, eTable.Render())
	}
}

func checkConfigbyComponent(dbconn *sql.DB, kvs []config.ConfigKV, component string) (cTable table.Table, eTable table.Table, err error) {
	log.Info(fmt.Sprintf("Start to check %s config.", component))
	var instanceCount int
	var correctTable = table.Table{}
	var errorTable = table.Table{}
	header := table.Row{"Component", "Instance", "Config Name", "Current Value", "Expect Value", "Result"}
	correctTable.AppendHeader(header)
	correctTable.SetTitle(fmt.Sprintf("%s Config PASS TABLE", strings.ToUpper(component)))
	errorTable.AppendHeader(header)
	errorTable.SetTitle(fmt.Sprintf("%s Config NOPASS TABLE", strings.ToUpper(component)))
	countsql := fmt.Sprintf("SELECT TYPE,INSTANCE FROM INFORMATION_SCHEMA.CLUSTER_INFO WHERE TYPE='%s'", component)
	countTable, err := pkg.QueryTable(dbconn, countsql)
	if err != nil {
		log.Errorf("Query error:%v", err)
		return correctTable, errorTable, err
	}
	instanceCount = len(countTable.RecordList)
	for _, val := range kvs {
		log.Debug(fmt.Sprintf("check config key=%s value=%s", val.Name, val.Value))
		querysql := fmt.Sprintf("SHOW CONFIG WHERE NAME = '%s' AND TYPE = '%s'", val.Name, component)
		log.Debug(querysql)
		resTable, err := pkg.QueryTable(dbconn, querysql)
		if err != nil {
			log.Errorf("Query error:%v", err)
			return correctTable, errorTable, err
		}
		if len(resTable.RecordList) != instanceCount {
			log.Error(fmt.Sprintf("Config value count is not match with instance count.[%d:%d]", len(resTable.RecordList), instanceCount))
		}
		for _, v := range resTable.RecordList {
			actvalue := v[3]
			//TODO
			// off = OFF / 24h0m0s = 24h / 30MiB = 30MB / 40GiB = 40GB
			if valueEqual(actvalue, val.Value) {
				correctTable.AppendRow(table.Row{v[0], v[1], v[2], v[3], val.Value, "PASS"})
			} else {
				errorTable.AppendRow(table.Row{v[0], v[1], v[2], v[3], val.Value, "NOPASS"})
			}

		}
	}
	return correctTable, errorTable, nil
}

type DatabaseConnect struct {
	Cfg  config.DBConfig
	Conn *sql.DB
}

// parseInstanceAddress splits the "host:port" string reported by
// INFORMATION_SCHEMA.CLUSTER_INFO. net.SplitHostPort is used so IPv6
// addresses such as [::1]:4000 keep working.
func parseInstanceAddress(instance string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(strings.TrimSpace(instance))
	if err != nil {
		return "", 0, fmt.Errorf("parse instance %q failed: %w", instance, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("parse instance port %q failed: %w", instance, err)
	}
	if host == "" || port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("invalid instance %q", instance)
	}
	return host, port, nil
}

// openTiDBInstances dials every TiDB instance reported by
// INFORMATION_SCHEMA.CLUSTER_INFO using that instance's own host:port.
// Instances that fail to connect are logged and skipped, so one unreachable
// node does not hide the results collected from the healthy ones.
func openTiDBInstances(records [][]string, dbconfig config.DBConfig) []DatabaseConnect {
	conns := make([]DatabaseConnect, 0, len(records))
	for _, v := range records {
		if len(v) == 0 {
			continue
		}
		instance := v[0]
		host, port, err := parseInstanceAddress(instance)
		if err != nil {
			log.Errorf("Skip tidb instance:%v", err)
			continue
		}
		dbcfg := config.DBConfig{
			Host:     host,
			Port:     port,
			User:     dbconfig.User,
			Password: dbconfig.Password,
			Database: dbconfig.Database,
		}
		conn, err := database.OpenMySQLDB(&dbcfg)
		if err != nil {
			log.Errorf("Connect tidb instance %s error:%v", instance, err)
			continue
		}
		conns = append(conns, DatabaseConnect{Cfg: dbcfg, Conn: conn})
	}
	return conns
}

func checkVariables(dbconn *sql.DB, dbconfig config.DBConfig, kvs []config.ConfigKV) (cTable table.Table, eTable table.Table, err error) {
	log.Info("Start to check TiDB Variables")
	var correctTable = table.Table{}
	var errorTable = table.Table{}
	header := table.Row{"Component", "Instance", "Variables Name", "Current Value", "Expect Value", "Result"}
	correctTable.AppendHeader(header)
	correctTable.SetTitle("TiDB Variables PASS TABLE")
	errorTable.AppendHeader(header)
	errorTable.SetTitle("TiDB Variables NOPASS TABLE")
	getTiDBsql := "SELECT INSTANCE FROM INFORMATION_SCHEMA.CLUSTER_INFO WHERE TYPE='tidb'"
	tidbTable, err := pkg.QueryTable(dbconn, getTiDBsql)
	if err != nil {
		log.Errorf("Query error:%v", err)
		return correctTable, errorTable, err
	}

	tidbConns := openTiDBInstances(tidbTable.RecordList, dbconfig)
	defer func() {
		for _, c := range tidbConns {
			if c.Conn != nil {
				c.Conn.Close()
			}
		}
	}()
	if len(tidbConns) == 0 {
		log.Errorf("No reachable TiDB instance found in cluster_info, skip variables check")
		return correctTable, errorTable, nil
	}

	for _, val := range kvs {
		log.Debug(fmt.Sprintf("check variables key=%s value=%s", val.Name, val.Value))
		for _, conn := range tidbConns {
			instance := fmt.Sprintf("%s:%d", conn.Cfg.Host, conn.Cfg.Port)
			querySQL := fmt.Sprintf("show variables like '%s'", val.Name)
			resTable, err := pkg.QueryTable(conn.Conn, querySQL)
			if err != nil {
				// one broken instance must not abort the whole report
				log.Errorf("Query variables on instance %s failed:%v", instance, err)
				continue
			}
			if len(resTable.RecordList) != 1 {
				log.Errorf("Show variables result rows is not one,sql:%s,instance:%s", querySQL, instance)
				continue
			}
			actvalue := resTable.RecordList[0][1]
			//TODO
			// off = OFF / 24h0m0s = 24h / 30MiB = 30MB / 40GiB = 40GB
			if valueEqual(actvalue, val.Value) {
				correctTable.AppendRow(table.Row{"tidb", instance, val.Name, actvalue, val.Value, "PASS"})
			} else {
				errorTable.AppendRow(table.Row{"tidb", instance, val.Name, actvalue, val.Value, "NOPASS"})
			}
		}
	}
	return correctTable, errorTable, nil
}

func WriteFile(filePath string, content string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		os.Create(filePath)
	}
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("%s\n", content))
	w.Flush()
	return nil
}

func valueEqual(val1 string, val2 string) bool {

	// if could convert to string
	duration1, err := time.ParseDuration(val1)
	if err == nil {
		duration2, err := time.ParseDuration(val2)
		if err != nil {
			log.Error(fmt.Sprintf("Strings convert to Duration failed. [%s == %s]", val1, val2))
		} else {
			return strings.EqualFold(duration1.String(), duration2.String())
		}
	}
	//
	if hasUnit(val1) && hasUnit(val2) {
		val1Value, val1Unit := parseValueAndUnit(val1)
		val2Value, val2Unit := parseValueAndUnit(val2)

		// 比较数值和单位是否相等
		if val1Value == val2Value && areEquivalentUnits(val1Unit, val2Unit) {
			return true
		}
	}

	return strings.EqualFold(val1, val2)
}

func hasUnit(value string) bool {
	return strings.ContainsAny(value, "KMGTPEZY")
}

func parseValueAndUnit(value string) (float64, string) {
	value = strings.TrimSpace(value)

	// 提取数值部分和单位部分
	numEndIndex := 0
	for numEndIndex < len(value) && (value[numEndIndex] == '.' || isDigit(value[numEndIndex])) {
		numEndIndex++
	}

	numStr := value[:numEndIndex]
	unit := value[numEndIndex:]

	numValue := parseFloat(numStr)

	return numValue, unit
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func areEquivalentUnits(unit1, unit2 string) bool {
	// GiB == GB == G
	unit1 = strings.Replace(strings.ToLower(unit1), "i", "", -1)
	unit2 = strings.Replace(strings.ToLower(unit2), "i", "", -1)
	unit1 = strings.TrimRight(unit1, "b")
	unit2 = strings.TrimRight(unit2, "b")
	return unit1 == unit2

}

func parseFloat(str string) float64 {
	value, _ := strconv.ParseFloat(str, 64)
	return value
}
