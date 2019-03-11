package util

import (
	"strconv"
	"strings"
)

const (
	mysqlSQLURLKey      = "ADMIN_AUTH_MYSQL_URL"
	mysqlSQLUserKey     = "ADMIN_AUTH_MYSQL_USER"
	mysqlSQLPasswordKey = "ADMIN_AUTH_MYSQL_PASS"
	mysqlSQLDatabaseKey = "ADMIN_AUTH_MYSQL_DB"

	mysqlAuditURLKey      = "ADMIN_AUTH_AUDIT_MYSQL_URL"
	mysqlAuditUserKey     = "ADMIN_AUTH_AUDIT_MYSQL_USER"
	mysqlAuditPasswordKey = "ADMIN_AUTH_AUDIT_MYSQL_PASS"
	mysqlAuditDatabaseKey = "ADMIN_AUTH_AUDIT_MYSQL_DB"

	mysqlBFURLKey      = "ADMIN_AUTH_BF_MYSQL_URL"
	mysqlBFUserKey     = "ADMIN_AUTH_BF_MYSQL_USER"
	mysqlBFPasswordKey = "ADMIN_AUTH_BF_MYSQL_PASS"
	mysqlBFDatabaseKey = "ADMIN_AUTH_BF_MYSQL_DB"

	ssoTypeKey        = "ADMIN_AUTH_SSO_TYPE"
	ssoValidateURLKey = "ADMIN_AUTH_SSO_VALIDATE"
	ssoLoginURLKey    = "ADMIN_UI_SSO_LOGIN_URL"
	ssoLogoutURLKey   = "ADMIN_UI_SSO_LOGOUT_URL"

	serverPortKey = "ADMIN_AUTH_PORT"
	serverURLKey  = "ADMIN_AUTH_URL"

	authUseCaptchaKey = "ADMIN_UI_USE_CAPTCHA"

	jwtTimeoutKey = "ADMIN_AUTH_TIMEOUT"
)

// SSOConfig is used to store config about sso
type SSOConfig struct {
	SSOType     string
	ValidateURL string
	LoginURL    string
	LogoutURL   string
}

// IsEnable is used to check SSO is enable or not
func (config SSOConfig) IsEnable() bool {
	return config.ValidateURL != "" && config.LoginURL != "" && config.LogoutURL != ""
}

// GetSSOConfig is used to get config of sso from env
func GetSSOConfig() *SSOConfig {
	ret := SSOConfig{
		SSOType:     GetStrEnv(ssoTypeKey, ""),
		ValidateURL: GetStrEnv(ssoValidateURLKey, ""),
		LoginURL:    GetStrEnv(ssoLoginURLKey, ""),
		LogoutURL:   GetStrEnv(ssoLogoutURLKey, ""),
	}
	if ret.SSOType == "" {
		return nil
	}
	return &ret
}

// GetMySQLConfig will get db init config from env
func GetMySQLConfig() (url string, port int, user string, password string, database string) {
	envURL := GetStrEnv(mysqlSQLURLKey, "localhost:3306")
	params := strings.Split(envURL, ":")
	if len(params) <= 1 {
		url = params[0]
		port = 3306
	} else {
		url = params[0]
		port, _ = strconv.Atoi(params[1])
	}
	user = GetStrEnv(mysqlSQLUserKey, "root")
	password = GetStrEnv(mysqlSQLPasswordKey, "123456")
	database = GetStrEnv(mysqlSQLDatabaseKey, "auth")

	//envURL := GetStrEnv(mysqlSQLURLKey, "10.10.10.86:3306")
	//params := strings.Split(envURL, ":")
	//if len(params) <= 1 {
	//	url = params[0]
	//	port = 3306
	//} else {
	//	url = params[0]
	//	port, _ = strconv.Atoi(params[1])
	//}
	//user = GetStrEnv(mysqlSQLUserKey, "root")
	//password = GetStrEnv(mysqlSQLPasswordKey, "emotibot")
	//database = GetStrEnv(mysqlSQLDatabaseKey, "auth_hx")
	return
}

// GetAuditMySQLConfig will get audit db init config from env
func GetAuditMySQLConfig() (url string, port int, user string, password string, database string) {
	envURL := GetStrEnv(mysqlAuditURLKey, "127.0.0.1:3306")
	params := strings.Split(envURL, ":")
	if len(params) <= 1 {
		url = params[0]
		port = 3306
	} else {
		url = params[0]
		port, _ = strconv.Atoi(params[1])
	}
	user = GetStrEnv(mysqlAuditUserKey, "root")
	password = GetStrEnv(mysqlAuditPasswordKey, "password")
	database = GetStrEnv(mysqlAuditDatabaseKey, "emotibot")
	return
}

// GetBFMySQLConfig will get BF db init config from env
func GetBFMySQLConfig() (url string, port int, user string, password string, database string) {
	envURL := GetStrEnv(mysqlBFURLKey, "172.17.0.1:3306")
	params := strings.Split(envURL, ":")
	if len(params) <= 1 {
		url = params[0]
		port = 3306
	} else {
		url = params[0]
		port, _ = strconv.Atoi(params[1])
	}
	user = GetStrEnv(mysqlBFUserKey, "root")
	password = GetStrEnv(mysqlBFPasswordKey, "password")
	database = GetStrEnv(mysqlBFDatabaseKey, "emotibot")
	return
}

// GetServerConfig will get server binding config from env
func GetServerConfig() (url string, port int) {
	port = GetIntEnv(serverPortKey, 8088)
	url = GetStrEnv(serverURLKey, "0.0.0.0")
	return
}

// GetJWTExpireTimeConfig will get timeout in jwt token
func GetJWTExpireTimeConfig() (timeout int) {
	timeout = GetIntEnv(jwtTimeoutKey, 3600)
	return
}

// GetCaptchaStatus will get status of captcha in auth system
func GetCaptchaStatus() bool {
	status := GetStrEnv(authUseCaptchaKey, "0")
	status = strings.ToLower(status)
	return status == "1" || status == "true"
}
