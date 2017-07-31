package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

const (
	const_mysql_timeout       string = "10s"
	const_mysql_write_timeout string = "30s"
	const_mysql_read_timeout  string = "30s"
)

type mysqlWrapper struct {
	mysql *sql.DB
}

type mysqlWrapperITF interface {
	Init(db_url string, db_user string, db_pass string, db_name string) error
	Query(cmd string) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error) // TODO(mike): wrapper.Result
	Destroy() error
}

func (dao *mysqlWrapper) Init(db_url string, db_user string, db_pass string, db_name string) error {
	// TODO(mike): input check
	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s", db_user, db_pass, db_url, db_name, const_mysql_timeout, const_mysql_read_timeout, const_mysql_write_timeout)
	log.Printf("url: %s", url) // log.info
	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Printf("open db(%s) failed: %s", url, err) //log.error
		return err
	}
	dao.mysql = db
	return dao.mysql.Ping()
}

func (dao *mysqlWrapper) Query(cmd string) (*sql.Rows, error) {
	// TODO(mike): cmd parser / validation
	if cmd == "" {
		return nil, errors.New("invalid sql command")
	}
	log.Printf("command: %s", cmd)
	return dao.mysql.Query(cmd)
}

func (dao *mysqlWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	// TODO(mike): cmd parser / validation
	if query == "" {
		return nil, errors.New("invalid sql command")
	}
	return dao.mysql.Exec(query, args...)
}

func (dao *mysqlWrapper) Destroy() error {
	if dao != nil {
		if err := dao.mysql.Close(); err != nil {
			log.Printf("close mysql failed: %s", err)
			return err
		}
	}
	return nil
}

// TODO(mike): should wrap as a dao struct
func GetDao(db_url string, db_user string, db_pass string, db_name string) (*mysqlWrapper, error) {
	dao := mysqlWrapper{}
	return &dao, dao.Init(db_url, db_user, db_pass, db_name)
}
