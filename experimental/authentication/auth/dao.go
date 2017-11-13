package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	const_mysql_timeout       string = "10s"
	const_mysql_write_timeout string = "30s"
	const_mysql_read_timeout  string = "30s"
)

type stConsulConfig struct {
	consul_url string
}

type DaoWrapper struct {
	mysql  *sql.DB
	consul *stConsulConfig
}

func DaoInit(db_url string, db_user string, db_pass string, db_name string, consul_url string) (d *DaoWrapper, err error) {
	if len(db_url) == 0 || len(db_user) == 0 || len(db_pass) == 0 || len(db_name) == 0 {
		return nil, errors.New("invalid parameters!")
	}
	LogInfo.Printf("db_url: %s, db_user: %s, db_pass: %s, db_name: %s", db_url, db_user, db_pass, db_name)

	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=true", db_user, db_pass, db_url, db_name, const_mysql_timeout, const_mysql_read_timeout, const_mysql_write_timeout)
	LogInfo.Printf("url: %s", url)

	db, err := sql.Open("mysql", url)
	if err != nil {
		LogInfo.Printf("open db(%s) failed: %s", url, err)
		return nil, err
	}

	// init consul client
	consul := stConsulConfig{consul_url}

	dao := DaoWrapper{db, &consul}
	return &dao, db.Ping()
}

// ===== appid related =====
func (d *DaoWrapper) GetValidAppIdCount(appid string) (int, error) {
	var c int
	LogInfo.Printf("appid: %s", appid)
	fmt.Printf("%s", appid)
	if err := d.mysql.QueryRow("select count(app_id) from appid_list where app_id=? and activation=true", appid).Scan(&c); err != nil {
		fmt.Printf("err: %v", err)
		return 0, err
	}
	return c, nil
}

// ===== user related =====
func (d *DaoWrapper) GetUserByName(user_name string, password string) (*UserLoginProp, error) {
	LogInfo.Printf("user_name: %s, password: %s", user_name, password)

	var u UserLoginProp
	if err := d.mysql.QueryRow("select el.app_id,ul.user_id,ul.user_type,ul.enterprise_id,rl.privilege,rl.role_name from (select user_id,user_type,enterprise_id,role_id from user_list where user_name=? and password=?) as ul left join role_list rl on (ul.role_id=rl.role_id) left join enterprise_list el on (el.enterprise_id=ul.enterprise_id)", user_name, password).Scan(&u.AppId, &u.UserId, &u.UserType, &u.EnterpriseId, &u.Privilege, &u.RoleName); err != nil {
		LogError.Printf("err: %s", err)
		return nil, err
	}
	u.UserName = user_name
	return &u, nil
}

// ===== enterprise related =====
func (d *DaoWrapper) GetEnterpriseByName(enterprise_name string) (*EnterpriseUserProp, error) {
	LogInfo.Printf("enterprise_name: %s", enterprise_name)

	// TODO(mike): enterprise_name should be unique
	var e EnterpriseUserProp
	if err := d.mysql.QueryRow("select enterprise_id,app_id from enterprise_list where enterprise_name=?", enterprise_name).Scan(&e.EnterpriseId, &e.AppId); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	return &e, nil
}

func (d *DaoWrapper) GetEnterpriseById(enterprise_id string) (*EnterpriseUserProp, error) {
	LogInfo.Printf("enterprise_id: %s", enterprise_id)

	// TODO(mike): enterprise_id should be unique
	var e EnterpriseUserProp
	if err := d.mysql.QueryRow("select el.enterprise_id,el.enterprise_name,el.created_time,el.industry,el.phone_number,el.address,el.people_numbers,el.app_id,ul.user_id,ul.user_name,ul.email from (select enterprise_id,enterprise_name,created_time,industry,phone_number,address,people_numbers,app_id from enterprise_list where enterprise_id=?) as el left join user_list ul on el.enterprise_id = ul.enterprise_id", enterprise_id).Scan(&e.EnterpriseId, &e.EnterpriseName, &e.CreatedTime, &e.Industry, &e.PhoneNumber, &e.Address, &e.PeopleNumbers, &e.AppId, &e.UserId, &e.UserName, &e.UserEmail); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	return &e, nil
}

func (d *DaoWrapper) GetEnterprises() ([]*EnterpriseUserProp, error) {
	rows, err := d.mysql.Query("select el.enterprise_id,el.enterprise_name,el.created_time,el.industry,el.phone_number,el.address,el.people_numbers,el.app_id,ul.user_id,ul.user_name,ul.email from enterprise_list el left join user_list ul on el.enterprise_id = ul.enterprise_id and ul.user_type = 0")
	LogInfo.Printf("rows: %s, err: %s", rows, err)
	if err != nil {
		return nil, err
	}

	got := []*EnterpriseUserProp{}
	for rows.Next() {
		var r EnterpriseUserProp
		err = rows.Scan(&r.EnterpriseId, &r.EnterpriseName, &r.CreatedTime, &r.Industry, &r.PhoneNumber, &r.Address, &r.PeopleNumbers, &r.AppId, &r.UserId, &r.UserName, &r.UserEmail)
		if err != nil {
			LogInfo.Println(err)
			continue
		}
		LogInfo.Println(r)
		got = append(got, &r)
	}
	return got, nil
}

// ===== enterprise register related =====
func (d *DaoWrapper) AddAppEntry(a *AppIdProp) error {
	_, err := d.mysql.Exec("insert into appid_list values(?, ?, ?, ?, ?, ?)", a.AppId, a.CreatedTime, a.ApiCnt, a.ExpirationTime, a.AnalysisDuration, a.Activation)
	LogInfo.Printf("add appid: %s, (%s)", a, err)
	return err
}

func (d *DaoWrapper) AddEnterpriseEntry(e *EnterpriseUserProp) error {
	_, err := d.mysql.Exec("insert into enterprise_list values(?, ?, ?, ?, ?, ?, ?, ?)", e.EnterpriseId, e.EnterpriseName, e.CreatedTime, e.Industry, e.PhoneNumber, e.Address, e.PeopleNumbers, e.AppId)
	LogInfo.Printf("add enterprise: %s, (%s)", e, err)
	return err
}

func (d *DaoWrapper) AddUserEntry(e *EnterpriseUserProp) error {
	_, err := d.mysql.Exec("insert into user_list values(?, ?, ?, ?, ?, ?, ?)", e.UserId, e.UserName, e.UserType, e.UserPass, nil, e.UserEmail, e.EnterpriseId)
	LogInfo.Printf("add user: %s, (%s)", e, err)
	return err
}

// ===== enterprise delete related =====
func (d *DaoWrapper) DeleteEnterprise(enterprise_id string) error {
	if rows, err := d.mysql.Exec("delete el,al from enterprise_list as el join appid_list as al on al.app_id=el.app_id where el.enterprise_id=?", enterprise_id); err != nil {
		LogError.Printf("err: %s", err)
		return err
	} else {
		LogInfo.Printf("affect rows: %s", rows)
	}
	if rows, err := d.mysql.Exec("delete from user_list where enterprise_id=?", enterprise_id); err != nil {
		LogError.Printf("err: %s", err)
		return err
	} else {
		LogInfo.Printf("affect rows: %s", rows)
	}
	// TODO(mike): deal with privilege and role
	return nil
}

// ==== admin apis: role related apis =====
func (d *DaoWrapper) GetRoles(enterprise_id string) ([]*RoleProp, error) {
	rows, err := d.mysql.Query("select role_id, role_name, privilege from role_list where enterprise_id = ?", enterprise_id)

	got := []*RoleProp{}
	for rows.Next() {
		var r RoleProp
		err = rows.Scan(&r.RoleId, &r.RoleName, &r.Privilege)
		r.EnterpriseId = enterprise_id
		if err != nil {
			LogInfo.Println(err)
			continue
		}
		got = append(got, &r)
	}
	return got, nil
}

func (d *DaoWrapper) AddRole(enterprise_id string, r *RoleProp) error {
	_, err := d.mysql.Exec("insert into role_list values(?, ?, ?, ?)", r.RoleId, r.RoleName, r.Privilege, enterprise_id)
	LogInfo.Printf("add role: %#v, (%s)", r, err)
	return err
}

func (d *DaoWrapper) GetRoleById(role_id string, enterprise_id string) (*RoleProp, error) {
	var r RoleProp
	if err := d.mysql.QueryRow("select role_id, role_name, privilege from role_list where enterprise_id = ? and role_id = ?", enterprise_id, role_id).Scan(&r.RoleId, &r.RoleName, &r.Privilege); err != nil {
		return nil, err
	}
	LogInfo.Printf("Get role: %#v", r)
	return &r, nil
}

func (d *DaoWrapper) DeleteRole(role_id string, enterprise_id string) error {
	LogInfo.Printf("delete role: %s enterprise: %s", role_id, enterprise_id)
	// 1. Change all user whose role is this, and change them into null
	_, err := d.mysql.Exec("update user_list set role_id = NULL where role_id = ? and enterprise_id = ?", role_id, enterprise_id)
	if err != nil {
		LogInfo.Print("Update user list error")
		return err
	}

	// 2. Remove role in role_list
	_, err = d.mysql.Exec("delete from role_list where role_id = ? and enterprise_id = ?", role_id, enterprise_id)
	return err
}

func (d *DaoWrapper) PatchRole(role_id string, enterprise_id string, r *RoleProp) (*RoleProp, error) {
	params := make([]string, 0)
	if r.RoleName != "" {
		params = append(params, fmt.Sprintf("role_name = '%s'", r.RoleName))
	}
	if r.Privilege != "" {
		params = append(params, fmt.Sprintf("privilege = '%s'", r.Privilege))
	}

	if len(params) == 0 {
		return r, nil
	}

	sets := strings.Join(params, ",")

	sql := fmt.Sprintf("update role_list set %s where role_id = ? and enterprise_id = ?", sets)
	_, err := d.mysql.Exec("update role_list set ? where role_id = ? and enterprise_id = ?", sets, r.RoleId, enterprise_id)

	log := fmt.Sprintf("update role: %#v, (%s), (%#v)\n", r, sql, err)
	LogInfo.Print(log)
	return r, err
}

// ===== admin apis: user management =====
func (d *DaoWrapper) GetUsers(enterprise_id string) ([]*UserProp, error) {
	rows, err := d.mysql.Query("select user_id, user_name, user_type, role_id, email from user_list where enterprise_id = ?", enterprise_id)
	LogInfo.Printf("rows: %#v, err: %s", rows, err)
	if err != nil {
		return nil, err
	}

	got := []*UserProp{}
	for rows.Next() {
		var r UserProp
		err = rows.Scan(&r.UserId, &r.UserName, &r.UserType, &r.RoleId, &r.Email)
		if err != nil {
			LogInfo.Println(err)
			continue
		}
		got = append(got, &r)
	}
	return got, nil
}

func (d *DaoWrapper) GetUserById(user_id string, enterprise_id string) (*UserProp, error) {
	LogInfo.Printf("user_id: %s, enterprise_id: %s", user_id, enterprise_id)

	var r UserProp
	if err := d.mysql.QueryRow("select user_id, user_name, user_type, role_id, email from user_list where enterprise_id = ? and user_id = ?", enterprise_id, user_id).Scan(&r.UserId, &r.UserName, &r.UserType, &r.RoleId, &r.Email); err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DaoWrapper) AddUser(enterprise_id string, r *UserProp) error {
	// if input is empty, insert NULL
	r.RoleId.Valid = r.RoleId.String != ""
	r.Email.Valid = r.Email.String != ""

	_, err := d.mysql.Exec("insert into user_list values(?, ?, ?, ?, ?, ?, ?)", r.UserId, r.UserName, r.UserType, r.Password, r.RoleId, r.Email, enterprise_id)
	LogInfo.Printf("add user: %q, (%s)", r, err)
	return err
}

func (d *DaoWrapper) DeleteUser(enterprise_id string, user_id string) error {
	// If user_id not existed, return success
	user, err := d.GetUserById(user_id, enterprise_id)
	if err != nil {
		return nil
	}

	if user.UserType == const_type_super_user {
		return errors.New("Cannot remove super user")
	}

	_, err = d.mysql.Exec("delete from user_list where user_id = ? and enterprise_id = ?", user_id, enterprise_id)
	LogInfo.Printf("delete user: $q, (%s)", user, err)

	return err
}

func (d *DaoWrapper) PatchUser(user_id string, enterprise_id string, u *UserProp) (*UserProp, error) {
	params := make([]string, 0)
	if u.RoleId.Valid {
		params = append(params, fmt.Sprintf("role_id = '%s'", u.RoleId.String))
	}
	if u.Password != "" {
		params = append(params, fmt.Sprintf("password = '%s'", u.Password))
	}
	if u.Email.Valid {
		params = append(params, fmt.Sprintf("email = '%s'", u.Email.String))
	}

	if len(params) == 0 {
		return u, nil
	}

	sets := strings.Join(params, ",")

	sql := fmt.Sprintf("update user_list set %s where user_id = ? and enterprise_id = ?", sets)
	_, err := d.mysql.Exec("update user_list set ? where user_id = ? and enterprise_id = ?", sets, u.UserId, enterprise_id)

	log := fmt.Sprintf("update user: %#v, (%s), (%#v)\n", u, sql, err)
	LogInfo.Print(log)
	return u, err
}

func (d *DaoWrapper) GetEnterpriseIdByAppId(appid string) (string, error) {
	LogInfo.Printf("appid: %s", appid)
	var enterprise_id string
	if err := d.mysql.QueryRow("select enterprise_id from enterprise_list where app_id = ?", appid).Scan(&enterprise_id); err != nil {
		return "", err
	}
	return enterprise_id, nil
}

func (d *DaoWrapper) GetPrivileges() ([]*PrivilegeProp, error) {
	rows, err := d.mysql.Query("select privilege_id, privilege_name from privilege_list")
	LogInfo.Printf("rows: %#v, err: %s", rows, err)
	if err != nil {
		return nil, err
	}

	got := []*PrivilegeProp{}
	for rows.Next() {
		var p PrivilegeProp
		err = rows.Scan(&p.Id, &p.Name)
		if err != nil {
			LogInfo.Println(err)
			continue
		}
		got = append(got, &p)
	}
	return got, nil
}

// ===== system setting related
func (d *DaoWrapper) AddLogo(enterprise_id string, img_b64_str string) error {
	// {consul-server-port}/v1/kv/voice-emotion/{enterprise_id}/system-setting/logo
	url := fmt.Sprintf("http://%s/v1/kv/%s/system/logo", d.consul.consul_url, enterprise_id)
	LogInfo.Printf("url: %s", url)

	return DoPut(url, img_b64_str)
}

func (d *DaoWrapper) GetLogo(enterprise_id string) (string, error) {
	// {consul-server-port}/v1/kv/voice-emotion/{enterprise_id}/system-setting/logo
	url := fmt.Sprintf("http://%s/v1/kv/%s/system/logo?raw", d.consul.consul_url, enterprise_id)
	LogInfo.Printf("url: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		LogError.Printf("failed to get from %s, %v", url, err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogError.Printf("failed to read from body, %v", err)
		return "", err
	}

	return string(body), nil
}

func (d *DaoWrapper) GetSystemSetting(enterprise_id string) (*SystemProp, error) {
	url := fmt.Sprintf("http://%s/v1/kv/%s/system/setting?raw", d.consul.consul_url, enterprise_id)
	LogInfo.Printf("url: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		LogError.Printf("failed to get from %s, err: %v", url, err)
		return nil, err
	}

	var s SystemProp
	if resp.StatusCode != http.StatusOK {
		LogError.Printf("failed to get from %s, resp: %v", url, resp)
		return &s, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogError.Printf("failed to read from body, %v", err)
		return nil, err
	}
	LogInfo.Printf("body(%d)", len(body))

	if err := json.Unmarshal(body, &s); err != nil {
		LogError.Printf("unmarshal %v failed. %v", body, err)
		return nil, err
	}

	LogInfo.Printf("system prop: %v", s)
	return &s, nil
}

func (d *DaoWrapper) PatchSystemSetting(enterprise_id string, s *SystemProp) (*SystemProp, error) {
	url := fmt.Sprintf("http://%s/v1/kv/%s/system/setting", d.consul.consul_url, enterprise_id)
	LogInfo.Printf("url: %s", url)
	if es, err := json.Marshal(s); err != nil {
		LogError.Printf("marshal failed.(%s)", err)
		return nil, err
	} else {
		if err := DoPut(url, string(es)); err != nil {
			LogError.Printf("put to %s failed. %s", url, err)
			return nil, err
		}
	}
	return d.GetSystemSetting(enterprise_id)
}
