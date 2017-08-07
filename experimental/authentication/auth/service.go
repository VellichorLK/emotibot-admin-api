package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// enterprise relatived
type EnterpriseUserProp struct {
	EnterpriseId   string    `json:"enterprise_id"`
	EnterpriseName string    `json:"enterprise_name"`
	CreatedTime    time.Time `json:"created_time"`
	Industry       string    `json:"industry"`
	PhoneNumber    string    `json:"phone_number"`
	Address        string    `json:"address"`
	PeopleNumbers  int       `json:"people_numbers"`
	AppId          string    `json:"app_id"`
	UserId         string    `json:"user_id"`
	UserName       string    `json:"user_name"`
	UserPass       string    `json:"user_pass"`
	UserType       int       `json:"user_type"`
	UserEmail      string    `json:"user_email"`
}

type AppIdProp struct {
	AppId            string    `json:"app_id"`
	ApiCnt           string    `json:"api_cnt"`
	CreatedTime      time.Time `json:"creted_time"`
	ExpirationTime   time.Time `json:"exp_time"`
	AnalysisDuration int       `json:"ana_duration"`
	Activation       bool      `json:"activation"`
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
	c, err := d.GetValidAppIdCount(appid)
	if err != nil {
		LogWarn.Printf("get appid %s failed. %s", appid, err)
		return false, nil
	}
	if c > 0 {
		return true, nil
	}
	return false, nil
}

// ==================== User Series Services ====================
func UserLoginValidation(user_name string, password string, d *DaoWrapper) (*UserLoginProp, error) {
	LogInfo.Printf("user_name: %s, password: %s", user_name, password)
	if len(user_name) == 0 || len(password) == 0 {
		return nil, errors.New("invalid parameters")
	}
	return d.GetUserByName(user_name, password)
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

	ent_prop, err := d.GetEnterpriseByName(p.EnterpriseName)
	if err != nil {
		return err
	}
	LogInfo.Printf("enterprise prop: %s", ent_prop)

	if len(ent_prop.EnterpriseId) == 0 {
		time_now := time.Now()
		a.AppId = GenAppId()
		a.CreatedTime = time_now
		if err := d.AddAppEntry(a); err != nil {
			return err
		}
		p.EnterpriseId = GenEnterpriseId()
		p.AppId = a.AppId
		p.CreatedTime = time_now
		if err := d.AddEnterpriseEntry(p); err != nil {
			// TODO(mike) roll back appid
			return err
		}
	} else {
		p.AppId = ent_prop.AppId
		p.EnterpriseId = ent_prop.EnterpriseId
	}
	LogInfo.Printf("enterprise_id: %s, appid: %s", p.EnterpriseId, p.AppId)
	p.UserId = GenUserId()
	if err := d.AddUserEntry(p); err != nil {
		// TODO(mike): rollback appid, and enterpriseid entries if need
		return err
	}
	return nil
}

func EnterprisesGet(d *DaoWrapper) ([]*EnterpriseUserProp, error) {
	if d == nil {
		return nil, errors.New("dao is nil")
	}
	return d.GetEnterprises()
}

func EnterpriseGetById(enterprise_id string, d *DaoWrapper) (*EnterpriseUserProp, error) {
	if !IsValidEnterpriseId(enterprise_id) {
		return nil, fmt.Errorf("invalid enterprise_id: %s", enterprise_id)
	}

	if d == nil {
		return nil, errors.New("dao is nil")
	}

	return d.GetEnterpriseById(enterprise_id)
}

func EnterpriseDeleteByIds(ent_ids []string, d *DaoWrapper) error {
	// TBD
	return nil
}
