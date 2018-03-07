package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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
	FieldNameMap    map[string]string          //col_id  = field_name
	DefaultValue    map[string][]*DefaultValue //appid  = default value
	SelectableValue map[string]map[string]bool // col_id  = selectable map
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
	DefaulUsrField.FieldNameMap = make(map[string]string)
	DefaulUsrField.DefaultValue = make(map[string][]*DefaultValue)
	DefaulUsrField.SelectableValue = make(map[string]map[string]bool)

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

			DefaulUsrField.FieldNameMap[key] = colName

			dvs, ok := DefaulUsrField.DefaultValue[appid]
			if !ok {
				dvs = make([]*DefaultValue, 0)

			}

			dv := &DefaultValue{colID, defValue}
			dvs = append(dvs, dv)
			DefaulUsrField.DefaultValue[appid] = dvs

		}

		if colValue.Valid {

			sv, ok := DefaulUsrField.SelectableValue[key]
			if !ok {
				sv = make(map[string]bool)
				DefaulUsrField.SelectableValue[key] = sv
			}
			sv[colValue.String] = true
		}

	}

	return rows.Err()

}

//GetUserColumn get the user defined column
func GetUserColumn(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if DefaulUsrField == nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := LoadUsrField()
	if err != nil {
		http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ucbs := make([]*UsrColumBlock, 0)
	dvs, ok := DefaulUsrField.DefaultValue[appid]
	if ok {
		for _, dv := range dvs {
			ucb := &UsrColumBlock{}
			ucb.SelValue = make([]string, 0)

			ucb.ColID = dv.ColID
			ucb.DefaultVal = dv.ColValue
			name, ok := DefaulUsrField.FieldNameMap[dv.ColID]
			if !ok {
				log.Println("Field name map has no colum id " + dv.ColID)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			ucb.ColName = name

			selMap, ok := DefaulUsrField.SelectableValue[dv.ColID]
			if ok {
				for selVal := range selMap {
					ucb.SelValue = append(ucb.SelValue, selVal)
				}
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
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if DefaulUsrField == nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ucubs := make([]*UsrColumnUpdateBlock, 0)
	//ucub := &UsrColumnUpdateBlock{}

	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&ucubs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	defer r.Body.Close()

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, ucub := range ucubs {
		//check if its fileID
		if isFileID(ucub.FileID) {
			if ucub.ColID != "" {
				if !checkSelectableVal(ucub.ColID, ucub.Value) {
					http.Error(w, "value("+ucub.Value+") can't be set", http.StatusBadRequest)
					return
				}
				_, err := updateColVal(appid, ucub.FileID, ucub.ColID, ucub.Value, tx)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			} else {
				http.Error(w, "No col_id given", http.StatusBadRequest)
				return
			}
		}
	}
	tx.Commit()

}

func checkSelectableVal(colID string, val string) bool {
	svs, ok := DefaulUsrField.SelectableValue[colID]
	if ok {
		_, ok := svs[val]
		if !ok {
			return false
		}
	}
	return true
}

func updateColVal(appid string, fileID string, colID string, val string, tx *sql.Tx) (int64, error) {

	stmt, err := tx.Prepare(UpdateColValSQL)
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
