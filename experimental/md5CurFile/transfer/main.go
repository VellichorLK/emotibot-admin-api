package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	var dbAddres, dbUser, dbPass, volum, dbName string
	flag.StringVar(&volum, "p", "", "dir to scan")
	flag.StringVar(&dbAddres, "db", "localhost:3306", "image db location")
	flag.StringVar(&dbName, "dbname", "emotibot", "database name")
	flag.StringVar(&dbUser, "u", "deployer", "db userName")
	flag.StringVar(&dbPass, "pass", "", "db password")
	flag.Parse()

	if _, err := os.Stat(volum); os.IsNotExist(err) {
		log.Fatal("Path " + volum + " is not exist")
	} else if err != nil {
		log.Fatalf("path checking err, %v\n", err)
	}
	fmt.Println("scanning path " + volum)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=%s",
		dbUser, dbPass, dbAddres, dbName, url.QueryEscape("Asia/Shanghai"),
	))
	if err != nil {
		log.Fatal("db init failed, " + err.Error())
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal("db connect failed, " + err.Error())
	}

	querySQL := "select id,fileName from images"
	rows, err := db.Query(querySQL)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	volum = path.Clean(volum)

	var fileName string
	var id uint64
	var countSuccess uint64
	var countFailed uint64
	var mismatchCount uint64
	mismatchFile := make([]string, 0)
	for rows.Next() {
		err := rows.Scan(&id, &fileName)
		if err != nil {
			log.Fatal(err)
		}
		oldFile := volum + "/" + fileName
		if _, err := os.Stat(oldFile); err == nil {
			newFile := volum + "/" + getImageName(id, fileName)
			log.Printf("find file %s, copy it to %s\n", oldFile, newFile)
			err := os.Link(oldFile, newFile)
			if err != nil {
				log.Printf("Copy file %s to %s has error %s\n", oldFile, newFile, err.Error())
				countFailed++
			} else {
				countSuccess++
			}
		} else {
			mismatchCount++
			mismatchFile = append(mismatchFile, oldFile)
		}
	}

	fmt.Println("------------------------------------------")
	fmt.Println("Result:")
	fmt.Printf("Success transfer count:%v, failed transfer count:%v\n", countSuccess, countFailed)
	fmt.Printf("miss file count:%v (database has records but local file doesn't exist)\nmiss list:\n", mismatchCount)
	for i := 0; i < len(mismatchFile); i++ {
		fmt.Printf("%s\n", mismatchFile[i])
	}

}

//Md5Uint64 do md5sum with uint64 input
func Md5Uint64(id uint64) string {
	h := md5.New()
	fmt.Fprint(h, id)
	encode := h.Sum(nil)
	return hex.EncodeToString(encode)
}
func getImageName(id uint64, fileName string) string {
	ext := path.Ext(fileName)
	return Md5Uint64(id) + ext
}
