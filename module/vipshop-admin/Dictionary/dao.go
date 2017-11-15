package Dictionary

import (
	"errors"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

// GetProcessStatus will get status of latest wordbank process
func GetProcessStatus(appid string) (string, error) {
	mySQL := util.GetMainDB()
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

// GetFullProcessStatus will get more status info from latest wordbank process
func GetFullProcessStatus(appid string) (*StatusInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT status, UNIX_TIMESTAMP(start_at), message from process_status where app_id = ? and module = 'wordbank' order by id desc limit 1", appid)
	if err != nil {
		return nil, err
	}

	status := StatusInfo{}
	ret := rows.Next()
	if !ret {
		return nil, nil
	}

	var timestamp int64
	if err := rows.Scan(&status.Status, &timestamp, &status.Message); err != nil {
		return nil, err
	}
	status.StartTime = time.Unix(timestamp, 0)

	emptyMsg := ""
	if status.Message == nil {
		status.Message = &emptyMsg
	}
	return &status, nil
}

// GetLastTwoSuccess will return last two record which status is success, order by time
func GetLastTwoSuccess(appid string) ([]*DownloadMeta, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT UNIX_TIMESTAMP(start_at),entity_file_name from process_status where app_id = ? and module = 'wordbank' and status = 'success' order by start_at desc limit 2", appid)
	if err != nil {
		return nil, err
	}

	ret := []*DownloadMeta{}
	for rows.Next() {
		var meta DownloadMeta
		var startTime int64
		if err := rows.Scan(&startTime, &meta.UploadFile); err != nil {
			return nil, err
		}

		meta.UploadTime = time.Unix(startTime, 0)

		ret = append(ret, &meta)
	}

	return ret, nil
}

// InsertNewProcess will create a file record into process_status, which status is running
func InsertProcess(appid string, status Status, filename string, message string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	_, err := mySQL.Query("insert process_status(app_id, module, status, entity_file_name, message) values (?, 'wordbank', ?, ?, ?)", appid, status, filename, message)
	if err != nil {
		return err
	}

	return nil
}
