package Switch

import (
	"errors"
	"fmt"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func getSwitchList(appid string) ([]*SwitchInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}
	lastOperation := time.Now()

	queryStr := fmt.Sprintf("select OnOff_Id, OnOff_Code, OnOff_Name, OnOff_Status, OnOff_Remark, OnOff_Scenario, OnOff_NumType, OnOff_Num, OnOff_Msg, OnOff_Flow, OnOff_WhiteList, OnOff_BlackList, UpdateTime from %s_onoff", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("get switch list in getSwitchList took: %s", time.Since(lastOperation))
	lastOperation = time.Now()

	ret := []*SwitchInfo{}
	for rows.Next() {
		var info SwitchInfo
		if err := rows.Scan(&info.ID, &info.Code, &info.Name, &info.Status, &info.Remark, &info.Scenario, &info.NumType, &info.Num, &info.Msg, &info.Flow, &info.WhiteList, &info.BlackList, &info.UpdateTime); err != nil {
			return nil, err
		}
		ret = append(ret, &info)
	}
	util.LogInfo.Printf("create switch list struct in getSwitchList took: %s", time.Since(lastOperation))

	return ret, nil
}

func getSwitch(appid string, id int) (*SwitchInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}
	lastOperation := time.Now()

	queryStr := fmt.Sprintf("SELECT OnOff_Id, OnOff_Code, OnOff_Name, OnOff_Status, OnOff_Remark, OnOff_Scenario, OnOff_NumType, OnOff_Num, OnOff_Msg, OnOff_Flow, OnOff_WhiteList, OnOff_BlackList, UpdateTime from %s_onoff where OnOff_Id = ?", appid)
	rows, err := mySQL.Query(queryStr, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("get switch in getSwitch took: %s", time.Since(lastOperation))

	var info SwitchInfo
	if rows.Next() {
		if err := rows.Scan(&info.ID, &info.Code, &info.Name, &info.Status, &info.Remark, &info.Scenario, &info.NumType, &info.Num, &info.Msg, &info.Flow, &info.WhiteList, &info.BlackList, &info.UpdateTime); err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}

	return &info, nil
}

func updateSwitch(appid string, id int, input *SwitchInfo) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	lastOperation := time.Now()

	queryStr := fmt.Sprintf("UPDATE %s_onoff SET OnOff_Code = ?, OnOff_Name = ?, OnOff_Status = ?, OnOff_Remark = ?, OnOff_Scenario = ?, OnOff_NumType = ?, OnOff_Num = ?, OnOff_Msg = ?, OnOff_Flow = ?, OnOff_WhiteList = ?, OnOff_BlackList = ?, UpdateTime = ? where OnOff_Id = ?", appid)
	_, err := mySQL.Exec(queryStr, input.Code, input.Name, input.Status, input.Remark, input.Scenario, input.NumType, input.Num, input.Msg, input.Flow, input.WhiteList, input.BlackList, input.UpdateTime, id)
	if err != nil {
		return err
	}
	util.LogInfo.Printf("update switch in updateSwitch took: %s", time.Since(lastOperation))


	return nil
}

func insertSwitch(appid string, input *SwitchInfo) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	lastOperation := time.Now()

	queryStat := fmt.Sprintf("INSERT INTO %s_onoff(OnOff_Code, OnOff_Name, OnOff_Status, OnOff_Remark, OnOff_Scenario, OnOff_NumType, OnOff_Num, OnOff_Msg, OnOff_Flow, OnOff_WhiteList, OnOff_BlackList, UpdateTime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)", appid)
	stmt, err := mySQL.Prepare(queryStat)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(input.Code, input.Name, input.Status, input.Remark, input.Scenario, input.NumType, input.Num, input.Msg, input.Flow, input.WhiteList, input.BlackList, input.UpdateTime)
	if err != nil {
		return err
	}
	util.LogInfo.Printf("insert switch in insertSwitch took: %s", time.Since(lastOperation))

	id, _ := res.LastInsertId()
	input.ID = int(id)

	return nil
}

func deleteSwitch(appid string, id int) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	lastOperation := time.Now()

	queryStr := fmt.Sprintf("DELETE FROM %s_onoff where OnOff_Id = ?", appid)
	_, err := mySQL.Exec(queryStr, id)
	if err != nil {
		return err
	}
	util.LogInfo.Printf("delete switch in deleteSwitch took: %s", time.Since(lastOperation))

	return nil
}
