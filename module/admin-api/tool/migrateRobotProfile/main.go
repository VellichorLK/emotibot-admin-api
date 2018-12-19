package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	"emotibot.com/emotigo/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

const (
	mySQLTimeout      string = "10s"
	mySQLWriteTimeout string = "30s"
	mySQLReadTimeout  string = "30s"
)

func main() {
	var err error
	logger.Init("MIGRATE", os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	if len(os.Args) != 5 {
		fmt.Printf("Usage: %s src-host src-appid target-host target-appid\n", os.Args[0])
		os.Exit(1)
	}

	type machine struct {
		Host  string
		Appid string
		DB    *sql.DB
	}

	source := &machine{os.Args[1], os.Args[2], nil}
	target := &machine{os.Args[3], os.Args[4], nil}

	fmt.Println("Init DB")
	source.DB, err = initDB(source.Host, "root", "password", "emotibot")
	if err != nil {
		fmt.Printf("Init DB of source (%s) fail: %s", source.Host, err.Error())
		os.Exit(2)
	}
	target.DB, err = initDB(target.Host, "root", "password", "emotibot")
	if err != nil {
		fmt.Printf("Init DB of target (%s) fail: %s", target.Host, err.Error())
		os.Exit(2)
	}

	fmt.Println("Get standard question from DB")
	srcQMap, _, err := getStdQ(source.DB)
	if err != nil {
		fmt.Printf("Get source std q err: %s\n", err.Error())
		os.Exit(3)
	}

	_, targetQMapRev, err := getStdQ(target.DB)
	if err != nil {
		fmt.Printf("Get target std q err: %s\n", err.Error())
		os.Exit(3)
	}

	fmt.Println("Get extend question from DB")
	srcExtendQ, err := getExtendQ(source.DB, source.Appid)
	if err != nil {
		fmt.Printf("Get source extend q err: %s\n", err.Error())
		os.Exit(4)
	}

	fmt.Println("Get answers from DB")
	srcAnswer, err := getAnswers(source.DB, source.Appid)
	if err != nil {
		fmt.Printf("Get source answer err: %s\n", err.Error())
		os.Exit(5)
	}

	extendQLen := 0
	answerLen := 0

	targetExtendQ := map[int][]string{}
	for origQid, extends := range srcExtendQ {
		origQ, ok := srcQMap[origQid]
		if !ok {
			continue
		}
		targetQid, ok := targetQMapRev[origQ]
		if !ok {
			continue
		}
		targetExtendQ[targetQid] = extends
		extendQLen += len(extends)
	}

	targetAnswers := map[int][]string{}
	for origQid, answers := range srcAnswer {
		origQ, ok := srcQMap[origQid]
		if !ok {
			continue
		}
		targetQid, ok := targetQMapRev[origQ]
		if !ok {
			continue
		}
		targetAnswers[targetQid] = answers
		answerLen += len(answers)
	}

	fmt.Printf(`
	=========================================================
	srcQ: %+v
	srcExtend: %+v
	srcAnswer: %+v
	tarQ: %+v
	=========================================================
	`, srcQMap, srcExtendQ, srcAnswer, targetQMapRev)

	fmt.Printf("Get total %d extend and %d answers, insert into target DB\n", extendQLen, answerLen)
	err = insertQA(target.DB, target.Appid, targetExtendQ, targetAnswers)
	if err != nil {
		fmt.Printf("Insert into target DB fail: %s\n", err.Error())
		os.Exit(6)
	}
}

func getStdQ(db *sql.DB) (map[int]string, map[string]int, error) {
	queryStr := "SELECT id, content FROM robot_profile_question"
	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, nil, err
	}

	stdQMap := map[int]string{}
	stdQMapRev := map[string]int{}
	for rows.Next() {
		var id int
		var content string
		err = rows.Scan(&id, &content)
		if err != nil {
			return nil, nil, err
		}
		stdQMapRev[content] = id
		stdQMap[id] = content
	}
	return stdQMap, stdQMapRev, nil
}

func getExtendQ(db *sql.DB, appid string) (map[int][]string, error) {
	queryStr := "SELECT qid, content FROM robot_profile_extend WHERE appid = ?"
	rows, err := db.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}

	ret := map[int][]string{}
	for rows.Next() {
		var qid int
		var content string
		err = rows.Scan(&qid, &content)
		if err != nil {
			return nil, err
		}
		if _, ok := ret[qid]; !ok {
			ret[qid] = []string{}
		}
		ret[qid] = append(ret[qid], content)
	}
	return ret, nil
}

func getAnswers(db *sql.DB, appid string) (map[int][]string, error) {
	queryStr := "SELECT qid, content FROM robot_profile_answer WHERE appid = ?"
	rows, err := db.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}

	ret := map[int][]string{}
	for rows.Next() {
		var qid int
		var content string
		err = rows.Scan(&qid, &content)
		if err != nil {
			return nil, err
		}
		if _, ok := ret[qid]; !ok {
			ret[qid] = []string{}
		}
		ret[qid] = append(ret[qid], content)
	}
	return ret, nil
}

func insertQA(db *sql.DB, appid string, extends map[int][]string, answers map[int][]string) error {
	queryStr := "INSERT INTO robot_profile_extend (appid, qid, content, status) VALUES (?, ?, ?, 1)"
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for qid, contents := range extends {
		for _, content := range contents {
			_, err := tx.Exec(queryStr, appid, qid, content)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	queryStr = "INSERT INTO robot_profile_answer (appid, qid, content, status) VALUES (?, ?, ?, 1)"
	for qid, contents := range answers {
		for _, content := range contents {
			_, err := tx.Exec(queryStr, appid, qid, content)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	err = tx.Commit()
	return err
}

func initDB(dbURL string, user string, pass string, db string) (*sql.DB, error) {
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
	logger.Trace.Println("Init DB: ", linkURL)
	openDB, err := sql.Open("mysql", linkURL)
	if err != nil {
		return nil, err
	}
	openDB.SetMaxIdleConns(0)
	return openDB, nil
}
