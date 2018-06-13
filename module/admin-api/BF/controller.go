package BF

import (
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

// This module will do some dirty work for BF compatible...
func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "bf",
		EntryPoints: []util.EntryPoint{
			// id must same with token-auth
			// id, name
			util.NewEntryPoint("POST", "enterprise", []string{}, handleAddEnterprise),
			// id
			util.NewEntryPoint("DELETE", "enterprise/{id}", []string{}, handleDeleteEnterprise),

			// id must same with token-auth
			// appid 	userid 	name
			util.NewEntryPoint("POST", "app", []string{}, handleAddApp),
			// appid
			util.NewEntryPoint("DELETE", "app/{id}", []string{}, handleDeleteApp),

			// id must same with token-auth
			// userid account password enterprise
			util.NewEntryPoint("POST", "user", []string{}, handleAddUser),
			// role
			util.NewEntryPoint("PUT", "user/{id}", []string{}, handleUpdateUserPassword),
			// role
			util.NewEntryPoint("PUT", "user/{id}/role", []string{}, handleUpdateUserRole),
			// userid
			util.NewEntryPoint("DELETE", "user/{id}", []string{}, handleDeleteUser),

			// use name as uuid of role in token-auth
			// id commands
			util.NewEntryPoint("POST", "role", []string{}, handleAddRole),
			// id commands
			util.NewEntryPoint("PUT", "role/{id}", []string{}, handleUpdateRole),
			// id
			util.NewEntryPoint("DELETE", "role/{id}", []string{}, handleDeleteRole),

			// appid
			util.NewEntryPoint("POST", "ssm-data", []string{}, handleInitSSM),
		},
	}
}
func handleAddEnterprise(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	name := r.FormValue("name")

	err := addEnterprise(id, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleDeleteEnterprise(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	err := deleteEnterprise(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleAddApp(w http.ResponseWriter, r *http.Request) {
	appid := r.FormValue("appid")
	userid := r.FormValue("userid")
	name := r.FormValue("name")

	err := addApp(appid, userid, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	appid := r.FormValue("appid")

	err := deleteApp(appid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleAddUser(w http.ResponseWriter, r *http.Request) {
	userid := r.FormValue("id")
	account := r.FormValue("account")
	password := r.FormValue("password")
	enterprise := r.FormValue("enterprise")

	err := addUser(userid, account, password, enterprise)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userid := util.GetMuxVar(r, "id")
	err := deleteUser(userid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleAddRole(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("id")
	commandStr := strings.TrimSpace(r.FormValue("commands"))

	commands := []string{}
	if commandStr != "" {
		commands = strings.Split(commandStr, ",")
	}

	err := addRole(uuid, commands)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	uuid := util.GetMuxVar(r, "id")
	commandStr := strings.TrimSpace(r.FormValue("commands"))

	commands := []string{}
	if commandStr != "" {
		commands = strings.Split(commandStr, ",")
	}

	err := updateRole(uuid, commands)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	uuid := util.GetMuxVar(r, "id")

	err := deleteRole(uuid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userid := util.GetMuxVar(r, "id")
	roleid := r.FormValue("role")
	enterprise := r.FormValue("enterprise")

	err := updateUserRole(enterprise, userid, roleid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	userid := util.GetMuxVar(r, "id")
	password := r.FormValue("password")
	enterprise := r.FormValue("enterprise")

	err := updateUserPassword(enterprise, userid, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func handleInitSSM(w http.ResponseWriter, r *http.Request) {
	appid := r.FormValue("appid")
	envs := util.GetEnvOf(ModuleInfo.ModuleName)
	url := envs["DAL_URL"]
	if url == "" {
		url = "http://127.0.0.1:8885/dal"
	}

	options := map[string]interface{}{
		"op":       "insert",
		"category": "app",
		"appid":    appid,
	}

	ret, err := util.HTTPPostJSON(url, options, 30)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(ret))
	}
}
