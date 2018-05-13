package main

import (
	"flag"
	"fmt"
	"os"

	"emotibot.com/emotigo/module/token-auth/dao"
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
	enterpriseID := ""
	flag.StringVar(&enterpriseID, "e", "", "enterpriseID (required)")
	flag.Parse()

	util.LogInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr, "AUTH")
	setUpDB()

	if enterpriseID == "" {
		util.LogError.Println("Invalid enterpriseID")
		flag.Usage()
		os.Exit(1)
	}

	err := service.DeleteEnterprise(enterpriseID)
	if err != nil {
		util.LogError.Println("Delete enterprise fail, ", err.Error())
		os.Exit(2)
	}
	fmt.Println("Delete enterprise success")
}
