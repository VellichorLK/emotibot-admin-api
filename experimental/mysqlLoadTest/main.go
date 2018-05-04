package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Host     string
	Port     string
	USER     string
	Password string
	Database string
	DoneNum  uint64
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	Host = os.Getenv("host")
	Port = os.Getenv("port")
	USER = os.Getenv("user")
	Database = os.Getenv("db")
	Password = os.Getenv("password")
	var dbGroup []*sql.DB
	for _, h := range strings.Split(Host, ",") {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=30s", USER, Password, h, Port, Database))
		if err != nil {
			log.Fatal(err)
		}
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(h + ":" + Port)
		dbGroup = append(dbGroup, db)
	}
	connections := flag.Int64("c", 1, "how many concurrency connection you want to have.")
	flag.Parse()
	fmt.Println("Using connection pool:" + strconv.Itoa(int(*connections)))
	for i := 0; int64(i) < *connections; i++ {
		// db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=30s", USER, Password, Host, Port, Database))
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// err = db.Ping()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		rotate := i % len(dbGroup)
		db := dbGroup[rotate]
		go WriteDB(db, 10000)
	}
	go func() {
		for {
			select {
			case <-time.Tick(time.Duration(5) * time.Second):
				fmt.Printf("Num:%d\n", DoneNum)
			}
		}
	}()
	sigChan := make(chan os.Signal, 1)
	cleanUpDone := make(chan bool)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for _ = range sigChan {
			fmt.Println("receive signal, clean up resource.")
			fmt.Printf("Num of %d loops complete\n", DoneNum)
			for _, db := range dbGroup {
				db.Close()
			}
			cleanUpDone <- true
		}
	}()
	<-cleanUpDone

}

func WriteDB(db *sql.DB, id int) {
	//SELECT row And Delete row And Insert a new row.
	for {
		tx, err := db.Begin()
		if err != nil {
			log.Println(err)
			tx.Rollback()
			continue
		}
		rows, err := tx.Query("SELECT content From vipshop_answer WHERE Question_Id = ?", id)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			continue
		}
		for rows.Next() {
			rows.Scan()
		}
		rows.Close()
		_, err = tx.Exec("DELETE FROM vipshop_answer WHERE Question_Id = ?", id)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			continue
		}

		_, err = tx.Exec("INSERT INTO vipshop_answer (Question_Id, Content, Content_String, Status) VALUE(?, ?, ?, ?)", id, "TestMariadbBenchmark", "", 0)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			continue
		}
		tx.Commit()
		atomic.AddUint64(&DoneNum, 1)
	}
}
