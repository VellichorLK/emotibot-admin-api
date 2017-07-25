package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	const_endpoint_appid_validate      string = "/auth/v1/appid/validate"
	const_endpoint_user_login          string = "/auth/v1/user/login"
	const_endpoint_enterprises         string = "/auth/v1/enterprises"
	const_endpoint_enterprise_register string = "/auth/v1/enterprise/register"

	const_appid_length    int = 32 // md5sum length
	const_type_super_user int = 0
)

// TODO(mike): split handler only to call service.{func}

func SetRoute(c *Configuration) error {
	// TODO(mike): input parameters verification

	log.Printf("config: %s", c)
	dao, err := GetDao(c)
	if err != nil {
		return err
	}
	log.Println(dao)
	http.HandleFunc(const_endpoint_appid_validate, func(w http.ResponseWriter, r *http.Request) {
		appidValidateHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_user_login, func(w http.ResponseWriter, r *http.Request) {
		userLoginHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_enterprise_register, func(w http.ResponseWriter, r *http.Request) {
		EnterpriseRegisterHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_enterprises, func(w http.ResponseWriter, r *http.Request) {
		EnterprisesHandler(w, r, c, dao)
	})
	return nil
}

func appidValidateHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *sql.DB) {
	if HandleHttpMethodError(r.Method, "GET") {
		return
	}
	// get appid
	appid := r.URL.Path[len(const_endpoint_appid_validate):]
	log.Printf("appid: %s", appid)

	// verify appid
	if len(appid) != const_appid_length {
		fmt.Fprint(w, false)
		return
	}

	// (TODO) check cache

	// check database if cache miss
	rows, err := dao.Query("select count(app_id) from appid_list where app_id = ? AND activation = ?", appid, true)
	if HandleHttpError(http.StatusInternalServerError, err, w) {
		return
	}
	var count int
	for rows.Next() {
		rows.Scan(&count)
	}
	log.Println(count)
	if count >= 1 {
		fmt.Fprint(w, true)
	} else {
		fmt.Fprint(w, false)
	}
}

func userLoginHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *sql.DB) {
	if HandleHttpMethodError(r.Method, "POST") {
		return
	}
	// parse param
	err := r.ParseForm()
	if HandleHttpError(http.StatusBadRequest, err, w) {
		return
	}
	user_name := r.FormValue("user_name")
	password := r.FormValue("password")
	log.Printf("user_name: %s, password: %s", user_name, password)
	if len(user_name) == 0 || len(password) == 0 {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}
	// (TODO)check cache

	// return user_id, user_type, enterprise_id, appid
	rows, err := dao.Query("select ul.user_id,ul.user_type,ul.enterprise_id,rl.privilege,rl.role_name from user_list ul left join role_list rl on (ul.role_id = rl.role_id and user_name = ? and password = ?)", user_name, password)
	if HandleError(-1, err, w) {
		return
	}

	for rows.Next() {
		type priv struct {
		}
		type row struct {
			User_id       string         `json:"user_id"`
			User_type     int            `json:"user_type"`
			Enterprise_id string         `json:"enterprise_id"`
			Privilege     interface{}    `json:"privilege"`
			Role_name     sql.NullString `json:"role_name"`
		}
		var r row
		err := rows.Scan(&r.User_id, &r.User_type, &r.Enterprise_id, &r.Privilege, &r.Role_name)
		if HandleError(-2, err, w) {
			return
		}
		RespJson(w, &r)
		return
	}
	http.Error(w, "invalid user", http.StatusForbidden)

}

func EnterpriseRegisterHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *sql.DB) {
	if HandleHttpMethodError(r.Method, "POST") {
		return
	}

	// 1. parse form
	err := r.ParseForm()
	if HandleHttpError(http.StatusBadRequest, err, w) {
		return
	}
	log.Println(r.Form)

	// 2. check required parameters & assign default value
	enterprise_name := r.FormValue("nickName")
	user_name := r.FormValue("account")
	password := r.FormValue("password")
	location := r.FormValue("location")
	people_number := r.FormValue("peopleNumber")
	industry := r.FormValue("industry")
	email := r.FormValue("linkEmail")
	phone := r.FormValue("linkPhone")
	api_cnt := r.FormValue("apiCnt")
	exp_time := r.FormValue("expTime")
	ana_duration := r.FormValue("anaDuration")
	activation := true
	user_type := const_type_super_user

	if len(enterprise_name) == 0 || len(user_name) == 0 || len(password) == 0 {
		HandleHttpError(http.StatusBadRequest, errors.New("lack of required parameters"), w)
		//http.Error(w, "lack of required parameters", http.StatusBadRequest)
		return
	}
	// 3. create enterprise entry, if enterprise exist, goto next stage
	/*
		if enterprise_name in enterprise_list
			goto next stage
		else
			generate app_id and insert into appid_list
			generate enterprise_id and insert into enterprise_list
			recovery if any of above 2 actions failed
	*/
	rows, err := dao.Query("select enterprise_id,app_id from enterprise_list where enterprise_name = ?", enterprise_name)
	if HandleError(-1, err, w) {
		return
	}
	var enterprise_id string
	var app_id string
	// TODO(mike): check if rows count > 2?
	for rows.Next() {
		err := rows.Scan(&enterprise_id, &app_id)
		if HandleError(-2, err, w) {
			// goto ?
			return
		}
		break
	}
	// add app_id entry and enterprise_id entry if need
	if len(enterprise_id) == 0 {
		// add app_id entry
		app_id = GenAppId()
		enterprise_id = GenEnterpriseId()
		stmtIns, err := dao.Prepare("insert into appid_list values(?, ?, ?, ?, ?, ?)")
		defer stmtIns.Close()
		if HandleError(-3, err, w) {
			// goto ?
			return
		}
		i, err := strconv.ParseInt(exp_time, 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)
		_, err = stmtIns.Exec(app_id, time.Now(), api_cnt, tm, ana_duration, activation)
		if HandleError(-4, err, w) {
			// goto ?
			return
		}
		// add enterprise_id entry
		stmtIns2, err2 := dao.Prepare("insert into enterprise_list values(?, ?, ?, ?, ?, ?, ?, ?)")
		defer stmtIns2.Close()
		if HandleError(-5, err2, w) {
			// goto ?
			return
		}
		_, err2 = stmtIns2.Exec(enterprise_id, enterprise_name, time.Now(), industry, phone, location, people_number, app_id)
		if HandleError(-6, err2, w) {
			// goto ?
			return
		}
	}
	// 4. create user entry, 409 if exist
	user_id := GenUserId()
	log.Printf("enterprise_id: %s, app_id: %s, user_id: %s", enterprise_id, app_id, user_id)
	stmtIns, err := dao.Prepare("insert into user_list values(?, ?, ?, ?, ?, ?, ?)") // user_id, user_name, user_type, password, role_id, email, enterprise_id
	defer stmtIns.Close()
	_, err = stmtIns.Exec(user_id, user_name, user_type, password, nil, email, enterprise_id)
	// TODO: deal with conflict
	if HandleError(-7, err, w) {
		// goto
		return
	}
}

// List all enterprises and its appid/login username, password
func EnterprisesHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *sql.DB) {
	if HandleHttpMethodError(r.Method, "POST") {
		return
	}
	// select all from enterprise_list join user_list on enterprise_id
	rows, err := dao.Query("select el.enterprise_id,el.enterprise_name,el.created_time,el.industry,el.phone_number,el.address,el.people_numbers,el.app_id,ul.user_id,ul.user_name,ul.email from enterprise_list el left join user_list ul on el.enterprise_id = ul.enterprise_id")
	if HandleError(-1, err, w) {
		return
	}
	type row struct {
		Enterprise_id   string `json:"enterprise_id"`
		Enterprise_name string `json:"enterprise_name"`
		Created_time    string `json:"created_time"`
		Industry        string `json:"industry"`
		Phone_number    string `json:"phone_number"`
		Address         string `json:"address"`
		People_numbers  int    `json:"people_numbers"`
		App_id          string `json:"app_id"`
		User_id         string `json:"user_id"`
		User_name       string `json:"user_name"`
		Email           string `json:"email"`
	}
	got := []*row{}
	for rows.Next() {
		var r row
		err := rows.Scan(&r.Enterprise_id, &r.Enterprise_name, &r.Created_time, &r.Industry, &r.Phone_number, &r.Address, &r.People_numbers, &r.App_id, &r.User_id, &r.User_name, &r.Email)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(r)
		got = append(got, &r)
	}
	RespJson(w, got)
}
