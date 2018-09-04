package sqlUtil

import (
	"JLog"
	_ "com.git.mysql"
	"container/list"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"
	"yxtTool"
	"fmt"
)

var (
	userName string
	password string
	host     string
	port     string
)

const DriverName string = "mysql"
const DataSourceName string = "root:123@tcp(10.11.1.220:3306)/zqcw?charset=utf8"

var (
	db  *sql.DB
	err error
)

func init() {
	keys := yxtTool.ReadProperty()
	if keys == nil {
		JLog.PrintSqlError("配置文件有误！！\r\n")
		os.Exit(0)
	}
	db, err = sql.Open(keys["DriverName"], keys["DataSourceName"])
	checkErr(err, "init")
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Minute * 15)
	db.SetMaxIdleConns(5)
	JLog.PrintSqlError("init sqlHanlder sql info")

}

func Insert(sqlStr string, params ...interface{}) int64 {
	sqlStr = parseStmtSql(sqlStr, params...)
	//JLog.WriteClientMsg(sqlStr+"\t","sqlInfo","sql")
	return insert(sqlStr)
}
func Delete(sqlStr string, params ...interface{}) bool {
	return execute(sqlStr, params...)
}
func Update(sqlStr string, params ...interface{}) bool {
	return execute(sqlStr, params...)
}

/**
参数化查询方法
*/
func QueryList(sqlStr string, params ...interface{}) (rs *list.List ,err error) {
	if len(params) < 1 {
		 rs,err=queryListNoParams(sqlStr)
		return
	}
	//jLog.Println(sqlStr)
	//JLog.WriteClientMsg(sqlStr+"\t","sqlInfo","sql")
	rs,err= queryList(sqlStr, params...)
	return
}

/**
删除、更新 表操作
*/
func execute(sqlStr string, params ...interface{}) bool {
	//jLog.Println(sqlStr)
	tx, err := db.Begin()
	defer tx.Commit()
	if checkErr(err, sqlStr) {
		return false
	}
	rs, err := tx.Exec(sqlStr, params...)
	if checkErr(err, sqlStr) {
		return false
	}
	count, err := rs.RowsAffected()
	if checkErr(err, sqlStr) {
		return false
	}
	//fmt.Printf("操作记录数：%d\r\n", count)

	return count > 0
}

/**
增加数据库操作
*/
func insert(sqlStr string, params ...interface{}) int64 {
	//jLog.Println(sqlStr)
	tx, err := db.Begin()
	defer tx.Commit()
	if checkErr(err, sqlStr) {
		return 0
	}
	//_, err = tx.Exec(sqlStr, params...)
	rs, err := tx.Exec(sqlStr)
	if checkErr(err, sqlStr) {
		return 0
	}
	id,err:=rs.LastInsertId()
	if err!=nil{
		fmt.Println(err)
	}else{
		//fmt.Println(id)

		//if checkErr(err,sqlStr) {return false}

		return id
	}

	return 0
}

/**
普通查询方法
*/
func queryListNoParams(sqlStr string) (rs *list.List ,rserr error) {
	//jLog.Println(sqlStr)
	rserr=nil
	tx, err := db.Begin()
	defer tx.Commit()
	rs = list.New()
	var rows *sql.Rows
	if checkErr(err, sqlStr) {
		rserr=err
		return
	}
	rows, err = tx.Query(sqlStr)
	if checkErr(err, sqlStr) {
		rserr=err
		return
	}
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}
	record := make(map[string]string)

	for rows.Next() {
		//将行数据保存到record字典
		err = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			} else {
				record[columns[i]] = ""
			}
		}
		rs.PushBack(record)
		//fmt.Println(record)
	}
	return
}

/**
参数化查询方法
*/
func queryList(sqlStr string, params ...interface{}) (rs *list.List,rserr error) {
	//jLog.Println(sqlStr)
	rs = list.New()
	rserr=nil
	tx, err := db.Begin()
	defer tx.Commit()
	if checkErr(err, sqlStr) {
		rserr=err
		return
	}
	rows, err := tx.Query(sqlStr, params...)
	if checkErr(err, sqlStr) {
		rserr=err
		return
	}

	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	for rows.Next() {
		record := make(map[string]string)
		//将行数据保存到record字典

		err = rows.Scan(scanArgs...)
		for i, col := range values {
			if col != nil {
				switch col.(type) {
				case []byte:
					record[columns[i]] = string(col.([]byte))
				case int64:
					record[columns[i]] = strconv.FormatInt(col.(int64), 10)
				case float64:
					record[columns[i]] = strconv.FormatFloat(col.(float64), 'f', -1, 64)
				}
			} else {
				record[columns[i]] = ""
			}
		}
		rs.PushBack(record)
		//fmt.Println(record)
	}
	return
}

func checkErr(err error, sqlStr string) bool {
	if err != nil {
		JLog.PrintSqlError("ERROR:" + err.Error())
		JLog.PrintSqlError("errSql:" + sqlStr + "\r\n")
		return true
	}
	return false
}
func parseStmtSql(sql string, params ...interface{}) string {
	for i := 0; strings.Contains(sql, "?") && i < len(params); i++ {
		s, ok := params[i].(string)
		if ok {
			sql = strings.Replace(sql, "?", "'"+s+"'", 1)
			continue
		}
		f, ok := params[i].(float64)
		if ok {
			sql = strings.Replace(sql, "?", "'"+strconv.FormatFloat(f, 'f', -1, 64)+"'", 1)
			continue
		}
		par, ok := params[i].(int)
		if ok {
			sql = strings.Replace(sql, "?", "'"+strconv.Itoa(par)+"'", 1)
			continue
		}
		u32, ok := params[i].(uint32)
		if ok {
			sql = strings.Replace(sql, "?", "'"+strconv.Itoa(int(u32))+"'", 1)
			continue
		}
	}
	return sql
}
