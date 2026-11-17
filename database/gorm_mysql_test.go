package database

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// TestGORMConn is an integration test; testDBConfig skips it unless
// TICHECK_TEST_TIDB is set.
func TestGORMConn(t *testing.T) {
	dbConfig := testDBConfig(t)

	if err := InitDB(*dbConfig); err != nil {
		t.Fatalf("init gorm db error:%v", err)
	}
	if DB == nil {
		t.Fatal("gorm DB handle is nil after InitDB")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		t.Fatalf("get underlying sql.DB error:%v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping gorm db error:%v", err)
	}
}
