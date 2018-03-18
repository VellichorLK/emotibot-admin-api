package Dictionary

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func addWordbankDir(appid string, paths []string) error {
	wb := &WordBank{}
	addWordbank(appid, paths, wb)
	return nil
}

func addWordbank(appid string, paths []string, wordbank *WordBank) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	vals := make([]interface{}, len(paths))
	for idx := range paths {
		vals[idx] = paths[idx]
	}

	vals = append(vals, wordbank.Name, wordbank.SimilarWords, wordbank.Answer)

	sqlStr := fmt.Sprintf(`INSERT INTO %s_entity
		(level1, level2, level3, level4, status_flag, entity_name, similar_words, answer)
		VALUES (?, ?, ?, ?, 1, ?, ?, ?)`, appid)

	_, err := mySQL.Exec(sqlStr, vals...)
	if err != nil {
		return err
	}
	return nil
}

func checkDirExist(appid string, paths []string) (bool, error) {
	return checkWordbankExist(appid, paths, "")
}

func checkWordbankExist(appid string, paths []string, name string) (bool, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return false, errors.New("DB not init")
	}

	conditions := []string{}
	vals := []interface{}{}
	for idx := range paths {
		if paths[idx] == "" {
			break
		}
		conditions = append(conditions, fmt.Sprintf("level%d = ?", idx+1))
		vals = append(vals, paths[idx])
	}

	if name != "" {
		// conditionStr += "entity_name = ?"
		conditions = append(conditions, "entity_name = ?")
		vals = append(vals, name)
	}

	sqlStr := fmt.Sprintf(`SELECT COUNT(*) from %s_entity WHERE %s`, appid, strings.Join(conditions, " and "))
	rows := mySQL.QueryRow(sqlStr, vals...)
	count := 0
	err := rows.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// getProcessStatus will get status of latest wordbank process
func getProcessStatus(appid string) (string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return "", errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT status from process_status where app_id = ? and module = 'wordbank' order by id desc limit 1", appid)
	if err != nil {
		return "", err
	}
	defer rows.Close()

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

// getFullProcessStatus will get more status info from latest wordbank process
func getFullProcessStatus(appid string) (*StatusInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT status, UNIX_TIMESTAMP(start_at), message from process_status where app_id = ? and module = 'wordbank' order by id desc limit 1", appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

// getLastTwoSuccess will return last two record which status is success, order by time
func getLastTwoSuccess(appid string) ([]*DownloadMeta, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT UNIX_TIMESTAMP(start_at),entity_file_name from process_status where app_id = ? and module = 'wordbank' and status = 'success' order by start_at desc limit 2", appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

// insertProcess will create a file record into process_status, which status is running
func insertProcess(appid string, status Status, filename string, message string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	_, err := mySQL.Exec("insert process_status(app_id, module, status, entity_file_name, message) values (?, 'wordbank', ?, ?, ?)", appid, status, filename, message)
	if err != nil {
		return err
	}

	return nil
}

func getEntities(appid string) ([]*WordBank, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT level1, level2, level3, level4, entity_name, similar_words, answer from %s_entity where status_flag = 1", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cache := make(map[int](map[string]*WordBank))
	for idx := 0; idx < 4; idx++ {
		cache[idx] = make(map[string]*WordBank)
	}
	for rows.Next() {
		categories := make([]sql.NullString, 4)
		var entityName sql.NullString
		var similarWord sql.NullString
		var answer sql.NullString

		args := make([]interface{}, 7)
		for i, category := range categories {
			args[i] = &category
		}
		args[4] = &entityName
		args[5] = &similarWord
		args[6] = &answer

		if err := rows.Scan(&categories[0], &categories[1], &categories[2], &categories[3], &entityName, &similarWord, &answer); err != nil {
			//if err := rows.Scan(args...); err != nil {
			return nil, err
		}

		var lastCategory *WordBank
		for idx, category := range categories {
			if !category.Valid || category.String == "" {
				break
			}

			if _, ok := cache[idx][category.String]; !ok {
				newWordBank := &WordBank{category.String, 0, make([]*WordBank, 0), "", ""}
				cache[idx][category.String] = newWordBank
				if lastCategory != nil {
					lastCategory.Children = append(lastCategory.Children, newWordBank)
				}
			}
			lastCategory = cache[idx][category.String]
		}
		if lastCategory == nil {
			util.LogError.Printf("Level 1 should not be empty in wordbank, skip it")
			continue
		}
		if entityName.Valid && entityName.String != "" {
			newWordBank := &WordBank{entityName.String, 1, make([]*WordBank, 0), "", ""}
			if similarWord.Valid && similarWord.String != "" {
				newWordBank.SimilarWords = similarWord.String
			}
			if answer.Valid && answer.String != "" {
				newWordBank.Answer = answer.String
			}
			lastCategory.Children = append(lastCategory.Children, newWordBank)
		}
	}

	ret := []*WordBank{}
	for _, wordbank := range cache[0] {
		ret = append(ret, wordbank)
	}

	return ret, nil
}
