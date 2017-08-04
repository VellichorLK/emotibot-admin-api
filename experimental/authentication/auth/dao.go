package auth

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	const_mysql_timeout       string = "10s"
	const_mysql_write_timeout string = "30s"
	const_mysql_read_timeout  string = "30s"
)

type DaoWrapper struct {
	mysql *sql.DB
}

// interface definition
type Query func(cmd string) (*sql.Rows, error)
type QueryRow func(cmd string) *sql.Row
type Exec func(cmd string) (sql.Result, error)
type Finalize func() error

// func implementation
func DaoMysqlInit(db_url string, db_user string, db_pass string, db_name string) (d *DaoWrapper, err error) {
	if len(db_url) == 0 || len(db_user) == 0 || len(db_pass) == 0 || len(db_name) == 0 {
		return nil, errors.New("invalid parameters!")
	}
	LogInfo.Printf("db_url: %s, db_user: %s, db_pass: %s, db_name: %s", db_url, db_user, db_pass, db_name)

	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s", db_user, db_pass, db_url, db_name, const_mysql_timeout, const_mysql_read_timeout, const_mysql_write_timeout)
	LogInfo.Printf("url: %s", url)

	db, err := sql.Open("mysql", url)
	if err != nil {
		LogInfo.Printf("open db(%s) failed: %s", url, err)
		return nil, err
	}
	dao := DaoWrapper{db}
	return &dao, db.Ping()
}

func (d *DaoWrapper) Query(cmd string) (*sql.Rows, error) {
	// TODO(mike): parameter check, isinstance(d, *sqlDB), query syntax check
	if len(cmd) == 0 || d == nil {
		return nil, errors.New("invalid parameter")
	}
	LogInfo.Printf("cmd: %s", cmd)
	return d.mysql.Query(cmd)
}

func (d *DaoWrapper) QueryRow(cmd string) *sql.Row {
	// TODO(mike): parameter check, isinstance(d, *sqlDB), query syntax check
	if len(cmd) == 0 || d == nil {
		return nil
	}
	LogInfo.Printf("cmd: %s", cmd)
	return d.mysql.QueryRow(cmd)
}

func (d *DaoWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	// TODO(mike): parameter check, isinstance(d, *sqlDB), query syntax check
	if len(query) == 0 || d == nil {
		return nil, errors.New("invalid parameter")
	}
	return d.mysql.Exec(query, args...)
}

func (d *DaoWrapper) Finalize() error {
	if d != nil {
		if err := d.mysql.Close(); err != nil {
			LogInfo.Printf("close mysql failed: %s", err)
			return err
		}
	}
	return nil
}
