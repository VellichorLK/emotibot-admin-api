package v2

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

var dao intentDaoInterface

// GetIntents will get all intents of appid and version with keyword
func GetIntents(appid string, version *int, keyword string) ([]*IntentV2, AdminErrors.AdminError) {
	intents, err := dao.GetIntents(appid, version, keyword)
	if err == sql.ErrNoRows {
		return []*IntentV2{}, nil
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intents, nil
}

// GetIntent will get intent of appid and intentID with keyword
func GetIntent(appid string, intentID int64, keyword string) (*IntentV2, AdminErrors.AdminError) {
	intent, err := dao.GetIntent(appid, intentID, keyword)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoNotFound, "")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intent, nil
}

func AddIntent(appid, name string, positive, negative []string) (*IntentV2, AdminErrors.AdminError) {
	intent, err := dao.AddIntent(appid, name, positive, negative)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, "Add fail")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intent, nil
}

func ModifyIntent(appid string, intentID int64, name string,
	updateSentence []*SentenceV2WithType, deleteSentences []int64) (*IntentV2, AdminErrors.AdminError) {
	err := dao.ModifyIntent(appid, intentID, name, updateSentence, deleteSentences)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoNotFound, "")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	intent, err := dao.GetIntent(appid, intentID, "")
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intent, nil
}

func DeleteIntent(appid string, intentID int64) AdminErrors.AdminError {
	err := dao.DeleteIntent(appid, intentID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return nil
}
