package UI

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
)

func getUIModules(appid string) ([]*Module, error) {
	var err error

	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	queryStr := "SELECT id, appid, code, url, status FROM ui_modules WHERE appid = ? OR appid = ''"
	rows, err := mySQL.Query(queryStr, appid)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return []*Module{}, nil
		}
		return nil, err
	}

	moduleMap := map[string]*Module{}
	for rows.Next() {
		mod := Module{}
		status := 0
		appid := ""
		err = rows.Scan(&mod.ID, &appid, &mod.Code, &mod.URL, &status)
		if err != nil {
			return nil, err
		}
		mod.Enable = (status != 0)

		if _, ok := moduleMap[mod.Code]; ok {
			// if there existed a custom setting of module with appid, skip default setting
			if appid == "" {
				continue
			}
		}
		moduleMap[mod.Code] = &mod
	}

	ret := []*Module{}
	for _, obj := range moduleMap {
		ret = append(ret, obj)
	}
	return ret, nil
}
