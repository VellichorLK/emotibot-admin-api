package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"
	_ "github.com/go-sql-driver/mysql"
)

func setUpDB() {
	db := dao.MYSQLController{}
	url, port, user, passwd, dbName := util.GetMySQLConfig()
	util.LogInfo.Printf("Connect mysql: %s:%s@%s:%d/%s\n", user, passwd, url, port, dbName)
	db.InitDB(url, port, dbName, user, passwd)
	service.SetDB(&db)
}

func main() {
	username := ""
	displayName := ""
	password := ""
	email := ""
	enterpriseName := ""
	flag.StringVar(&username, "u", "", "username of enterprise admin (required)")
	flag.StringVar(&displayName, "name", "", "display name of enterprise admin")
	flag.StringVar(&email, "email", "", "email of enterprise admin")
	flag.StringVar(&password, "p", "", "password of enterprise admin (required)")
	flag.StringVar(&enterpriseName, "e", "", "name of new enterprise (required)")
	flag.Parse()

	util.LogInit(ioutil.Discard, ioutil.Discard, os.Stdout, os.Stderr, "AUTH")
	setUpDB()

	if displayName == "" {
		displayName = username
	}
	md5Password := getMD5Hash(password)
	admin := &data.User{
		UserName:    &username,
		Password:    &md5Password,
		DisplayName: &displayName,
		Email:       &email,
		Type:        enum.AdminUser,
	}
	if !admin.IsValid() {
		util.LogError.Println("Invalid user parameter")
		flag.Usage()
		os.Exit(1)
	}
	if enterpriseName == "" {
		util.LogError.Println("Invalid enterprise name")
		flag.Usage()
		os.Exit(1)
	}
	enterprise := &data.Enterprise{
		Name: &enterpriseName,
	}
	enterpriseID, err := service.AddEnterprise(enterprise, admin)
	if err != nil {
		util.LogError.Println("Add enterprise fail, ", err.Error())
		os.Exit(2)
	}
	fmt.Println("Add enterprise success, get ID: ", enterpriseID)
}

func getMD5Hash(in string) string {
	hasher := md5.New()
	hasher.Write([]byte(in))
	return hex.EncodeToString(hasher.Sum(nil))
}
