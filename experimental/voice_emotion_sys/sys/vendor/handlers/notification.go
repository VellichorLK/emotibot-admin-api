package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/consul/api"
)

//UpdateAlertConfig update the configuration of alert system
func UpdateAlertConfig(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method == "POST" {
		config := &AlertConfig{}
		err := json.NewDecoder(r.Body).Decode(config)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		} else {

			if err := configCheck(config); err != nil {
				http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
				return
			}

			tx, err := db.Begin()
			if err != nil {
				log.Printf("Error getting transaction from db:%s\n", err)
				http.Error(w, "Internal server error ", http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()

			if config.Threshold != nil {
				err = updateThreshold(tx, *config.Threshold, appid)
				if err != nil {
					log.Printf("Error update threshold:%s\n", err)
					http.Error(w, "Internal server error ", http.StatusInternalServerError)
					return
				}
			}

			if config.EmailList != nil {
				err = updateMailList(tx, config.EmailList, appid)
				if err != nil {
					log.Printf("Error update mailList:%s\n", err)
					http.Error(w, "Internal server error ", http.StatusInternalServerError)
					return
				}
			}

			enableStr, err := getAlertSysStatus(appid)
			if err != nil {
				log.Printf("Error getting status from consul:%s\n", err)
				http.Error(w, "Internal server error ", http.StatusInternalServerError)
				return
			}

			enable := 0

			if enableStr != "" {
				enable, err = strconv.Atoi(enableStr)
				if err != nil {
					log.Printf("Error getting sys status from consul, has wrong value:%s(%s)\n", err, enableStr)
					http.Error(w, "Internal server error ", http.StatusInternalServerError)
					return
				}
			}

			if config.Enable != nil && enable != *config.Enable {
				err = activateAlertSys(*config.Enable, appid)
				if err != nil {
					log.Printf("Error put status to consul:%s\n", err)
					http.Error(w, "Internal server error ", http.StatusInternalServerError)
					return
				}
			}

			tx.Commit()

		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func configCheck(config *AlertConfig) error {
	if config == nil {
		return errors.New("empty config")
	}

	if config.Enable != nil && *config.Enable != 1 && *config.Enable != 0 {
		return errors.New("Wrong value of enable. Must be 0 or 1")
	}

	if config.Threshold != nil && (*config.Threshold <= 0 || *config.Threshold >= 100) {
		return errors.New("threshold out of range, 0 < threshold < 100")
	}

	for _, email := range config.EmailList {
		if !ValidateEmail(email) {
			return errors.New("Bad format of email " + email)
		}
	}

	return nil
}

func getAlertSysStatus(appid string) (string, error) {
	consulClient := GetConsulClient()
	consulKey := ConsulAlertKey + "/" + appid
	kv := consulClient.KV()

	// Lookup the pair
	pair, _, err := kv.Get(consulKey, nil)
	if err != nil {
		return "", err
	}
	if pair != nil {
		return string(pair.Value), nil
	}
	return "", nil
}

func activateAlertSys(enable int, appid string) error {
	consulClient := GetConsulClient()

	consulKey := ConsulAlertKey + "/" + appid
	kv := consulClient.KV()
	val := strconv.Itoa(enable)

	p := &api.KVPair{Key: consulKey, Value: []byte(val)}
	_, err := kv.Put(p, nil)
	return err
}

func updateThreshold(tx *sql.Tx, threshold int, appid string) error {
	querySQL := fmt.Sprintf("select %s from %s where %s=?", NID, ThresholdTable, NAPPID)
	rows, err := db.Query(querySQL, appid)
	if err != nil {
		return err
	}
	defer rows.Close()

	var updateSQL string
	var id uint64
	ids := make([]uint64, 0)

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		updateSQL = fmt.Sprintf("insert into %s (%s,%s,%s,%s) values (%d,%d,%d,'%s'),(%d,%d,%d,'%s')",
			ThresholdTable, NCHANNEL, NType, NSCORE, NAPPID,
			1, 1, threshold, appid,
			2, 1, threshold, appid)
	} else {
		updateSQL = fmt.Sprintf("update %s set %s=%d where %s in (", ThresholdTable, NSCORE, threshold, NID)
		for idx, val := range ids {
			if idx != 0 {
				updateSQL += ","
			}
			updateSQL += strconv.FormatUint(val, 10)
		}
		updateSQL += ")"
	}

	_, err = tx.Exec(updateSQL)

	return err

}

func updateMailList(tx *sql.Tx, mailList []string, appid string) error {

	emailIDMap, err := queryMailList(appid)
	if err != nil {
		return err
	}

	addMailList := make([]string, 0)
	deleteMailIDs := make([]uint64, 0)

	for _, email := range mailList {
		if _, ok := emailIDMap[email]; ok {
			delete(emailIDMap, email)
		} else {
			addMailList = append(addMailList, email)
		}
	}

	for _, id := range emailIDMap {
		deleteMailIDs = append(deleteMailIDs, id)
	}

	err = addMail(tx, addMailList, appid)
	if err != nil {
		return err
	}

	return deleteMail(tx, deleteMailIDs)

}

func addMail(tx *sql.Tx, mailList []string, appid string) error {
	insertSQL := fmt.Sprintf("insert into %s (%s,%s) values (?,?)", EmailTable, NAPPID, NEmail)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, mail := range mailList {
		_, err = stmt.Exec(appid, mail)
		if err != nil {
			break
		}
	}

	return err
}

func deleteMail(tx *sql.Tx, ids []uint64) error {
	deleteSQL := fmt.Sprintf("delete from %s where %s=?", EmailTable, NID)
	stmt, err := tx.Prepare(deleteSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, id := range ids {
		_, err = stmt.Exec(id)
		if err != nil {
			break
		}
	}
	return err
}

func queryMailList(appid string) (map[string]uint64, error) {
	querySQL := fmt.Sprintf("select %s,%s from %s where %s=?", NID, NEmail, EmailTable, NAPPID)
	rows, err := db.Query(querySQL, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id uint64
	var email string

	emailIDMap := make(map[string]uint64)

	for rows.Next() {
		err = rows.Scan(&id, &email)
		if err != nil {
			break
		}
		emailIDMap[email] = id
	}

	return emailIDMap, err
}
