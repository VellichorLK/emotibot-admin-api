package Robot

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

// ==========================================
// Functions for using mysql
// ==========================================
func GetDBFunctions(appid string, version int) ([]*Function, int, error) {
	functions, err := getDBFunctions(appid, version)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return functions, ApiError.SUCCESS, nil
}

func UpdateDBFunction(appid string, function string, active bool, version int) (int, error) {
	_, err := getDBFunction(appid, function, version)
	if err != nil {
		if err == sql.ErrNoRows {
			return ApiError.REQUEST_ERROR, err
		}
		return ApiError.DB_ERROR, err
	}
	_, err = setDBFunctionActiveStatus(appid, function, active, version)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, err
}

func UpdateMultiDBFunction(appid string, active map[string]bool, version int) (int, error) {
	_, err := setDBMultiFunctionActiveStatus(appid, active, version)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, err
}

// ==========================================
// Functions for old method, mount files
// ==========================================
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

func InitRobotFunction(appid string) error {
	return initRobotFunctionData(appid)
}
