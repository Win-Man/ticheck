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
	"io"
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
	PDAPI_PDMEMBERS      = "/pd/api/v1/members"
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
			cfg, err := config.InitDRCheckConfig(configPath)
			if err != nil {
				return err
			}
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
	cfgBody, err := io.ReadAll(cfgResp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("io.ReadAll error:%v", err))
		os.Exit(1)
	}
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
	rulesBody, err := io.ReadAll(rulesResp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("io.ReadAll error:%v", err))
		os.Exit(1)
	}
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
		log.Error(fmt.Sprintf("Http get response code get %d, not %d", rulesResp.StatusCode, http.StatusOK))
		os.Exit(1)
	}

	// prepare table header
	locationLabels := cfgInfo.Replication.LocationLabels
	if len(locationLabels) == 0 {
		// dr-auto-sync cannot tell the two data centers apart without
		// location-labels, and the table loses its zone columns. Report it
		// instead of silently printing a label-less table.
		fmt.Fprintln(os.Stderr, "WARNING: [replication].location-labels is empty in the PD config. "+
			"dr-auto-sync needs it (usually \"zone,host\") to group stores by data center, "+
			"so the store table below has no label columns.")
	}
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
	storeBody, err := io.ReadAll(storeResp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("io.ReadAll error:%v", err))
		os.Exit(1)
	}
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
	storeRows := buildStoreRows(storeInfo.Stores, locationLabels, rulesInfo)

	drInfoTable.AppendRows(storeRows)
	// merge cell and set stype
	var colConfigs []table.ColumnConfig
	for idx := 0; idx < len(locationLabels); idx++ {
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

	drstate := getDRState(cfg.DRCfg.PDAddr)
	fmt.Printf("DR_AUTO_SYNC State is [%s]\n", drstate)

	if replicationMode == "dr-auto-sync" {
		//fmt.Printf("label-key: %s \tprimary: %s \nwait-store-timeout:%v \n", labelKey, primary, waitStoreTimeout)
		fmt.Printf("Primary label is [%s = %s]\n", labelKey, primary)
	}
	// pdLeader := getPDLeader(cfg.DRCfg.PDAddr)
	// fmt.Println(fmt.Sprintf("PD leader address is %s", pdLeader))

	members := getPDMembers(cfg.DRCfg.PDAddr)
	fmt.Printf("PD member info :\n%s\n", strings.Join(members, "\n"))

	fmt.Printf("Configs info:\nwait-store-timeout = %v\n", waitStoreTimeout)

}

func getDRState(pdaddr string) string {
	// get DR state
	statusResp, err := http.Get(fmt.Sprintf("http://%s%s", pdaddr, PDAPI_DRSTATUS))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", pdaddr, PDAPI_CONFIG), err))
		return ""
	}
	defer statusResp.Body.Close()
	statusBody, err := io.ReadAll(statusResp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("io.ReadAll error:%v", err))
		return ""
	}
	statusInfo := pkg.HTTPReplicationStatus{}
	if statusResp.StatusCode == http.StatusOK {
		err = json.Unmarshal([]byte(string(statusBody)), &statusInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			return ""
		}
		log.Debug(fmt.Sprintf("Get status:%v", statusInfo))
		return statusInfo.DrAutoSync.State
		//fmt.Printf("DR_AUTO_SYNC State is [%s]\n", statusInfo.DrAutoSync.State)
		//log.Debug(fmt.Sprintf("Get location labels:%s", strings.Join(cfgInfo.Replication.LocationLabels, ",")))
	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", statusResp.StatusCode, http.StatusOK))
		return ""
	}
}

// func getPDLeader(pdaddr string) string {
// 	var pdleader string
// 	// get leader info
// 	leaderResp, err := http.Get(fmt.Sprintf("http://%s%s", pdaddr, PDAPI_PDLEADER))
// 	if err != nil {
// 		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", pdaddr, PDAPI_PDLEADER), err))
// 		return ""
// 	}
// 	defer leaderResp.Body.Close()
// 	storeBody, err := io.ReadAll(leaderResp.Body)
// 	if err != nil {
// 		log.Error(fmt.Sprintf("io.ReadAll error:%v", err))
// 		return ""
// 	}
// 	leaderInfo := pkg.Member{}
// 	// check response status code
// 	if leaderResp.StatusCode == http.StatusOK {
// 		err = json.Unmarshal([]byte(string(storeBody)), &leaderInfo)
// 		if err != nil {
// 			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
// 			return ""
// 		}
// 		log.Debug(fmt.Sprintf("Get pd leader info:%v", leaderInfo))
// 		//log.Debug(fmt.Sprintf("pd leader %s ", leaderInfo.ClientUrls))
// 		pdleader = leaderInfo.ClientUrls[0]
// 	} else {
// 		log.Error(fmt.Sprintf("Http get response code get %d , not %d", leaderResp.StatusCode, http.StatusOK))
// 		return ""
// 	}
// 	return pdleader
// }

func getPDMembers(pdaddr string) []string {
	var pdleader string
	var memberUrls []string
	// get members info
	membersResp, err := http.Get(fmt.Sprintf("http://%s%s", pdaddr, PDAPI_PDMEMBERS))
	if err != nil {
		log.Error(fmt.Sprintf("Http GET request %s failed. Error:%v", fmt.Sprintf("http://%s%s", pdaddr, PDAPI_PDLEADER), err))
		return memberUrls
	}
	defer membersResp.Body.Close()
	membersBody, err := io.ReadAll(membersResp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("io.ReadAll error:%v", err))
		return memberUrls
	}
	membersInfo := pkg.GetMembersResponse{}
	// check response status code
	if membersResp.StatusCode == http.StatusOK {

		err = json.Unmarshal([]byte(string(membersBody)), &membersInfo)
		if err != nil {
			log.Error(fmt.Sprintf("json unmarshal failed. Error :%v", err))
			// return ""
		}
		log.Debug(fmt.Sprintf("Get pd members info:%v", membersInfo))
		// A PD cluster in the middle of an election reports no leader, and
		// client-urls can be empty for a member that has not started serving
		// yet. Neither should crash the check.
		if membersInfo.Leader != nil && len(membersInfo.Leader.ClientUrls) > 0 {
			pdleader = membersInfo.Leader.ClientUrls[0]
		}
		for _, m := range membersInfo.Members {
			if m == nil || len(m.ClientUrls) == 0 {
				log.Warn("Skip pd member without client-urls")
				continue
			}
			url := m.ClientUrls[0]
			leader_priority := m.LeaderPriority
			memurl := fmt.Sprintf("%s(priority:%3d)", url, leader_priority)
			if strings.EqualFold(url, pdleader) {
				memurl = memurl + " | Leader"
			}
			memberUrls = append(memberUrls, memurl)
		}

	} else {
		log.Error(fmt.Sprintf("Http get response code get %d , not %d", membersResp.StatusCode, http.StatusOK))
	}

	sort.Strings(memberUrls)
	return memberUrls
}

// storeLabelMap expands a store's labels into a lookup map. Every configured
// location label is pre-filled with an empty value, so a store missing that
// label still renders an empty cell instead of shifting the columns.
func storeLabelMap(store *pkg.StoreInfo, locationLabels []string) map[string]string {
	mmap := make(map[string]string, len(locationLabels))
	for _, val := range locationLabels {
		mmap[val] = ""
	}
	if store == nil || store.Store == nil {
		return mmap
	}
	for _, lab := range store.Store.Labels {
		if lab == nil {
			continue
		}
		mmap[lab.Key] = lab.Value
	}
	return mmap
}

// matchStoreRole returns the role of the last placement rule matching the
// store. Several rules can match the same store; PD applies them by rule index
// while this keeps the historical "last match wins" behaviour.
// TODO: switch to "smallest (group, index) wins" after verifying the output
// against a real dr-auto-sync cluster.
func matchStoreRole(store *pkg.StoreInfo, rules []pkg.Rule) pkg.PeerRoleType {
	var role pkg.PeerRoleType
	if store == nil || store.Store == nil {
		return role
	}
	for _, rule := range rules {
		for _, constraint := range rule.LabelConstraints {
			if constraint.MatchStore(store) {
				role = rule.Role
			}
		}
	}
	return role
}

// buildStoreRow renders one table row: the location label columns first, then
// instance, role, region count and leader count. The number of leading label
// columns always follows locationLabels and is never hard coded.
func buildStoreRow(store *pkg.StoreInfo, locationLabels []string, role pkg.PeerRoleType) table.Row {
	mmap := storeLabelMap(store, locationLabels)

	row := make(table.Row, 0, len(locationLabels)+4)
	for _, val := range locationLabels {
		row = append(row, mmap[val])
	}

	var (
		address     string
		regionCount int
		leaderCount int
	)
	if store != nil && store.Store != nil {
		address = store.Store.Address
	}
	if store != nil && store.Status != nil {
		regionCount = store.Status.RegionCount
		leaderCount = store.Status.LeaderCount
	}

	return append(row, address, role, regionCount, leaderCount)
}

// sortRowsByLabels sorts rows by the leading location label columns with a
// stable sort, so stores stay grouped by location in the rendered table.
// Every label column takes part in the comparison, including the first one, and
// the column count follows the cluster's actual location-labels.
// Comparison goes through fmt.Sprint on purpose: a row mixes string,
// pkg.PeerRoleType and int values, and the previous implementation asserted the
// compared column to string, which panicked as soon as a location label column
// was absent (index 1 then lands on the Role column).
func sortRowsByLabels(rows []table.Row, labelCount int) {
	sort.SliceStable(rows, func(i, j int) bool {
		n := labelCount
		if n > len(rows[i]) {
			n = len(rows[i])
		}
		if n > len(rows[j]) {
			n = len(rows[j])
		}
		for k := 0; k < n; k++ {
			av, bv := fmt.Sprint(rows[i][k]), fmt.Sprint(rows[j][k])
			if av != bv {
				return av < bv
			}
		}
		return false
	})
}

// buildStoreRows renders and sorts every store row, skipping TiFlash nodes
// because dr-auto-sync only tracks TiKV replicas.
func buildStoreRows(stores []*pkg.StoreInfo, locationLabels []string, rules []pkg.Rule) []table.Row {
	rows := make([]table.Row, 0, len(stores))
	for _, store := range stores {
		if store == nil || store.Store == nil {
			continue
		}
		mmap := storeLabelMap(store, locationLabels)
		if mmap["engine"] == "tiflash" {
			log.Debug(fmt.Sprintf("Skip tiflash store:%s", store.Store.Address))
			continue
		}
		log.Debug(fmt.Sprintf("Get store infomation:%s", store.Store.Address))
		log.Debug(fmt.Sprintf("store label info:%v", mmap))
		rows = append(rows, buildStoreRow(store, locationLabels, matchStoreRole(store, rules)))
	}
	sortRowsByLabels(rows, len(locationLabels))
	return rows
}
