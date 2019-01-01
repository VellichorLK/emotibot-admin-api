package qi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/qic-api/model/v1"

	"emotibot.com/emotigo/module/admin-api/util"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

//HandleGetTags handle the get request for tag.
func HandleGetTags(w http.ResponseWriter, r *http.Request) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	limit, page, err := getPageLimit(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("get page&limit failed, %v", err))
		return
	}
	resp, err := Tags(enterpriseID, limit, page)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, err.Error())
		return
	}

	util.WriteJSON(w, resp)
}

func HandlePostTags(w http.ResponseWriter, r *http.Request) {

	modelTag, err := extractTag(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("bad input, %v", err))
		return
	}

	id, err := NewTag(*modelTag)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("new tag failed, %v", err))
		return
	}
	util.WriteJSON(w, tag{
		TagID: id,
	})
}

func HandlePutTags(w http.ResponseWriter, r *http.Request) {
	modeltag, err := extractTag(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("bad input, %v", err))
		return
	}
	uuid, found := mux.Vars(r)["uuid"]
	if !found {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("bad input, path variable uuid is not found"))
		return
	}
	modeltag.UUID = uuid
	_, err = UpdateTag(*modeltag)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("update tag failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	uuid, found := mux.Vars(r)["uuid"]
	if !found {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("bad input, path variable uuid is not found"))
		return
	}
	var err error
	err = DeleteTag(uuid)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("delete tag failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func TagType(typ string) (int8, error) {
	var typNo int8
	for no, ttyp := range tagTypeDict {
		if typ == ttyp {
			typNo = no
		}
	}
	if typNo == 0 {
		return 0, fmt.Errorf("bad request, type %s is not valid", typ)
	}
	return typNo, nil
}

func extractTag(r *http.Request) (*model.Tag, error) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body failed, %v", err)
	}
	defer r.Body.Close()
	var reqBody NewTagReq
	err = json.Unmarshal(data, &reqBody)
	if err != nil {
		return nil, fmt.Errorf("unmarshal req body failed, %v", err)
	}
	typno, err := TagType(reqBody.TagType)
	if err != nil {
		return nil, fmt.Errorf("get tag type failed, %v", err)
	}

	posSentences, _ := json.Marshal(reqBody.PosSentences)
	negSentences, _ := json.Marshal(reqBody.NegSentences)
	timestamp := time.Now().Unix()
	return &model.Tag{
		UUID:             reqBody.TagUUID,
		Enterprise:       enterpriseID,
		Name:             reqBody.TagName,
		Typ:              typno,
		PositiveSentence: string(posSentences),
		NegativeSentence: string(negSentences),
		CreateTime:       timestamp,
		UpdateTime:       timestamp,
	}, nil
}

type NewTagReq struct {
	TagUUID      string   `json:"tag_uuid"`
	TagName      string   `json:"tag_name"`
	TagType      string   `json:"tag_type"`
	PosSentences []string `json:"pos_sentences"`
	NegSentences []string `json:"neg_sentences"`
}
