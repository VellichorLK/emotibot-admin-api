package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type prohibitedWords struct {
	Words string `json:"words"`
}

//ManipulateProhibitedWords entrypoint of get and add prohibited words api
func ManipulateProhibitedWords(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body []byte
	var err error

	switch r.Method {
	case "GET":
		body, err = getProhibitedWords(appid)
	case "POST":
		wording := &prohibitedWords{}
		err = json.NewDecoder(r.Body).Decode(wording)
		if err != nil {
			http.Error(w, "Request body invalid", http.StatusBadRequest)
			return
		}
		_, err = addProhibitedWords(appid, 1, wording.Words)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if body != nil {
		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func getProhibitedWords(appid string) ([]byte, error) {
	prohibitedWordsMap, err := getProhitbitWords(appid)

	if err != nil {
		return nil, err
	}

	type prohibitedRsp struct {
		ID    uint64 `json:"prohibited_id"`
		Words string `json:"words"`
	}
	prohibitedWords := make([]prohibitedRsp, len(prohibitedWordsMap))
	keys := make([]uint64, len(prohibitedWordsMap))

	//sort id
	count := 0
	for id := range prohibitedWordsMap {
		keys[count] = id
		count++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	//package to slice
	count = 0
	for _, id := range keys {
		prohibitedWords[count].ID = id
		prohibitedWords[count].Words = prohibitedWordsMap[id]
		count++
	}
	return json.Marshal(prohibitedWords)
}

//addProhibitedWords insert a new reord of prohibited words
func addProhibitedWords(appid string, priority uint, words string) (int64, error) {
	addProhibitedSQL := fmt.Sprintf("insert into %s (%s,%s,%s) values(?,?,?)",
		ProhibitedTable, NJAPPID, NPRIORITY, NPROHIBIT)
	return execUpdate(addProhibitedSQL, appid, priority, words)
}

//ModifyProhibitedWords the entrypoint of modify api, update and delete
func ModifyProhibitedWords(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := getProhibitID(r.URL.Path)
	if err != nil {
		http.Error(w, "Bad request, id ", http.StatusBadRequest)
		return
	}
	var updateCount int64

	switch r.Method {
	case "DELETE":
		updateCount, err = deleteProhibitedWords(id, appid)
	case "PATCH":
		wording := &prohibitedWords{}
		err = json.NewDecoder(r.Body).Decode(wording)
		if err != nil {
			http.Error(w, "Request body invalid", http.StatusBadRequest)
			return
		}
		updateCount, err = updateProhibitedWords(id, appid, wording.Words)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if updateCount == 0 {
		http.Error(w, "Bad request, no such id ", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getProhibitID(path string) (uint64, error) {
	idPos := 6
	element := strings.Split(path, "/")
	if len(element) != idPos {
		return 0, errors.New("wrong path")
	}
	id, err := strconv.ParseUint(element[idPos-1], 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

//deleteProhibitedWords delete the prohibited words record
func deleteProhibitedWords(id uint64, appid string) (int64, error) {
	deleteSQL := fmt.Sprintf("delete from %s where %s=? and %s=?",
		ProhibitedTable, NID, NJAPPID)
	return execUpdate(deleteSQL, id, appid)
}

//updateProhibitedWords update the prohibited words
func updateProhibitedWords(id uint64, appid string, words string) (int64, error) {
	updateSQL := fmt.Sprintf("update %s set %s=? where %s=? and %s=?",
		ProhibitedTable, NPROHIBIT, NID, NJAPPID)
	return execUpdate(updateSQL, words, id, appid)
}

func execUpdate(execSQL string, params ...interface{}) (int64, error) {
	res, err := db.Exec(execSQL, params...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
