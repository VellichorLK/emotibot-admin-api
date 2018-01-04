package Switch

import (
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func getSwitchList(appid string) ([]*SwitchInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("select OnOff_Id, OnOff_Code, OnOff_Name, OnOff_Status, OnOff_Remark, OnOff_Scenario, OnOff_NumType, OnOff_Num, OnOff_Msg, OnOff_Flow, OnOff_WhiteList, OnOff_BlackList, UpdateTime from %s_onoff", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*SwitchInfo{}
	for rows.Next() {
		var info SwitchInfo
		if err := rows.Scan(&info.ID, &info.Code, &info.Name, &info.Status, &info.Remark, &info.Scenario, &info.NumType, &info.Num, &info.Msg, &info.Flow, &info.WhiteList, &info.BlackList, &info.UpdateTime); err != nil {
			return nil, err
		}
		ret = append(ret, &info)
	}

	return ret, nil
}

func getSwitch(appid string, id int) (*SwitchInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT OnOff_Id, OnOff_Code, OnOff_Name, OnOff_Status, OnOff_Remark, OnOff_Scenario, OnOff_NumType, OnOff_Num, OnOff_Msg, OnOff_Flow, OnOff_WhiteList, OnOff_BlackList, UpdateTime from %s_onoff where OnOff_Id = ?", appid)
	rows, err := mySQL.Query(queryStr, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	queryStr := fmt.Sprintf("UPDATE %s_onoff SET OnOff_Code = ?, OnOff_Name = ?, OnOff_Status = ?, OnOff_Remark = ?, OnOff_Scenario = ?, OnOff_NumType = ?, OnOff_Num = ?, OnOff_Msg = ?, OnOff_Flow = ?, OnOff_WhiteList = ?, OnOff_BlackList = ?, UpdateTime = ? where OnOff_Id = ?", appid)
	_, err := mySQL.Exec(queryStr, input.Code, input.Name, input.Status, input.Remark, input.Scenario, input.NumType, input.Num, input.Msg, input.Flow, input.WhiteList, input.BlackList, input.UpdateTime, id)
	if err != nil {
		return err
	}

	return nil
}

func insertSwitch(appid string, input *SwitchInfo) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

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

	id, _ := res.LastInsertId()
	input.ID = int(id)

	return nil
}

func deleteSwitch(appid string, id int) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("DELETE FROM %s_onoff where OnOff_Id = ?", appid)
	_, err := mySQL.Exec(queryStr, id)
	if err != nil {
		return err
	}

	return nil
}
