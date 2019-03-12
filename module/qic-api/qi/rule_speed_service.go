package qi

import (
	"encoding/json"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	ruleSpeedDao model.SpeedRuleDao = &model.SpeedRuleSQLDao{}
)

//NewRuleSpeed creates the new speed rule
func NewRuleSpeed(r *model.SpeedRule, enterprise string) (string, error) {
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
	_, err = ruleSpeedDao.Add(dbLike.Conn(), r)
	return r.UUID, err
}

//GetRuleSpeeds gets the list of speed rule
func GetRuleSpeeds(q *model.GeneralQuery, p *model.Pagination) ([]*model.SpeedRule, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	return ruleSpeedDao.Get(dbLike.Conn(), q, p)
}

//CountRuleSpeed counts the total number of speed rule
func CountRuleSpeed(q *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	return ruleSpeedDao.Count(dbLike.Conn(), q)
}

//RuleSpeedExceptionContent is used to return the exception
type RuleSpeedExceptionContent struct {
	Customer []model.SimpleSentence `json:"customer"`
}

//RuleSpeedException is used to return the speed rule exception
type RuleSpeedException struct {
	Under RuleSpeedExceptionContent `json:"under"`
	Over  RuleSpeedExceptionContent `json:"over"`
	// Under  struct {
	// 	Customer []model.SimpleSentence `json:"customer"`
	// } `json:"under"`
	// Over  struct {
	// 	Customer []model.SimpleSentence `json:"customer"`
	// } `json:"over"`
}

//GetRuleSpeedException gets the speed rule exception
func GetRuleSpeedException(r *model.SpeedRule) (*RuleSpeedException, error) {

	if r == nil {
		return &RuleSpeedException{
			Under: RuleSpeedExceptionContent{Customer: make([]model.SimpleSentence, 0)},
			Over:  RuleSpeedExceptionContent{Customer: make([]model.SimpleSentence, 0)},
		}, nil
	}

	if dbLike == nil {
		return nil, ErrNilCon
	}

	var sentenceUUIDs []string

	var underExcpt, overExcpt RuleExceptionInteral
	if r.ExceptionUnder != "" {
		err := json.Unmarshal([]byte(r.ExceptionUnder), &underExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d under exception failed\n", r.ID)
			return nil, err
		}
	} else {
		underExcpt.Customer = make([]string, 0)
	}
	if r.ExceptionOver != "" {
		err := json.Unmarshal([]byte(r.ExceptionOver), &overExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d over exception failed\n", r.ID)
			return nil, err
		}
	} else {
		overExcpt.Customer = make([]string, 0)
	}

	sentenceUUIDs = append(sentenceUUIDs, underExcpt.Customer...)
	sentenceUUIDs = append(sentenceUUIDs, overExcpt.Customer...)

	sentences, err := sentenceDao.GetSentences(dbLike.Conn(),
		&model.SentenceQuery{UUID: sentenceUUIDs, Enterprise: &r.Enterprise})

	if err != nil {
		logger.Error.Printf("get sentence %+v failed. %s\n", sentenceUUIDs, err)
		return nil, err
	}

	uuidSentenceMap := make(map[string]*model.Sentence, len(sentences))
	for _, v := range sentences {
		uuidSentenceMap[v.UUID] = v
	}

	resp := RuleSpeedException{
		Under: RuleSpeedExceptionContent{Customer: make([]model.SimpleSentence, 0)},
		Over:  RuleSpeedExceptionContent{Customer: make([]model.SimpleSentence, 0)},
	}
	for _, v := range underExcpt.Customer {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID}
			resp.Under.Customer = append(resp.Under.Customer, ss)
		} else {
			logger.Warn.Printf("Sentence with uuid:%s doesn't exist, but is recorded in the under exception of speed rule with id:%d\n", v, r.ID)
		}
	}

	for _, v := range overExcpt.Customer {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID}
			resp.Over.Customer = append(resp.Over.Customer, ss)
		} else {
			logger.Warn.Printf("Sentence with uuid:%s doesn't exist, but is recorded in the over exception of speed rule with id:%d\n", v, r.ID)
		}
	}
	return &resp, nil

}

//DeleteRuleSpeed deletes the speed rule
func DeleteRuleSpeed(q *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNoID
	}
	return ruleSpeedDao.SoftDelete(dbLike.Conn(), q)
}

//UpdateRuleSpeed updates the speed rule
func UpdateRuleSpeed(q *model.GeneralQuery, d *model.SpeedUpdateSet) (int64, error) {
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

	current, err := ruleSpeedDao.Get(tx, q, &model.Pagination{Limit: 10})
	if err != nil {
		return 0, err
	}
	if len(current) == 0 {
		return 0, ErrNoSuchID
	}

	q.ID = append(q.ID, current[0].ID)
	newID, err := ruleSpeedDao.Copy(tx, q)
	if err != nil {
		logger.Error.Printf("copy new record %v failed. %s\n", *q, err)
		return 0, err
	}

	affected, err := ruleSpeedDao.SoftDelete(tx, q)
	if err != nil {
		logger.Error.Printf("delete failed. %s\n", err)
		return 0, err
	}
	if affected == 0 {
		return 0, ErrNoSuchID
	}

	newQuery := &model.GeneralQuery{ID: []int64{newID}}
	affected, err = ruleSpeedDao.Update(tx, newQuery, d)
	if err != nil {
		logger.Error.Printf("update failed. %s\n", err)
		return 0, err
	}
	tx.Commit()
	return affected, err

}
