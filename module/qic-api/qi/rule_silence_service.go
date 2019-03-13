package qi

import (
	"encoding/json"
	"time"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	ruleSilenceDao model.SilenceRuleDao = &model.SilenceRuleSQLDao{}
)

//NewRuleSilence creates the new rule
func NewRuleSilence(r *model.SilenceRule, enterprise string) (string, error) {
	if r == nil {
		return "", ErrNoArgument
	}
	if dbLike == nil {
		return "", ErrNilCon
	}
	uuid, err := general.UUID()
	if err != nil {
		return "", err
	}
	r.Enterprise = enterprise
	r.CreateTime = time.Now().Unix()
	r.UpdateTime = r.CreateTime
	r.UUID = uuid
	_, err = ruleSilenceDao.Add(dbLike.Conn(), r)
	return r.UUID, err
}

//GetRuleSilences gets the list of rule
func GetRuleSilences(q *model.GeneralQuery, p *model.Pagination) ([]*model.SilenceRule, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	return ruleSilenceDao.Get(dbLike.Conn(), q, p)
}

//CountRuleSilence counts the total number of rule
func CountRuleSilence(q *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	return ruleSilenceDao.Count(dbLike.Conn(), q)
}

//RuleExceptionInteral is used to stored the exception uuid in db
type RuleExceptionInteral struct {
	Staff    []string `json:"staff"`
	Customer []string `json:"customer"`
}

//RuleException is used to return the exception
type RuleException struct {
	Staff    []model.SimpleSentence `json:"staff"`
	Customer []model.SimpleSentence `json:"customer"`
}

//RuleSilenceException is used to return the exception
type RuleSilenceException struct {
	Before RuleException `json:"before"`
	After  struct {
		Staff []model.SimpleSentence `json:"staff"`
	} `json:"after"`
}

//GetRuleSilenceException gets the rule silence exception
func GetRuleSilenceException(r *model.SilenceRule) (*RuleSilenceException, error) {

	if r == nil {
		return &RuleSilenceException{
			Before: RuleException{Staff: make([]model.SimpleSentence, 0), Customer: make([]model.SimpleSentence, 0)},
			After: struct {
				Staff []model.SimpleSentence `json:"staff"`
			}{
				Staff: make([]model.SimpleSentence, 0),
			},
		}, nil
	}

	if dbLike == nil {
		return nil, ErrNilCon
	}

	var exception []string

	var lowerExcpt, upperExcpt RuleExceptionInteral
	if r.ExceptionBefore != "" {
		err := json.Unmarshal([]byte(r.ExceptionBefore), &lowerExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d exception before failed\n", r.ID)
			return nil, err
		}
	} else {
		lowerExcpt.Customer = make([]string, 0)
		lowerExcpt.Staff = make([]string, 0)
	}
	if r.ExceptionAfter != "" {
		err := json.Unmarshal([]byte(r.ExceptionAfter), &upperExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d exception after failed\n", r.ID)
			return nil, err
		}
	} else {
		upperExcpt.Staff = make([]string, 0)
	}

	exception = append(exception, lowerExcpt.Customer...)
	exception = append(exception, lowerExcpt.Staff...)
	exception = append(exception, upperExcpt.Staff...)

	sentences, err := sentenceDao.GetSentences(dbLike.Conn(),
		&model.SentenceQuery{UUID: exception, Enterprise: &r.Enterprise})

	if err != nil {
		logger.Error.Printf("get sentence %+v failed. %s\n", exception, err)
		return nil, err
	}

	uuidSentenceMap := make(map[string]*model.Sentence, len(sentences))
	for _, v := range sentences {
		uuidSentenceMap[v.UUID] = v
	}

	resp := RuleSilenceException{Before: RuleException{Staff: make([]model.SimpleSentence, 0), Customer: make([]model.SimpleSentence, 0)},
		After: struct {
			Staff []model.SimpleSentence `json:"staff"`
		}{Staff: make([]model.SimpleSentence, 0)}}
	for _, v := range lowerExcpt.Staff {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID}
			resp.Before.Staff = append(resp.Before.Staff, ss)
		} else {
			logger.Warn.Printf("Cannot find %s sentence, but it exists in %d before staff exception\n", v, r.ID)
		}
	}

	for _, v := range lowerExcpt.Customer {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID}
			resp.Before.Customer = append(resp.Before.Customer, ss)
		} else {
			logger.Warn.Printf("Cannot find %s sentence, but it exists in %d before customer exception\n", v, r.ID)
		}
	}

	for _, v := range upperExcpt.Staff {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID}
			resp.After.Staff = append(resp.After.Staff, ss)
		} else {
			logger.Warn.Printf("Cannot find %s sentence, but it exists in %d after staff exception\n", v, r.ID)
		}
	}

	return &resp, nil

}

//DeleteRuleSilence deletes the rule
func DeleteRuleSilence(q *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNoID
	}
	return ruleSilenceDao.SoftDelete(dbLike.Conn(), q)
}

//UpdateRuleSilence updates the rule
func UpdateRuleSilence(q *model.GeneralQuery, d *model.SilenceUpdateSet) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNoID
	}
	tx, err := dbLike.Begin()
	if err != nil {
		logger.Error.Printf("create session failed. %s\n", err)
		return 0, err
	}
	defer tx.Rollback()

	current, err := ruleSilenceDao.Get(tx, q, &model.Pagination{Limit: 10})
	if err != nil {
		return 0, err
	}
	if len(current) == 0 {
		return 0, ErrNoSuchID
	}

	q.ID = append(q.ID, current[0].ID)
	newID, err := ruleSilenceDao.Copy(tx, q)
	if err != nil {
		logger.Error.Printf("copy new record %v failed. %s\n", *q, err)
		return 0, err
	}

	affected, err := ruleSilenceDao.SoftDelete(tx, q)
	if err != nil {
		logger.Error.Printf("delete failed. %s\n", err)
		return 0, err
	}
	if affected == 0 {
		return 0, ErrNoSuchID
	}

	newQuery := &model.GeneralQuery{ID: []int64{newID}}
	affected, err = ruleSilenceDao.Update(tx, newQuery, d)
	if err != nil {
		logger.Error.Printf("update failed. %s\n", err)
		return 0, err
	}
	tx.Commit()
	return affected, err

}
