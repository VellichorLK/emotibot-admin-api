package Robot

import (
	"errors"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

func GetRobotQuestionCnt(appid string) (int, error) {
	count, err := getAllRobotQACnt(appid)
	if err != nil {
		return 0, err
	}

	return count, err
}

func GetRobotQA(appid string, id int) (*QAInfo, int, error) {
	ret, err := getRobotQA(appid, id)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return ret, ApiError.SUCCESS, err
}

func GetRobotQAList(appid string) (*RetQAInfo, int, error) {
	list, err := getAllRobotQAList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	ret := RetQAInfo{
		Count: len(list),
		Infos: list,
	}

	return &ret, ApiError.SUCCESS, err
}

func GetRobotQAPage(appid string, page int, listPerPage int) (*RetQAInfo, int, error) {
	list, err := getRobotQAListPage(appid, page, listPerPage)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	count, err := getAllRobotQACnt(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	ret := RetQAInfo{
		Count: count,
		Infos: list,
	}

	return &ret, ApiError.SUCCESS, err
}

func UpdateRobotQA(appid string, id int, info *QAInfo) (int, error) {
	err := updateRobotQA(appid, id, info)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}

func GetRobotChatInfoList(appid string) ([]*ChatDescription, int, error) {
	ret, err := getRobotChatInfoList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func GetRobotChat(appid string, id int) (*ChatInfo, int, error) {
	ret, err := getRobotChat(appid, id)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func GetRobotChatList(appid string) ([]*ChatInfo, int, error) {
	ret, err := getRobotChatList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func GetMultiRobotChat(appid string, id []int) ([]*ChatInfo, int, error) {
	ret, err := getMultiRobotChat(appid, id)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return ret, ApiError.SUCCESS, nil
}

func UpdateMultiChat(appid string, input []*ChatInfoInput) (int, error) {
	if len(input) <= 0 {
		return ApiError.REQUEST_ERROR, errors.New("Invalid request")
	}

	err := updateMultiRobotChat(appid, input)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}
