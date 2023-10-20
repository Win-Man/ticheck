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
	"os"
	"sort"
	"strings"

	"github.com/Win-Man/ticheck/config"
	"github.com/Win-Man/ticheck/pkg"
	"github.com/Win-Man/ticheck/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	PDAPI_STORE          = "/pd/api/v1/stores"
	PDAPI_CONFIG         = "/pd/api/v1/config"
	PDAPI_DRSTATUS       = "/pd/api/v1/replication_mode/status"
	PDAPI_PDLEADER       = "/pd/api/v1/leader"
	PDAPI_PLACEMENTRULES = "/pd/api/v1/config/rules"
)

type ConstraintInfo struct {
	ConKey    string
	ConOp     pkg.LabelConstraintOp
	ConValues []string
	ConRole   pkg.PeerRoleType
}

func newDRCheckCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "dr-check",
		Short: "dr-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.InitDRCheckConfig(configPath)
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

func executeDRCheck(cfg config.DRCheckConfig) {
	// prepare table
	var drInfoTable = table.Table{}
	extraHeader := []string{"Instance", "Role", "Region_count", "Leader_count"}

	// get location labels info
	cfgResp, err := http.Get(fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_CONFIG))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_CONFIG), err))
		os.Exit(1)
	}
	defer cfgResp.Body.Close()
	cfgBody, err := ioutil.ReadAll(cfgResp.Body)
	cfgInfo := pkg.Config{}
	if cfgResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(cfgBody)), &cfgInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			os.Exit(1)
		}
		log.Debug(fmt.Sprintf("Get config:%v", cfgInfo))
		//log.Debug(fmt.Sprintf("Get location labels:%s", strings.Join(cfgInfo.Replication.LocationLabels, ",")))
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", cfgResp.StatusCode, http.StatusOK))
		os.Exit(1)
	}
	// get placement rules
	rulesResp, err := http.Get(fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_PLACEMENTRULES))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_PLACEMENTRULES), err))
		os.Exit(1)
	}
	defer rulesResp.Body.Close()
	rulesBody, err := ioutil.ReadAll(rulesResp.Body)
	rulesInfo := []pkg.Rule{}
	if rulesResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(rulesBody)), &rulesInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			os.Exit(1)
		}
		//log.Debug(fmt.Sprintf("Get rules count:%d", len(rulesInfo)))
		log.Debug(fmt.Sprintf("Get placement rules:%v", rulesInfo))
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", cfgResp.StatusCode, http.StatusOK))
		os.Exit(1)
	}

	// prepare table header
	locationLabels := cfgInfo.Replication.LocationLabels
	header := table.Row{}
	for _, val := range locationLabels {
		header = append(header, val)
	}
	for _, val := range extraHeader {
		header = append(header, val)
	}
	drInfoTable.AppendHeader(header)

	// prepare table rows

	// get stores info
	storeResp, err := http.Get(fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_STORE))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_STORE), err))
		os.Exit(1)
	}
	defer storeResp.Body.Close()
	storeBody, err := ioutil.ReadAll(storeResp.Body)
	storeInfo := pkg.StoresInfo{}
	// check response status code
	if storeResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(storeBody)), &storeInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			os.Exit(1)
		}
		log.Debug(fmt.Sprintf("Get stores info:%v", storeInfo))
		//log.Debug(fmt.Sprintf("Get %d stores info", storeInfo.Count))
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", storeResp.StatusCode, http.StatusOK))
		os.Exit(1)
	}
	stores := storeInfo.Stores
	var storeRows []table.Row
	for _, store := range stores {
		log.Debug(fmt.Sprintf("Get store infomation:%s", store.Store.Address))
		mmap := make(map[string]string)
		for _, val := range locationLabels {
			mmap[val] = ""
		}
		storeLabels := store.Store.Labels
		for _, lab := range storeLabels {
			mmap[lab.Key] = lab.Value
		}
		var storeRole pkg.PeerRoleType
		log.Debug("Start to match constaint rules.........")
		for _, constra := range rulesInfo {

			for _, t := range constra.LabelConstraints {
				log.Debug(fmt.Sprintf("MatchStore Func store %s label value %s,constraint:%+v", store.Store.Address, store.GetLabelValue(t.Key), t))
				if t.MatchStore(store) {
					storeRole = constra.Role
					log.Debug(fmt.Sprintf("Constraint match！！！ store:%s %s", constra.Role, store.Store.Address))
				}
			}
		}

		address := store.Store.Address
		leaderCount := store.Status.LeaderCount
		regionCount := store.Status.RegionCount

		storeRow := table.Row{}
		for _, val := range locationLabels {
			storeRow = append(storeRow, mmap[val])
		}

		storeRow = append(storeRow, address)
		storeRow = append(storeRow, storeRole)
		storeRow = append(storeRow, regionCount)
		storeRow = append(storeRow, leaderCount)
		// expect tiflash
		if mmap["engine"] != "tiflash" {
			//drInfoTable.AppendRow(storeRow)
			storeRows = append(storeRows, storeRow)
		}

		log.Debug(fmt.Sprintf("store label info:%v", mmap))
	}
	sort.Slice(storeRows, func(i, j int) bool {
		// 比较每个内部切片的前三个元素
		for k := 1; k < 3; k++ {
			if storeRows[i][k] != storeRows[j][k] {
				// 使用类型断言将接口值转换为可比较的类型
				return storeRows[i][k].(string) < storeRows[j][k].(string)
			}
		}
		return false // 所有元素相等时，保持原顺序
	})

	drInfoTable.AppendRows(storeRows)
	// merge cell and set stype
	var colConfigs []table.ColumnConfig
	for idx, _ := range locationLabels {
		colConfigs = append(colConfigs, table.ColumnConfig{
			//Name:      strings.ToUpper(locationLabels[0]),
			Number:    idx + 1,
			AutoMerge: true,
			Align:     text.AlignCenter,
			VAlign:    text.VAlignMiddle,
		})
	}
	drInfoTable.SetColumnConfigs(colConfigs)
	drInfoTable.Style().Options.SeparateRows = true

	// check replication mode
	replicationMode := cfgInfo.ReplicationMode.ReplicationMode
	labelKey := cfgInfo.ReplicationMode.DRAutoSync.LabelKey
	primary := cfgInfo.ReplicationMode.DRAutoSync.Primary
	waitStoreTimeout := cfgInfo.ReplicationMode.DRAutoSync.WaitStoreTimeout
	println(drInfoTable.Render())
	fmt.Printf("TiDB Cluster replication mode is [%s]\n", replicationMode)

	// get DR state
	statusResp, err := http.Get(fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_DRSTATUS))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", cfg.DRCfg.PDAddr, PDAPI_CONFIG), err))
		os.Exit(1)
	}
	defer statusResp.Body.Close()
	statusBody, err := ioutil.ReadAll(statusResp.Body)
	statusInfo := pkg.HTTPReplicationStatus{}
	if statusResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(statusBody)), &statusInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			os.Exit(1)
		}
		log.Debug(fmt.Sprintf("Get status:%v", statusInfo))
		fmt.Printf("DR_AUTO_SYNC State is [%s]\n", statusInfo.DrAutoSync.State)
		//log.Debug(fmt.Sprintf("Get location labels:%s", strings.Join(cfgInfo.Replication.LocationLabels, ",")))
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", cfgResp.StatusCode, http.StatusOK))
		os.Exit(1)
	}

	if replicationMode == "dr-auto-sync" {
		//fmt.Printf("label-key: %s \tprimary: %s \nwait-store-timeout:%v \n", labelKey, primary, waitStoreTimeout)
		fmt.Printf("Primary label is [%s = %s]\n", labelKey, primary)
	}
	pdLeader := getPDLeader(cfg.DRCfg.PDAddr)
	fmt.Println(fmt.Sprintf("PD leader address is %s", pdLeader))
	fmt.Println(fmt.Sprintf("wait-store-timeout = %v", waitStoreTimeout))

}

func getPDLeader(pdaddr string) string {
	var pdleader string
	// get stores info
	leaderResp, err := http.Get(fmt.Sprintf("http://%s%s", pdaddr, PDAPI_PDLEADER))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", pdaddr, PDAPI_PDLEADER), err))
		os.Exit(1)
	}
	defer leaderResp.Body.Close()
	storeBody, err := ioutil.ReadAll(leaderResp.Body)
	leaderInfo := pkg.Member{}
	// check response status code
	if leaderResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(storeBody)), &leaderInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			os.Exit(1)
		}
		log.Debug(fmt.Sprintf("Get pd leader info:%v", leaderInfo))
		//log.Debug(fmt.Sprintf("pd leader %s ", leaderInfo.ClientUrls))
		pdleader = leaderInfo.ClientUrls[0]
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", leaderResp.StatusCode, http.StatusOK))
		os.Exit(1)
	}
	return pdleader
}
