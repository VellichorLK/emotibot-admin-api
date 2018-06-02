package Robot

import (
	"database/sql"
	"errors"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

func GetRobotWords(appid string) ([]*ChatInfoV2, int, error) {
	ret, err := getRobotWords(appid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ApiError.NOT_FOUND_ERROR, err
		}
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func GetRobotWord(appid string, id int) (*ChatInfoV2, int, error) {
	ret, err := getRobotWord(appid, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ApiError.NOT_FOUND_ERROR, err
		}
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func UpdateRobotWord(appid string, id int, contents []string) ([]*ChatContentInfoV2, int, error) {
	if appid == "" || contents == nil {
		return nil, ApiError.REQUEST_ERROR, errors.New("Invalid parameter")
	}

	_, err := getRobotWord(appid, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ApiError.NOT_FOUND_ERROR, err
		}
		return nil, ApiError.DB_ERROR, err
	}

	ret, err := updateRobotWord(appid, id, contents)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ApiError.NOT_FOUND_ERROR, err
		}
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func AddRobotWordContent(appid string, id int, content string) (*ChatContentInfoV2, int, error) {
	_, err := getRobotWord(appid, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ApiError.NOT_FOUND_ERROR, err
		}
		return nil, ApiError.DB_ERROR, err
	}

	ret, err := addRobotWordContent(appid, id, content)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ApiError.NOT_FOUND_ERROR, err
		}
		if err == errDuplicate {
			return nil, ApiError.REQUEST_ERROR, err
		}
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func UpdateRobotWordContent(appid string, id int, cid int, content string) (int, error) {
	robotWord, err := getRobotWord(appid, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ApiError.NOT_FOUND_ERROR, err
		}
		return ApiError.DB_ERROR, err
	}

	found := false
	for _, c := range robotWord.Contents {
		if cid == c.ID {
			found = true
			break
		}
	}
	if !found {
		return ApiError.NOT_FOUND_ERROR, errors.New("Content not found")
	}

	err = updateRobotWordContent(appid, id, cid, content)
	if err != nil {
		if err == sql.ErrNoRows {
			return ApiError.NOT_FOUND_ERROR, err
		}
		if err == errDuplicate {
			return ApiError.REQUEST_ERROR, err
		}
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}

func DeleteRobotWordContent(appid string, id int, cid int) (int, error) {
	err := deleteRobotWordContent(appid, id, cid)
	if err != nil {
		if err != sql.ErrNoRows {
			return ApiError.DB_ERROR, err
		}
	}

	return ApiError.SUCCESS, nil
}
