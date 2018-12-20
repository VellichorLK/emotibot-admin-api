package qi

var (
	serviceDAO DAO = &sqlDAO{}
)

func GetGroups() (groups []Group, err error) {
	groups, err = serviceDAO.GetGroups()
	if err != nil {
		return
	}
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