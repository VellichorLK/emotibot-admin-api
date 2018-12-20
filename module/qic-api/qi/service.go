package qi

var (
	serviceDAO DAO = &sqlDAO{}
)

func GetGroups() (groups []Group, err error) {
	groups, err = serviceDAO.GetGroups()
	return
}

func CreateGroup(group *Group) (createdGroup *Group, err error) {
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

func GetGroupBy(id int64) (group *Group, err error) {
	group, err = serviceDAO.GetGroupBy(id)
	if err != nil || group == nil {
		return
	}

	// TODO: set channel name by code
	group.Condition.LeftChannel = "staff"
	group.Condition.RightChannel = "client"
	return
}

func UpdateGroup(id int64, gruop *Group) (err error) {
	tx, err := serviceDAO.Begin()
	if err != nil {
		return
	}
	defer serviceDAO.ClearTranscation(tx)

	err = serviceDAO.UpdateGroup(id, gruop, tx)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}