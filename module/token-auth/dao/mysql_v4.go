package dao

import (
	"database/sql"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/lang"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"encoding/hex"
	"fmt"
	"github.com/satori/go.uuid"
	"strconv"
	"strings"
)

// GetOAuthClient will get client info with clientID, if ID is invalid, return nil
func (controller MYSQLController) GetOAuthClient(clientID string) (*data.OAuthClient, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := `
		SELECT secret, redirect_uri, status
		FROM product
		WHERE id = ?`
	row := controller.connectDB.QueryRow(queryStr, clientID)

	status := 0
	ret := data.OAuthClient{ID: clientID}
	err = row.Scan(&ret.Secret, &ret.RedirectURI, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	ret.Active = status > 0

	return &ret, nil
}

func (controller MYSQLController) AddEnterpriseV4(enterprise *data.EnterpriseV3, modules []string,
	adminUser *data.UserDetailV3, dryRun, active bool) (enterpriseID string, err error) {
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

	queryStr := fmt.Sprintf("SELECT user_name, email FROM %s WHERE user_name = ? OR email = ?", userTableV3)
	mail, name := "", ""
	err = t.QueryRow(queryStr, adminUser.UserName, adminUser.Email).Scan(&name, &mail)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if mail == adminUser.Email {
		return "", util.ErrUserEmailExists
	} else if name == adminUser.UserName {
		return "", util.ErrUserNameExists
	}

	if dryRun {
		return "", nil
	}

	adminUserUUID, err := uuid.NewV4()
	if err != nil {
		util.LogDBError(err)
		return
	}
	adminUserID := hex.EncodeToString(adminUserUUID[:])

	// Insert human table entry
	queryStr = fmt.Sprintf("INSERT INTO %s (uuid) VALUES (?)", humanTableV3)
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
		(uuid, name, description, status)
		VALUES (?, ?, ?, ?)`,
		enterpriseTableV3)
	statusInt := 0
	if active {
		statusInt = 1
	}
	_, err = t.Exec(queryStr, enterpriseID, enterprise.Name, enterprise.Description, statusInt)
	if err != nil {
		util.LogDBError(err)
		return
	}

	queryStr = fmt.Sprintf(`
		UPDATE enterprises
		SET secret = concat(md5(concat(now(), uuid)), sha1(rand()))
		WHERE uuid = ?;`)
	_, err = t.Exec(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s
		(uuid, display_name, user_name, email, enterprise, type, password, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, userTableV3)
	_, err = t.Exec(queryStr, adminUserID, adminUser.DisplayName, adminUser.UserName,
		adminUser.Email, enterpriseID, adminUser.Type, adminUser.Password, statusInt)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = addModulesEnterpriseWithTxV3(modules, enterpriseID, t)
	if err != nil {
		util.LogDBError(err)
		return
	}

	err = controller.addBFEnterprise(enterpriseID, enterprise.Name, adminUserID, adminUser.UserName, *adminUser.Password)
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

func (controller MYSQLController) AddAppV4(enterpriseID string, app *data.AppDetailV4) (appID string, err error) {
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

	robotCount, err := controller.GetAppCount(enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return "", err
	}

	limitCount, err := controller.GetRobotLimitPerEnterprise(enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return "", err
	}

	if robotCount >= limitCount {
		return "", util.ErrOperationForbidden
	}

	appUUID, err := uuid.NewV4()
	if err != nil {
		util.LogDBError(err)
		return
	}
	appID = hex.EncodeToString(appUUID[:])

	// Insert machine table entry
	queryStr := fmt.Sprintf("INSERT INTO %s (uuid) VALUES (?)", machineTableV3)
	_, err = t.Exec(queryStr, appID)
	if err != nil {
		return
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s
		(uuid, name, description, enterprise, status, app_type)
		VALUES (?, ?, ?, ?, 1, ?)`, appTableV3)

	_, err = t.Exec(queryStr, appID, app.Name, app.Description, enterpriseID, app.AppType)
	if err != nil {
		return
	}

	err = t.Commit()
	if err != nil {
		util.LogDBError(err)
		return
	}

	_, secretErr := controller.RenewAppSecretV3(appID)
	if secretErr != nil {
		util.LogError.Println("Create app secret fail, auth may need migration")
	}

	return
}
func (controller MYSQLController) GetAppsV4(enterpriseID string) ([]*data.AppDetailV4, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT uuid, name, status, description, app_type
		FROM %s
		WHERE enterprise = ?`, appTableV3)
	rows, err := controller.connectDB.Query(queryStr, enterpriseID)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	defer rows.Close()

	apps := make([]*data.AppDetailV4, 0)
	for rows.Next() {
		app := data.AppDetailV4{}
		err := rows.Scan(&app.ID, &app.Name, &app.Status, &app.Description, &app.AppType)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}
		apps = append(apps, &app)
	}

	return apps, nil
}

func (controller MYSQLController) UpdateEnterpriseStatusV4(enterpriseID string, active bool) (err error) {
	defer func() {
		if err != nil {
			util.LogDBError(err)
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

	statusInt := 0
	if active {
		statusInt = 1
	}
	queryStr := "UPDATE enterprises SET status = ? WHERE uuid = ?"
	_, err = t.Exec(queryStr, statusInt, enterpriseID)
	if err != nil {
		return
	}

	queryStr = "UPDATE users SET status = ? WHERE enterprise = ?"
	_, err = t.Exec(queryStr, statusInt, enterpriseID)
	if err != nil {
		return
	}
	err = t.Commit()
	return
}

func (controller MYSQLController) ActivateEnterpriseV4(enterpriseID string, username string, password string) (err error) {
	defer func() {
		if err != nil {
			util.LogDBError(err)
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

	statusInt := 1
	queryStr := "UPDATE enterprises SET status = ? WHERE uuid = ?"
	_, err = t.Exec(queryStr, statusInt, enterpriseID)
	if err != nil {
		return
	}

	queryStr = "UPDATE users SET status = ? WHERE enterprise = ?"
	_, err = t.Exec(queryStr, statusInt, enterpriseID)
	if err != nil {
		return
	}

	if password != "" {
		targetUser := username
		if username == "" {
			queryStr = "SELECT user_name FROM users WHERE enterprise = ?"
			row := t.QueryRow(queryStr, enterpriseID)
			err = row.Scan(&targetUser)
			if err != nil {
				return err
			}
		}
		queryStr = "UPDATE users SET password = ? WHERE user_name = ? AND enterprise = ?"
		_, err = t.Exec(queryStr, password, targetUser, enterpriseID)
		if err != nil {
			return err
		}
		err = controller.setBFUserPassword(targetUser, password)
	}

	err = t.Commit()
	return
}

// Belongings will be functions which will set data in origin BF2 database

func (controller MYSQLController) addBFEnterprise(id, name, userid, account, password string) error {
	ok, err := controller.checkBFDB()
	if !ok {
		util.LogDBError(err)
		return err
	}

	tx, err := controller.bfDB.Begin()

	queryStr := `
		INSERT INTO api_enterprise
		(id, enterprise_name, account_type, account_status, create_time, modify_time, enterprise_type)
		VALUES
		(?, ?, 2, 0, CURRENT_TIMESTAMP(), CURRENT_TIMESTAMP(), 1)`
	_, err = tx.Exec(queryStr, id, name)
	if err != nil {
		return err
	}

	err = addBFUserWithTx(tx, userid, account, password, id)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func addBFUserWithTx(tx *sql.Tx, userid, account, password, enterprise string) error {
	var err error
	queryStr := ""
	if enterprise == "" {
		queryStr = `
			INSERT INTO api_user
			(UserId, Email, CreatedTime, Password, NickName, Type, Status, UpdatedTime, enterprise_id)
			VALUES
			(?, ?, CURRENT_TIMESTAMP(), ?, ?, 0, 1, CURRENT_TIMESTAMP(), NULL)`
		_, err = tx.Exec(queryStr, userid, account, password, account)
	} else {
		queryStr = `
			INSERT INTO api_user
			(UserId, Email, CreatedTime, Password, NickName, Type, Status, UpdatedTime, enterprise_id)
			VALUES
			(?, ?, CURRENT_TIMESTAMP(), ?, ?, 0, 1, CURRENT_TIMESTAMP(), ?)`
		_, err = tx.Exec(queryStr, userid, account, password, account, enterprise)
	}
	if err != nil {
		return err
	}
	return nil
}

func (controller MYSQLController) setBFUserPassword(username, password string) error {
	ok, err := controller.checkBFDB()
	if !ok {
		util.LogDBError(err)
		return err
	}
	queryStr := "UPDATE api_user SET Password = ? WHERE Email = ?"
	_, err = controller.bfDB.Exec(queryStr, password, username)
	return err
}

func (controller MYSQLController) GetModulesV4(enterpriseID string, isShow int, local string, prefix string) ([]*data.ModuleDetailV4, error) {
	var err error
	var params []interface{}

	//select mc.*
	//from auth.modules_cmds as mc
	//where sort > 0
	//and mc.code in (
	//	select m1.code
	//	from auth.modules as m1
	//	left join (
	//		select m2.code
	//		from auth.modules as m2
	//		where m2.enterprise = "d8fd73954131493ea5b0494e20615908"
	//		and status = 0
	//	) as m3 on m1.code = m3.code
	//	where 1
	//	and m3.code is null
	//	and m1.enterprise is null
	//	and m1.status = 1
	//)
	//ORDER BY mc.parent_id, mc.sort

	sql := `
		select mc.* 
		from modules_cmds as mc 
		where sort > ? 
		and mc.code in (
			select m1.code
			from modules as m1
	`
	if len(enterpriseID) > 0 {
		tmpStr := `
			left outer join (
				select m2.code 
				from modules as m2 
				where m2.enterprise = "%s" 
				and status = 0 
			) as m3 on m1.code = m3.code 
			where 1 
			and m3.code is null
		`
		sql += fmt.Sprintf(tmpStr, enterpriseID)
	} else {
		sql += `
			where 1 
		`
	}
	sql += `
			and m1.enterprise is null
			and m1.status = 1
		)
	`

	if isShow == 1 {
		sql += `
			AND mc.is_show = ?
		`
		params = append(params, 0, isShow)
	} else {
		params = append(params, 0)
	}

	sql += `
		ORDER BY mc.parent_id, mc.sort
	`

	res, err := controller.queryDB(sql, params...)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}

	return controller.processMenuData(res, local, prefix)
}

func (controller MYSQLController) processMenuData(menuRes []map[string]string, locale string, prefix string) ([]*data.ModuleDetailV4, error) {
	if menuRes == nil {
		return nil, nil
	}

	parentMap := map[int]int{}
	var ret []*data.ModuleDetailV4
	for _, v := range menuRes {
		parentId, _ := strconv.Atoi(v["parent_id"])
		cmdId, _ := strconv.Atoi(v["id"])

		m := renderModuleDetail(v, locale, prefix)

		if parentId == 0 {
			ret = append(ret, &m)
			parentMap[cmdId] = len(ret) - 1
		} else {
			if _, ok := parentMap[parentId]; ok {
				ret[parentMap[parentId]].SubCmd = append(ret[parentMap[parentId]].SubCmd, &m)
			}
		}
	}

	return ret, nil
}

func renderModuleDetail(mLine map[string]string, locale string, prefix string) data.ModuleDetailV4 {
	var m data.ModuleDetailV4

	m.ID, _ = strconv.Atoi(mLine["parent_id"])
	m.ParentId, _ = strconv.Atoi(mLine["id"])
	m.ParentCmd = mLine["parent_cmd"]
	m.Code = mLine["code"]
	m.Cmd = mLine["cmd"]
	m.CmdKey = getCmdKey(mLine["code"], mLine["cmd"])
	m.Name = lang.Get(locale, prefix+m.CmdKey)
	m.Sort, _ = strconv.Atoi(mLine["sort"])
	m.Position = mLine["position"]
	m.Icon = mLine["icon"]
	m.Route = mLine["route"]
	m.IsLink, _ = strconv.ParseBool(mLine["is_link"])
	m.IsShow, _ = strconv.ParseBool(mLine["is_show"])
	m.CreateTime = mLine["create_time"]

	return m
}

func getCmdKey(code string, cmd string) string {
	if len(cmd) == 0 {
		return code
	} else {
		return code + "_" + cmd
	}
}

func (controller MYSQLController) GetRolesV4(enterpriseID string) ([]*data.RoleV4, error) {
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

	roles := make([]*data.RoleV4, 0)
	for rows.Next() {
		role := data.NewRoleV4()
		err = rows.Scan(&role.ID, &role.UUID, &role.Name, &role.Description)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		roles = append(roles, role)
	}
	rows.Close()

	for _, role := range roles {
		controller.getRoleUserCountV4(role)
		controller.getRolePrivilegesV4(role)
	}

	return roles, nil
}

func (controller MYSQLController) getRoleUserCountV4(role *data.RoleV4) error {
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

func (controller MYSQLController) getRolePrivilegesV4(role *data.RoleV4) error {
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

	//role.Privileges = make(map[string][]string, 0)
	for rows.Next() {
		var code, cmdList string
		err := rows.Scan(&code, &cmdList)
		if err != nil {
			util.LogDBError(err)
			return err
		}

		tmp := strings.Split(cmdList, ",")
		for _, v := range tmp {
			cmdKey := code + "_" + v
			role.Privileges = append(role.Privileges, cmdKey)
		}
	}
	rows.Close()

	return nil
}

func (controller MYSQLController) GetRoleV4(enterpriseID string, roleID string, userInfo *data.UserDetailV3) (*data.RoleV4, error) {
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

	role := data.RoleV4{}
	err = row.Scan(&role.ID, &role.UUID, &role.Name, &role.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	controller.getRoleUserCountV4(&role)
	controller.getRolePrivilegesV4(&role)
	controller.getMenuV4(enterpriseID, &role, userInfo)

	return &role, nil
}

func (controller MYSQLController) getMenuV4(enterpriseID string, role *data.RoleV4, userInfo *data.UserDetailV3) error {
	role.Menu, _ = controller.GetModulesV4(enterpriseID, 1, "", "auth_")

	if userInfo.Type > 1 {
		for _, v := range role.Menu {
			for _, vv := range v.SubCmd {
				hasPrivilege := false
				for _, pri := range role.Privileges {
					if vv.CmdKey == pri {
						hasPrivilege = true
						break
					}
				}
				if hasPrivilege {
					//	delete menu node
				}
			}
		}
		for _, v := range role.Menu {
			if v.SubCmd == nil {
				//	delete menu node

			}
		}
	}

	return nil
}

func (controller MYSQLController) GetMenuV4(userInfo *data.UserDetailV3, local string, enterpriseID string, appId string) (map[string][]*data.ModuleDetailV4, error) {
	var menus []*data.ModuleDetailV4
	if userInfo.Type < 2 {
		if userInfo.Enterprise == nil && len(enterpriseID) == 0 {
			menus, _ = controller.GetModulesV4("", 0, local, "menu_")
		} else {
			if len(enterpriseID) == 0 {
				menus, _ = controller.GetModulesV4(*userInfo.Enterprise, 0, local, "menu_")
			} else {
				menus, _ = controller.GetModulesV4(enterpriseID, 0, local, "menu_")
			}
		}
	} else {
		//	user -> role -> privileges
		sql := `
		    select p.module, p.cmd_list, m.code
		    from privileges as p 
		    left join roles as r on r.id = p.role
		    left join user_privileges as up on up.role = r.uuid
			left join modules as m on m.id = p.module
	    	where up.human = ?
			and up.machine in (
				select a.uuid
				from apps as a 
				where uuid = ? 
				union 
				select ag.robot_group 
				from app_group as ag 
				where ag.app = ? 
			)
			group by p.module, p.cmd_list
    	`
		var params []interface{}
		params = append(params, userInfo.ID, appId, appId)

		res, err := controller.queryDB(sql, params...)
		if err != nil {
			return nil, err
		}

		moduleCmd := []map[string]string{}
		for _, v := range res {
			cmds := strings.Split(v["cmd_list"], ",")
			fmt.Println(cmds)
			if len(cmds) > 0 {
				for _, vv := range cmds {
					tmp := map[string]string{}
					tmp["code"] = v["code"]
					tmp["cmd"] = vv
					moduleCmd = append(moduleCmd, tmp)
				}
			}
		}

		sql = `
			SELECT * 
			FROM modules_cmds 
			where (parent_id = ? and is_show = ?) 
		`

		for _, v := range moduleCmd {
			sql += fmt.Sprintf("or (code = \"%s\" and cmd = \"%s\") ", v["code"], v["cmd"])
		}

		sql += `
			order by parent_id, sort 
		`
		var params1 []interface{}
		params1 = append(params1, 0, 1)

		res, err = controller.queryDB(sql, params1...)
		if err != nil {
			util.LogDBError(err)
			return nil, err
		}

		var retMenus []*data.ModuleDetailV4
		retMenus, _ = controller.processMenuData(res, local, "menu_")
		for _, v := range retMenus {
			if v.SubCmd != nil {
				menus = append(menus, v)
			}
		}
	}

	positionMenus := make(map[string][]*data.ModuleDetailV4)
	for _, v := range menus {
		positionMenus[v.Position] = append(positionMenus[v.Position], v)
	}

	return positionMenus, nil
}

func (controller MYSQLController) GetEnterpriseMenuV4(userInfo *data.UserDetailV3, locale string) ([]*data.ModuleDetailV4, error) {
	if userInfo.Type != 0 {
		return nil, nil
	}

	sql := `
		select * 
		from modules_cmds
		where cmd = "view"
		and parent_cmd = ""
        and sort > ?
		order by parent_id, sort
	`
	params := make([]interface{}, 1)
	params[0] = 0

	res, err := controller.queryDB(sql, params...)
	if err != nil {
		util.LogDBError(err)
		return nil, err
	}
	fmt.Println(res)

	var ret []*data.ModuleDetailV4
	for _, v := range res {
		m := renderModuleDetail(v, locale, "menu_")
		ret = append(ret, &m)
	}

	return ret, nil
}

func (controller MYSQLController) queryDB(sql string, params ...interface{}) ([]map[string]string, error) {
	var err error
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	rows, err := controller.connectDB.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	cols, _ := rows.Columns()
	// TODO
	//colTypes, _ := rows.ColumnTypes()
	//for _, v := range colTypes {
	//	fmt.Println(v.Name(), v.DatabaseTypeName(), v.ScanType())
	//}

	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}

	ret := []map[string]string{}

	for rows.Next() {
		if err := rows.Scan(scans...); err != nil {
			return nil, err
		}

		row := map[string]string{}

		for k, v := range values {
			key := cols[k]
			row[key] = string(v)
		}
		ret = append(ret, row)
	}
	defer rows.Close()

	return ret, nil
}
