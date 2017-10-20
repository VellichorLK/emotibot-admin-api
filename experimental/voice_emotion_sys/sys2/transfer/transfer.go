package main

import (
	"database/sql"
	"flag"
	"fmt"
	"handlers"
	"log"
	"strconv"
)

//new table insert sql
const (
	InsertFileInfoSQL = "insert into " + handlers.MainTable + " (" + handlers.NFILEID + ", " + handlers.NFILEPATH + "," + handlers.NFILENAME + ", " + handlers.NFILETYPE + ", " +
		handlers.NSIZE + "," + handlers.NDURATION + ", " + handlers.NFILET + ", " + handlers.NCHECKSUM + ", " + handlers.NPRIORITY + ", " + handlers.NAPPID + ", " + handlers.NANAST + ", " +
		handlers.NANAET + ", " + handlers.NANARES + ", " + handlers.NUPT + ", " + handlers.NRDURATION + ")" + " values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	InsertTagSQL = "insert into " + handlers.UserDefinedTagsTable + " (" + handlers.NID + "," + handlers.NTAG + ") values(?,?)"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	iip := flag.String("i", "127.0.0.1", "database source ip (voice sys 1.0)")
	ipo := flag.Int("ipo", 3306, "port of source databse (voice sys 1.0)")
	ipp := flag.String("ip", "password", "password of souce ip")
	iu := flag.String("iu", "root", "user name of souce ip")

	oip := flag.String("o", "127.0.0.1", "database destination ip (voice sys 2.0)")
	opo := flag.Int("opo", 3306, "port of destination of database")
	op := flag.String("op", "password", "password of destination ip")
	ou := flag.String("ou", "root", "user name of destination ip")

	//src := "root:password@tcp(127.0.0.1:3307)/" + handlers.DataBase
	flag.Parse()

	src := *iu + ":" + *ipp + "@tcp(" + *iip + ":" + strconv.Itoa(*ipo) + ")/" + handlers.DataBase

	db, err := sql.Open("mysql", src)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//newSrc := "root:password@tcp(127.0.0.1:3306)/" + handlers.DataBase
	newSrc := *ou + ":" + *op + "@tcp(" + *oip + ":" + strconv.Itoa(*opo) + ")/" + handlers.DataBase
	newdb, err := sql.Open("mysql", newSrc)
	if err != nil {
		log.Fatal(err)
	}
	defer newdb.Close()

	query := "select * from " + handlers.MainTable //+ " where file_id='3800001e4ccb42fc89d7a0a02af616cb'"
	queryAna := "select * from " + handlers.AnalysisTable + " where " + handlers.NID + "=? order by " + handlers.NCHANNEL + "," + handlers.NSEGST
	queryCh := "select * from " + handlers.ChannelTable + " where " + handlers.NID + "=?"
	queryEmo := "select * from " + handlers.EmotionTable + " where " + handlers.NSEGID + "=?"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	newValues := make([]interface{}, 15)

	insertStmt, err := newdb.Prepare(InsertFileInfoSQL)
	if err != nil {
		log.Fatal(err)
	}
	insertTagStmt, err := newdb.Prepare(InsertTagSQL)
	if err != nil {
		log.Fatal(err)
	}
	insertAnaStmt, err := newdb.Prepare(handlers.InsertAnalysisSQL)
	if err != nil {
		log.Fatal(err)
	}
	insertEmoStmt, err := newdb.Prepare(handlers.InsertEmotionSQL)
	if err != nil {
		log.Fatal(err)
	}
	insertChaStmt, err := newdb.Prepare(handlers.InsertChannelScoreSQL)
	if err != nil {
		log.Fatal(err)
	}
	var c uint64
	for rows.Next() {

		rows.Scan(valuePtrs...)

		idx := 0

		tags := make([]string, 0)
		for i := range columns {

			b, ok := values[i].([]byte)
			var vs string
			if ok {
				vs = string(b)
			} else {
				log.Fatal("byte convert failed")
			}

			if i == 0 {

			} else if i == 9 || i == 10 { //tags
				if len(vs) > 0 {
					tags = append(tags, vs)
				}
			} else if i == 13 || i == 14 { //analysis time

				vv, err := strconv.ParseUint(vs, 10, 64)
				if err != nil {
					log.Fatal(err)
				}
				newValues[idx] = vv / 1000
				idx++
			} else {
				newValues[idx] = values[i]
				idx++
			}

		}

		//insert file info into new table
		res, err := insertStmt.Exec(newValues...)
		if err != nil {
			//log.Fatal(err)
			fmt.Println(err)
			continue
		}
		newid, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		//insert tag into new table
		for _, v := range tags {
			_, err := insertTagStmt.Exec(newid, v)
			if err != nil {
				log.Fatal(err)
			}
		}

		//query analysis
		anaRow, err := db.Query(queryAna, valuePtrs[0])
		if err != nil {
			log.Fatal(err)
		}

		anaCol, err := anaRow.Columns()
		if err != nil {
			log.Fatal(err)
		}

		anaValues := make([]interface{}, len(anaCol))
		anaValuePtr := make([]interface{}, len(anaCol))
		for i := range anaCol {
			anaValues[i] = &anaValuePtr[i]
		}

		for anaRow.Next() {
			err := anaRow.Scan(anaValues...)
			if err != nil {
				log.Fatal(err)
			}
			anaVV := anaValuePtr[1:]

			segID := anaValuePtr[0]

			anaVV[0] = newid
			res, err := insertAnaStmt.Exec(anaVV...)
			if err != nil {
				log.Fatal(err)
			}
			newSegID, err := res.LastInsertId()
			if err != nil {
				log.Fatal(err)
			}

			//query emotion
			emoRow, err := db.Query(queryEmo, segID)
			if err != nil {
				log.Fatal(err)
			}

			emoCol, err := emoRow.Columns()
			if err != nil {
				log.Fatal(err)
			}

			emoValues := make([]interface{}, len(emoCol))
			emoValuePtr := make([]interface{}, len(emoCol))
			for i := range emoCol {
				emoValues[i] = &emoValuePtr[i]
			}

			for emoRow.Next() {
				err := emoRow.Scan(emoValues...)
				if err != nil {
					log.Fatal(err)
				}
				emoVV := emoValuePtr[1:]
				emoVV[0] = newSegID
				_, err = insertEmoStmt.Exec(emoVV...)
				if err != nil {
					log.Fatal(err)
				}
			}

			emoRow.Close()

		}

		anaRow.Close()

		//query channel score
		chRow, err := db.Query(queryCh, valuePtrs[0])
		if err != nil {
			log.Fatal(err)
		}

		chCol, err := chRow.Columns()
		if err != nil {
			log.Fatal(err)
		}

		chValues := make([]interface{}, len(chCol))
		chValuePtr := make([]interface{}, len(chCol))
		for i := range chCol {
			chValues[i] = &chValuePtr[i]
		}

		for chRow.Next() {
			err := chRow.Scan(chValues...)
			if err != nil {
				log.Fatal(err)
			}
			chVV := chValuePtr[1:]
			chVV[0] = newid

			_, err = insertChaStmt.Exec(chVV...)
			if err != nil {
				log.Fatal(err)
			}

		}
		chRow.Close()
		c++
		log.Printf("finished %d rows transfer\n", c)
	}

	log.Printf("done!\n")

}
