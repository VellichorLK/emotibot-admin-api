package imagesManager

import (
	"database/sql"
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/go-sql-driver/mysql"
)

var mainDB *sql.DB

//Save copy image struct into SQL Database, return nil if success
func Save(image Image) (int64, error) {
	if image.Location == "" {
		return 0, fmt.Errorf("image location should not be empty")
	}
	if image.ID == 0 {
		return 0, fmt.Errorf("image ID should not be zero")
	}
	if image.FileName == "" {
		return 0, fmt.Errorf("image FileName should not be empty")
	}

	result, err := mainDB.Exec("INSERT INTO (id, filename, location, createdTime, lastmodified) VALUES (?, ?, ?, ?, ?)", image.ID, image.FileName, image.Location, image.CreatedTime, image.LastModifiedTime)
	if err != nil {
		return 0, fmt.Errorf("sql exec failed, %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return id, fmt.Errorf("get insert id failed, %v", err)
	}

	return id, nil

}

func getLocationID(location string) (uint64, error) {
	sql := "insert into " + locationTable + "(" + attrLocation + ")" + " values(?)"
	res, err := SqlExec(sql, location)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok { // Now the error number is accessible directly
			if driverErr.Number == ErDupEntry {
				return getExistLocationID(location)
			}
			return 0, err
		}
	}
	id, err := res.LastInsertId()
	return uint64(id), err

}

func getExistLocationID(location string) (uint64, error) {
	sql := "select " + attrID + " from " + locationTable + " where " + attrLocation + "=?"
	rows, err := SqlQuery(sql, location)
	if err != nil {
		return 0, nil
	}
	defer rows.Close()

	var id uint64
	if rows.Next() {
		err = rows.Scan(&id)
	}
	return id, err
}

func SqlQuery(sql string, params ...interface{}) (*sql.Rows, error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}
	return db.Query(sql, params...)
}

func GetTx() (*sql.Tx, error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}
	return db.Begin()
}

func SqlExec(sql string, params ...interface{}) (sql.Result, error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}

	stmt, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	return ExecStmt(stmt, params...)
}

func ExecStmt(stmt *sql.Stmt, params ...interface{}) (sql.Result, error) {
	return stmt.Exec(params...)
}
