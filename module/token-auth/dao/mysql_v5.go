package dao

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/satori/go.uuid"
)

func (controller MYSQLController) AddAppV5(enterpriseID string, app *data.AppDetailV5) (appID string, err error) {
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

	var params []interface{}
	sql := fmt.Sprintf(`INSERT INTO %s (app_id, prop_id) VALUES `, appPropsRelTableV5)
	for _, v := range app.Props {
		sql += `(?, ?),`
		params = append(params, appID, v.ID)
	}
	sql = sql[:len(sql)-1]
	logger.Info.Println(sql)

	_, err = t.Exec(sql, params...)
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

func (controller MYSQLController) GetAppsV5(enterpriseID string) ([]*data.AppDetailV5, error) {
	var apps []*data.AppDetailV5
	sql := fmt.Sprintf(`
		SELECT uuid as id, name, status, description, app_type
		FROM %s
		WHERE enterprise = ?`, appTableV3)
	var params []interface{}
	params = append(params, enterpriseID)

	err := controller.queryDBToType(&apps, sql, params...)
	if err != nil {
		return nil, err
	}

	if len(apps) > 0 {
		var appProps data.AppPropRelsV5
		sql = `
			select a.*, b.app_id 
			from %s as a 
			left join %s as b on a.id = b.prop_id
			where ? and b.app_id in ( 
		`
		sql = fmt.Sprintf(sql, appPropsTableV5, appPropsRelTableV5)
		appIds := ""
		for _, v := range apps {
			appIds += fmt.Sprintf(`"%s",`, v.ID)
		}
		sql += appIds[:len(appIds)-1]
		sql += `)`

		var params1 []interface{}
		params1 = append(params1, 1)

		err := controller.queryDBToType(&appProps, sql, params1...)
		if err != nil {
			return nil, err
		}

		logger.Info.Println(appProps)
		if len(appProps) > 0 {
			appProps.GetAppPropsName("zh-cn")
			mapPropsToApps := map[string]data.AppPropsV5{}
			for _, v := range appProps {
				var tmp data.AppPropV5
				tmp.ID = v.ID
				tmp.PKey = v.PKey
				tmp.PValue = v.PValue
				tmp.PName = v.PName
				mapPropsToApps[v.AppId] = append(mapPropsToApps[v.AppId], &tmp)
			}
			logger.Info.Println(mapPropsToApps)

			for k, v := range apps {
				if _, ok := mapPropsToApps[v.ID]; ok {
					apps[k].Props = mapPropsToApps[v.ID]
				}
			}
		}
	}

	return apps, nil
}

func (controller MYSQLController) GetAppV5(enterpriseID string, appID string) (*data.AppDetailV5, error) {
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

	var app data.AppDetailV5
	err = row.Scan(&app.ID, &app.Name, &app.Description, &app.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		util.LogDBError(err)
		return nil, err
	}

	sql := `
		select a.* 
		from %s as a 
		left join %s as b on a.id = b.prop_id
		where b.app_id = ? 
	`
	sql = fmt.Sprintf(sql, appPropsTableV5, appPropsRelTableV5)
	logger.Info.Println("sql: ", sql)

	var params []interface{}
	params = append(params, appID)

	res, err := controller.queryDB(sql, params...)
	if err != nil {
		return nil, err
	}
	logger.Info.Println("res: ", res)

	var appProps data.AppPropsV5
	err = controller.renderDbResultToType(res, &appProps)
	if err != nil {
		return nil, err
	}
	err = appProps.GetAppPropsName("zh-cn")
	if err != nil {
		return nil, err
	}
	logger.Info.Println("appProps: ", appProps)

	app.Props = appProps

	return &app, nil
}

func (controller MYSQLController) AppPropsGetV5(pKey string) ([]*data.AppPropV5, error) {
	sql := `
		select * 
		from app_props 
		where ? 
	`
	var params []interface{}
	params = append(params, 1)

	if len(pKey) > 0 {
		sql += `and p_key = ? `
		params = append(params, pKey)
	}

	res, err := controller.queryDB(sql, params...)
	if err != nil {
		return nil, err
	}

	var appProps data.AppPropsV5
	err = controller.renderDbResultToType(res, &appProps)
	if err != nil {
		return nil, err
	}
	err = appProps.GetAppPropsName("zh-cn")
	if err != nil {
		return nil, err
	}
	logger.Info.Println("appProps: ", appProps)

	return appProps, nil
}
