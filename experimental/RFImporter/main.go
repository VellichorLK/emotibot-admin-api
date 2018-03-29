package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	//For Consul client
	util.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	f, err := os.Open("./data.csv")
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(f)
	defer f.Close()

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	var data []string
	for _, record := range records {
		if record[0] != "" {
			data = append(data, record[0])
		}
	}

	var addr, appID, consulAddr string
	flag.StringVar(&addr, "addr", "root:password@tcp(127.0.0.1:3306)", "address for db connection. template: user:password@tcp(ip:port)")
	flag.StringVar(&consulAddr, "consul", "http://127.0.0.1:8500/v1/kv/idc", "address for consul key store.")
	flag.StringVar(&appID, "app", "vipshop", "appid for table")
	flag.Parse()
	u, err := url.Parse(consulAddr)
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s/emotibot?parseTime=true", addr))
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	tx.Exec("TRUNCATE " + appID + "_removeFeedbackQuestion")

	stmt, err := tx.Prepare("INSERT INTO " + appID + "_removeFeedbackQuestion(Question_Content) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}

	size := len(data)
	for i, d := range data {
		_, err := stmt.Exec(d)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\rInserted %d/%d", i, size)
	}

	client := util.NewConsulClient(u)
	unixTime := time.Now().UnixNano() / 1000000
	_, err = client.ConsulUpdateVal("vipshopdata/RFQuestion", unixTime)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Finish import, total %d of rows insert.\n", size)
	// For sorting bad data.
	// sort.Strings(data)
	// nf, _ := os.Create("./gd.csv")
	// defer nf.Close()
	// w := csv.NewWriter(nf)
	// for _, d := range data {
	// 	w.Write([]string{d})
	// }
	// if err = w.Error(); err != nil {
	// 	log.Fatal(err)
	// }
	// w.Flush()
}
