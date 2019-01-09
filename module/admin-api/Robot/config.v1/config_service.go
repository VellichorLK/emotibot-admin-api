package config

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

var dao configDaoInterface

func init() {
	dao = configMySQL{}
}

// GetDefaultConfigs will get configs of system default
func GetDefaultConfigs() ([]*Config, AdminErrors.AdminError) {
	configs, err := dao.GetDefaultConfigs()
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoInitfailed, "Empty system robot config")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return configs, nil
}

// GetConfigs will get all config set with specific appid. If the config is not set,
// use default value of system
func GetConfigs(appid string) ([]*Config, AdminErrors.AdminError) {
	configs, err := dao.GetConfigs(appid)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoInitfailed, "Empty system robot config")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return configs, nil
}

// GetConfig can get value of specific config
func GetConfig(appid, configName string) (*Config, AdminErrors.AdminError) {
	config, err := dao.GetConfig(appid, configName)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoNotFound, "config not found")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return config, nil
}

// SetConfig can set value of specific config
func SetConfig(appid, module, configName, value string) AdminErrors.AdminError {
	err := dao.SetConfig(appid, module, configName, value)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return nil
}

// SetConfigToDefault will set specific config to system default
func SetConfigToDefault(appid, configName string) AdminErrors.AdminError {
	err := dao.SetConfigToDefault(appid, configName)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return nil
}
