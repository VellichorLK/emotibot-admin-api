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
	service.SetDBV3(&db)
}

func main() {
	username := ""
	displayName := ""
	password := ""
	email := ""
	enterpriseID := ""
	isAdmin := false
	flag.StringVar(&username, "u", "", "username")
	flag.StringVar(&displayName, "name", "", "display name")
	flag.StringVar(&email, "email", "", "email")
	flag.StringVar(&password, "p", "", "password (required)")
	flag.StringVar(&enterpriseID, "e", "", "enterpriseID")
	flag.BoolVar(&isAdmin, "admin", false, "if admin, set this flag")
	flag.Parse()

	util.LogInit(ioutil.Discard, ioutil.Discard, os.Stdout, os.Stderr, "AUTH")
	setUpDB()

	if displayName == "" {
		displayName = username
	}
	md5Password := getMD5Hash(password)
	// user := &data.User{
	// 	UserName:    &username,
	// 	Password:    &md5Password,
	// 	DisplayName: &displayName,
	// 	Email:       &email,
	// 	Type:        enum.AdminUser,
	// }
	user := &data.UserDetailV3{
		UserV3: data.UserV3{
			UserName:    username,
			DisplayName: displayName,
			Email:       email,
			Type:        enum.NormalUser,
		},
		Password: &md5Password,
	}
	if isAdmin {
		user.Type = enum.AdminUser
	}

	if !user.IsValid() {
		util.LogError.Println("Invalid user parameter")
		flag.Usage()
		os.Exit(1)
	}
	userID, err := service.AddUserV3(enterpriseID, user)
	if err != nil {
		util.LogError.Println("Add enterprise fail, ", err.Error())
		os.Exit(2)
	}
	fmt.Printf("Add user in %s success, get ID: %s\n", enterpriseID, userID)
}

func getMD5Hash(in string) string {
	hasher := md5.New()
	hasher.Write([]byte(in))
	return hex.EncodeToString(hasher.Sum(nil))
}
