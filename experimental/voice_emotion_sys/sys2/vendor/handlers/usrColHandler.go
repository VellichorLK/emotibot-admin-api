package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

//user defined column table
const (
	GetColumSQL     = "select " + NCOLID + "," + NCOLNAME + "," + NDEDAULT + " from " + UsrColTable + " where " + NAPPID + "=?"
	NewColumnSQL    = "insert into " + UsrColTable + " (" + NCOLTYPE + "," + NCOLNAME + "," + NAPPID + "," + NDEDAULT + ") values (?,?,?,?)"
	NewSelSQL       = "insert into " + UsrSelValTable + "(" + NCOLID + "," + NSELVAL + ") values (?,?)"
	UpdateColValSQL = "update " + UsrColValTable + " set " + NCOLVAL + "=? where " +
		NID + " = ( select " + NID + " from " + MainTable + " where " + NAPPID + "=? and " + NFILEID + " = ?) and " +
		NCOLID + "=?"
	GetSelValueSQL = "select a." + NCOLID + ",a." + NCOLNAME + ",a." + NAPPID + ",a." + NDEDAULT + ",b." + NSELVAL + " from " + UsrColTable + " as a left join " + UsrSelValTable +
		" as b on a." + NCOLID + "=b." + NCOLID + " where " + NAPPID + "=? " + " order by a." + NAPPID
)

type UsrColHandler struct {
	FieldNameMap    sync.Map //col_id  -> field_name
	DefaultValue    sync.Map //appid  -> default value
	SelectableValue sync.Map // col_id  -> selectable map
	FieldOwner      sync.Map //col_id -> appid
}

type DefaultValue struct {
	ColID    string
	ColValue string
	ColName  string
}

type UsrColumBlock struct {
	ColName     *string  `json:"col_name"`
	ColID       string   `json:"col_id"`
	SelValue    []string `json:"selectable_value,omitempty"`
	OrgSelValue []string `json:"org_selectable_value,omitempty"`
	DefaultVal  *string  `json:"default_value"`
}

type UsrColumnUpdateBlock struct {
	FileID string `json:"file_id"`
	ColID  string `json:"col_id"`
	Value  string `json:"col_value"`
}

//var DefaulUsrField *UsrColHandler

func (uch *UsrColHandler) LoadUsrField(query string, params ...interface{}) error {

	rows, err := db.Query(query, params...)
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

		uch.FieldOwner.Store(colID, appid)

		_, ok := colIDMap[key]
		if !ok {
			colIDMap[key] = true

			uch.FieldNameMap.Store(key, colName)

			var dvs []*DefaultValue

			dvsInterface, ok := uch.DefaultValue.Load(appid)
			if !ok {
				dvs = make([]*DefaultValue, 0)
			} else {
				dvs = dvsInterface.([]*DefaultValue)
			}

			dv := &DefaultValue{colID, defValue, colName}
			dvs = append(dvs, dv)

			uch.DefaultValue.Store(appid, dvs)

		}

		if colValue.Valid {

			svInterface, ok := uch.SelectableValue.Load(key)
			if !ok {
				sv := &sync.Map{}
				sv.Store(colValue.String, true)
				uch.SelectableValue.Store(key, sv)
			} else {
				sv := svInterface.(*sync.Map)
				sv.Store(colValue.String, true)

			}

		}

	}

	return rows.Err()

}

func UserColumnOperation(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		GetUserColumn(w, r)
	case "PUT":
		AddUserColumn(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func UserColumnModifier(w http.ResponseWriter, r *http.Request) {

	paths := strings.SplitN(r.URL.Path, "/", MaxSlash)

	switch r.Method {
	case "DELETE":
		DeleteUserColumn(w, r, paths[MaxSlash-1])
	case "PATCH":
		UpdateUsrColumn(w, r, paths[MaxSlash-1])
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

}

func AddUserColumn(w http.ResponseWriter, r *http.Request) {

	appid := r.Header.Get(HXAPPID)
	ucb := &UsrColumBlock{}

	err := json.NewDecoder(r.Body).Decode(ucb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if ucb.ColName == nil {
		http.Error(w, "No col assigned", http.StatusBadRequest)
		return
	}

	var dvs []*DefaultValue

	uch := &UsrColHandler{}
	err = uch.LoadUsrField(GetSelValueSQL, appid)
	if err != nil {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
		return
	}

	dvsInterface, ok := uch.DefaultValue.Load(appid)
	if ok {
		dvs = dvsInterface.([]*DefaultValue)
	} else {
		dvs = make([]*DefaultValue, 0)
	}

	if len(dvs)+1 > LIMITUSRCOL {
		http.Error(w, "Over limit ", http.StatusBadRequest)
		return
	}

	usrType := UsrColumString
	if ucb.SelValue != nil {
		if len(ucb.SelValue) > LIMITUSRSEL {
			http.Error(w, "Over limit selectable value", http.StatusBadRequest)
			return
		}
		if hasDupString(ucb.SelValue) {
			http.Error(w, "has duplicate selectable value", http.StatusBadRequest)
			return
		}
		usrType = UsrColumSel
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	rows, err := tx.Exec(NewColumnSQL, usrType, *ucb.ColName, appid, ucb.DefaultVal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	colID, err := rows.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	colIDString := strconv.FormatInt(colID, 10)
	selMap := &sync.Map{}
	if ucb.SelValue != nil {
		for _, sel := range ucb.SelValue {
			_, err := tx.Exec(NewSelSQL, colID, sel)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			selMap.Store(sel, true)
		}
	}

	err = batchInsert(tx, colIDString, *ucb.DefaultVal, appid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dv := &DefaultValue{colIDString, *ucb.DefaultVal, *ucb.ColName}
	dvs = append(dvs, dv)

	ucb.ColID = colIDString

	encodeRes, err := json.Marshal(ucb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	contentType := ContentTypeJSON

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(encodeRes)

}

func DeleteUserColumn(w http.ResponseWriter, r *http.Request, colID string) {

	appid := r.Header.Get(HXAPPID)

	uch := &UsrColHandler{}
	err := uch.LoadUsrField(GetSelValueSQL, appid)
	if err != nil {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
		return
	}

	owner, ok := uch.FieldOwner.Load(colID)
	if !ok || strings.Compare(owner.(string), appid) != 0 {
		http.Error(w, "No such col_id", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	sql := "delete from " + UsrColTable + " where " + NAPPID + " = ? and " + NCOLID + "=?"
	resp, err := tx.Exec(sql, appid, colID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	aff, err := resp.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if aff == 0 {
		http.Error(w, "No such col_id", http.StatusBadRequest)
		return
	}

	sql = "delete from " + UsrColValTable + " where " + NCOLID + "=?"
	//" and " + NID + " in ( select " + NID + " from " + MainTable + " where " + NAPPID + "=?)"
	_, err = tx.Exec(sql, colID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql = "delete from " + UsrSelValTable + " where " + NCOLID + "=?"
	_, err = tx.Exec(sql, colID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func UpdateUsrColumn(w http.ResponseWriter, r *http.Request, colID string) {
	appid := r.Header.Get(HXAPPID)

	uch := &UsrColHandler{}
	err := uch.LoadUsrField(GetSelValueSQL, appid)
	if err != nil {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
		return
	}

	owner, ok := uch.FieldOwner.Load(colID)
	if !ok || strings.Compare(owner.(string), appid) != 0 {
		http.Error(w, "No such col_id", http.StatusBadRequest)
		return
	}

	ucb := &UsrColumBlock{}
	err = json.NewDecoder(r.Body).Decode(ucb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	sql := "update " + UsrColTable
	var count int
	params := make([]interface{}, 0)
	colType := UsrColumString

	if ucb.SelValue != nil && len(ucb.SelValue) > 0 {

		if ucb.OrgSelValue != nil && len(ucb.OrgSelValue) > len(ucb.SelValue) {
			http.Error(w, "length of selectable_value must greater than  length of org_selectable_value", http.StatusBadRequest)
			return
		}

		if len(ucb.SelValue) > LIMITUSRSEL {
			http.Error(w, "Over limit of selectable value", http.StatusBadRequest)
			return
		}

		selMapInterface, ok := uch.SelectableValue.Load(colID)
		var selMap *sync.Map
		if ok {
			selMap = selMapInterface.(*sync.Map)
		}
		var idx int
		for idx = 0; idx < len(ucb.OrgSelValue); idx++ {

			if selMap != nil {
				_, ok := selMap.Load(ucb.OrgSelValue[idx])
				if !ok {
					http.Error(w, "no selectable value "+ucb.OrgSelValue[idx], http.StatusBadRequest)
					return
				}
			}

			selSQL := "update " + UsrSelValTable + " set " + NSELVAL + "=? where " + NSELVAL + "=? and " +
				NCOLID + " in (select " + NCOLID + " from " + UsrColTable + " where " + NAPPID + "=?)"
			_, err := tx.Exec(selSQL, ucb.SelValue[idx], ucb.OrgSelValue[idx], appid)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if ok {
			selMap := selMapInterface.(*sync.Map)
			var countDef int
			selMap.Range(func(k, v interface{}) bool {
				countDef++
				return true
			})

			diff := len(ucb.SelValue) - idx

			if countDef+diff > LIMITUSRSEL {
				http.Error(w, "Over limit of selectable value", http.StatusBadRequest)
				return
			}
		}

		for i := idx; i < len(ucb.SelValue); i++ {
			_, err := tx.Exec(NewSelSQL, colID, ucb.SelValue[i])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}

		colType = UsrColumSel
		sql += " set " + NCOLTYPE + "=?"
		params = append(params, colType)
		count++
	}

	if ucb.ColName != nil {
		if count > 0 {
			sql += ", "
		} else {
			sql += " set "
		}
		sql += NCOLNAME + "=?"
		params = append(params, *ucb.ColName)
		count++
	}

	if ucb.DefaultVal != nil {
		if count > 0 {
			sql += ", "
		} else {
			sql += " set "
		}
		sql += NDEDAULT + "=?"
		params = append(params, *ucb.DefaultVal)
		count++
	}

	if count > 0 {
		sql += " where " + NAPPID + "=? and " + NCOLID + "=?"
		params = append(params, appid, colID)

		_, err := tx.Exec(sql, params...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()
}

func batchInsert(tx *sql.Tx, colID string, colVal string, appid string) error {
	sql := "select id from fileInformation where appid=?"
	rows, err := db.Query(sql, appid)
	if err != nil {
		return err
	}
	defer rows.Close()

	var id string

	var count uint64
	var values string

	insertSQL := "insert into userColumnValue (id,col_id,col_value) values "

	for rows.Next() {

		if count >= 100000 {
			count = 0
			values = ""
			_, err = tx.Exec(insertSQL + values)
			if err != nil {
				return err
			}
		}

		err := rows.Scan(&id)
		if err != nil {
			return err
		}

		if count != 0 {
			values += ","
		}

		values += "(" + id
		values += "," + colID + "," + "\"" + colVal + "\""
		values += ")"
		count++
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	_, err = tx.Exec(insertSQL + values)
	if err != nil {
		return err
	}

	return nil

}

//GetUserColumn get the user defined column
func GetUserColumn(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	uch := &UsrColHandler{}
	err := uch.LoadUsrField(GetSelValueSQL, appid)
	if err != nil {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
		return
	}

	ucbs := make([]*UsrColumBlock, 0)
	dvsInterface, ok := uch.DefaultValue.Load(appid)

	if ok {
		dvs := dvsInterface.([]*DefaultValue)
		for _, dv := range dvs {
			ucb := &UsrColumBlock{}
			ucb.SelValue = make([]string, 0)

			ucb.ColID = dv.ColID
			colVal := dv.ColValue
			ucb.DefaultVal = &colVal
			nameInterface, ok := uch.FieldNameMap.Load(dv.ColID)

			if !ok {
				log.Println("Field name map has no colum id " + dv.ColID)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			name := nameInterface.(string)
			ucb.ColName = &name

			selMapInterface, ok := uch.SelectableValue.Load(dv.ColID)
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
	uch := &UsrColHandler{}
	err := uch.LoadUsrField(GetSelValueSQL, appid)
	if err != nil {
		http.Error(w, "internal server error ", http.StatusInternalServerError)
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
			if !checkSelectableVal(ucub.ColID, ucub.Value, uch) {
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

func checkSelectableVal(colID string, val string, uch *UsrColHandler) bool {

	selMapInterface, ok := uch.SelectableValue.Load(colID)

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
