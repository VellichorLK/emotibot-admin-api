package qi

import (
	"encoding/hex"
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
	"github.com/satori/go.uuid"
)

//HandleGetTags handle the get request for tag.
func HandleGetTags(w http.ResponseWriter, r *http.Request) {
	query, err := getTagQuery(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("parse request failed, %v", err))
		return
	}
	resp, err := Tags(*query)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, err.Error())
		return
	}

	util.WriteJSON(w, resp)
}

func getTagQuery(r *http.Request) (*model.TagQuery, error) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	query := &model.TagQuery{
		Enterprise: &enterpriseID,
	}
	page, limit, err := getPageLimit(r)
	if err != nil {
		return nil, fmt.Errorf("get page & limit failed")
	}
	if limit != 0 {
		query.Paging = &model.Pagination{
			Limit: limit,
			Page:  page,
		}
	}
	r.ParseForm()
	types := r.Form["tag_type"]
	if len(types) > 0 {
		for _, typ := range types {
			typno, err := TagType(typ)
			if err != nil {
				return nil, fmt.Errorf("invalid tag type '%s'", typ)
			}
			query.TagType = append(query.TagType, typno)
		}
	}
	name := r.Form.Get("keyword")
	if name != "" {
		name = fmt.Sprintf("%%%s%%", name)
		query.Name = &name
	}
	return query, nil
}

func HandleGetTag(w http.ResponseWriter, r *http.Request) {
	tag, err := tagFromRequest(r)
	if ae, ok := err.(adminError); ok {
		util.ReturnError(w, ae.ErrorNo(), ae.Error())
	} else if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get tag by request failed, %v", err))
		return
	}

	t, err := toTag(*tag)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("validating tag failed, %v", err))
		return
	}
	util.WriteJSON(w, t[0])
}

func HandlePostTags(w http.ResponseWriter, r *http.Request) {

	modelTag, err := extractTag(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("bad input, %v", err))
		return
	}
	uuid, err := uuid.NewV4()
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, fmt.Sprintf("generate uuid failed, %v", err))
		return
	}
	modelTag.UUID = hex.EncodeToString(uuid[:])
	_, err = NewTag(*modelTag)
	if ce, ok := err.(adminError); ok {
		util.ReturnError(w, ce.ErrorNo(), fmt.Sprintf("new tag failed, %v", ce.Error()))
	} else if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("new tag failed, %v", err))
		return
	}
	util.WriteJSON(w, tag{
		TagUUID: modelTag.UUID,
	})
}

func HandlePutTags(w http.ResponseWriter, r *http.Request) {
	tg, err := tagFromRequest(r)
	if ae, ok := err.(adminError); ok {
		util.ReturnError(w, ae.ErrorNo(), ae.Error())
	} else if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("get tag by request failed, %v", err))
		return
	}
	updateTag, err := extractTag(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("bad input, %v", err))
		return
	}
	_, err = UpdateTag(model.Tag{
		ID:               tg.ID,
		UUID:             tg.UUID,
		Enterprise:       updateTag.Enterprise,
		Name:             updateTag.Name,
		Typ:              updateTag.Typ,
		PositiveSentence: updateTag.PositiveSentence,
		NegativeSentence: updateTag.NegativeSentence,
		CreateTime:       tg.CreateTime,
		UpdateTime:       updateTag.UpdateTime,
	})
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("update tag failed, %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	uuid, found := mux.Vars(r)["tag_id"]
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

// HandleIncreUpdateTagSentences is a special handler for labeling system api
// It add/delete sentences from specify tags.
func HandleIncreUpdateTagSentences(w http.ResponseWriter, r *http.Request) {
	type IncreUpdateRequest struct {
		TagID     string   `json:"tag_id"`
		Op        string   `json:"op"`
		Sentences []string `json:"sentences"`
	}
	var requests []IncreUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&requests)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("Bad request body, %v", err))
		return
	}
	r.Body.Close()
	var uuid []string
	for _, req := range requests {
		uuid = append(uuid, req.TagID)
	}
	cmd, err := NewTagUpdateCmd(requestheader.GetEnterpriseID(r), uuid)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("prepare update failed, %v", err))
		return
	}
	for _, req := range requests {
		err := cmd.AddSentenceUpdate(req.TagID, UpdateOp(req.Op), req.Sentences)
		if err != nil {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("prepare request failed, %v", err))
			return
		}
	}
	if err = cmd.Update(); err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("execute update failed, %v", err))
		return
	}

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

// tagFromRequest find the tag by the request's path variable "tag_id" and other infos.
// If no tags is found or request is invalid, a controllerError will return.
// If no error is found in the process, first tag it found will return.
func tagFromRequest(r *http.Request) (*model.Tag, error) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	uuid, found := mux.Vars(r)["tag_id"]
	if !found {
		return nil, controllerError{
			error: fmt.Errorf("path variable tag_id is not found"),
			errNo: AdminErrors.ErrnoRequestError,
		}
	}
	tags, err := tagDao.Tags(nil, model.TagQuery{
		UUID:       []string{uuid},
		Enterprise: &enterpriseID,
	})
	if err != nil {
		return nil, fmt.Errorf("tag by query failed, %v", err)
	}
	if len(tags) < 1 {
		return nil, controllerError{
			error: fmt.Errorf("tag is not exist"),
			errNo: AdminErrors.ErrnoRequestError,
		}
	}
	return &tags[0], nil
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
	TagName      string   `json:"tag_name"`
	TagType      string   `json:"tag_type"`
	PosSentences []string `json:"pos_sentences"`
	NegSentences []string `json:"neg_sentences"`
}
