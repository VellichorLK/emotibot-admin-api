package dao

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	uuid "github.com/satori/go.uuid"
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
