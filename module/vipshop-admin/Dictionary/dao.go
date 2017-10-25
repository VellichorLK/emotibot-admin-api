package Dictionary

import (
	"database/sql"
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/util"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mySQL *sql.DB
)

const (
	mySQLTimeout      string = "10s"
	mySQLWriteTimeout string = "30s"
	mySQLReadTimeout  string = "30s"
)

// DaoInit should be called before all db operation
func DaoInit(mysqlURL string, mysqlUser string, mysqlPass string, mysqlDB string) error {
	if len(mysqlURL) == 0 || len(mysqlUser) == 0 || len(mysqlPass) == 0 || len(mysqlDB) == 0 {
		return errors.New("invalid parameters")
	}
	util.LogInfo.Printf("mysqlURL: %s, mysqlUser: %s, mysqlPass: %s, mysql_name: %s", mysqlURL, mysqlUser, mysqlPass, mysqlDB)

	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=true", mysqlUser, mysqlPass, mysqlURL, mysqlDB, mySQLTimeout, mySQLReadTimeout, mySQLWriteTimeout)
	util.LogInfo.Printf("url: %s", url)

	var err error
	mySQL, err = sql.Open("mysql", url)
	if err != nil {
		util.LogInfo.Printf("open db(%s) failed: %s", url, err)
		return err
	}
	return nil
}

// GetProcessStatus will get status of latest wordbank process
func GetProcessStatus(appid string) (string, error) {
	if mySQL == nil {
		return "", errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT status from process_status where app_id = ? and module = 'wordbank' order by id desc limit 1", appid)
	if err != nil {
		return "", err
	}

	var status string
	ret := rows.Next()
	if !ret {
		return "", nil
	}
	if err := rows.Scan(&status); err != nil {
		return "", err
	}

	return status, nil
}

// GetLastTwoSuccess will return last two record which status is success, order by time
func GetLastTwoSuccess(appid string) ([]*DownloadMeta, error) {
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	// rows, err := mySQL.Query("SELECT start_at,entity_file_name from process_status where app_id = ? and module = 'wordbank' and status = 'success' order by start_at desc limit 2", appid)
	rows, err := mySQL.Query("SELECT start_at,entity_file_name from process_status")
	if err != nil {
		return nil, err
	}

	ret := []*DownloadMeta{}
	for rows.Next() {
		var meta DownloadMeta
		if err := rows.Scan(&meta.UploadTime, &meta.UploadFile); err != nil {
			return nil, err
		}
		ret = append(ret, &meta)
	}

	return ret, nil
}

// InsertNewProcess will create a file record into process_status, which status is running
func InsertProcess(appid string, status Status, filename string, message string) error {
	if mySQL == nil {
		return errors.New("DB not init")
	}

	_, err := mySQL.Query("insert process_status(app_id, module, status, entity_file_name, message) values (?, 'wordbank', ?, ?, ?)", appid, status, filename, message)
	if err != nil {
		return err
	}

	return nil
}
