package util

import (
	"database/sql"
	"fmt"
	"net/url"

	"emotibot.com/emotigo/pkg/logger"
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
	db, err := InitDB(mysqlURL, mysqlUser, mysqlPass, mysqlDB)
	if err != nil {
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
	db, err := InitDB(auditURL, auditUser, auditPass, auditDB)
	if err != nil {
		return err
	}
	allDB[auditDBKey] = db
	return nil
}

func InitDB(dbURL string, user string, pass string, db string) (*sql.DB, error) {
	linkURL := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=true&loc=%s",
		user,
		pass,
		dbURL,
		db,
		mySQLTimeout,
		mySQLReadTimeout,
		mySQLWriteTimeout,
		url.QueryEscape("Asia/Shanghai"), //A quick dirty fix to ensure time.Time parsing
	)

	if len(dbURL) == 0 || len(user) == 0 || len(pass) == 0 || len(db) == 0 {
		return nil, fmt.Errorf("invalid parameters in initDB: %s", linkURL)
	}

	var err error
	logger.Trace.Println("Init DB: ", linkURL)
	openDB, err := sql.Open("mysql", linkURL)
	if err != nil {
		return nil, err
	}
	openDB.SetMaxIdleConns(0)
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
	return nil
}

func SetDB(key string, db *sql.DB) {
	allDB[key] = db
}

func ClearTransition(tx *sql.Tx) {
	rollbackRet := tx.Rollback()
	if rollbackRet != sql.ErrTxDone && rollbackRet != nil {
		logger.Error.Printf("Critical db error in rollback: %s", rollbackRet.Error())
	}
}

func ShowError(err error) {
	if err != nil {
		logger.Error.Printf("DB error: %s\n", err.Error())
		PrintRuntimeStack(5)
	}
}

func GetDBStatus() map[string]string {
	ret := map[string]string{}
	for key := range allDB {
		pingError := allDB[key].Ping()
		if pingError == nil {
			ret[key] = "health"
		} else {
			ret[key] = pingError.Error()
		}
	}
	return ret
}
