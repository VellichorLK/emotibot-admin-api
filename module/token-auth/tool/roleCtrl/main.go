package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"
	_ "github.com/go-sql-driver/mysql"
)

const (
	success = iota
	errEnterpriseID
	errModules
	errPrivilegeStr
	errAddFail
	errCmdNotFound
	errDeleteFail
	errRoleID
	errGetFail
)

var errMap = map[int]string{
	errEnterpriseID: "Invalid enterprise",
	errModules:      "Get modules of enterprise fail",
	errPrivilegeStr: "Invalid privileges str",
	errAddFail:      "Add role to enterprise fail",
	errCmdNotFound:  "No command found",
	errDeleteFail:   "Delete role of enterprise fail",
	errRoleID:       "Invalid role ID",
	errGetFail:      "Get role info of enterprise fail",
}

func showError(retCode int, err error) {
	if retCode != 0 {
		errMsg := ""
		if msg, ok := errMap[retCode]; ok {
			errMsg = msg
		} else {
			errMsg = fmt.Sprintf("Unknown error %d", retCode)
		}
		if err == nil {
			util.LogError.Println(errMsg)
		} else {
			util.LogError.Printf("%s, %s\n", errMsg, err.Error())
		}
		os.Exit(retCode)
	}
}

func setUpDB() {
	db := dao.MYSQLController{}
	url, port, user, passwd, dbName := util.GetMySQLConfig()
	db.InitDB(url, port, dbName, user, passwd)
	service.SetDB(&db)
}

func printUsage() {
	fmt.Printf(`Usage: %s <command>
command list:
  delete: delete a role
  add: add a role
  list: list all role and it's id
  priv: show privileges of role
`, os.Args[0])
}

func main() {
	if len(os.Args) <= 1 {
		printUsage()
		os.Exit(-1)
	}
	util.LogInit(ioutil.Discard, ioutil.Discard, os.Stdout, os.Stderr, "AUTH")
	setUpDB()

	command := os.Args[1]
	remainArgs := os.Args[2:]
	switch command {
	case "add":
		showError(handleAdd(remainArgs))
	case "delete":
		showError(handleDelete(remainArgs))
	case "list":
		showError(handleList(remainArgs))
	case "priv":
		showError(handleGetPriv(remainArgs))
	default:
		fmt.Println(errMap[errCmdNotFound])
		printUsage()
	}
}

func handleGetPriv(args []string) (retCode int, err error) {
	enterpriseID := ""
	roleID := ""
	showHelp := false

	deleteFlagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	deleteFlagSet.StringVar(&enterpriseID, "eid", "", "target enterprise ID of added role (required)")
	deleteFlagSet.StringVar(&roleID, "rid", "", "the id of role (required)")
	deleteFlagSet.BoolVar(&showHelp, "h", false, "show help messages")
	deleteFlagSet.Parse(args)
	if showHelp {
		fmt.Println("Usage of delete:")
		deleteFlagSet.PrintDefaults()
		return
	}

	_, retCode, err = checkEnterprise(enterpriseID)
	if retCode != success || err != nil {
		return
	}
	if roleID == "" {
		retCode, err = errRoleID, nil
		return
	}

	role, err := service.GetRole(enterpriseID, roleID)
	if err != nil {
		retCode = errGetFail
		return
	}
	ret, err := json.MarshalIndent(role.Privileges, "", "  ")
	if err != nil {
		retCode = errPrivilegeStr
		return
	}

	fmt.Printf("%s\n", ret)
	return
}

func handleList(args []string) (retCode int, err error) {
	enterpriseID := ""
	showHelp := false

	listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	listFlagSet.StringVar(&enterpriseID, "eid", "", "target enterprise ID of added role (required)")
	listFlagSet.BoolVar(&showHelp, "h", false, "show help messages")
	listFlagSet.Parse(args)
	if showHelp {
		fmt.Println("Usage of delete:")
		listFlagSet.PrintDefaults()
		return
	}

	_, retCode, err = checkEnterprise(enterpriseID)
	if retCode != success || err != nil {
		return
	}

	roles, err := service.GetRoles(enterpriseID)
	if err != nil {
		retCode = errGetFail
		return
	}

	if len(roles) == 0 {
		fmt.Println("No role found in enterprise")
		return
	}

	for _, role := range roles {
		fmt.Printf("%s %s\n", role.Name, role.UUID)
	}
	return
}

func handleDelete(args []string) (retCode int, err error) {
	enterpriseID := ""
	roleID := ""
	showHelp := false

	deleteFlagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	deleteFlagSet.StringVar(&enterpriseID, "eid", "", "target enterprise ID of added role (required)")
	deleteFlagSet.StringVar(&roleID, "rid", "", "the id of role (required)")
	deleteFlagSet.BoolVar(&showHelp, "h", false, "show help messages")
	deleteFlagSet.Parse(args)
	if showHelp {
		fmt.Println("Usage of delete:")
		deleteFlagSet.PrintDefaults()
		return
	}

	_, retCode, err = checkEnterprise(enterpriseID)
	if retCode != success || err != nil {
		return
	}
	if roleID == "" {
		retCode, err = errRoleID, nil
		return
	}

	ret, err := service.DeleteRole(enterpriseID, roleID)
	if ret != true {
		retCode = errDeleteFail
	}
	return
}

func handleAdd(args []string) (retCode int, err error) {
	addFlagSet := flag.NewFlagSet("add", flag.ContinueOnError)
	roleName := ""
	roleDesc := ""
	enterpriseID := ""
	privilegeStr := ""
	hasAllPrivilege := false
	showHelp := false
	addFlagSet.StringVar(&enterpriseID, "eid", "", "target enterprise ID of added role (required)")
	addFlagSet.StringVar(&roleName, "name", "", "the name of added role (required)")
	addFlagSet.StringVar(&roleDesc, "d", "", "the description of added role")
	addFlagSet.StringVar(&privilegeStr, "p", "", `privileges of added role in json format
        ex: "{'moduleA': ['view', 'edit'], 'moduleB': ['view']}"`)
	addFlagSet.BoolVar(&hasAllPrivilege, "a", false, `If this parameter is set,the role will has all
        privileges in this enterprise, and privileges setting will be ignored`)
	addFlagSet.BoolVar(&showHelp, "h", false, "show help messages")
	addFlagSet.Parse(args)

	if showHelp {
		fmt.Println("Usage of add:")
		addFlagSet.PrintDefaults()
		return
	}

	role := &data.Role{
		Name:        roleName,
		Description: roleDesc,
		UserCount:   0,
	}

	_, retCode, err = checkEnterprise(enterpriseID)
	if retCode != success || err != nil {
		return
	}

	privileges := map[string][]string{}
	if hasAllPrivilege {
		var modules []*data.Module
		modules, err = service.GetModules(enterpriseID)
		if err != nil {
			retCode = errModules
			return
		}
		for _, mod := range modules {
			privileges[mod.Code] = mod.Commands
		}
	} else {
		err = json.Unmarshal([]byte(privilegeStr), &privileges)
		if err != nil {
			retCode = errPrivilegeStr
			return
		}
	}

	role.Privileges = privileges
	roleID, err := service.AddRole(enterpriseID, role)
	if err != nil {
		retCode = errAddFail
		return
	}

	fmt.Println("Add role to enterprise success, get ID: ", roleID)
	return
}
