package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

//user defined column table
const (
	GetColumSQL     = "select " + NCOLID + "," + NCOLNAME + "," + NDEDAULT + " from " + UsrColTable + " where " + NAPPID + "=?"
	UpdateColValSQL = "update " + UsrColValTable + " set " + NCOLVAL + "=? where " +
		NID + " = ( select " + NID + " from " + MainTable + " where " + NAPPID + "=? and " + NFILEID + " = ?) and " +
		NCOLID + "=?"
	GetSelValueSQL = "select a." + NCOLID + ",a." + NCOLNAME + ",a." + NAPPID + ",a." + NDEDAULT + ",b." + NSELVAL + " from " + UsrColTable + " as a left join " + UsrSelValTable +
		" as b on a." + NCOLID + "=b." + NCOLID + " order by a." + NAPPID
)

type UsrColHandler struct {
	FieldNameMap    sync.Map //col_id  -> field_name
	DefaultValue    sync.Map //appid  -> default value
	SelectableValue sync.Map // col_id  -> selectable map
}

type DefaultValue struct {
	ColID    string
	ColValue string
}

type UsrColumBlock struct {
	ColName    string   `json:"col_name"`
	ColID      string   `json:"col_id"`
	SelValue   []string `json:"selectable_value,omitempty"`
	DefaultVal string   `json:"default_value"`
}

type UsrColumnUpdateBlock struct {
	FileID string `json:"file_id"`
	ColID  string `json:"col_id"`
	Value  string `json:"col_value"`
}

var DefaulUsrField *UsrColHandler

func LoadUsrField() error {
	DefaulUsrField = &UsrColHandler{}

	rows, err := db.Query(GetSelValueSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	colIDMap := make(map[string]bool)

	for rows.Next() {

		var colID, appid, colName, defValue string
		var colValue sql.NullString

		err := rows.Scan(&colID, &colName, &appid, &defValue, &colValue)
		if err != nil {
			return err
		}

		key := colID

		_, ok := colIDMap[key]
		if !ok {
			colIDMap[key] = true

			DefaulUsrField.FieldNameMap.Store(key, colName)

			var dvs []*DefaultValue

			dvsInterface, ok := DefaulUsrField.DefaultValue.Load(appid)
			if !ok {
				dvs = make([]*DefaultValue, 0)
			} else {
				dvs = dvsInterface.([]*DefaultValue)
			}

			dv := &DefaultValue{colID, defValue}
			dvs = append(dvs, dv)

			DefaulUsrField.DefaultValue.Store(appid, dvs)

		}

		if colValue.Valid {

			svInterface, ok := DefaulUsrField.SelectableValue.Load(key)
			if !ok {
				sv := &sync.Map{}
				sv.Store(colValue.String, true)
				DefaulUsrField.SelectableValue.Store(key, sv)
			} else {
				sv := svInterface.(*sync.Map)
				sv.Store(colValue.String, true)

			}

		}

	}

	return rows.Err()

}

//GetUserColumn get the user defined column
func GetUserColumn(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	if DefaulUsrField == nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ucbs := make([]*UsrColumBlock, 0)
	dvsInterface, ok := DefaulUsrField.DefaultValue.Load(appid)

	if ok {
		dvs := dvsInterface.([]*DefaultValue)
		for _, dv := range dvs {
			ucb := &UsrColumBlock{}
			ucb.SelValue = make([]string, 0)

			ucb.ColID = dv.ColID
			ucb.DefaultVal = dv.ColValue
			nameInterface, ok := DefaulUsrField.FieldNameMap.Load(dv.ColID)

			if !ok {
				log.Println("Field name map has no colum id " + dv.ColID)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			name := nameInterface.(string)
			ucb.ColName = name

			selMapInterface, ok := DefaulUsrField.SelectableValue.Load(dv.ColID)
			if ok {
				selMap := selMapInterface.(*sync.Map)

				selMap.Range(func(k, v interface{}) bool {
					ucb.SelValue = append(ucb.SelValue, k.(string))
					return true
				})

			}

			ucbs = append(ucbs, ucb)
		}
	}

	resp, err := json.Marshal(ucbs)
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	contentType := "application/json; charset=utf-8"
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

//UpdateColumnVal update the column value
func UpdateColumnVal(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	if DefaulUsrField == nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ucub := &UsrColumnUpdateBlock{}

	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(ucub)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	defer r.Body.Close()

	//check if its fileID
	if isFileID(ucub.FileID) {
		if ucub.ColID != "" {
			if !checkSelectableVal(ucub.ColID, ucub.Value) {
				http.Error(w, "value("+ucub.Value+") can't be set", http.StatusBadRequest)
				return
			}
			_, err := updateColVal(appid, ucub.FileID, ucub.ColID, ucub.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "No col_id given", http.StatusBadRequest)
			return
		}
	}
}

func checkSelectableVal(colID string, val string) bool {

	selMapInterface, ok := DefaulUsrField.SelectableValue.Load(colID)

	//svs, ok := DefaulUsrField.SelectableValue[colID]
	if ok {
		svs := selMapInterface.(*sync.Map)
		_, ok := svs.Load(val)
		if !ok {
			return false
		}
	}
	return true
}

func updateColVal(appid string, fileID string, colID string, val string) (int64, error) {

	stmt, err := db.Prepare(UpdateColValSQL)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	resp, err := stmt.Exec(val, appid, fileID, colID)
	if err != nil {
		return 0, err
	}
	return resp.RowsAffected()

}
