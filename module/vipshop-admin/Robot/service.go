package Robot

import (
	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
)

func GetFunctions(appid string) (map[string]*FunctionInfo, int, error) {
	list, err := getFunctionList(appid)
	if err != nil {
		return nil, ApiError.IO_ERROR, err
	}

	return list, ApiError.SUCCESS, nil
}

func UpdateFunction(appid string, function string, newInfo *FunctionInfo) (int, error) {
	infos, code, err := GetFunctions(appid)
	if err != nil {
		return code, err
	}

	infos[function] = newInfo
	err = updateFunctionList(appid, infos)
	if err != nil {
		return ApiError.IO_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func UpdateFunctions(appid string, newInfos map[string]*FunctionInfo) (int, error) {
	err := updateFunctionList(appid, newInfos)
	if err != nil {
		return ApiError.IO_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func GetRobotQuestionCnt(appid string) (int, error) {
	count, err := getAllRobotQACnt(appid)
	if err != nil {
		return 0, err
	}

	return count, err
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
