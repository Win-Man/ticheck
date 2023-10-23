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
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Win-Man/ticheck/config"
	"github.com/Win-Man/ticheck/pkg"
	"github.com/Win-Man/ticheck/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	CDCAPI_CAPTURES    = "/api/v2/captures"
	CDCAPI_CHANGEFEEDS = "/api/v2/changefeeds"
)

func newCDCCheckCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "cdc-check",
		Short: "cdc-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.InitCDCCheckConfig(configPath)
			logger.InitLogger(logLevel, logPath, cfg.Log)
			log.Info("Welcome to cdc-check")
			log.Debug(fmt.Sprintf("Flags:%+v", cmd.Flags()))
			log.Debug(fmt.Sprintf("arguments:%s", strings.Join(args, ",")))

			executeCDCCheck(cfg)

			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "C", "", "config file path")
	cmd.Flags().StringVarP(&logLevel, "log-level", "L", "", "log level: info, debug, warn, error, fatal")
	cmd.Flags().StringVar(&logPath, "log-path", "", "The path of log file")

	return cmd
}

func executeCDCCheck(cfg config.CDCCheckConfig) {
	captureTable, err := getCapturesInfo(cfg.CDCCfg.CDCAddr)
	if err != nil {
		log.Error(fmt.Sprintf("Get CDC Capture info failed:%v", err))
	} else {
		fmt.Println(captureTable.Render())
	}

	changefeedTable, err := getChangefeedsInfo(cfg.CDCCfg.CDCAddr)
	if err != nil {
		log.Error(fmt.Sprintf("Get CDC Changefeed info failed:%v", err))
	} else {
		fmt.Println(changefeedTable.Render())
	}

}

func getCapturesInfo(cdcaddr string) (table.Table, error) {
	// get captures info
	var resTable table.Table
	resTable.AppendHeader(table.Row{"address", "is_owner", "cluster_id"})
	capturesResp, err := http.Get(fmt.Sprintf("http://%s%s", cdcaddr, CDCAPI_CAPTURES))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", cdcaddr, CDCAPI_CAPTURES), err))
		return resTable, err
	}
	defer capturesResp.Body.Close()
	capturesBody, err := ioutil.ReadAll(capturesResp.Body)
	if err != nil {
		return resTable, err
	}
	capturesInfo := pkg.ListResponse[pkg.Capture]{}
	// check response status code
	if capturesResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(capturesBody)), &capturesInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			return resTable, err
		}
		log.Debug(fmt.Sprintf("Get cdc captures info:%v", capturesInfo))
		for _, c := range capturesInfo.Items {
			resTable.AppendRow(table.Row{c.AdvertiseAddr, c.IsOwner, c.ClusterID})
		}
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", capturesResp.StatusCode, http.StatusOK))
	}
	resTable.SetTitle("Captures Info")
	return resTable, nil
}

func getChangefeedsInfo(cdcaddr string) (table.Table, error) {
	// get captures info
	var resTable table.Table
	resTable.AppendHeader(table.Row{"id", "namespace", "state", "checkpoint_time", "error_code", "error_message"})
	changefeedsResp, err := http.Get(fmt.Sprintf("http://%s%s", cdcaddr, CDCAPI_CHANGEFEEDS))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", cdcaddr, CDCAPI_CAPTURES), err))
		return resTable, err
	}
	defer changefeedsResp.Body.Close()
	changefeedsBody, err := ioutil.ReadAll(changefeedsResp.Body)
	if err != nil {
		return resTable, err
	}
	changefeedsInfo := pkg.ListResponse[pkg.ChangefeedCommonInfo]{}
	// check response status code
	if changefeedsResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(changefeedsBody)), &changefeedsInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			return resTable, err
		}
		log.Debug(fmt.Sprintf("Get cdc changefeeds info:%v", changefeedsInfo))
		for _, c := range changefeedsInfo.Items {
			if c.RunningError != nil {
				resTable.AppendRow(table.Row{c.ID, c.Namespace, c.FeedState, c.CheckpointTime, c.RunningError.Code, c.RunningError.Message})
			} else {
				resTable.AppendRow(table.Row{c.ID, c.Namespace, c.FeedState, c.CheckpointTime, "", ""})
			}

		}
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", changefeedsResp.StatusCode, http.StatusOK))
	}
	resTable.SetTitle("Changefeeds Info")
	return resTable, nil
}
