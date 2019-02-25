package feedback

import (
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

var (
	serviceDao Dao
)

// GetReasons is use to get reasons of feedback, which is store in database
func GetReasons(appid string) ([]*Reason, AdminErrors.AdminError) {
	if appid == "" {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError, "Invalid appid paramter")
	}
	if serviceDao == nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, "dao is not inited")
	}

	reasons, err := serviceDao.GetReasons(appid)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return reasons, nil
}

// AddReason will add a feedback reason to the robot with appid, and return the
// reason added
func AddReason(appid string, content string) (*Reason, AdminErrors.AdminError) {
	if appid == "" {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError, "Invalid appid paramter")
	}
	if content == "" {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError, "Invalid content paramter")
	}
	if serviceDao == nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, "dao is not inited")
	}

	id, err := serviceDao.AddReason(appid, content)
	if err != nil {
		if err == ErrDuplicateContent {
			return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
				fmt.Sprintf("%s: %s", err.Error(), content))
		}
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	ret := Reason{
		ID:      id,
		Content: content,
	}
	return &ret, nil
}

// DeleteReason will delete reason of feed with reason id
func DeleteReason(appid string, id int64) AdminErrors.AdminError {
	if appid == "" {
		return AdminErrors.New(AdminErrors.ErrnoRequestError, "Invalid appid paramter")
	}
	if id <= 0 {
		return AdminErrors.New(AdminErrors.ErrnoRequestError, "Invalid id paramter")
	}
	if serviceDao == nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, "dao is not inited")
	}

	err := serviceDao.DeleteReason(appid, id)
	if err != nil {
		if err == ErrIDNotExisted {
			return AdminErrors.New(AdminErrors.ErrnoNotFound, err.Error())
		}
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return nil
}
