package validate

import (
	"emotibot.com/emotigo/module/admin-api/util"
)

func IsValidAppID(id string) bool {
	db := util.GetAuditDB()
	if db == nil {
		return false
	}

	appsql := "SELECT uuid FROM auth.apps WHERE uuid = ?"
	nums, err := db.Query(appsql, id)
	if err != nil {
		return false
	}
	if !nums.Next() {
		return false
	}
	return len(id) > 0 && HasOnlyNumEngDash(id)
}
