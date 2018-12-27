package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
)

var (
	serviceDAO model.GroupDAO = &model.GroupSQLDao{}
)

func CreateGroup(group *model.GroupWCond) (createdGroup *model.GroupWCond, err error) {
	if group == nil || group.Condition == nil {
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("error while create uuid in CreateGroup, err: %s", err.Error())
		return
	}

	tx, err := serviceDAO.Begin()
	if err != nil {
		return
	}
	defer serviceDAO.ClearTranscation(tx)

	// TODO: set code left channel & right channel
	group.Condition.LeftChannelCode = 0
	group.Condition.RightChannelCode = 1

	group.Enabled = 1
	group.UUID = uuid.String()
	group.UUID = strings.Replace(group.UUID, "-", "", -1)

	createdGroup, err = serviceDAO.CreateGroup(group, tx)
	if err != nil {
		return
	}

	serviceDAO.Commit(tx)
	return
}

func GetGroupBy(id string) (group *model.GroupWCond, err error) {
	group, err = serviceDAO.GetGroupBy(id)
	if err != nil || group == nil {
		return
	}

	// TODO: set channel name by code
	group.Condition.LeftChannel = "staff"
	group.Condition.RightChannel = "client"
	return
}

func GetGroupsByFilter(filter *model.GroupFilter) (total int64, groups []model.GroupWCond, err error) {
	total, err = serviceDAO.CountGroupsBy(filter)
	if err != nil {
		return
	}

	groups, err = serviceDAO.GetGroupsBy(filter)
	return
}

func UpdateGroup(id string, group *model.GroupWCond) (err error) {
	tx, err := serviceDAO.Begin()
	if err != nil {
		return
	}
	defer serviceDAO.ClearTranscation(tx)

	err = serviceDAO.DeleteGroup(id, tx)
	if err != nil {
		return
	}

	group.UUID = id
	_, err = serviceDAO.CreateGroup(group, tx)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func DeleteGroup(id string) (err error) {
	tx, err := serviceDAO.Begin()
	if err != nil {
		return
	}
	defer serviceDAO.ClearTranscation(tx)

	err = serviceDAO.DeleteGroup(id, tx)
	return
}
