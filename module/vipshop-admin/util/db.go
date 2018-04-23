package util

import (
	"database/sql"
	"fmt"
	"net/url"
	"crypto/sha256"
	"encoding/hex"

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
		LogError.Printf("Critical db error in rollback: %s", rollbackRet.Error())
	}
}

func HashContent(content string) (contentHash string) {
	hash := sha256.New()
	hash.Write([]byte(content))
	md := hash.Sum(nil)
	contentHash = hex.EncodeToString(md)
	return
}
