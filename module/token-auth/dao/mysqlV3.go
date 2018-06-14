package dao

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"emotibot.com/emotigo/module/token-auth/internal/util"

	uuid "github.com/satori/go.uuid"
)

const (
	enterpriseTableV3     = "enterprises"
	userTableV3           = "users"
	userInfoTableV3       = "user_info"
	userPrivilegesTableV3 = "user_privileges"
	appTableV3            = "apps"
	appGroupTableV3       = "app_group"
	groupTableV3          = "robot_groups"
	roleTableV3           = "roles"
	rolePrivilegeTableV3  = "privileges"
	humanTableV3          = "human"
	machineTableV3        = "machine"
	columnTableV3         = "columns"
	moduleTableV3         = "modules"
)

func (controller MYSQLController) GetEnterprisesV3() ([]*data.EnterpriseV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	enterprises := make([]*data.EnterpriseV3, 0)
	rows, err := controller.connectDB.Query(fmt.Sprintf("SELECT uuid, name, description FROM %s", enterpriseTableV3))
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		enterprise := data.EnterpriseV3{}
		err := rows.Scan(&enterprise.ID, &enterprise.Name, &enterprise.Description)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		enterprises = append(enterprises, &enterprise)
	}

	return enterprises, nil
}

func (controller MYSQLController) GetEnterpriseV3(enterpriseID string) (*data.EnterpriseDetailV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	enterprise := data.EnterpriseDetailV3{}
	queryStr := fmt.Sprintf(`
		SELECT uuid, name, description
		FROM %s
		WHERE uuid = ?`, enterpriseTableV3)
	err = controller.connectDB.QueryRow(queryStr, enterpriseID).Scan(&enterprise.ID, &enterprise.Name, &enterprise.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	queryStr = fmt.Sprintf(`
		SELECT code, name, status
		FROM %s
		WHERE enterprise = ?`, moduleTableV3)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	modules := make([]*data.ModuleV3, 0)
	for rows.Next() {
		module := data.ModuleV3{}
		err := rows.Scan(&module.Code, &module.Name, &module.Status)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		modules = append(modules, &module)
	}

	enterprise.Modules = modules
	return &enterprise, nil
}

func (controller MYSQLController) AddEnterpriseV3(enterprise *data.EnterpriseV3, modules []string,
	adminUser *data.UserDetailV3) (enterpriseID string, err error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return
	}
	defer util.ClearTransition(t)

	adminUserUUID, err := uuid.NewV4()
	if err != nil {
		util.LogDBError(err)
		return
	}
	adminUserID := hex.EncodeToString(adminUserUUID[:])

	// Insert human table entry
	queryStr := fmt.Sprintf("INSERT IGNORE INTO %s (uuid) VALUES (?)", humanTableV3)
	_, err = t.Exec(queryStr, adminUserID)
	if err != nil {
		util.LogDBError(err)
		return
	}

	enterpriseUUID, err := uuid.NewV4()
	if err != nil {
		util.LogDBError(err)
		return
	}
	enterpriseID = hex.EncodeToString(enterpriseUUID[:])

	queryStr = fmt.Sprintf(`
		INSERT INTO %s
		(uuid, name, description)
		VALUES (?, ?, ?)`,
		enterpriseTableV3)
	_, err = t.Exec(queryStr, enterpriseID, enterprise.Name, enterprise.Description)
	if err != nil {
		util.LogDBError(err)
		return
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s
		(uuid, display_name, user_name, email, enterprise, type, password)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, userTableV3)
	_, err = t.Exec(queryStr, adminUserID, adminUser.DisplayName, adminUser.UserName,
		adminUser.Email, enterpriseID, adminUser.Type, adminUser.Password)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = addModulesEnterpriseWithTxV3(modules, enterpriseID, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return
	}

	return
}

func (controller MYSQLController) UpdateEnterpriseV3(enterpriseID string,
	enterprise *data.EnterpriseV3, modules []string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET name = ?, description = ?
		WHERE uuid = ?`,
		enterpriseTableV3)
	_, err = t.Exec(queryStr, enterprise.Name, enterprise.Description, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = updateModulesEnterpriseWithTxV3(modules, enterpriseID, t)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}

	return nil
}

func (controller MYSQLController) DeleteEnterpriseV3(enterpriseID string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	// Delete enterprise users
	userUUIDs := make([]string, 0)
	queryStr := fmt.Sprintf(`
		SELECT uuid
		FROM %s
		WHERE enterprise = ?`, userTableV3)
	rows, err := t.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userUUID string
		err = rows.Scan(&userUUID)
		if err != nil {
			util.LogDBError(err)
			return err
		}

		userUUIDs = append(userUUIDs, userUUID)
	}
	rows.Close()

	for _, userUUID := range userUUIDs {
		err = deleteUserWithTxV3(userUUID, t)
		if err != nil {
			util.LogDBError(err)
			return err
		}
	}

	// Delete enterprise groups
	groupUUIDs := make([]string, 0)
	queryStr = fmt.Sprintf(`
		SELECT uuid
		FROM %s
		WHERE enterprise = ?`, groupTableV3)
	rows, err = t.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var groupUUID string
		err = rows.Scan(&groupUUID)
		if err != nil {
			util.LogDBError(err)
			return err
		}

		groupUUIDs = append(groupUUIDs, groupUUID)
	}
	rows.Close()

	for _, groupUUID := range groupUUIDs {
		err = deleteGroupWithTxV3(groupUUID, t)
		if err != nil {
			util.LogDBError(err)
			return err
		}
	}

	// Delete enterprise apps
	appUUIDs := make([]string, 0)
	queryStr = fmt.Sprintf(`
		SELECT uuid
		FROM %s
		WHERE enterprise = ?`, appTableV3)
	rows, err = t.Query(queryStr, enterpriseID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var appUUID string
		err = rows.Scan(&appUUID)
		if err != nil {
			util.LogDBError(err)
			return err
		}

		appUUIDs = append(appUUIDs, appUUID)
	}
	rows.Close()

	for _, appUUID := range appUUIDs {
		err = deleteAppWithTxV3(appUUID, t)
		if err != nil {
			util.LogDBError(err)
			return err
		}
	}

	// Delete enterprise roles
	roleIDs := make([]int, 0)
	roleUUIDs := make([]string, 0)
	queryStr = fmt.Sprintf(`
		SELECT id, uuid
		FROM %s
		WHERE enterprise = ?`, roleTableV3)
	rows, err = t.Query(queryStr, enterpriseID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var roleID int
		var roleUUID string
		err = rows.Scan(&roleID, &roleUUID)
		if err != nil {
			util.LogDBError(err)
			return err
		}

		roleIDs = append(roleIDs, roleID)
		roleUUIDs = append(roleUUIDs, roleUUID)
	}
	rows.Close()

	for i, roleID := range roleIDs {
		roleUUID := roleUUIDs[i]

		err = deleteRoleWithTx(roleID, roleUUID, t)
		if err != nil {
			util.LogDBError(err)
			return err
		}
	}

	queryStr = fmt.Sprintf(`
		DELETE FROM %s
		WHERE uuid = ?`, enterpriseTableV3)
	_, err = t.Exec(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = t.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (controller MYSQLController) EnterpriseExistsV3(enterpriseID string) (bool, error) {
	queryStr := fmt.Sprintf("SELECT 1 FROM %s WHERE uuid = ?", enterpriseTableV3)
	return controller.rowExists(queryStr, enterpriseID)
}

func (controller MYSQLController) GetUsersV3(enterpriseID string, admin bool) ([]*data.UserV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	users := make([]*data.UserV3, 0)
	var queryStr string
	var queryParams []interface{}

	if admin {
		queryStr = fmt.Sprintf(`
			SELECT uuid, user_name, display_name, email, phone, type
			FROM %s
			WHERE type = %d`, userTableV3, enum.SuperAdminUser)
		queryParams = []interface{}{}
	} else {
		queryStr = fmt.Sprintf(`
			SELECT uuid, user_name, display_name, email, phone, type
			FROM %s
			WHERE enterprise = ? AND type != %d`, userTableV3, enum.SuperAdminUser)
		queryParams = []interface{}{enterpriseID}
	}

	rows, err := controller.connectDB.Query(queryStr, queryParams...)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := data.UserV3{}
		err := rows.Scan(&user.ID, &user.UserName, &user.DisplayName, &user.Email, &user.Phone, &user.Type)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		users = append(users, &user)
	}
	rows.Close()

	for _, user := range users {
		roles, err := controller.getUserRolesV3(user.ID)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		user.Roles = roles
	}

	return users, nil
}

func (controller MYSQLController) GetUserV3(enterpriseID string,
	userID string) (*data.UserDetailV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	var queryStr string
	var queryParams []interface{}

	if enterpriseID == "" {
		queryStr = fmt.Sprintf(`
		SELECT uuid, user_name, display_name, email, phone, type, enterprise, status, password
		FROM %s
		WHERE uuid = ?`, userTableV3)
		queryParams = []interface{}{userID}
	} else {
		queryStr = fmt.Sprintf(`
		SELECT uuid, user_name, display_name, email, phone, type, enterprise, status, password
		FROM %s
		WHERE enterprise = ? AND uuid = ?`, userTableV3)
		queryParams = []interface{}{enterpriseID, userID}
	}

	row := controller.connectDB.QueryRow(queryStr, queryParams...)

	user := data.UserDetailV3{}
	err = row.Scan(&user.ID, &user.UserName, &user.DisplayName, &user.Email, &user.Phone, &user.Type,
		&user.Enterprise, &user.Status, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	roles, err := controller.getUserRolesV3(user.ID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}

	user.Roles = roles

	return &user, nil
}

func (controller MYSQLController) GetAuthUserV3(account string, passwd string) (*data.UserDetailV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT uuid, user_name, display_name, email, phone, type, enterprise, status
		FROM %s
		WHERE (user_name = ? OR email = ?) AND password = ?`,
		userTableV3)
	row := controller.connectDB.QueryRow(queryStr, account, account, passwd)

	user := data.UserDetailV3{}
	err = row.Scan(&user.ID, &user.UserName, &user.DisplayName, &user.Email, &user.Phone, &user.Type,
		&user.Enterprise, &user.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	roles, err := controller.getUserRolesV3(user.ID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}

	user.Roles = roles

	return &user, nil
}

func (controller MYSQLController) AddUserV3(enterpriseID string,
	user *data.UserDetailV3) (userID string, err error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return
	}
	defer util.ClearTransition(t)

	count := 0
	queryStr := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE user_name = ? OR (email = ? AND email != '')`, userTableV3)
	row := t.QueryRow(queryStr, user.UserName, user.Email)
	err = row.Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		util.LogDBError(err)
		return
	}
	if count > 0 {
		err = errors.New("Conflict user")
		return
	}

	userUUID, _ := uuid.NewV4()
	userID = hex.EncodeToString(userUUID[:])

	// Insert human table entry
	queryStr = fmt.Sprintf("INSERT IGNORE INTO %s (uuid) VALUES (?)", humanTableV3)
	_, err = t.Exec(queryStr, userID)
	if err != nil {
		util.LogDBError(err)
		return
	}

	var queryParams []interface{}

	switch user.Type {
	case enum.SuperAdminUser:
		queryStr = fmt.Sprintf(`
			INSERT INTO %s
			(uuid, user_name, display_name, email, phone, type, password)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, userTableV3)
		queryParams = []interface{}{userID, user.UserName, user.DisplayName,
			user.Email, user.Phone, user.Type, user.Password}
	case enum.AdminUser:
		fallthrough
	case enum.NormalUser:
		queryStr = fmt.Sprintf(`
			INSERT INTO %s
			(uuid, user_name, display_name, email, phone, enterprise, type, password)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, userTableV3)
		queryParams = []interface{}{userID, user.UserName, user.DisplayName,
			user.Email, user.Phone, enterpriseID, user.Type, user.Password}
	}

	_, err = t.Exec(queryStr, queryParams...)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = addUserPrivilegesWithTxV3(userID, user, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	// If custom info not set, no need to check custom columns
	if user.CustomInfo == nil {
		err = t.Commit()
		if err != nil {
			util.LogDBError(err)
			return
		}
		return
	}

	err = insertCustomInfoWithTxV3(enterpriseID, userID, *user.CustomInfo, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return
	}

	return
}

func (controller MYSQLController) UpdateUserV3(enterpriseID string,
	userID string, user *data.UserDetailV3) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	var queryStr string
	var queryParams []interface{}

	switch user.Type {
	case enum.SuperAdminUser:
		enterpriseID = ""

		if user.Password == nil || *user.Password == "" {
			queryStr = fmt.Sprintf(`
				UPDATE %s
				SET user_name = ?, display_name = ?, email = ?, phone = ?
				WHERE uuid = ?`, userTableV3)
			queryParams = []interface{}{user.UserName, user.DisplayName, user.Email, user.Phone, userID}
		} else {
			queryStr = fmt.Sprintf(`
				UPDATE %s
				SET user_name = ?, display_name = ?, email = ?, phone = ?, password = ?
				WHERE uuid = ?`, userTableV3)
			queryParams = []interface{}{user.UserName, user.DisplayName, user.Email, user.Phone,
				user.Password, userID}
		}
	case enum.AdminUser:
		fallthrough
	case enum.NormalUser:
		if user.Password == nil || *user.Password == "" {
			queryStr = fmt.Sprintf(`
				UPDATE %s
				SET user_name = ?, display_name = ?, email = ?, phone = ?
				WHERE uuid = ?`, userTableV3)
			queryParams = []interface{}{user.UserName, user.DisplayName, user.Email, user.Phone,
				userID}
		} else {
			queryStr = fmt.Sprintf(`
				UPDATE %s
				SET user_name = ?, display_name = ?, email = ?, phone = ?, password = ?
				WHERE uuid = ?`, userTableV3)
			queryParams = []interface{}{user.UserName, user.DisplayName, user.Email, user.Phone,
				user.Password, userID}
		}
	}

	_, err = t.Exec(queryStr, queryParams...)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = deleteUserPrivilegesWithTxV3(userID, t)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = addUserPrivilegesWithTxV3(userID, user, t)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	if user.CustomInfo == nil {
		err = t.Commit()
		if err != nil {
			return err
		}
		return nil
	}

	queryStr = fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", userInfoTable)
	_, err = t.Exec(queryStr, user.UserName)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = insertCustomInfoWithTxV3(enterpriseID, user.UserName, *user.CustomInfo, t)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	return nil
}

func (controller MYSQLController) DeleteUserV3(enterpriseID string, userID string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	err = deleteUserWithTxV3(userID, t)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}

	return nil
}

func (controller MYSQLController) UserExistsV3(userID string) (bool, error) {
	queryStr := fmt.Sprintf("SELECT 1 FROM %s WHERE uuid = ?", userTableV3)
	return controller.rowExists(queryStr, userID)
}

func (controller MYSQLController) GetAppsV3(enterpriseID string) ([]*data.AppV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT uuid, name, status
		FROM %s
		WHERE enterprise = ?`, appTableV3)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	apps := make([]*data.AppV3, 0)
	for rows.Next() {
		app := data.AppV3{}
		err := rows.Scan(&app.ID, &app.Name, &app.Status)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		apps = append(apps, &app)
	}

	return apps, nil
}

func (controller MYSQLController) GetAppV3(enterpriseID string, appID string) (*data.AppDetailV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT uuid, name, description, status
		FROM %s
		WHERE enterprise = ? and uuid = ?`, appTableV3)
	row := controller.connectDB.QueryRow(queryStr, enterpriseID, appID)

	app := data.AppDetailV3{}
	err = row.Scan(&app.ID, &app.Name, &app.Description, &app.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	return &app, nil
}

func (controller MYSQLController) AddAppV3(enterpriseID string, app *data.AppDetailV3) (appID string, err error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return
	}
	defer util.ClearTransition(t)

	appUUID, err := uuid.NewV4()
	if err != nil {
		util.LogDBError(err)
		return
	}
	appID = hex.EncodeToString(appUUID[:])

	// Insert machine table entry
	queryStr := fmt.Sprintf("INSERT IGNORE INTO %s (uuid) VALUES (?)", machineTableV3)
	_, err = t.Exec(queryStr, appID)
	if err != nil {
		return
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s
		(uuid, name, description, enterprise, status)
		VALUES (?, ?, ?, ?, 1)`, appTableV3)

	_, err = t.Exec(queryStr, appID, app.Name, app.Description, enterpriseID)
	if err != nil {
		return
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return
	}

	return
}

func (controller MYSQLController) UpdateAppV3(enterpriseID string, appID string,
	app *data.AppDetailV3) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET name = ?, description = ?
		WHERE uuid = ?`, appTableV3)
	_, err = controller.connectDB.Exec(queryStr, app.Name, app.Description, appID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	return nil
}

func (controller MYSQLController) DeleteAppV3(enterpriseID string, appID string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	err = deleteAppWithTxV3(appID, t)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}

	return nil
}

func (controller MYSQLController) AppExistsV3(appID string) (bool, error) {
	queryStr := fmt.Sprintf("SELECT 1 FROM %s WHERE uuid = ?", appTableV3)
	return controller.rowExists(queryStr, appID)
}

func (controller MYSQLController) GetGroupsV3(enterpriseID string) ([]*data.GroupDetailV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT uuid, name, status
		FROM %s
		WHERE enterprise = ?`, groupTableV3)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	groups := make([]*data.GroupDetailV3, 0)
	for rows.Next() {
		group := data.GroupDetailV3{}
		err := rows.Scan(&group.ID, &group.Name, &group.Status)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		groups = append(groups, &group)
	}
	rows.Close()

	for _, group := range groups {
		queryStr := fmt.Sprintf(`
			SELECT a.uuid, a.name, a.status
			FROM %s AS ag
			LEFT JOIN %s AS a
			ON a.uuid = ag.app
			WHERE ag.robot_group = ?`, appGroupTableV3, appTableV3)
		rows, err := controller.connectDB.Query(queryStr, group.ID)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		group.Apps = make([]*data.AppV3, 0)

		for rows.Next() {
			app := data.AppV3{}
			err := rows.Scan(&app.ID, &app.Name, &app.Status)
			if err != nil {
				util.LogDBError(err)
				return nil, err
			}

			group.Apps = append(group.Apps, &app)
		}
	}

	return groups, nil
}

func (controller MYSQLController) GetGroupV3(enterpriseID string, groupID string) (*data.GroupDetailV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT g.uuid, g.name, g.status
		FROM %s AS g
		WHERE enterprise = ? AND g.uuid = ?`, groupTableV3)
	row := controller.connectDB.QueryRow(queryStr, enterpriseID, groupID)

	group := data.GroupDetailV3{}
	err = row.Scan(&group.ID, &group.Name, &group.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	queryStr = fmt.Sprintf(`
		SELECT a.uuid, a.name
		FROM %s AS a
		INNER JOIN %s AS ag
		ON ag.app = a.uuid
		WHERE enterprise = ? AND ag.robot_group = ?`, appTableV3, appGroupTableV3)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID, groupID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}

	group.Apps = make([]*data.AppV3, 0)

	for rows.Next() {
		app := data.AppV3{}
		err := rows.Scan(&app.ID, &app.Name)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		group.Apps = append(group.Apps, &app)
	}

	return &group, nil
}

func (controller MYSQLController) AddGroupV3(enterpriseID string, group *data.GroupDetailV3,
	apps []string) (groupID string, err error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)

		return "", err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return "", err
	}
	defer util.ClearTransition(t)

	groupUUID, err := uuid.NewV4()
	if err != nil {
		util.LogDBError(err)
		return
	}
	groupID = hex.EncodeToString(groupUUID[:])

	// Insert machine table entry
	queryStr := fmt.Sprintf("INSERT IGNORE INTO %s (uuid) VALUES (?)", machineTableV3)
	_, err = t.Exec(queryStr, groupID)
	if err != nil {
		util.LogDBError(err)
		return
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s (uuid, name, enterprise, status)
		VALUES (?, ?, ?, 1)`,
		groupTableV3)
	_, err = t.Exec(queryStr, groupID, group.Name, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return
	}

	// Insert app_group table
	queryStr = fmt.Sprintf("INSERT INTO %s (robot_group, app) VALUES (?, ?)", appGroupTableV3)

	// Update app_group
	for _, app := range apps {
		_, err = t.Exec(queryStr, groupID, app)
		if err != nil {
			util.LogDBError(err)
			return
		}
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return
	}

	return
}

func (controller MYSQLController) UpdateGroupV3(enterpriseID string, groupID string,
	group *data.GroupDetailV3, apps []string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf("UPDATE %s SET name = ? WHERE uuid = ?", groupTableV3)
	_, err = t.Exec(queryStr, group.Name, groupID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	// Update app_group table
	queryStr = fmt.Sprintf("DELETE FROM %s WHERE robot_group = ?", appGroupTableV3)
	_, err = t.Exec(queryStr, groupID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	// Update app_group table
	queryStr = fmt.Sprintf("INSERT INTO %s (robot_group, app) VALUES (?, ?)", appGroupTableV3)

	for _, app := range apps {
		_, err = t.Exec(queryStr, groupID, app)
		if err != nil {
			util.LogDBError(err)
			return err
		}
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}

	return nil
}

func (controller MYSQLController) DeleteGroupV3(enterpriseID string, groupID string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	err = deleteGroupWithTxV3(groupID, t)
	if err != nil {
		util.LogDBError(err)
		return nil
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return nil
	}

	return nil
}

func (controller MYSQLController) GroupExistsV3(groupID string) (bool, error) {
	queryStr := fmt.Sprintf("SELECT 1 FROM %s WHERE uuid = ?", groupTableV3)
	return controller.rowExists(queryStr, groupID)
}

func (controller MYSQLController) GetRolesV3(enterpriseID string) ([]*data.RoleV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT id, uuid, name, description
		FROM %s
		WHERE enterprise = ?`, roleTableV3)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	roles := make([]*data.RoleV3, 0)
	for rows.Next() {
		role := data.NewRoleV3()
		err = rows.Scan(&role.ID, &role.UUID, &role.Name, &role.Description)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		roles = append(roles, role)
	}
	rows.Close()

	for _, role := range roles {
		getRoleUserCount(controller, role)
		getRolePrivileges(controller, role)
	}

	return roles, nil
}

func (controller MYSQLController) GetRoleV3(enterpriseID string, roleID string) (*data.RoleV3, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT id, uuid, name, description
		FROM %s
		WHERE enterprise = ? AND uuid = ?`, roleTableV3)
	row := controller.connectDB.QueryRow(queryStr, enterpriseID, roleID)

	role := data.RoleV3{}
	err = row.Scan(&role.ID, &role.UUID, &role.Name, &role.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	getRoleUserCount(controller, &role)
	getRolePrivileges(controller, &role)

	return &role, nil
}

func (controller MYSQLController) AddRoleV3(enterpriseID string, role *data.RoleV3) (roleUUID string, err error) {
	defer func() {
		util.LogTrace.Println("Add role ret uuid: ", roleUUID)
		if err != nil {
			util.LogTrace.Println("Add role ret: ", err.Error())
		}
	}()
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return
	}
	defer util.ClearTransition(t)

	_roleUUID, err := uuid.NewV4()
	if err != nil {
		util.LogDBError(err)
		return
	}
	roleUUID = hex.EncodeToString(_roleUUID[:])

	queryStr := fmt.Sprintf(`
		INSERT INTO %s (uuid, name, enterprise, description)
		VALUES (?, ?, ?, ?)`, roleTable)
	ret, err := t.Exec(queryStr, roleUUID, role.Name, enterpriseID, role.Description)
	if err != nil {
		util.LogDBError(err)
		return
	}

	roleID, err := ret.LastInsertId()
	if err != nil {
		util.LogDBError(err)
		return
	}

	moduleMap := map[string]*data.ModuleDetailV3{}
	modules, err := controller.GetModulesV3(enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return
	}

	for _, module := range modules {
		moduleMap[module.Code] = module
	}

	for priv, cmds := range role.Privileges {
		if module, ok := moduleMap[priv]; ok {
			queryStr = fmt.Sprintf(`
				INSERT INTO %s (role, module, cmd_list)
				VALUES (?, ?, ?)`, rolePrivilegeTable)
			_, err = t.Exec(queryStr, roleID, module.ID, strings.Join(cmds, ","))
			if err != nil {
				util.LogDBError(err)
				return
			}
		}
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return
	}

	return
}

func (controller MYSQLController) UpdateRoleV3(enterpriseID string, roleUUID string, role *data.RoleV3) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	var roleID int
	queryStr := fmt.Sprintf("SELECT id FROM %s WHERE uuid = ?", roleTable)
	err = t.QueryRow(queryStr, roleUUID).Scan(&roleID)
	if err != nil {
		// Check the existence
		if err == sql.ErrNoRows {
			return nil
		}

		util.LogDBError(err)
		return err
	}

	queryStr = fmt.Sprintf(`
		UPDATE %s
		SET name = ?, description = ?
		WHERE enterprise = ? AND id = ?`, roleTable)
	_, err = t.Exec(queryStr, role.Name, role.Description, enterpriseID, roleID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	moduleMap := map[string]*data.ModuleDetailV3{}
	modules, err := controller.GetModulesV3(enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	for _, module := range modules {
		moduleMap[module.Code] = module
	}

	queryStr = fmt.Sprintf(`DELETE FROM %s WHERE role = ?`, rolePrivilegeTable)
	_, err = t.Exec(queryStr, roleID)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	for priv, cmds := range role.Privileges {
		if module, ok := moduleMap[priv]; ok {
			queryStr = fmt.Sprintf(`
				INSERT INTO %s (role, module, cmd_list)
				VALUES (?, ?, ?)`, rolePrivilegeTable)
			_, err = t.Exec(queryStr, roleID, module.ID, strings.Join(cmds, ","))
			if err != nil {
				util.LogDBError(err)
				return err
			}
		}
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}

	return nil
}

func (controller MYSQLController) DeleteRoleV3(enterpriseID string, roleID string) error {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	queryStr := fmt.Sprintf("SELECT id FROM %s WHERE enterprise = ? and uuid = ?", roleTable)
	roleRow := controller.connectDB.QueryRow(queryStr, enterpriseID, roleID)

	var id int
	err = roleRow.Scan(&id)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	t, err := controller.connectDB.Begin()
	if err != nil {
		util.LogDBError(err)
		return err
	}
	defer util.ClearTransition(t)

	err = deleteRoleWithTx(id, roleID, t)
	if err != nil {
		util.LogDBError(err)
		return err
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return err
	}

	return nil
}

func (controller MYSQLController) RoleExistsV3(roleID string) (bool, error) {
	queryStr := fmt.Sprintf("SELECT 1 FROM %s WHERE uuid = ?", roleTableV3)
	return controller.rowExists(queryStr, roleID)
}

func (controller MYSQLController) GetUsersCountOfRoleV3(roleID string) (count int, err error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return
	}

	queryStr := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE role = ?`, userPrivilegesTableV3)
	err = controller.connectDB.QueryRow(queryStr, roleID).Scan(&count)
	return
}

func (controller MYSQLController) GetModulesV3(enterpriseID string) ([]*data.ModuleDetailV3, error) {
	var err error
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT id, code, name, status, description, cmd_list
		FROM %s
		WHERE enterprise = ?`, moduleTableV3)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	modules := make([]*data.ModuleDetailV3, 0)
	for rows.Next() {
		module := data.ModuleDetailV3{}
		var commands string
		err := rows.Scan(&module.ID, &module.Code, &module.Name, &module.Status, &module.Description, &commands)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		module.Commands = strings.Split(commands, ",")
		modules = append(modules, &module)
	}

	sort.Sort(ByIDV3(modules))
	return modules, nil
}

type ByIDV3 []*data.ModuleDetailV3

func (m ByIDV3) Len() int      { return len(m) }
func (m ByIDV3) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m ByIDV3) Less(i, j int) bool {
	switch {
	case m[i] == nil:
		return true
	case m[j] == nil:
		return true
	default:
		return m[i].ID < m[j].ID
	}
}

func (controller MYSQLController) getUserRolesV3(userID string) (roles *data.UserRolesV3,
	err error) {
	// Get user's group roles
	queryStr := fmt.Sprintf(`
		SELECT g.uuid , g.name , r.uuid
		FROM %s AS u
		INNER JOIN %s AS p
		ON p.human = u.uuid
		INNER JOIN %s AS g
		ON p.machine = g.uuid
		INNER JOIN %s AS r
		ON p.role = r.uuid
		WHERE u.uuid = ?
	`, userTableV3, userPrivilegesTableV3, groupTableV3, roleTableV3)
	rows, err := controller.connectDB.Query(queryStr, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	groupRoles := make([]*data.UserGroupRoleV3, 0)
	for rows.Next() {
		groupRole := data.UserGroupRoleV3{}
		err = rows.Scan(&groupRole.ID, &groupRole.Name, &groupRole.Role)
		if err != nil {
			return
		}

		groupRoles = append(groupRoles, &groupRole)
	}
	rows.Close()

	// Get user's app roles
	queryStr = fmt.Sprintf(`
		SELECT a.uuid , a.name , r.uuid
		FROM %s AS u
		INNER JOIN %s AS p
		ON p.human = u.uuid
		INNER JOIN %s AS a
		ON p.machine = a.uuid
		INNER JOIN %s AS r
		ON p.role = r.uuid
		WHERE u.uuid = ?`,
		userTableV3, userPrivilegesTableV3, appTableV3, roleTableV3)
	rows, err = controller.connectDB.Query(queryStr, userID)
	if err != nil {
		return
	}
	defer rows.Close()

	appRoles := make([]*data.UserAppRoleV3, 0)
	for rows.Next() {
		appRole := data.UserAppRoleV3{}
		err = rows.Scan(&appRole.ID, &appRole.Name, &appRole.Role)
		if err != nil {
			return
		}

		appRoles = append(appRoles, &appRole)
	}
	rows.Close()

	roles = &data.UserRolesV3{
		groupRoles,
		appRoles,
	}

	return
}

func addUserPrivilegesWithTxV3(userID string, user *data.UserDetailV3, tx *sql.Tx) error {
	if roles := user.Roles; roles != nil {
		if groupRoles := roles.GroupRoles; groupRoles != nil {
			for _, groupRole := range groupRoles {
				queryStr := fmt.Sprintf(`
					INSERT INTO %s (human, machine, role)
					VALUES (?, ?, ?)`, userPrivilegesTableV3)
				_, err := tx.Exec(queryStr, userID, groupRole.ID, groupRole.Role)
				if err != nil {
					return err
				}
			}
		}

		if appRoles := roles.AppRoles; appRoles != nil {
			for _, appRole := range appRoles {
				queryStr := fmt.Sprintf(`
					INSERT INTO %s (human, machine, role)
					VALUES (?, ?, ?)`, userPrivilegesTableV3)
				_, err := tx.Exec(queryStr, userID, appRole.ID, appRole.Role)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func deleteUserPrivilegesWithTxV3(userID string, tx *sql.Tx) error {
	queryStr := fmt.Sprintf(`
		DELETE FROM %s
		WHERE human = ?`, userPrivilegesTableV3)
	_, err := tx.Exec(queryStr, userID)
	return err
}

func insertCustomInfoWithTxV3(enterpriseID string, userID string,
	customInfo map[string]string, tx *sql.Tx) (err error) {
	queryStr := fmt.Sprintf(`SELECT id, info.column FROM %s AS info WHERE enterprise = ?`, columnTableV3)
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
		queryStr = fmt.Sprintf("INSERT INTO %s (user_id, column_id, value) VALUES (?, ?, ?)",
			userInfoTableV3)
		_, err = tx.Exec(queryStr, param...)
		if err != nil {
			return
		}
	}
	return
}

func addModulesEnterpriseWithTxV3(modules []string, enterpriseID string, t *sql.Tx) (err error) {
	// Copy system level modules for enterprise, set status to 0 by default
	queryStr := fmt.Sprintf(`
		INSERT INTO %s (code, name, description, enterprise, cmd_list, status)
		SELECT code, name, description, ?, cmd_list, 0
		FROM %s
		WHERE enterprise IS NULL AND status = 1`, moduleTableV3, moduleTableV3)
	_, err = t.Exec(queryStr, enterpriseID)
	if err != nil {
		return
	}

	if len(modules) > 0 {
		// Update status flags
		queryStr = fmt.Sprintf(`
			UPDATE %s
			SET status = 1
			WHERE enterprise = ? AND code IN (?%s)`,
			moduleTableV3, strings.Repeat(",?", len(modules)-1))
		queryParams := []interface{}{enterpriseID}
		for _, module := range modules {
			queryParams = append(queryParams, module)
		}
		_, err = t.Exec(queryStr, queryParams...)
		if err != nil {
			return
		}
	}

	return
}

func updateModulesEnterpriseWithTxV3(modules []string, enterpriseID string, t *sql.Tx) (err error) {
	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET status = 1
		WHERE enterprise = ? AND code IN (?)`, moduleTableV3)
	_, err = t.Exec(queryStr, enterpriseID, strings.Join(modules, ","))
	if err != nil {
		return
	}

	queryStr = fmt.Sprintf(`
		UPDATE %s
		SET status = 0
		WHERE enterprise = ? AND code NOT IN (?)`, moduleTableV3)
	_, err = t.Exec(queryStr, enterpriseID, strings.Join(modules, ","))
	if err != nil {
		return
	}

	return
}

func getRoleUserCount(controller MYSQLController, role *data.RoleV3) error {
	queryStr := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE role = ?`, userPrivilegesTableV3)
	err := controller.connectDB.QueryRow(queryStr, role.ID).Scan(&role.UserCount)
	if err != nil {
		return err
	}

	return nil
}

func getRolePrivileges(controller MYSQLController, role *data.RoleV3) error {
	queryStr := fmt.Sprintf(`
		SELECT m.code, p.cmd_list
		FROM %s AS p
		INNER JOIN %s AS m
		ON p.module = m.id
		WHERE p.role = ?`, rolePrivilegeTableV3, moduleTableV3)
	rows, err := controller.connectDB.Query(queryStr, role.ID)
	if err != nil {
		return err
	}

	role.Privileges = make(map[string][]string, 0)
	for rows.Next() {
		var code, cmdList string
		err := rows.Scan(&code, &cmdList)
		if err != nil {
			util.LogDBError(err)
			return err
		}

		role.Privileges[code] = strings.Split(cmdList, ",")
	}
	rows.Close()

	return nil
}

func (controller MYSQLController) rowExists(query string, args ...interface{}) (bool, error) {
	ok, err := controller.checkDB()
	if !ok {
		return false, err
	}

	var exists bool
	queryStr := fmt.Sprintf("SELECT EXISTS (%s)", query)
	err = controller.connectDB.QueryRow(queryStr, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}

func deleteUserWithTxV3(userID string, t *sql.Tx) error {
	// Delete user_privileges table entry
	err := deleteUserPrivilegesWithTxV3(userID, t)
	if err != nil {
		return err
	}

	// TODO: Delete user_info table entry

	queryStr := fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", userTableV3)
	_, err = t.Exec(queryStr, userID)
	if err != nil {
		return err
	}

	// Delete human table entry
	queryStr = fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", humanTableV3)
	_, err = t.Exec(queryStr, userID)
	if err != nil {
		return nil
	}

	return nil
}

func deleteAppWithTxV3(appID string, t *sql.Tx) error {
	// Delete app_group table entry
	queryStr := fmt.Sprintf(`
		DELETE FROM %s
		WHERE app = ?`, appGroupTableV3)
	_, err := t.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		DELETE FROM %s
		WHERE uuid = ?`, appTableV3)
	_, err = t.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	// Delete machine table entry
	queryStr = fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", machineTableV3)
	_, err = t.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	return nil
}

func deleteGroupWithTxV3(groupID string, t *sql.Tx) error {
	// Delete app_group table entry
	queryStr := fmt.Sprintf(`
		DELETE FROM %s
		WHERE robot_group = ?`, appGroupTableV3)
	_, err := t.Exec(queryStr, groupID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		DELETE FROM %s
		WHERE uuid = ?`, groupTableV3)
	_, err = t.Exec(queryStr, groupID)
	if err != nil {
		return err
	}

	// Delete machine table entry
	queryStr = fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", machineTableV3)
	_, err = t.Exec(queryStr, groupID)
	if err != nil {
		return err
	}

	return nil
}

func deleteRoleWithTx(roleID int, roleUUID string, t *sql.Tx) error {
	// Delete privileges table entry
	queryStr := fmt.Sprintf("DELETE FROM %s WHERE role = ?", rolePrivilegeTable)
	_, err := t.Exec(queryStr, roleID)
	if err != nil {
		return err
	}

	// Delete user_priviliges table entry
	queryStr = fmt.Sprintf("DELETE FROM %s WHERE role = ?", userPrivilegesTableV3)
	_, err = t.Exec(queryStr, roleUUID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf("DELETE FROM %s WHERE uuid = ?", roleTable)
	_, err = t.Exec(queryStr, roleUUID)
	if err != nil {
		return err
	}

	return nil
}
