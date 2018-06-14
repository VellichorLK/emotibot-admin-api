package BF

import (
	"database/sql"
	"strconv"
)

func GetCmds(appid string) (*CmdClass, error) {
	return getCmds(appid)
}

func GetCmdsOfLabel(appid string, labelID int) ([]*Cmd, error) {
	return getCmdsOfLabel(appid, labelID)
}

func GetCmd(appid string, id int) (*Cmd, error) {
	return getCmd(appid, id)
}

func DeleteCmd(appid string, id int) error {
	err := deleteCmd(appid, id)
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func AddCmd(appid string, cmd *Cmd, cid int) (int, error) {
	return addCmd(appid, cmd, cid)
}

func UpdateCmd(appid string, id int, cmd *Cmd) error {
	return updateCmd(appid, id, cmd)
}

func GetLabelsOfCmd(appid string, cmdID int) ([]*Label, error) {
	labels, err := getLabelsOfCmd(appid, cmdID)
	if err != nil {
		return nil, err
	}
	countMap, err := GetCmdCountOfLabels(appid)
	if err != nil {
		return nil, err
	}
	for _, l := range labels {
		lid, err := strconv.Atoi(l.ID)
		if err != nil {
			return nil, err
		}
		if count, ok := countMap[lid]; ok {
			l.CmdCount = count
		}
	}
	return labels, nil
}

func GetCmdCountOfLabels(appid string) (map[int]int, error) {
	return getCmdCountOfLabels(appid)
}
