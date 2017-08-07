package auth

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

const (
	// endpoint: auth
	const_endpoint_appid_validate string = "/auth/v1/appid/validate"
	const_endpoint_user_login     string = "/auth/v1/user/login"

	// endpoint: enterprise management
	const_endpoint_enterprises         string = "/auth/v1/enterprises"         // GET /auth/v1/enterprises
	const_endpoint_enterprise_register string = "/auth/v1/enterprise/register" // POST /auth/v1/enterprise/register
	const_endpoint_enterprise_id       string = "/auth/v1/enterprise/"         // GET|PATCH|DELETE /auth/v1/enterprise/<enterprise_id>

	// endpoint: role management
	// endpoint: general user management
	// endpoint: sys settings

	const_type_super_user int = 0
)

// TODO(mike): split handler only to call service.{func}

func SetRoute(c *Configuration) error {
	// TODO(mike): input parameters verification

	LogInfo.Printf("config: %s", c)
	// dao, err := GetDao(c.DbUrl, c.DbUser, c.DbPass, c.DbName)
	dao, err := DaoMysqlInit(c.DbUrl, c.DbUser, c.DbPass, c.DbName)
	if err != nil {
		return err
	}
	LogInfo.Println(dao)
	// auth management
	http.HandleFunc(const_endpoint_appid_validate, func(w http.ResponseWriter, r *http.Request) {
		appidValidateHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_user_login, func(w http.ResponseWriter, r *http.Request) {
		userLoginHandler(w, r, c, dao)
	})
	// enterprise management
	http.HandleFunc(const_endpoint_enterprise_register, func(w http.ResponseWriter, r *http.Request) {
		EnterpriseRegisterHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_enterprises, func(w http.ResponseWriter, r *http.Request) {
		EnterprisesHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_enterprise_id, func(w http.ResponseWriter, r *http.Request) {
		EnterpriseIdHandler(w, r, c, dao)
	})
	// role management
	LogInfo.Printf("Set route OK.")
	return nil
}

// return 200 OK: pass
// return 403 Forbidden: invalid appid
// reeturn 500 InternalServerError: database query failure
func appidValidateHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"GET"}, w) {
		return
	}

	// get appid
	appid := r.Header.Get("Authorization")
	LogInfo.Printf("appid: %s", appid)

	// TODO(mike): check cache

	if valid, err := AppIdValidation(appid, dao); err != nil || !valid {
		if !valid {
			HandleHttpError(http.StatusForbidden, errors.New("Forbidden"), w)
		} else {
			HandleHttpError(http.StatusInternalServerError, err, w)
		}
	}
}

// return 200 with user infor. or 403
func userLoginHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"POST"}, w) {
		return
	}

	// parse param
	if err := r.ParseForm(); err != nil {
		HandleHttpError(http.StatusBadRequest, err, w)
		return
	}
	LogInfo.Println(r.Form)

	user_name := r.FormValue("user_name")
	password := r.FormValue("password")
	LogInfo.Printf("user_name: %s, password: %s", user_name, password)

	u, err := UserLoginValidation(user_name, password, dao)
	if err != nil {
		HandleHttpError(http.StatusForbidden, err, w)
		return
	}
	RespJson(w, &u)
}

func EnterpriseRegisterHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"POST"}, w) {
		return
	}

	// 1. parse form
	if err := r.ParseForm(); err != nil {
		HandleHttpError(http.StatusBadRequest, err, w)
		return
	}
	LogInfo.Println(r.Form)

	var p EnterpriseUserProp
	var a AppIdProp

	// 2. check required parameters & assign default value
	p.EnterpriseName = r.FormValue("nickName")
	p.UserName = r.FormValue("account")
	p.UserPass = r.FormValue("password")
	p.Address = r.FormValue("location")
	p.PeopleNumbers, _ = strconv.Atoi(r.FormValue("peopleNumber"))
	p.Industry = r.FormValue("industry")
	p.UserEmail = r.FormValue("linkEmail")
	p.PhoneNumber = r.FormValue("linkPhone")
	p.UserType = const_type_super_user
	a.ApiCnt = r.FormValue("apiCnt")
	exp_time, _ := strconv.ParseInt(r.FormValue("expTime"), 10, 64)
	a.ExpirationTime = time.Unix(exp_time, 0)
	a.AnalysisDuration, _ = strconv.Atoi(r.FormValue("anaDuration"))
	a.Activation = true

	if err := EnterpriseRegister(&p, &a, dao); err != nil {
		HandleHttpError(http.StatusInternalServerError, err, w)
		return
	}
}

// List all enterprises and its appid/login username, password
// GET /auth/v1/enterprises
// return 200 / 200 with error {} / 410
func EnterprisesHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"GET"}, w) {
		return
	}
	ents, err := EnterprisesGet(dao)
	if err != nil {
		// 200 with error
		HandleError(-1, err, w)
		return
	}
	RespJson(w, ents)
}

// Handle GET / DELETE / PATCH enterprise_id
// GET /auth/v1/enterprise/<enterprise_id>, return 200 | 404 | 500
// DELETE /auth/v1/enterprise/<enterprise_id>,  return 204 | 404 | 500
// PATCH /auth/v1/enterprise/<enterprise_id>, return 204 | 400 | 404 | 500
func EnterpriseIdHandler(w http.ResponseWriter, r *http.Request, c *Configuration, d *DaoWrapper) {
	// request method check
	if HandleHttpMethodError(r.Method, []string{"GET", "DELETE", "PATCH"}, w) {
		return
	}

	// get enterprise_id
	enterprise_id := r.URL.Path[len(const_endpoint_enterprise_id):]
	if IsValidEnterpriseId(enterprise_id) == false {
		HandleHttpError(http.StatusBadRequest, errors.New("invalid parameters"), w)
		return
	}
	LogInfo.Printf("enterprise_id: %s", enterprise_id)

	// apploy get/delte/patch
	if r.Method == "GET" {
		// enterprise_service.get_enterprise(enterprise_id)
		// select all from enterprise_list join user_list on enterprise_id
		ent, err := EnterpriseGetById(enterprise_id, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		RespJson(w, ent)
	} else if r.Method == "DELETE" {
		err := EnterpriseDeleteByIds([]string{enterprise_id}, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		// TODO(mike)
		// delete appid_list where enterprise_id=enterprise_id
		// delete user_list where enterprise_id=enterprise_id
		// delete enterprise_list
	} else {
		// TODO(mike)
		// PATCH
	}
}
