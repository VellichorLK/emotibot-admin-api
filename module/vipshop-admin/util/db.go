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

func EscapeQuery(sqlStr string) string {
	// escape the following characters
	// \0     An ASCII NUL (0x00) character.
	// \'     A single quote (“'”) character.
	// \"     A double quote (“"”) character.
	// \b     A backspace character.
	// \n     A newline (linefeed) character.
	// \r     A carriage return character.
	// \t     A tab character.
	// \Z     ASCII 26 (Control-Z). See note following the table.
	// \\     A backslash (“\”) character.
	// \%     A “%” character. See note following the table.
	// \_     A “_” character. See note following the table.

	// the implementation is copied from https://gist.github.com/siddontang/8875771
	// very appreciate for his working
	dest := make([]byte, 0, 2*len(sqlStr))
	var escape byte
	for i := 0; i < len(sqlStr); i++ {
		c := sqlStr[i]

		escape = 0

		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
			break
		case '\n': /* Must be escaped for logs */
			escape = 'n'
			break
		case '\r':
			escape = 'r'
			break
		case '\\':
			escape = '\\'
			break
		case '\'':
			escape = '\''
			break
		case '"': /* Better safe than sorry */
			escape = '"'
			break
		case '%':
			escape = '%'
			break
		case '\b':
			escape = 'b'
			break
		case '\t':
			escape = 't'
			break
		case '_':
			escape = '_'
			break
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}
	return string(dest)
}
