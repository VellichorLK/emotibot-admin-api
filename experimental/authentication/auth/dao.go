package auth

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func GetDao(c *Configuration) (*sql.DB, error) {
	log.Printf("config: %s", c)
	url := fmt.Sprintf("%s:%s@tcp(%s)/%s", c.DbUser, c.DbPass, c.DbUrl, c.DbName)
	log.Printf("url: %s", url)
	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Printf("open db %s failed. %s", db, err.Error())
		return nil, err
	}
	return db, db.Ping()
}
