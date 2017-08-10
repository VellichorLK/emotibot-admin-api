package auth

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	const_mysql_timeout       string = "10s"
	const_mysql_write_timeout string = "30s"
	const_mysql_read_timeout  string = "30s"
)

type DaoWrapper struct {
	mysql *sql.DB
}

func DaoMysqlInit(db_url string, db_user string, db_pass string, db_name string) (d *DaoWrapper, err error) {
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
	dao := DaoWrapper{db}
	return &dao, db.Ping()
}

// ===== appid related =====
func (d *DaoWrapper) GetValidAppIdCount(appid string) (int, error) {
	var c int
	cmd := fmt.Sprintf("select count(app_id) from appid_list where app_id=\"%s\" and activation=true", appid)
	LogInfo.Printf("appid: %s, cmd: %s", appid, cmd)

	if err := d.mysql.QueryRow(cmd).Scan(&c); err != nil {
		return 0, err
	}
	return c, nil
}

// ===== user related =====
func (d *DaoWrapper) GetUserByName(user_name string, password string) (*UserLoginProp, error) {
	LogInfo.Printf("user_name: %s, password: %s", user_name, password)

	var u UserLoginProp
	cmd := fmt.Sprintf("select el.app_id,ul.user_id,ul.user_type,ul.enterprise_id,rl.privilege,rl.role_name from (select user_id,user_type,enterprise_id,role_id from user_list where user_name=\"%s\" and password=\"%s\") as ul left join role_list rl on (ul.role_id=rl.role_id) left join enterprise_list el on (el.enterprise_id=ul.enterprise_id)", user_name, password)
	if err := d.mysql.QueryRow(cmd).Scan(&u.AppId, &u.UserId, &u.UserType, &u.EnterpriseId, &u.Privilege, &u.RoleName); err != nil {
		// none existed also a kind of error
		//if err == sql.ErrNoRows {
		//	return nil, nil
		//}
		LogError.Printf("cmd: %s, err: %s", cmd, err)
		return nil, err
	}
	u.UserName = user_name
	return &u, nil
}

// ===== enterprise related =====
func (d *DaoWrapper) GetEnterpriseByName(enterprise_name string) (*EnterpriseUserProp, error) {
	cmd := fmt.Sprintf("select enterprise_id,app_id from enterprise_list where enterprise_name=\"%s\"", enterprise_name)
	LogInfo.Printf("enterprise_name: %s, cmd: %s", enterprise_name, cmd)

	// TODO(mike): enterprise_name should be unique
	var e EnterpriseUserProp
	if err := d.mysql.QueryRow(cmd).Scan(&e.EnterpriseId, &e.AppId); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	return &e, nil
}

func (d *DaoWrapper) GetEnterpriseById(enterprise_id string) (*EnterpriseUserProp, error) {
	cmd := fmt.Sprintf("select el.enterprise_id,el.enterprise_name,el.created_time,el.industry,el.phone_number,el.address,el.people_numbers,el.app_id,ul.user_id,ul.user_name,ul.email from (select enterprise_id,enterprise_name,created_time,industry,phone_number,address,people_numbers,app_id from enterprise_list where enterprise_id=\"%s\") as el left join user_list ul on el.enterprise_id = ul.enterprise_id", enterprise_id)
	LogInfo.Printf("enterprise_id: %s, cmd: %s", enterprise_id, cmd)

	// TODO(mike): enterprise_id should be unique
	var e EnterpriseUserProp
	if err := d.mysql.QueryRow(cmd).Scan(&e.EnterpriseId, &e.EnterpriseName, &e.CreatedTime, &e.Industry, &e.PhoneNumber, &e.Address, &e.PeopleNumbers, &e.AppId, &e.UserId, &e.UserName, &e.UserEmail); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	return &e, nil
}

func (d *DaoWrapper) GetEnterprises() ([]*EnterpriseUserProp, error) {
	rows, err := d.mysql.Query("select el.enterprise_id,el.enterprise_name,el.created_time,el.industry,el.phone_number,el.address,el.people_numbers,el.app_id,ul.user_id,ul.user_name,ul.email from enterprise_list el left join user_list ul on el.enterprise_id = ul.enterprise_id")
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
	cmd := fmt.Sprintf("delete el,al from enterprise_list join on appid_list al where el.enterprise_id=\"%s\" and al.app_id=el.app_id", enterprise_id)
	if rows, err := d.mysql.Exec(cmd); err != nil {
		LogErro.Printf("cmd: %s, err: %s", cmd, err)
		return err
	} else {
		LogInfo.Printf("cmd: %s, affect rows: %s", cmd, rows)
	}
	cmd = fmt.Sprintf("delete from user_list where enterprise_id=\"%s\"", enterprise_id)
	if rows, err := d.mysql.Exec(cmd); err != nil {
		LogErro.Printf("cmd: %s, err: %s", cmd, err)
		return err
	} else {
		LogInfo.Printf("cmd: %s, affect rows: %s", cmd, rows)
	}
	// TODO(mike): deal with privilege and role
	return nil
}

// ==== admin apis: role related apis =====
func (d *DaoWrapper) GetRoles(enterprise_id string) ([]*RoleProp, error) {
	return nil, nil
}

func (d *DaoWrapper) GetRoleById(enterprise_id string, role_id string) (*RoleProp, error) {
	return nil, nil
}

func (d *DaoWrapper) AddRole(enterprie_id string, r *RoleProp) error {
	return nil
}

func (d *DaoWrapper) DeleteRole(enterprise_id string, role_id string) error {
	return nil
}

func (d *DaoWrapper) PatchRole(enterprise_id string, r *RoleProp) error {
	return nil
}

// ===== admin apis: user management =====
func (d *DaoWrapper) GetUsers(enterprise_id string) ([]*UserLoginProp, error) {
	return nil, nil
}

func (d *DaoWrapper) GetUserbyId(enterprise_id string, role_id string) (*UserLoginProp, error) {
	return nil, nil
}

func (d *DaoWrapper) AddUser(enterprie_id string, r *UserLoginProp) error {
	return nil
}

func (d *DaoWrapper) DeleteUser(enterprise_id string, user_id string) error {
	return nil
}

func (d *DaoWrapper) PatchUser(user_id string, r *UserLoginProp) error {
	return nil
}
