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
	const_endpoint_roles         string = "/admin/v1/roles"         // GET /admin/v1/roles
	const_endpoint_role_register string = "/admin/v1/role/register" // POST /admin/v1/role/register
	const_endpoint_roles_id      string = "/admin/v1/role/"         // GET|PATCH|DELETE /admin/v1/role/<rold_id>

	// endpoint: general user management
	const_endpoint_users         string = "/admin/v1/users"         // GET /admin/v1/users
	const_endpoint_user_register string = "/admin/v1/user/register" // POST /admin/v1/user/register
	const_endpoint_user_id       string = "/admin/v1/user/"         // GET|PATCH|DELETE /admin/v1/user/user/<user_id>

	// endpoint: privilege list
	const_endpoint_privileges string = "/admin/v1/privileges"

	// endpoint: sys settings

	const_type_super_user   int = 0
	const_type_invalid_user int = -1

	const_type_authorization_header_key string = "Authorization"
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
	// user management
	http.HandleFunc(const_endpoint_user_register, func(w http.ResponseWriter, r *http.Request) {
		UserRegisterHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_users, func(w http.ResponseWriter, r *http.Request) {
		UsersHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_user_id, func(w http.ResponseWriter, r *http.Request) {
		UserIdHandler(w, r, c, dao)
	})
	// role management
	http.HandleFunc(const_endpoint_role_register, func(w http.ResponseWriter, r *http.Request) {
		RoleRegisterHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_roles, func(w http.ResponseWriter, r *http.Request) {
		RolesHandler(w, r, c, dao)
	})
	http.HandleFunc(const_endpoint_roles_id, func(w http.ResponseWriter, r *http.Request) {
		RoleIdHandler(w, r, c, dao)
	})
	// privilege list
	http.HandleFunc(const_endpoint_privileges, func(w http.ResponseWriter, r *http.Request) {
		PrivilegesHandler(w, r, c, dao)
	})

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

	} else {
		if err := r.ParseForm(); err != nil {
			HandleHttpError(http.StatusBadRequest, err, w)
			return
		}
		// TODO(mike)
		// parse EnterpsieUserProp and AppIdProp
		// PATCH
	}
}
func UserRegisterHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"POST"}, w) {
		return
	}

	// 0. Check header if it is valid
	appid := r.Header.Get(const_type_authorization_header_key)
	if appid == "" || !IsValidAppId(appid) {
		HandleHttpError(http.StatusUnauthorized, nil, w)
		return
	}

	// 1. Parse form to UserProp
	user := getUserFromForm(r, true)
	if user == nil {
		HandleHttpError(http.StatusBadRequest, errors.New("Invalid input"), w)
		return
	}

	// check role id existed
	if err := UserRegister(user, appid, dao); err != nil {
		HandleHttpError(http.StatusInternalServerError, err, w)
		return
	}
	HandleSuccess(w, user)
}

// List all users and its appid/login username, password
// GET /auth/v1/users
// return 200 / 200 with error {} / 410
func UsersHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"GET"}, w) {
		return
	}

	// Check header if it is valid
	appid := r.Header.Get(const_type_authorization_header_key)
	if appid == "" || !IsValidAppId(appid) {
		HandleHttpError(http.StatusUnauthorized, nil, w)
		return
	}

	ents, err := UsersGet(dao, appid)
	if err != nil {
		// 200 with error
		HandleError(-1, err, w)
		return
	}
	HandleSuccess(w, ents)
}

// UserIdHandler Handle GET / DELETE / PATCH user_id
// GET /auth/v1/user/<user_id>, return 200 | 404 | 500
// DELETE /auth/v1/user/<user_id>,  return 204 | 404 | 500
// PATCH /auth/v1/user/<user_id>, return 204 | 400 | 404 | 500
func UserIdHandler(w http.ResponseWriter, r *http.Request, c *Configuration, d *DaoWrapper) {
	// request method check
	if HandleHttpMethodError(r.Method, []string{"GET", "DELETE", "PATCH"}, w) {
		return
	}

	// 0. Check header if it is valid
	appid := r.Header.Get(const_type_authorization_header_key)
	if appid == "" || !IsValidAppId(appid) {
		HandleHttpError(http.StatusUnauthorized, nil, w)
		return
	}

	// get enterprise_id
	user_id := r.URL.Path[len(const_endpoint_user_id):]
	if IsValidUserId(user_id) == false {
		HandleHttpError(http.StatusBadRequest, errors.New("invalid parameters"), w)
		return
	}
	LogInfo.Printf("user_id: %s", user_id)

	// apploy get/delte/patch
	if r.Method == "GET" {
		// enterprise_service.get_enterprise(enterprise_id)
		// select all from enterprise_list join user_list on enterprise_id
		ent, err := UserGetById(user_id, appid, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		HandleSuccess(w, ent)
	} else if r.Method == "DELETE" {
		err := UserDeleteById(user_id, appid, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		HandleSuccess(w, nil)
	} else if r.Method == "PATCH" {
		user := getUserFromForm(r, false)
		patchedUser, err := UserPatchById(user_id, appid, user, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		HandleSuccess(w, patchedUser)
	} else {
		HandleHttpError(http.StatusMethodNotAllowed, errors.New("Method not allowed"), w)
	}
}

func getUserFromForm(r *http.Request, checkValue bool) *UserProp {
	if err := r.ParseForm(); err != nil {
		return nil
	}
	LogInfo.Println(r.Form)

	var p UserProp

	p.UserName = r.FormValue("username")
	userType, err := strconv.ParseInt(r.FormValue("type"), 10, 64)
	if err != nil {
		if checkValue {
			LogInfo.Println(err)
			return nil
		}
		p.UserType = const_type_invalid_user
	} else {
		p.UserType = int(userType)
	}
	p.Password = r.FormValue("password")
	p.RoleId.String = r.FormValue("role_id")
	p.Email.String = r.FormValue("email")

	if p.RoleId.String == "" {
		p.RoleId.Valid = false
	} else {
		p.RoleId.Valid = true
	}

	if p.Email.String == "" {
		p.Email.Valid = false
	} else {
		p.Email.Valid = true
	}

	if p.UserName == "" || p.Password == "" {
		if checkValue {
			return nil
		}
	}
	return &p
}

func RoleRegisterHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"POST"}, w) {
		return
	}

	// 0. Check header if it is valid
	appid := r.Header.Get(const_type_authorization_header_key)
	if appid == "" || !IsValidAppId(appid) {
		HandleHttpError(http.StatusUnauthorized, nil, w)
		return
	}

	role := getRoleFromForm(r, true)
	if role == nil {
		HandleHttpError(http.StatusBadRequest, errors.New("Invalid input"), w)
		return
	}

	// check role id existed
	if err := RoleRegister(role, appid, dao); err != nil {
		HandleHttpError(http.StatusInternalServerError, err, w)
		return
	}
	HandleSuccess(w, role)
}

// List all users and its appid/login username, password
// GET /auth/v1/users
// return 200 / 200 with error {} / 410
func RolesHandler(w http.ResponseWriter, r *http.Request, c *Configuration, dao *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"GET"}, w) {
		return
	}

	// 0. Check header if it is valid
	appid := r.Header.Get(const_type_authorization_header_key)
	if appid == "" || !IsValidAppId(appid) {
		HandleHttpError(http.StatusUnauthorized, nil, w)
		return
	}

	ents, err := RolesGet(appid, dao)
	if err != nil {
		// 200 with error
		HandleError(-1, err, w)
		return
	}
	HandleSuccess(w, ents)
}

// UserIdHandler Handle GET / DELETE / PATCH user_id
// GET /auth/v1/user/<user_id>, return 200 | 404 | 500
// DELETE /auth/v1/user/<user_id>,  return 204 | 404 | 500
// PATCH /auth/v1/user/<user_id>, return 204 | 400 | 404 | 500
func RoleIdHandler(w http.ResponseWriter, r *http.Request, c *Configuration, d *DaoWrapper) {
	// request method check
	if HandleHttpMethodError(r.Method, []string{"GET", "DELETE", "PATCH"}, w) {
		return
	}

	// 0. Check header if it is valid
	appid := r.Header.Get(const_type_authorization_header_key)
	if appid == "" || !IsValidAppId(appid) {
		HandleHttpError(http.StatusUnauthorized, nil, w)
		return
	}

	// get role_id
	role_id := r.URL.Path[len(const_endpoint_roles_id):]
	if IsValidRoleId(role_id) == false {
		HandleHttpError(http.StatusBadRequest, errors.New("invalid parameters"), w)
		return
	}
	LogInfo.Printf("role_id: %s", role_id)

	// apploy get/delte/patch
	if r.Method == "GET" {
		// enterprise_service.get_enterprise(enterprise_id)
		// select all from enterprise_list join user_list on enterprise_id
		ent, err := RoleGetById(role_id, appid, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		HandleSuccess(w, ent)
	} else if r.Method == "DELETE" {
		err := RoleDeleteById(role_id, appid, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		HandleSuccess(w, nil)
	} else if r.Method == "PATCH" {
		user := getRoleFromForm(r, false)
		patchedRole, err := RolePatchById(role_id, appid, user, d)
		if err != nil {
			HandleError(-1, err, w)
			return
		}
		HandleSuccess(w, patchedRole)
	} else {
		HandleHttpError(http.StatusMethodNotAllowed, errors.New("Method not allowed"), w)
	}
}

func getRoleFromForm(r *http.Request, checkValue bool) *RoleProp {
	if err := r.ParseForm(); err != nil {
		return nil
	}
	LogInfo.Println(r.Form)

	var role RoleProp

	role.RoleName = r.FormValue("role_name")
	role.Privilege = r.FormValue("privilege")

	if role.RoleName == "" && checkValue {
		return nil
	}

	if role.Privilege == "" && checkValue {
		role.Privilege = "{}"
	}

	return &role
}

func PrivilegesHandler(w http.ResponseWriter, r *http.Request, c *Configuration, d *DaoWrapper) {
	if HandleHttpMethodError(r.Method, []string{"GET"}, w) {
		return
	}

	// 0. Check header if it is valid
	appid := r.Header.Get(const_type_authorization_header_key)
	if appid == "" || !IsValidAppId(appid) {
		HandleHttpError(http.StatusUnauthorized, nil, w)
		return
	}

	privileges, err := PrivilegesGet(appid, d)
	if err != nil {
		// 200 with error
		HandleError(-1, err, w)
		return
	}
	HandleSuccess(w, privileges)
}
