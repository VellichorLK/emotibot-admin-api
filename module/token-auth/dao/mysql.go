package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

const (
	enterpriseTable    = "enterprises"
	userTable          = "users"
	appTable           = "apps"
	userInfoTable      = "user_info"
	userColumnTable    = "user_column"
	roleTable          = "roles"
	rolePrivilegeTable = "privileges"
	moduleTable        = "modules"
)

var (
	userColumnList = []string{"uuid", "display_name", "user_name", "email", "enterprise", "type", "status", "role"}
)

type MYSQLController struct {
	connectDB *sql.DB
	auditDB   *sql.DB
}

func (controller *MYSQLController) InitDB(host string, port int, dbName string, account string, password string) bool {
	var dbString string
	if port == 0 {
		dbString = fmt.Sprintf("%s:%s@%s/%s", account, password, host, dbName)
	} else {
		dbString = fmt.Sprintf("%s:%s@(%s:%d)/%s", account, password, host, port, dbName)
	}
	util.LogTrace.Printf("Connect to db [%s]\n", dbString)
	db, err := sql.Open("mysql", dbString)

	if err != nil {
		util.LogError.Printf("Connect to db[%s] fail: [%s]\n", dbString, err.Error())
		return false
	}

	controller.connectDB = db
	return true
}

func (controller MYSQLController) checkDB() (bool, error) {
	if controller.connectDB == nil {
		util.LogError.Fatalln("connectDB is nil, db is !initialized properly")
		return false, fmt.Errorf("DB hasn't init")
	}
	controller.connectDB.Ping()
	return true, nil
}

func (controller *MYSQLController) InitAuditDB(host string, port int, dbName string, account string, password string) bool {
	var dbString string
	if port == 0 {
		dbString = fmt.Sprintf("%s:%s@%s/%s", account, password, host, dbName)
	} else {
		dbString = fmt.Sprintf("%s:%s@(%s:%d)/%s", account, password, host, port, dbName)
	}
	util.LogTrace.Printf("Connect to audit db [%s]\n", dbString)
	db, err := sql.Open("mysql", dbString)

	if err != nil {
		util.LogError.Printf("Connect to audit db[%s] fail: [%s]\n", dbString, err.Error())
		return false
	}

	controller.auditDB = db
	return true
}

func (controller MYSQLController) checkAuditDB() (bool, error) {
	if controller.auditDB == nil {
		util.LogError.Fatalln("auditDB is nil, audit db is !initialized properly")
		return false, fmt.Errorf("Audit DB hasn't init")
	}
	controller.auditDB.Ping()
	return true, nil
}

func (controller MYSQLController) GetEnterprises() (*data.Enterprises, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}
	enterprises := make(data.Enterprises, 0)
	rows, err := controller.connectDB.Query(fmt.Sprintf("SELECT uuid,name from %s", enterpriseTable))
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		enterprise := data.Enterprise{}
		err := rows.Scan(&enterprise.ID, &enterprise.Name)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
			return nil, err
		}
		enterprises = append(enterprises, enterprise)
	}

	return &enterprises, nil
}
func (controller MYSQLController) GetEnterprise(enterpriseID string) (*data.Enterprise, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}
	rows, err := controller.connectDB.Query(fmt.Sprintf("SELECT uuid,name from %s where uuid = ?", enterpriseTable), enterpriseID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		enterprise := data.Enterprise{}
		err := rows.Scan(&enterprise.ID, &enterprise.Name)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
			return nil, err
		}
		return &enterprise, nil
	}

	return nil, nil
}
func (controller MYSQLController) AddEnterprise(enterprise *data.Enterprise, user *data.User) (enterpriseID string, err error) {
	ok, err := controller.checkDB()
	if !ok {
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf(`
		INSERT INTO %s
		(uuid, display_name, user_name, email, enterprise, type, password, role)
		VALUES (UUID(), ?, ?, ?, ?, ?, ?, ?)`, userTable)
	ret, err := t.Exec(queryStr, user.DisplayName, user.UserName, user.Email, "", user.Type, user.Password, "")
	if err != nil {
		util.LogDBError(err)
		return
	}

	id, err := ret.LastInsertId()
	if err != nil {
		return
	}

	userID, err := getUUIDWithTx(int(id), t, userTable)
	if err != nil {
		return
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s
		(uuid, name, admin_user)
		VALUES (UUID(), ?, ?)
		`, enterpriseTable)
	ret, err = t.Exec(queryStr, enterprise.Name, userID)
	if err != nil {
		util.LogDBError(err)
		return
	}

	id, err = ret.LastInsertId()
	if err != nil {
		return
	}
	enterpriseID, err = getUUIDWithTx(int(id), t, enterpriseTable)
	if err != nil {
		return
	}
	err = updateUserEnterpriseWithTx(userID, enterpriseID, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = t.Commit()
	return
}

func getUUIDWithTx(id int, tx *sql.Tx, table string) (string, error) {
	if tx == nil {
		return "", errors.New("tx is nil")
	}

	queryStr := fmt.Sprintf("SELECT uuid FROM %s WHERE id = ?", table)
	row := tx.QueryRow(queryStr, id)
	uuid := ""
	err := row.Scan(&uuid)
	if err != nil {
		return "", err
	}
	return uuid, nil
}

func updateUserEnterpriseWithTx(userID string, enterpriseID string, tx *sql.Tx) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	queryStr := fmt.Sprintf("UPDATE %s SET enterprise = ? WHERE uuid = ?", userTable)
	ret, err := tx.Exec(queryStr, enterpriseID, userID)
	if err != nil {
		return err
	}

	cnt, err := ret.RowsAffected()
	if err != nil {
		return err
	}

	if cnt == 0 {
		return sql.ErrNoRows
	}

	return err
}

func (controller MYSQLController) DeleteEnterprise(enterpriseID string) (ret bool, err error) {
	ok, err := controller.checkDB()
	if !ok {
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return
	}
	defer util.ClearTransition(t)

	err = deleteEnterpriseAppsWithTx(enterpriseID, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = deleteEnterpriseWithTx(enterpriseID, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = deleteEnterpriseUsersWithTx(enterpriseID, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = t.Commit()
	if err != nil {
		return
	}
	ret = true
	return
}
func deleteEnterpriseAppsWithTx(enterpriseID string, tx *sql.Tx) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	queryStr := fmt.Sprintf("DELETE FROM %s WHERE enterprise = ?", appTable)
	_, err := tx.Exec(queryStr, enterpriseID)
	return err
}
func deleteEnterpriseWithTx(enterpriseID string, tx *sql.Tx) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	queryStr := fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", enterpriseTable)
	_, err := tx.Exec(queryStr, enterpriseID)
	return err
}
func deleteEnterpriseUsersWithTx(enterpriseID string, tx *sql.Tx) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	queryStr := fmt.Sprintf("DELETE FROM %s WHERE enterprise = ?", userTable)
	_, err := tx.Exec(queryStr, enterpriseID)
	return err
}

func scanSingleRowToUser(row *sql.Row) (*data.User, error) {
	user := data.User{}
	var err error
	err = row.Scan(&user.ID, &user.DisplayName, &user.UserName, &user.Email, &user.Enterprise, &user.Type, &user.Status, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func scanRowToUser(rows *sql.Rows) (*data.User, error) {
	user := data.User{}
	err := rows.Scan(&user.ID, &user.DisplayName, &user.UserName, &user.Email, &user.Enterprise, &user.Type, &user.Status, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func getUserColumnList(tableAlias string) string {
	if tableAlias == "" {
		return strings.Join(userColumnList, ",")
	}
	temp := make([]string, len(userColumnList))
	for idx, col := range userColumnList {
		temp[idx] = tableAlias + "." + col
	}
	return strings.Join(temp, ",")
}

func (controller MYSQLController) GetUsers(enterpriseID string) (*data.Users, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}

	userInfoMap, err := controller.getUsersInfo(enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}

	users := make(data.Users, 0)
	rows, err := controller.connectDB.Query(fmt.Sprintf("SELECT %s FROM %s WHERE enterprise = ?",
		getUserColumnList(""), userTable), enterpriseID)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := scanRowToUser(rows)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
			return nil, err
		}

		if _, ok := userInfoMap[user.ID]; ok {
			info := userInfoMap[user.ID]
			user.CustomInfo = &info
		}

		users = append(users, *user)
	}

	return &users, nil
}
func (controller MYSQLController) GetUser(enterpriseID string, userID string) (*data.User, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}

	queryStr := fmt.Sprintf(`SELECT %s FROM %s WHERE enterprise = ? and uuid = ?`,
		getUserColumnList(""), userTable)
	row := controller.connectDB.QueryRow(queryStr, enterpriseID, userID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	user, err := scanSingleRowToUser(row)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	info, err := controller.getUserInfo(enterpriseID, userID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	user.CustomInfo = info

	return user, nil
}
func (controller MYSQLController) GetAdminUser(enterpriseID string) (*data.User, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}
	queryStr := fmt.Sprintf(
		"SELECT %s FROM %s as u LEFT JOIN %s as e ON e.admin_user = u.uuid AND e.uuid = ?",
		getUserColumnList("u"), userTable, enterpriseTable)
	row := controller.connectDB.QueryRow(queryStr, enterpriseID)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	user, err := scanSingleRowToUser(row)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	return user, nil
}
func (controller MYSQLController) GetAuthUser(account string, passwd string) (*data.User, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}

	queryStr := fmt.Sprintf(`SELECT %s FROM %s WHERE (user_name = ? OR email = ?) AND password = ?`,
		getUserColumnList(""), userTable)
	rows, err := controller.connectDB.Query(queryStr, account, account, passwd)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		user, err := scanRowToUser(rows)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
			return nil, err
		}
		info, err := controller.getUserInfo(*user.Enterprise, user.ID)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		user.CustomInfo = info
		return user, nil
	}

	return nil, nil
}
func (controller MYSQLController) AddUser(enterpriseID string, user *data.User, roleID string) (uuid string, err error) {
	defer func() {
		if err != nil {
			_, file, line, _ := runtime.Caller(1)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		}
	}()
	ok, err := controller.checkDB()
	if !ok {
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	count := 0
	queryStr := fmt.Sprintf("SELECT count(*) FROM %s WHERE user_name = ? OR (email = ? AND email != '')", userTable)
	row := t.QueryRow(queryStr, user.UserName, user.Email)
	err = row.Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return
	}
	if count > 0 {
		err = errors.New("Conflict user")
		return
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s
		(uuid, display_name, user_name, email, enterprise, type, password, role)
		VALUES (UUID(), ?, ?, ?, ?, ?, ?, ?)`, userTable)
	ret, err := t.Exec(queryStr, user.DisplayName, user.UserName, user.Email, enterpriseID, user.Type, user.Password, roleID)
	if err != nil {
		return
	}

	id, err := ret.LastInsertId()
	if err != nil {
		return
	}

	uuid, err = getUserUUIDWithTx(enterpriseID, id, t)
	if err != nil {
		return
	}

	// if custom info not set, no need to check custom columns
	if user.CustomInfo == nil {
		err = t.Commit()
		return
	}
	err = insertCustomInfoWithTx(enterpriseID, uuid, *user.CustomInfo, t)
	if err != nil {
		return
	}

	err = t.Commit()
	return
}
func (controller MYSQLController) getUserUUID(enterpriseID string, userID int64) (string, error) {
	ok, err := controller.checkDB()
	if !ok {
		return "", err
	}

	queryStr := fmt.Sprintf("SELECT uuid from %s WHERE enterprise = ? and id = ?", userTable)
	row := controller.connectDB.QueryRow(queryStr, enterpriseID, userID)

	ret := ""
	err = row.Scan(&ret)
	return ret, err
}
func getUserUUIDWithTx(enterpriseID string, userID int64, tx *sql.Tx) (string, error) {
	queryStr := fmt.Sprintf("SELECT uuid from %s WHERE enterprise = ? and id = ?", userTable)
	row := tx.QueryRow(queryStr, enterpriseID, userID)

	ret := ""
	err := row.Scan(&ret)
	return ret, err
}
func insertCustomInfoWithTx(enterpriseID, userID string, customInfo map[string]string, tx *sql.Tx) (err error) {
	queryStr := fmt.Sprintf(`SELECT id, info.column FROM %s AS info WHERE enterprise = ?`, userColumnTable)
	colRows, err := tx.Query(queryStr, enterpriseID)
	if err != nil {
		return
	}
	defer colRows.Close()

	paramList := [][]interface{}{}

	for colRows.Next() {
		colID, colName := 0, ""
		err = colRows.Scan(&colID, &colName)
		if err != nil {
			return
		}
		if val, ok := customInfo[colName]; ok {
			paramList = append(paramList, []interface{}{userID, colID, val})
		}
	}

	for _, param := range paramList {
		queryStr = fmt.Sprintf(`INSERT INTO %s (user_id, column_id, value) VALUES (?, ?, ?)`, userInfoTable)
		_, err = tx.Exec(queryStr, param...)
		if err != nil {
			return
		}
	}
	return
}

func (controller MYSQLController) UpdateUser(enterpriseID string, user *data.User) (err error) {
	ok, err := controller.checkDB()
	if !ok {
		return err
	}
	defer func() {
		util.LogDBError(err)
	}()

	t, err := controller.connectDB.Begin()
	if err != nil {
		return
	}

	var queryStr string
	var params []interface{}
	if user.Password == nil || *user.Password == "" {
		queryStr = fmt.Sprintf(`UPDATE %s SET
			display_name = ?, email = ?, type = ?, role = ?
			WHERE uuid = ? AND enterprise = ?`, userTable)
		params = []interface{}{user.DisplayName, user.Email, user.Type, user.Role, user.ID, user.Enterprise}
	} else {
		queryStr = fmt.Sprintf(`UPDATE %s SET
			display_name = ?, email = ?, type = ?, role = ?,
			password = ? WHERE uuid = ? AND enterprise = ?`, userTable)
		params = []interface{}{user.DisplayName, user.Email, user.Type, user.Role, user.Password, user.ID, user.Enterprise}
	}
	_, err = t.Exec(queryStr, params...)
	if err != nil {
		return
	}

	if user.CustomInfo == nil {
		err = t.Commit()
		return
	}

	queryStr = fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", userInfoTable)
	_, err = t.Exec(queryStr, user.ID)
	if err != nil {
		return
	}
	err = insertCustomInfoWithTx(enterpriseID, user.ID, *user.CustomInfo, t)
	if err != nil {
		return
	}

	err = t.Commit()
	return err
}
func (controller MYSQLController) DisableUser(enterpriseID string, userID string) (bool, error) {
	return false, nil
}
func (controller MYSQLController) DeleteUser(enterpriseID string, userID string) (bool, error) {
	ok, err := controller.checkDB()
	if !ok {
		return false, err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		return false, err
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", userInfoTable)
	_, err = t.Exec(queryStr, userID)
	if err != nil {
		return false, err
	}

	queryStr = fmt.Sprintf("DELETE FROM %s WHERE enterprise = ? AND uuid = ?", userTable)
	_, err = t.Exec(queryStr, enterpriseID, userID)
	if err != nil {
		return false, err
	}
	err = t.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}
func (controller MYSQLController) getUserInfo(enterpriseID string, userID string) (ret *map[string]string, err error) {
	err = nil
	ok, err := controller.checkDB()
	if !ok {
		return
	}

	queryStr := fmt.Sprintf(`SELECT col.column, info.value
		FROM %s as col, %s as info
		WHERE info.column_id = col.id AND info.user_id = ? AND col.enterprise = ?`, userColumnTable, userInfoTable)
	rows, err := controller.connectDB.Query(queryStr, userID, enterpriseID)
	if err != nil {
		return
	}
	defer rows.Close()

	infoMap := make(map[string]string)
	for rows.Next() {
		var key string
		var val string
		err = rows.Scan(&key, &val)
		if err != nil {
			return
		}
		infoMap[key] = val
	}
	ret = &infoMap
	return
}
func (controller MYSQLController) getUsersInfo(enterpriseID string) (ret map[string]map[string]string, err error) {
	err = nil
	ok, err := controller.checkDB()
	if !ok {
		return
	}

	queryStr := fmt.Sprintf(`SELECT info.user_id, col.column, info.value
		FROM %s as col, %s as info
		WHERE info.column_id = col.id AND col.enterprise = ?`, userColumnTable, userInfoTable)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		return
	}
	defer rows.Close()

	ret = make(map[string]map[string]string)
	for rows.Next() {
		var userID string
		var key string
		var val string
		err = rows.Scan(&userID, &key, &val)
		if err != nil {
			return
		}
		if userInfo, ok := ret[userID]; !ok {
			ret[userID] = map[string]string{
				key: val,
			}
		} else {
			userInfo[key] = val
		}
	}
	return
}

func (controller MYSQLController) GetApps(enterpriseID string) (*data.Apps, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}
	apps := make(data.Apps, 0)
	queryStr := fmt.Sprintf("SELECT uuid,name,UNIX_TIMESTAMP(start),UNIX_TIMESTAMP(end),UNIX_TIMESTAMP(count),status from %s where enterprise = ?", appTable)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		app := data.App{}
		err := rows.Scan(&app.ID, &app.Name, &app.ValidStart, &app.ValidEnd, &app.ValidCount, &app.Status)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
			return nil, err
		}
		apps = append(apps, app)
	}

	return &apps, nil
}
func (controller MYSQLController) GetApp(enterpriseID string, AppID string) (*data.App, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}
	queryStr := fmt.Sprintf("SELECT uuid,name,UNIX_TIMESTAMP(start),UNIX_TIMESTAMP(end),UNIX_TIMESTAMP(count),status from %s where enterprise = ? and uuid = ?", appTable)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID, AppID)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		app := data.App{}
		err := rows.Scan(&app.ID, &app.Name, &app.ValidStart, &app.ValidEnd, &app.ValidCount, &app.Status)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
			return nil, err
		}
		return &app, nil
	}

	return nil, nil
}
func (controller MYSQLController) AddApp(enterpriseID string, app *data.App) (string, error) {
	ok, err := controller.checkDB()
	if !ok {
		return "", err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		return "", err
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf(`
		INSERT INTO %s
		(uuid, name, start, end, count, enterprise, status)
		VALUES (UUID(), ?, ?, ?, ?, ?, 1)`, appTable)
	ret, err := t.Exec(queryStr,
		app.Name, app.ValidStart, app.ValidEnd, app.ValidCount, enterpriseID)
	if err != nil {
		return "", err
	}

	id, err := ret.LastInsertId()
	if err != nil {
		return "", err
	}
	uuid, err := getUUIDWithTx(int(id), t, appTable)
	if err != nil {
		return "", err
	}

	err = t.Commit()
	return uuid, err
}
func (controller MYSQLController) UpdateApp(enterpriseID string, app data.App) (*data.App, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}
	return nil, nil
}
func (controller MYSQLController) DisableApp(enterpriseID string, AppID string) (bool, error) {
	panic("TODO")
}
func (controller MYSQLController) DeleteApp(enterpriseID string, AppID string) (bool, error) {
	panic("TODO")
}

func (controller MYSQLController) GetRoles(enterpriseID string) ([]*data.Role, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}

	queryStr := fmt.Sprintf("SELECT id, uuid, name, discription FROM %s WHERE enterprise = ?", roleTable)
	roleRows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()

	roleIDs := []string{}
	roleMap := map[int]*data.Role{}
	roleUUIDMap := map[string]*data.Role{}
	ret := []*data.Role{}
	for roleRows.Next() {
		var id int
		temp := data.Role{}
		err = roleRows.Scan(&id, &temp.UUID, &temp.Name, &temp.Description)
		if err != nil {
			return nil, err
		}
		temp.Privileges = map[string][]string{}
		ret = append(ret, &temp)
		roleIDs = append(roleIDs, fmt.Sprintf("%d", id))
		roleMap[id] = &temp
		roleUUIDMap[temp.UUID] = &temp
	}

	if len(roleIDs) == 0 {
		return ret, nil
	}

	queryStr = fmt.Sprintf("SELECT role, count(*) FROM %s WHERE enterprise = ? AND role != '' GROUP BY role", userTable)
	countRows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer countRows.Close()

	for countRows.Next() {
		id := ""
		count := 0
		err = countRows.Scan(&id, &count)
		if role, ok := roleUUIDMap[id]; ok {
			role.UserCount = count
		}
	}

	queryStr = fmt.Sprintf(`
		SELECT priv.role, priv.module, priv.cmd_list, module.code
		FROM %s as priv, %s as module
		WHERE module.id = priv.module and priv.role in (%s)`,
		rolePrivilegeTable, moduleTable, strings.Join(roleIDs, ","))
	privRows, err := controller.connectDB.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer privRows.Close()

	for privRows.Next() {
		var roleID int
		var module string
		var cmdList string
		var moduleCode string
		err := privRows.Scan(&roleID, &module, &cmdList, &moduleCode)
		if err != nil {
			return nil, err
		}
		cmds := strings.Split(cmdList, ",")
		if role, ok := roleMap[roleID]; ok {
			role.Privileges[moduleCode] = cmds
		}
	}

	return ret, nil
}
func (controller MYSQLController) GetRole(enterpriseID string, roleID string) (*data.Role, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}

	queryStr := fmt.Sprintf("SELECT id, name, discription FROM %s WHERE enterprise = ? and uuid = ?", roleTable)
	roleRow := controller.connectDB.QueryRow(queryStr, enterpriseID, roleID)
	ret := data.Role{}
	ret.UUID = roleID
	var id int
	err = roleRow.Scan(&id, &ret.Name, &ret.Description)
	ret.Privileges = map[string][]string{}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	queryStr = fmt.Sprintf("SELECT count(*) FROM %s WHERE enterprise = ? AND role = ? GROUP BY role", userTable)
	countRow := controller.connectDB.QueryRow(queryStr, enterpriseID, id)
	err = countRow.Scan(&ret.UserCount)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	queryStr = fmt.Sprintf(`
		SELECT priv.id, priv.module, priv.cmd_list, module.code
		FROM %s as priv, %s as module
		WHERE module.id = priv.module and priv.role = ?`, rolePrivilegeTable, moduleTable)
	privRows, err := controller.connectDB.Query(queryStr, id)
	if err != nil {
		return nil, err
	}
	defer privRows.Close()

	for privRows.Next() {
		var id int
		var module string
		var cmdList string
		var moduleCode string
		err := privRows.Scan(&id, &module, &cmdList, &moduleCode)
		if err != nil {
			return nil, err
		}
		cmds := strings.Split(cmdList, ",")
		ret.Privileges[moduleCode] = cmds
	}

	return &ret, nil
}
func (controller MYSQLController) getRoleUUIDById(id int) (uuid string, err error) {
	ok, err := controller.checkDB()
	if !ok {
		return
	}
	queryStr := fmt.Sprintf("SELECT uuid FROM %s WHERE id = ?", roleTable)
	row := controller.connectDB.QueryRow(queryStr, id)
	err = row.Scan(&uuid)
	return
}
func (controller MYSQLController) getRoleUUIDByIdWidthTx(id int, t *sql.Tx) (uuid string, err error) {
	ok, err := controller.checkDB()
	if !ok {
		return
	}
	queryStr := fmt.Sprintf("SELECT uuid FROM %s WHERE id = ?", roleTable)
	row := t.QueryRow(queryStr, id)
	err = row.Scan(&uuid)
	return
}

func (controller MYSQLController) AddRole(enterprise string, role *data.Role) (uuid string, err error) {
	defer func() {
		util.LogTrace.Println("Add role ret uuid: ", uuid)
		if err != nil {
			util.LogTrace.Println("Add role ret: ", err.Error())
		}
	}()
	ok, err := controller.checkDB()
	if !ok {
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf("INSERT INTO %s (uuid, name, enterprise, discription) VALUES (UUID(), ?, ?, ?)", roleTable)
	ret, err := t.Exec(queryStr, role.Name, enterprise, role.Description)
	if err != nil {
		return
	}
	roleID, err := ret.LastInsertId()
	if err != nil {
		return
	}

	moduleMap := map[string]*data.Module{}
	modules, err := controller.GetModules(enterprise)
	if err != nil {
		return
	}
	for _, mod := range modules {
		moduleMap[mod.Code] = mod
	}

	for priv, cmds := range role.Privileges {
		if mod, ok := moduleMap[priv]; ok {
			modID := mod.ID
			queryStr = fmt.Sprintf(`
				INSERT INTO %s (role, module, cmd_list)
				VALUES (?, ?, ?)`, rolePrivilegeTable)
			_, err = t.Exec(queryStr, roleID, modID, strings.Join(cmds, ","))
			if err != nil {
				return
			}
		}
	}

	err = t.Commit()
	if err != nil {
		return
	}
	uuid, err = controller.getRoleUUIDById(int(roleID))
	return
}
func (controller MYSQLController) UpdateRole(enterprise string, roleUUID string, role *data.Role) (result bool, err error) {
	ok, err := controller.checkDB()
	if !ok {
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf("SELECT id FROM %s WHERE uuid = ?", roleTable)
	row := t.QueryRow(queryStr, roleUUID)
	var roleID int
	err = row.Scan(&roleID)
	if err != nil {
		return
	}

	queryStr = fmt.Sprintf(`
		UPDATE %s SET name = ?, discription = ?
		WHERE enterprise = ? AND id = ?`, roleTable)
	_, err = t.Exec(queryStr, role.Name, role.Description, enterprise, roleID)
	if err != nil {
		return
	}

	moduleMap := map[string]*data.Module{}
	modules, err := controller.GetModules(enterprise)
	if err != nil {
		return
	}
	for _, mod := range modules {
		moduleMap[mod.Code] = mod
	}

	queryStr = fmt.Sprintf(`DELETE FROM %s WHERE role = ?`, rolePrivilegeTable)
	_, err = t.Exec(queryStr, roleID)
	if err != nil {
		return
	}

	for priv, cmds := range role.Privileges {
		if mod, ok := moduleMap[priv]; ok {
			modID := mod.ID
			queryStr = fmt.Sprintf(`
				INSERT INTO %s (role, module, cmd_list)
				VALUES (?, ?, ?)`, rolePrivilegeTable)
			_, err = t.Exec(queryStr, roleID, modID, strings.Join(cmds, ","))
			if err != nil {
				return
			}
		}
	}

	t.Commit()
	result = true
	return
}
func (controller MYSQLController) DeleteRole(enterpriseID string, roleID string) (bool, error) {
	ok, err := controller.checkDB()
	if !ok {
		return false, err
	}

	queryStr := fmt.Sprintf("SELECT id from %s WHERE enterprise = ? and uuid = ?", roleTable)
	roleRow := controller.connectDB.QueryRow(queryStr, enterpriseID, roleID)
	var id int
	err = roleRow.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		return false, err
	}
	defer util.ClearTransition(t)

	queryStr = fmt.Sprintf("DELETE FROM %s WHERE role = ?", rolePrivilegeTable)
	_, err = t.Exec(queryStr, id)
	if err != nil {
		return false, err
	}

	queryStr = fmt.Sprintf("DELETE FROM %s WHERE enterprise = ? and uuid = ?", roleTable)
	_, err = t.Exec(queryStr, enterpriseID, roleID)
	if err != nil {
		return false, err
	}

	err = t.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}
func (controller MYSQLController) GetUsersOfRole(enterpriseID string, roleUUID string) (*data.Users, error) {
	ok, err := controller.checkDB()
	if !ok {
		return nil, err
	}

	userInfoMap, err := controller.getUsersInfo(enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}

	users := make(data.Users, 0)
	rows, err := controller.connectDB.Query(fmt.Sprintf("SELECT %s FROM %s WHERE enterprise = ? and role = ?",
		getUserColumnList(""), userTable), enterpriseID, roleUUID)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user, err := scanRowToUser(rows)
		if err != nil {
			_, file, line, _ := runtime.Caller(0)
			util.LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
			return nil, err
		}

		if _, ok := userInfoMap[user.ID]; ok {
			info := userInfoMap[user.ID]
			user.CustomInfo = &info
		}

		users = append(users, *user)
	}

	return &users, nil
}

func (controller MYSQLController) GetModules(enterpriseID string) ([]*data.Module, error) {
	var err error
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	parseModuleRow := func(row *sql.Rows) (*data.Module, error) {
		temp := &data.Module{}
		commands := ""
		status := 0
		err = row.Scan(&temp.ID, &temp.Code, &temp.Name, &commands, &status)
		if err != nil {
			return nil, err
		}
		temp.Code = strings.TrimSpace(temp.Code)
		temp.Name = strings.TrimSpace(temp.Name)

		commands = strings.TrimSpace(commands)
		if commands != "" {
			temp.Commands = strings.Split(commands, ",")
		} else {
			temp.Commands = []string{}
		}
		temp.Status = (status > 0)
		return temp, nil
	}

	// if enterprise is empty, that is all system modules
	// if enterprise is not empty, use status column to check module's status
	queryStr := fmt.Sprintf(`
		SELECT
			id, code, name, cmd_list, status
		FROM %s WHERE enterprise = ''`, moduleTable)
	moduleRows, err := controller.connectDB.Query(queryStr)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer moduleRows.Close()

	moduleMap := map[string]*data.Module{}
	for moduleRows.Next() {
		temp, err := parseModuleRow(moduleRows)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		moduleMap[temp.Code] = temp
	}

	queryStr = fmt.Sprintf(`
		SELECT
			id, code, name, cmd_list, status
		FROM %s WHERE enterprise = ?`, moduleTable)
	privateRows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer privateRows.Close()

	for privateRows.Next() {
		temp, err := parseModuleRow(privateRows)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		if dftModule, ok := moduleMap[temp.Code]; ok {
			if temp.Name != "" {
				dftModule.Name = temp.Name
			}
			if len(temp.Commands) != 0 {
				dftModule.Commands = temp.Commands
			}
			dftModule.Status = temp.Status
		}
	}

	ret := []*data.Module{}
	for _, mod := range moduleMap {
		if mod.Status {
			ret = append(ret, mod)
		}
	}
	sort.Sort(ByID(ret))

	return ret, nil
}

type ByID []*data.Module

func (m ByID) Len() int      { return len(m) }
func (m ByID) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m ByID) Less(i, j int) bool {
	switch {
	case m[i] == nil:
		return true
	case m[j] == nil:
		return true
	default:
		return m[i].ID < m[j].ID
	}
}
