package pkg

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const NILVALUE = "NULL"

type Table struct {
	SQLStr       string
	ColumnHeader []string
	RecordList   [][]string
}

func (t *Table) AddRecord(slist []string) {
	t.RecordList = append(t.RecordList, slist)
}

func (t *Table) String() string {
	resStr := fmt.Sprintf("%s\n%s", t.SQLStr, strings.Join(t.ColumnHeader, "\t"))
	for _, rlist := range t.RecordList {
		resStr = fmt.Sprintf("%s\n%s", resStr, strings.Join(rlist, "\t"))
	}
	return resStr
}

func (t *Table) ResultString() string {
	resStr := fmt.Sprintf("%s", strings.Join(t.ColumnHeader, "\t"))
	for _, rlist := range t.RecordList {
		resStr = fmt.Sprintf("%s\n%s", resStr, strings.Join(rlist, "\t"))
	}
	return resStr
}

// QueryTable runs an arbitrary SQL statement and renders every cell as a
// string, so the check reports can print results of queries they know nothing
// about.
func QueryTable(db *sql.DB, query string) (*Table, error) {
	var outTable Table
	outTable.SQLStr = query

	if db == nil {
		return &outTable, fmt.Errorf("database handle is nil")
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Error(err)
		return &outTable, err
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.Error(err)
		return &outTable, fmt.Errorf("get column types of %q failed: %w", query, err)
	}
	header := make([]string, 0, len(columnTypes))
	for _, colType := range columnTypes {
		header = append(header, strings.ToUpper(colType.Name()))
	}
	outTable.ColumnHeader = header

	// Scan into *any: the driver then hands back its own value ([]byte for text
	// columns, int64 / float64 / time.Time for numbers and timestamps, nil for
	// SQL NULL) and cellToString takes care of the formatting.
	values := make([]any, len(columnTypes))
	dest := make([]any, len(columnTypes))
	for i := range values {
		dest[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			log.Error(err)
			return &outTable, fmt.Errorf("scan row of %q failed: %w", query, err)
		}
		record := make([]string, len(columnTypes))
		for i := range values {
			record[i] = cellToString(values[i])
		}
		outTable.AddRecord(record)
	}
	// A half-read result set must not be reported as a successful check: the
	// caller has no other way to tell a short answer from a complete one.
	if err := rows.Err(); err != nil {
		log.Error(err)
		return &outTable, fmt.Errorf("read rows of %q failed: %w", query, err)
	}
	return &outTable, nil
}

// cellToString renders one scanned cell as the string the reports show.
//
// The driver returns []byte for text columns and int64 / float64 / bool /
// time.Time for columns of those types, so the previous unconditional
// `rowValue.([]byte)` assertion panicked with
// "interface conversion: interface {} is int64, not []uint8" as soon as a
// check query returned a number (for example SELECT COUNT(*)). Unknown types
// fall back to fmt.Sprint rather than panicking.
func cellToString(v any) string {
	switch val := v.(type) {
	case nil:
		return NILVALUE
	case []byte:
		return string(val)
	case string:
		return val
	case int64:
		return strconv.FormatInt(val, 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case time.Time:
		// Only reached when the DSN enables parseTime; the MySQL textual form
		// keeps the output identical to the default ([]byte) path.
		return val.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprint(val)
	}
}
