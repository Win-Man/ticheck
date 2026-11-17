package pkg

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

// --- minimal fake driver, so QueryTable can be exercised without a real MySQL ---

type fakeResult struct {
	cols []string
	vals [][]driver.Value
	err  error // returned by Next once all rows are consumed
}

type fakeDriver struct{ res fakeResult }

func (d fakeDriver) Open(string) (driver.Conn, error) { return &fakeConn{res: d.res}, nil }

type fakeConnector struct{ res fakeResult }

func (c fakeConnector) Connect(context.Context) (driver.Conn, error) {
	return &fakeConn{res: c.res}, nil
}

func (c fakeConnector) Driver() driver.Driver { return fakeDriver{res: c.res} }

type fakeConn struct{ res fakeResult }

func (c *fakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported by the fake driver")
}

func (c *fakeConn) Close() error { return nil }

func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, errors.New("begin is not supported by the fake driver")
}

func (c *fakeConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if c.res.cols == nil {
		return nil, errors.New("query failed")
	}
	return &fakeRows{res: c.res}, nil
}

type fakeRows struct {
	res fakeResult
	idx int
}

func (r *fakeRows) Columns() []string { return r.res.cols }

func (r *fakeRows) Close() error { return nil }

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.res.vals) {
		if r.res.err != nil {
			return r.res.err
		}
		return io.EOF
	}
	copy(dest, r.res.vals[r.idx])
	r.idx++
	return nil
}

func openFakeDB(t *testing.T, res fakeResult) *sql.DB {
	t.Helper()
	db := sql.OpenDB(fakeConnector{res: res})
	t.Cleanup(func() { db.Close() })
	return db
}

// --- QueryTable ---

// TestQueryTableMixedColumnTypes is the regression test for the unconditional
// `rowValue.([]byte)` assertion: the moment a column comes back as int64,
// float64 or time.Time the old code panicked with
// "interface conversion: interface {} is int64, not []uint8".
func TestQueryTableMixedColumnTypes(t *testing.T) {
	db := openFakeDB(t, fakeResult{
		cols: []string{"name", "cnt", "ratio", "created", "nothing", "flag"},
		vals: [][]driver.Value{{
			[]byte("tidb"),
			int64(3),
			float64(0.5),
			time.Date(2026, 11, 17, 10, 30, 0, 0, time.UTC),
			nil,
			true,
		}},
	})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("QueryTable() panicked on a mixed-type result: %v", r)
		}
	}()

	got, err := QueryTable(db, "select mixed")
	if err != nil {
		t.Fatalf("QueryTable() error: %v", err)
	}

	wantHeader := []string{"NAME", "CNT", "RATIO", "CREATED", "NOTHING", "FLAG"}
	if !reflect.DeepEqual(got.ColumnHeader, wantHeader) {
		t.Fatalf("header mismatch\n got: %v\nwant: %v", got.ColumnHeader, wantHeader)
	}

	wantRow := []string{"tidb", "3", "0.5", "2026-11-17 10:30:00", NILVALUE, "true"}
	if len(got.RecordList) != 1 {
		t.Fatalf("expected 1 row, got %d: %v", len(got.RecordList), got.RecordList)
	}
	if !reflect.DeepEqual(got.RecordList[0], wantRow) {
		t.Fatalf("row mismatch\n got: %v\nwant: %v", got.RecordList[0], wantRow)
	}
}

// TestQueryTableTextProtocolValues pins the behaviour the MySQL driver actually
// produces for `db.Query` without arguments: every column arrives as []byte and
// NULL as nil.
func TestQueryTableTextProtocolValues(t *testing.T) {
	db := openFakeDB(t, fakeResult{
		cols: []string{"variable_name", "value"},
		vals: [][]driver.Value{
			{[]byte("tidb_enable_async_commit"), []byte("ON")},
			{[]byte("max_connections"), nil},
		},
	})

	got, err := QueryTable(db, "show variables")
	if err != nil {
		t.Fatalf("QueryTable() error: %v", err)
	}

	want := [][]string{
		{"tidb_enable_async_commit", "ON"},
		{"max_connections", NILVALUE},
	}
	if !reflect.DeepEqual(got.RecordList, want) {
		t.Fatalf("rows mismatch\n got: %v\nwant: %v", got.RecordList, want)
	}
	if got.SQLStr != "show variables" {
		t.Fatalf("SQLStr not kept, got %q", got.SQLStr)
	}
}

// TestQueryTableRowsErr checks that a result set failing half way is reported as
// an error: a check must not print the rows it managed to read and claim success.
func TestQueryTableRowsErr(t *testing.T) {
	db := openFakeDB(t, fakeResult{
		cols: []string{"k"},
		vals: [][]driver.Value{{[]byte("a")}},
		err:  errors.New("connection reset by peer"),
	})

	got, err := QueryTable(db, "select k")
	if err == nil {
		t.Fatalf("expected an error for a truncated result set, got rows: %v", got.RecordList)
	}
	if !strings.Contains(err.Error(), "connection reset by peer") {
		t.Fatalf("error should wrap the driver cause, got %v", err)
	}
	if got.SQLStr != "select k" {
		t.Fatalf("table should still carry the sql, got %q", got.SQLStr)
	}
}

func TestQueryTableQueryError(t *testing.T) {
	db := openFakeDB(t, fakeResult{})

	got, err := QueryTable(db, "select broken")
	if err == nil {
		t.Fatalf("expected an error, got %v", got)
	}
	if got == nil || got.SQLStr != "select broken" {
		t.Fatalf("error path should still return the table with its sql, got %+v", got)
	}
}

// TestQueryTableNilDB guards the nil pointer that follows a failed
// database.OpenMySQLDB being passed through unchecked.
func TestQueryTableNilDB(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("QueryTable(nil, ...) panicked: %v", r)
		}
	}()

	if _, err := QueryTable(nil, "select 1"); err == nil {
		t.Fatal("expected an error for a nil database handle")
	}
}

// --- cellToString ---

func TestCellToString(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"nil is NULL", nil, NILVALUE},
		{"bytes", []byte("ON"), "ON"},
		{"empty bytes", []byte(""), ""},
		{"string", "OFF", "OFF"},
		{"int64", int64(-42), "-42"},
		{"uint64", uint64(18446744073709551615), "18446744073709551615"},
		{"float64", 1.25, "1.25"},
		{"float64 integral", float64(3), "3"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{
			"time",
			time.Date(2026, 11, 17, 8, 5, 9, 0, time.UTC),
			"2026-11-17 08:05:09",
		},
		{
			// an unknown type must be printed, never asserted
			"unknown type",
			struct{ A int }{A: 7},
			"{7}",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("cellToString(%#v) panicked: %v", c.in, r)
				}
			}()
			if got := cellToString(c.in); got != c.want {
				t.Fatalf("cellToString(%#v) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
