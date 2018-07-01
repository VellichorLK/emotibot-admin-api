package Robot

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

func GetRobotQAListV3(appid string) ([]*QAInfoV3, int, error) {
	qainfos, err := getRobotQAListV3(appid)
	if err == sql.ErrNoRows {
		return []*QAInfoV3{}, ApiError.SUCCESS, nil
	} else if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return qainfos, ApiError.SUCCESS, nil
}

func GetRobotQAV3(appid string, qid int) (*QAInfoV3, int, error) {
	qainfo, err := getRobotQAV3(appid, qid)
	if err == sql.ErrNoRows {
		return nil, ApiError.NOT_FOUND_ERROR, nil
	} else if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return qainfo, ApiError.SUCCESS, nil
}

func AddRobotQAAnswerV3(appid string, qid int, answer string) (int, int, error) {
	id, err := addRobotQAAnswerV3(appid, qid, answer)
	if err == sql.ErrNoRows {
		return 0, ApiError.NOT_FOUND_ERROR, util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
	} else if err == util.ErrDuplicated {
		return 0, ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["Answer"])
	} else if err != nil {
		return 0, ApiError.DB_ERROR, err
	}
	return id, ApiError.SUCCESS, nil
}

func GetBasicQusetionV3(qid int) (*InfoV3, error) {
	ret, err := getBasicQuestionV3(qid)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetRobotQAAnswerV3(appid string, qid, aid int) (*InfoV3, error) {
	ret, err := getRobotQAAnswerV3(appid, qid, aid)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdateRobotQAAnswerV3(appid string, qid int, aid int, answer string) (int, error) {
	err := updateRobotQAAnswerV3(appid, qid, aid, answer)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err == util.ErrDuplicated {
		return ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["Answer"])
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func DeleteRobotQAAnswerV3(appid string, qid, aid int) (int, error) {
	err := deleteRobotQAAnswerV3(appid, qid, aid)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func AddRobotQARQuestionV3(appid string, qid int, question string) (int, int, error) {
	id, err := addRobotQARQuestionV3(appid, qid, question)
	if err == sql.ErrNoRows {
		return 0, ApiError.NOT_FOUND_ERROR, util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
	} else if err == util.ErrDuplicated {
		return 0, ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["RelateQuestion"])
	} else if err != nil {
		return 0, ApiError.DB_ERROR, err
	}
	return id, ApiError.SUCCESS, nil
}

func GetRobotQARQuestionV3(appid string, qid, rQid int) (*InfoV3, error) {
	ret, err := getRobotQARQuestionV3(appid, qid, rQid)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdateRobotQARQuestionV3(appid string, qid, rQid int, relateQuestion string) (int, error) {
	err := updateRobotQARQuestionV3(appid, qid, rQid, relateQuestion)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err == util.ErrDuplicated {
		return ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["RelateQuestion"])
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func DeleteRobotQARQuestionV3(appid string, qid, rQid int) (int, error) {
	err := deleteRobotQARQuestionV3(appid, qid, rQid)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}
