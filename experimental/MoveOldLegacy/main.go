package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	var dbAddres, dbUser, dbPass, loc, path string
	flag.StringVar(&path, "p", "", "dir to scan")
	flag.StringVar(&dbAddres, "db", "localhost:3306", "image db location")
	flag.StringVar(&dbUser, "u", "deployer", "db userName")
	flag.StringVar(&dbPass, "pass", "", "db password")
	flag.StringVar(&loc, "loc", "", "location value, which used for image insert.")
	flag.Parse()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatal("Path " + path + " is not exist")
	} else if err != nil {
		log.Fatalf("path checking err, %v\n", err)
	}
	fmt.Println("scanning path " + path)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/emotibot?parseTime=true&loc=%s",
		dbUser, dbPass, dbAddres, url.QueryEscape("Asia/Shanghai"),
	))
	fmt.Println(db)
	if err != nil {
		log.Fatal("db init failed, " + err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("db connect failed, " + err.Error())
	}
	fmt.Println(db)
	var validFileExt = map[string]bool{
		".jpeg": true,
		".jpg":  true,
		".png":  true,
		".gif":  true,
	}
	infos, err := getFilesInDir(path, newExtfilter(validFileExt))
	if err != nil {
		log.Fatal("Get Image err, " + err.Error())
	}
	locID, err := getLocID(db, loc)
	if err != nil {
		log.Fatal("Get Location failed, " + err.Error())
	}
	tx, err := db.Begin()
	count, err := insertImageEntity(tx, locID, infos)
	defer tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	fmt.Printf("Insert %d rows of image. \n", count)
}

func getLocID(db *sql.DB, loc string) (id int, err error) {
	err = db.QueryRow("SELECT id FROM image_location WHERE location = ?", loc).Scan(&id)
	return
}

func insertImageEntity(tx *sql.Tx, locID int, infos []os.FileInfo) (counts int, err error) {
	rawQuery := "INSERT INTO images (fileName, size, location_id) VALUES(?, ?, ?)"
	stmt, err := tx.Prepare(rawQuery)
	if err != nil {
		return 0, fmt.Errorf("prepare failed, %v", err)
	}
	for _, f := range infos {
		_, err := stmt.Exec(f.Name(), f.Size(), locID)
		if err != nil {
			return counts, fmt.Errorf("insert image failed, %v", err)
		}
		counts++
	}

	return counts, nil
}

type filter func(f os.FileInfo) bool

// Create a filter base on file extension.
// It will filter directory, map key value should contain dot in front.
func newExtfilter(validFileExt map[string]bool) filter {
	return func(f os.FileInfo) bool {
		if f.IsDir() {
			return false
		}

		if ext := strings.ToLower(filepath.Ext(f.Name())); validFileExt[ext] != true {
			return false
		}

		return true
	}
}

func getFilesInDir(path string, filters ...filter) ([]os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file failed, %v", err)
	}
	infos, err := f.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("scan dir failed, %v", err)
	}
	for i := len(infos) - 1; i >= 0; i-- {
		for _, fs := range filters {
			if ok := fs(infos[i]); !ok {
				infos = append(infos[:i], infos[i+1:]...)
			}
		}
	}
	return infos, nil

}
