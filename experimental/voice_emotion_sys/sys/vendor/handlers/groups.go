package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-sql-driver/mysql"
)

//GroupSet group data
type GroupSet struct {
	ID        uint64   `json:"id"`
	GroupName string   `json:"group_name"`
	GroupVal  []string `json:"seat_id"`
}

//GroupAPI the entrypoint of group api
func GroupAPI(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		groups, err := getGroups(appid, nil)

		if err != nil {
			log.Printf("Error: %s\n", err)
			http.Error(w, "Internal server error ", http.StatusInternalServerError)
		} else {
			resp, err := json.Marshal(groups)
			if err != nil {
				http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			contentType := "application/json; charset=utf-8"

			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	case "PATCH":
		var groups []*GroupSet
		err := json.NewDecoder(r.Body).Decode(&groups)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		} else {
			err = updateGroup(appid, groups)
			if err != nil {
				if code, ok := mysqlErrCode(err); ok && code == 1062 {
					http.Error(w, "Bad Request: duplicate entry ", http.StatusBadRequest)
				} else {
					log.Printf("Error: %s\n", err)
					http.Error(w, "Internal server error ", http.StatusInternalServerError)
				}

			}
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func getGroups(appid string, groupID []uint64) ([]*GroupSet, error) {
	querySQL := fmt.Sprintf("select a.%s,a.%s,b.%s from %s as a left join %s as b on a.%s=b.%s where a.%s=?",
		NID, GroupName, GroupVal,
		GroupTable, GroupValTable,
		NID, GroupID, NAPPID)

	params := make([]interface{}, 0)
	params = append(params, appid)
	if groupID != nil && len(groupID) > 0 {
		querySQL += " and a." + NID + " in (?" + strings.Repeat(",?", len(groupID)-1) + ")"
		for _, id := range groupID {
			params = append(params, id)
		}
	}

	querySQL += " order by " + NID

	rows, err := db.Query(querySQL, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]*GroupSet, 0)
	groupsMap := make(map[uint64]*GroupSet)

	var id uint64
	var name, val string

	for rows.Next() {
		err = rows.Scan(&id, &name, &val)
		if err != nil {
			break
		}

		var group *GroupSet
		var ok bool
		if group, ok = groupsMap[id]; ok {
			group.GroupVal = append(group.GroupVal, val)
		} else {
			group = &GroupSet{ID: id, GroupName: name}
			group.GroupVal = make([]string, 1, 1)
			group.GroupVal[0] = val
			groups = append(groups, group)
			groupsMap[id] = group
		}

	}

	return groups, err

}

//UpdateGroup update the group setting by appid, currently just delete all and add all
func updateGroup(appid string, groups []*GroupSet) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = deleteGroup(tx, appid)
	if err != nil {
		return err
	}

	err = addGroup(tx, appid, groups)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func addGroup(tx *sql.Tx, appid string, groups []*GroupSet) error {
	var err error
	if len(groups) > 0 {
		err = addGroupName(tx, appid, groups)
		if err != nil {
			return err
		}
		err = addGroupVal(tx, groups)
	}
	return err
}

func addGroupName(tx *sql.Tx, appid string, groups []*GroupSet) error {
	insertSQL := fmt.Sprintf("insert into %s (%s,%s) values(?,?)", GroupTable, GroupName, NAPPID)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var res sql.Result
	for idx, group := range groups {
		res, err = stmt.Exec(group.GroupName, appid)
		if err != nil {
			break
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil
		}
		groups[idx].ID = uint64(id)
	}
	return err
}
func addGroupVal(tx *sql.Tx, groups []*GroupSet) error {

	insertSQL := fmt.Sprintf("insert into %s (%s,%s) values(?,?)", GroupValTable, GroupID, GroupVal)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, group := range groups {
		for _, val := range group.GroupVal {
			_, err = stmt.Exec(group.ID, val)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteGroup(tx *sql.Tx, appid string) error {
	ids, err := getGroupID(appid)
	if err != nil {
		return err
	}
	if len(ids) > 0 {
		_, err = deleteGroupByID(tx, ids)
		if err != nil {
			return err
		}
		_, err = deleteGroupValByGroupID(tx, ids)
	}

	return err
}
func getGroupID(appid string) ([]interface{}, error) {
	querySQL := fmt.Sprintf("select %s from %s where %s=?", NID, GroupTable, NAPPID)
	rows, err := db.Query(querySQL, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var id uint64
	ids := make([]interface{}, 0, 5)
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			break
		}
		ids = append(ids, id)
	}

	return ids, err

}
func deleteGroupByID(tx *sql.Tx, ids []interface{}) (int64, error) {
	return deleteRecordByID(tx, GroupTable, NID, ids)
}
func deleteGroupValByGroupID(tx *sql.Tx, ids []interface{}) (int64, error) {
	return deleteRecordByID(tx, GroupValTable, GroupID, ids)
}
func deleteRecordByID(tx *sql.Tx, tableName string, idName string, ids []interface{}) (int64, error) {
	num := len(ids)
	if num > 0 {
		deleteSQL := fmt.Sprintf("delete from %s where %s in (?%s)", tableName, idName, strings.Repeat(",?", num-1))
		res, err := tx.Exec(deleteSQL, ids...)
		if err != nil {
			return 0, nil
		}
		return res.RowsAffected()
	}
	return 0, nil
}

func mysqlErrCode(err error) (uint16, bool) {
	val, ok := err.(*mysql.MySQLError)
	if !ok {
		return 0, false
	}
	return val.Number, true
}
