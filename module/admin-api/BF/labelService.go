package BF

import (
	"database/sql"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
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

func GetCmdClass(appid string, classID int) (*CmdClass, int, error) {
	class, err := getCmdClass(appid, classID)
	if err == sql.ErrNoRows {
		return nil, ApiError.NOT_FOUND_ERROR, err
	} else if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	if class == nil {
		return nil, ApiError.NOT_FOUND_ERROR, err
	}
	return class, ApiError.SUCCESS, nil
}

func UpdateCmdClass(appid string, classID int, newClassName string) (int, error) {
	err := updateCmdClass(appid, classID, newClassName)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, err
	} else if err == errDuplicate {
		return ApiError.REQUEST_ERROR, err
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func AddCmdClass(appid string, pid *int, className string) (int, int, error) {
	id, err := addCmdClass(appid, pid, className)
	if err == errDuplicate {
		return 0, ApiError.REQUEST_ERROR, err
	} else if err != nil {
		return 0, ApiError.DB_ERROR, err
	}
	return id, ApiError.SUCCESS, nil
}

func DeleteCmdClass(appid string, classID int) error {
	return deleteCmdClass(appid, classID)
}
