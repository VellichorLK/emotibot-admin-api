package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

//{"tag":<tag>,"new_tag":<new_tag>, "file_id":<file_id>}

func TagsOperation(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET": //get all distinct tags
		getTags(w, r)
	case "POST": //update tag
		fallthrough
	case "PATCH": //DANGER!! update all tags from axxx to bxxx
		updateTag(w, r)
	case "PUT": //add new tag
		addTag(w, r)
	case "DELETE": //delete tag
		deleteTag(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func getTags(w http.ResponseWriter, r *http.Request) {

	appid := r.Header.Get(HXAPPID)
	tags, err := QueryTags(QueryTagsByAppidSQL, appid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encodeRes, err := json.Marshal(tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contentType := ContentTypeJSON

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(encodeRes)
}

func updateTag(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	if r.Body == nil {
		http.Error(w, "No request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var updateInfo TagOp
	err := json.NewDecoder(r.Body).Decode(&updateInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if updateInfo.NewTag == updateInfo.Tag {
		http.Error(w, "the same tag", http.StatusBadRequest)
		return
	}

	if updateInfo.NewTag == "" {
		http.Error(w, "new tag cannot be empty", http.StatusBadRequest)
		return
	}

	var status int
	if r.Method == "POST" {
		status, err = doUpdateTag(UpdateTagByFileIDSQL, updateInfo.NewTag, appid, updateInfo.FileID, updateInfo.Tag)
	} else {
		status, err = doUpdateTag(UpdateTagByAppidSQL, updateInfo.NewTag, appid, updateInfo.Tag)
	}

	if err != nil {
		http.Error(w, err.Error(), status)
	}
}

func addTag(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	if r.Body == nil {
		http.Error(w, "No request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var updateInfo TagOp
	err := json.NewDecoder(r.Body).Decode(&updateInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if updateInfo.NewTag == "" {
		http.Error(w, "new tag cannot be empty", http.StatusBadRequest)
		return
	}

	var status int

	status, err = doAddTag(appid, updateInfo.FileID, updateInfo.NewTag)
	if err != nil {
		http.Error(w, err.Error(), status)
	}

}

func deleteTag(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)
	if r.Body == nil {
		http.Error(w, "No request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var updateInfo TagOp
	err := json.NewDecoder(r.Body).Decode(&updateInfo)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var status int

	status, err = doDeleteTag(appid, updateInfo.FileID, updateInfo.Tag)
	if err != nil {
		http.Error(w, err.Error(), status)
	}

}

func doUpdateTag(query string, params ...interface{}) (int, error) {
	res, err := UpdateTags(query, params...)
	if err != nil {
		log.Println(err)
		return http.StatusBadRequest, err
	}
	return checkRowAffected(res)
}

func doAddTag(appid string, fileID string, tag string) (int, error) {

	var status int
	var err error

	id, tags, err := GetTagsByFileID(appid, fileID)
	if err == nil {
		if len(tags) >= LIMITUSERTAGS {
			status = http.StatusBadRequest
			err = errors.New("over limited tags")
		} else if id == "" {
			return http.StatusBadRequest, errors.New("No such file")
		} else {
			err = InsertUserDefinedTags(id, []string{tag}, nil)
			if err != nil {
				//assume its duplicate
				status = http.StatusBadRequest
			}
		}
	} else {
		status = http.StatusInternalServerError
	}

	return status, err
}

func doDeleteTag(params ...interface{}) (int, error) {
	res, err := ExecuteSQL(nil, DeleteTagByFileIDSQL, params...)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return checkRowAffected(res)
}

func checkRowAffected(res sql.Result) (int, error) {
	ra, err := res.RowsAffected()
	if err != nil {
		return http.StatusInternalServerError, err
	} else if ra == 0 {
		return http.StatusBadRequest, errors.New("No such file or tag")
	}

	return http.StatusOK, nil
}
