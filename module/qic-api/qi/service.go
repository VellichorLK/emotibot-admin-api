package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var (
	serviceDAO model.GroupDAO = &model.GroupSQLDao{}
)

func GetGroups() (total int64, groups []model.GroupWCond, err error) {
	filter := model.GroupFilter{
		Deal: -1,
	}

	total, err = serviceDAO.CountGroupsBy(&filter)
	if err != nil {
		return
	}

	groups, err = serviceDAO.GetGroupsBy(&filter)
	return
}

func CreateGroup(group *model.GroupWCond) (createdGroup *model.GroupWCond, err error) {
	if group == nil || group.Condition == nil {
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

	createdGroup, err = serviceDAO.CreateGroup(group, tx)
	if err != nil {
		return
	}

	serviceDAO.Commit(tx)
	return
}

func GetGroupBy(id int64) (group *model.GroupWCond, err error) {
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
	groups, err = serviceDAO.GetGroupsBy(filter)
	return
}

func UpdateGroup(id int64, gruop *model.GroupWCond) (err error) {
	// tx, err := serviceDAO.Begin()
	// if err != nil {
	// 	return
	// }
	// defer serviceDAO.ClearTranscation(tx)

	// err = serviceDAO.UpdateGroup(id, gruop, tx)
	// if err != nil {
	// 	return
	// }

	// err = tx.Commit()
	return
}

func DeleteGroup(id int64) (err error) {
	err = serviceDAO.DeleteGroup(id)
	return
}
