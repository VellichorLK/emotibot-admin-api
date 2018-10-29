package auth

import (
	"fmt"
	"testing"

	"emotibot.com/emotigo/module/admin-api/util"
)

func TestGetUsername(t *testing.T) {
	dao, err := util.InitDB("172.16.101.98", "root", "password", "auth")
	if err != nil {
		t.Error("Cannot init auth db")
		return
	}

	util.SetDB(ModuleInfo.ModuleName, dao)

	usernames, err := GetUserNames([]string{"4b21158a395311e88a710242ac110003"})
	if err != nil {
		t.Error("GetUsernames fail", err.Error())
	}
	fmt.Printf("Get usernames: %+v\n", usernames)

	allnames, err := GetAllUserNames("csbot")
	if err != nil {
		t.Error("GetUsernames fail", err.Error())
	}
	fmt.Printf("Get usernames: %+v\n", allnames)
}

func TestGetUUID(t *testing.T) {
	dao, err := util.InitDB("172.16.101.98", "root", "password", "auth")
	if err != nil {
		t.Error("Cannot init auth db")
		return
	}

	util.SetDB(ModuleInfo.ModuleName, dao)
	uuid, err := GetUserID("csbotadmin")
	if err != nil {
		t.Error("GetUserID fail", err.Error())
	}
	if uuid != "4b21158a395311e88a710242ac110003" {
		t.Errorf("GetUserID except %+v, get %+v", "4b21158a395311e88a710242ac110003", uuid)
	}
}

func TestGetUUID(t *testing.T) {
	dao, err := util.InitDB("172.16.101.147", "root", "password", "authentication")
	if err != nil {
		t.Error("Cannot init auth db")
		return
	}

	util.SetDB(ModuleInfo.ModuleName, dao)
	uuid, err := GetUserID("csbotadmin")
	if err != nil {
		t.Error("GetUserID fail", err.Error())
	}
	if uuid != "4b21158a-3953-11e8-8a71-0242ac110003" {
		t.Errorf("GetUserID except %+v, get %+v", "4b21158a-3953-11e8-8a71-0242ac110003", uuid)
	}
}
