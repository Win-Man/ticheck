/*
 * Created: 2022-01-26 15:52:52
 * Author : Win-Man
 * Email : gang.shen0423@gmail.com
 * -----
 * Last Modified:
 * Modified By:
 * -----
 * Description:
 */

package database

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/Win-Man/ticheck/config"
	"github.com/Win-Man/ticheck/pkg"
	_ "github.com/go-sql-driver/mysql"
)

// testDBConfig returns the TiDB connection used by the integration tests below.
// Both tests are skipped unless TICHECK_TEST_TIDB is set to "host:port", so
// `go test ./...` stays green on a machine without a TiDB cluster.
//
// Optional overrides: TICHECK_TEST_TIDB_USER (default root),
// TICHECK_TEST_TIDB_PASSWORD (default empty), TICHECK_TEST_TIDB_DATABASE
// (default test).
func testDBConfig(t *testing.T) *config.DBConfig {
	t.Helper()

	addr := os.Getenv("TICHECK_TEST_TIDB")
	if addr == "" {
		t.Skip("skipping integration test: set TICHECK_TEST_TIDB=host:port to enable")
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("invalid TICHECK_TEST_TIDB %q: %v", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("invalid port in TICHECK_TEST_TIDB %q: %v", addr, err)
	}

	user := os.Getenv("TICHECK_TEST_TIDB_USER")
	if user == "" {
		user = "root"
	}
	database := os.Getenv("TICHECK_TEST_TIDB_DATABASE")
	if database == "" {
		database = "test"
	}

	return &config.DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: os.Getenv("TICHECK_TEST_TIDB_PASSWORD"),
		Database: database,
	}
}

func TestMySQLConn(t *testing.T) {
	dbConfig := testDBConfig(t)

	db, err := OpenMySQLDB(dbConfig)
	if err != nil {
		t.Fatalf("connect mysql db error:%v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("ping mysql db error:%v", err)
	}
}

func TestMySQLSelect(t *testing.T) {
	dbConfig := testDBConfig(t)

	db, err := OpenMySQLDB(dbConfig)
	if err != nil {
		t.Fatalf("connect mysql db error:%v", err)
	}
	defer db.Close()

	sqlStr := fmt.Sprintf("select * from %s.t", dbConfig.Database)

	mysqlTable, err := pkg.QueryTable(db, sqlStr)
	if err != nil {
		t.Fatalf("execute sql error during query:%v", err)
	}

	t.Logf("%s", mysqlTable)
}
