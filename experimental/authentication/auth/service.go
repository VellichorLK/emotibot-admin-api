package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// enterprise relatived
type EnterpriseUserProp struct {
	EnterpriseId   string `json:"enterprise_id"`
	EnterpriseName string `json:"enterprise_name"`
	CreatedTime    string `json:"created_time"`
	Industry       string `json:"industry"`
	PhoneNumber    string `json:"phone_number"`
	Address        string `json:"address"`
	PeopleNumbers  int    `json:"people_numbers"`
	AppId          string `json:"app_id"`
	UserId         string `json:"user_id"`
	UserName       string `json:"user_name"`
	UserPass       string `json:"user_pass"`
	UserType       int    `json:"user_type"`
	UserEmail      string `json:"user_email"`
}

type AppIdProp struct {
	ApiCnt           string `json:"api_cnt"`
	ExpirationTime   string `json:"exp_time"`
	AnalysisDuration string `json:"ana_duration"`
	Activation       bool   `json:"activation"`
}

type UserLoginProp struct {
	AppId        string         `json:"appid"`
	UserId       string         `json:"user_id"`
	UserName     string         `json:"user_id"`
	UserType     int            `json:"user_type"`
	EnterpriseId string         `json:"enterprise_id"`
	Privilege    interface{}    `json:"privilege"`
	RoleName     sql.NullString `json:"role_name"`
}

// ==================== AppId Series Services ====================
func AppIdValidation(appid string, d *DaoWrapper) (bool, error) {
	if !IsValidAppId(appid) {
		return false, errors.New("invalid appid")
	}

	var count int
	cmd := fmt.Sprintf("select count(app_id) from appid_list where app_id=\"%s\" AND activation=true", appid)
	err := d.QueryRow(cmd).Scan(&count)
	LogInfo.Printf("cmd: %s, count: %d, err: %s,", cmd, count, err)

	if err != nil || count == 0 {
		return false, err
	}
	return true, nil
}

// ==================== User Series Services ====================
func UserLoginValidation(user_name string, password string, d *DaoWrapper) (*UserLoginProp, error) {
	if len(user_name) == 0 || len(password) == 0 {
		return nil, errors.New("invalid parameters")
	}
	cmd := fmt.Sprintf("select el.app_id,ul.user_id,ul.user_type,ul.enterprise_id,rl.privilege,rl.role_name from (select user_id,user_type,enterprise_id,role_id from user_list where user_name=\"%s\" and password=\"%s\") as ul left join role_list rl on (ul.role_id=rl.role_id) left join enterprise_list el on (el.enterprise_id=ul.enterprise_id)", user_name, password)

	var u UserLoginProp
	err := d.QueryRow(cmd).Scan(&u.AppId, &u.UserId, &u.UserType, &u.EnterpriseId, &u.Privilege, &u.RoleName)

	LogInfo.Printf("cmd: %s, err: %s", cmd, err)
	if err != nil {
		return nil, err
	}
	u.UserName = user_name
	return &u, nil
}

// ==================== Enterprise Series Services ====================
func EnterpriseRegister(p *EnterpriseUserProp, a *AppIdProp, d *DaoWrapper) error {
	// TODO(mike) change to begain / end transaction and define rollback
	/*
		if enterprise name in enterprise list
			goto next stage
		else
			create appid
			create enterprise id / user id

	*/
	// TODO(mike) each parameter checking
	if p == nil || d == nil {
		return errors.New("invalid property or dao wrapper")
	}
	LogInfo.Printf("enterprise prop: %s", p)

	// check if existed enterprise_name
	cmd := fmt.Sprintf("select enterprise_id,app_id from enterprise_list where enterprise_name=\"%s\"", p.EnterpriseName)
	var enterprise_id string
	var app_id string
	err := d.QueryRow(cmd).Scan(&enterprise_id, &app_id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	LogInfo.Printf("cmd: %s, enterprise_id: %s, app_id: %s", cmd, enterprise_id, app_id)

	// create enterprise_id & app_id if enterprise_id is empty
	if len(enterprise_id) == 0 {
		app_id = GenAppId()
		enterprise_id = GenEnterpriseId()
		i, err := strconv.ParseInt(a.ExpirationTime, 10, 64)
		if err != nil {
			return err
		}
		exp_timestamp := time.Unix(i, 0)
		var created_time = time.Now()

		// insert appid
		_, err = d.Exec("insert into appid_list values(?, ?, ?, ?, ?, ?)", app_id, created_time, a.ApiCnt, exp_timestamp, a.AnalysisDuration, a.Activation)
		LogInfo.Printf("create app_id: %s, err: %s", app_id, err)
		if err != nil {
			// defer rollback
			return err
		}

		// insert enterprise_id
		_, err = d.Exec("insert into enterprise_list values(?, ?, ?, ?, ?, ?, ?, ?)", enterprise_id, p.EnterpriseName, created_time, p.Industry, p.PhoneNumber, p.Address, p.PeopleNumbers, app_id)
		if err != nil {
			// TODO(mike): defer rollback
			return err
		}
	}

	// deal with user case
	// TODO(mike): deal with conflict case or check user first
	user_id := GenUserId()
	_, err = d.Exec("insert into user_list values(?, ?, ?, ?, ?, ?, ?)", user_id, p.UserName, p.UserType, p.UserPass, nil, p.UserEmail, enterprise_id)
	LogInfo.Printf("insert new user, err: %s", err)
	if err != nil {
		return nil
	}
	return nil
}

func EnterprisesGet(d *DaoWrapper) ([]*EnterpriseUserProp, error) {
	if d == nil {
		return nil, errors.New("dao is nil")
	}
	// select all from enterprise_list join user_list on enterprise_id
	rows, err := d.Query("select el.enterprise_id,el.enterprise_name,el.created_time,el.industry,el.phone_number,el.address,el.people_numbers,el.app_id,ul.user_id,ul.user_name,ul.email from enterprise_list el left join user_list ul on el.enterprise_id = ul.enterprise_id")
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

func EnterpriseGetById(enterprise_id string, d *DaoWrapper) (*EnterpriseUserProp, error) {
	if !IsValidEnterpriseId(enterprise_id) {
		return nil, fmt.Errorf("invalid enterprise_id: %s", enterprise_id)
	}

	if d == nil {
		return nil, errors.New("dao is nil")
	}

	cmd := fmt.Sprintf("select el.enterprise_id,el.enterprise_name,el.created_time,el.industry,el.phone_number,el.address,el.people_numbers,el.app_id,ul.user_id,ul.user_name,ul.email from (select enterprise_id,enterprise_name,created_time,industry,phone_number,address,people_numbers,app_id from enterprise_list where enterprise_id=\"%s\") as el left join user_list ul on el.enterprise_id = ul.enterprise_id", enterprise_id)

	// TODO(mike): should split this logic to static method to reuse by GetEnterprises and GetEnterprisebyId
	var r EnterpriseUserProp
	if err := d.QueryRow(cmd).Scan(&r.EnterpriseId, &r.EnterpriseName, &r.CreatedTime, &r.Industry, &r.PhoneNumber, &r.Address, &r.PeopleNumbers, &r.AppId, &r.UserId, &r.UserName, &r.UserEmail); err != nil {
		LogInfo.Printf("scan failed. %s", err)
		return nil, err
	}

	return &r, nil
}
