package pkg

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

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

func QueryTable(db *sql.DB, sql string) (*Table, error) {
	var outTable Table
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
		return &outTable, err
	}
	columnTypes, _ := rows.ColumnTypes()
	//fmt.Printf("columnTypes:%v\n", columnTypes)
	var rowParam = make([]interface{}, len(columnTypes)) // 传入到 rows.Scan 的参数 数组
	var rowValue = make([]interface{}, len(columnTypes)) // 接收数据一行列的数组

	var columnHeader []string
	for i, colType := range columnTypes {
		columnHeader = append(columnHeader, strings.ToUpper(colType.Name()))
		rowValue[i] = reflect.New(colType.ScanType())           // 跟据数据库参数类型，创建默认值 和类型
		rowParam[i] = reflect.ValueOf(&rowValue[i]).Interface() // 跟据接收的数据的类型反射出值的地址
	}
	var list []map[string]string
	outTable.ColumnHeader = columnHeader
	outTable.SQLStr = sql
	for rows.Next() {
		_ = rows.Scan(rowParam...)
		item := make(map[string]string)
		var valueList []string
		for i, colType := range columnTypes {
			//fmt.Printf("colType:%+v colscanType:%s\n", colType, colType.ScanType().String())
			typeName := colType.DatabaseTypeName()
			if rowValue[i] == nil {
				item[colType.Name()] = NILVALUE
				valueList = append(valueList, NILVALUE)
			} else {
				switch typeName {
				case "VARCHAR", "CHAR", "DATETIME", "TIMESTAMP", "INT":
					//item[colType.Name()] = reflect.ValueOf(rowValue[i]).String()
					item[colType.Name()] = string(rowValue[i].([]byte))
				// case "FLOAT":
				// 	item[colType.Name()], _ = strconv.ParseFloat(string(rowValue[i].([]byte)), 64)
				// 	// item[colType.Name()], _ = rowValue[i].(float64)
				default:
					item[colType.Name()] = string(rowValue[i].([]byte))
				}
				valueList = append(valueList, item[colType.Name()])
			}
		}
		list = append(list, item)
		outTable.AddRecord(valueList)
	}
	rows.Close()
	return &outTable, nil
}
