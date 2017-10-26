package util

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var (
	allDB = make(map[string]*sql.DB)
)

const (
	mainDBKey  = "main"
	auditDBKey = "audit"
)

const (
	mySQLTimeout      string = "10s"
	mySQLWriteTimeout string = "30s"
	mySQLReadTimeout  string = "30s"
)

// InitMainDB will add a db handler in allDB, which key is main
func InitMainDB(mysqlURL string, mysqlUser string, mysqlPass string, mysqlDB string) error {
	LogInfo.Print("Start init main db")
	db, err := initDB(mysqlURL, mysqlUser, mysqlPass, mysqlDB)
	if err != nil {
		LogInfo.Printf("Main db init fail [%s]", err.Error())
		return err
	}
	allDB[mainDBKey] = db
	return nil
}

// GetMainDB will return main db in allDB
func GetMainDB() *sql.DB {
	return GetDB(mainDBKey)
}

// InitAuditDB should be called before insert all audit log
func InitAuditDB(auditURL string, auditUser string, auditPass string, auditDB string) error {
	LogInfo.Print("Start init audit db")
	db, err := initDB(auditURL, auditUser, auditPass, auditDB)
	if err != nil {
		LogInfo.Printf("Audit db init fail [%s]", err.Error())
		return err
	}
	allDB[auditDBKey] = db
	return nil
}

func initDB(url string, user string, pass string, db string) (*sql.DB, error) {
	if len(url) == 0 || len(user) == 0 || len(pass) == 0 || len(db) == 0 {
		return nil, errors.New("invalid parameters")
	}
	LogInfo.Printf("url: %s, user: %s, pass: %s, db_name: %s", url, user, pass, db)

	linkURL := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=true", user, pass, url, db, mySQLTimeout, mySQLReadTimeout, mySQLWriteTimeout)
	LogInfo.Printf("dburl: %s", linkURL)

	var err error
	openDB, err := sql.Open("mysql", linkURL)
	if err != nil {
		LogInfo.Printf("open db(%s) failed: %s", url, err)
		return nil, err
	}
	return openDB, nil
}

// GetAuditDB will return audit db in allDB
func GetAuditDB() *sql.DB {
	return GetDB(auditDBKey)
}

// GetDB will return db has assigned key in allDB
func GetDB(key string) *sql.DB {
	if db, ok := allDB[key]; ok {
		return db
	}
	LogInfo.Printf("Get DB %s fail: %#v", key, allDB)
	return nil
}
