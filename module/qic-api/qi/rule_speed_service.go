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
}

//GetRuleSpeedException gets the speed rule exception
func GetRuleSpeedException(r *model.SpeedRule) (*RuleSpeedException, map[uint64][]uint64, error) {

	if r == nil {
		return &RuleSpeedException{
			Under: RuleSpeedExceptionContent{Customer: make([]model.SimpleSentence, 0)},
			Over:  RuleSpeedExceptionContent{Customer: make([]model.SimpleSentence, 0)},
		}, nil, nil
	}

	if dbLike == nil {
		return nil, nil, ErrNilCon
	}

	var sentenceUUIDs []string

	var underExcpt, overExcpt RuleExceptionInteral
	if r.ExceptionUnder != "" {
		err := json.Unmarshal([]byte(r.ExceptionUnder), &underExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d under exception failed\n", r.ID)
			return nil, nil, err
		}
	} else {
		underExcpt.Customer = make([]string, 0)
	}
	if r.ExceptionOver != "" {
		err := json.Unmarshal([]byte(r.ExceptionOver), &overExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d over exception failed\n", r.ID)
			return nil, nil, err
		}
	} else {
		overExcpt.Customer = make([]string, 0)
	}

	sentenceUUIDs = append(sentenceUUIDs, underExcpt.Customer...)
	sentenceUUIDs = append(sentenceUUIDs, overExcpt.Customer...)

	sentencesCriteria := make(map[uint64][]uint64)

	sentences, err := sentenceDao.GetSentences(dbLike.Conn(),
		&model.SentenceQuery{UUID: sentenceUUIDs, Enterprise: &r.Enterprise})

	if err != nil {
		logger.Error.Printf("get sentence %+v failed. %s\n", sentenceUUIDs, err)
		return nil, nil, err
	}

	uuidSentenceMap := make(map[string]*model.Sentence, len(sentences))
	for _, v := range sentences {
		uuidSentenceMap[v.UUID] = v
		sentencesCriteria[v.ID] = append(sentencesCriteria[v.ID], v.TagIDs...)
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
	return &resp, sentencesCriteria, nil

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

type SpeedRuleWithException struct {
	RuleGroupID int64
	model.SpeedRule
	RuleSpeedException
	sentences map[uint64][]uint64
}

//RuleSpeedCheck checks the speed rule
func RuleSpeedCheck(ruleGroup model.Group, tagMatchDat []*MatchedData, segs []*SegmentWithSpeaker,
	staffSpeed float64) ([]RulesException, error) {

	if len(segs) == 0 {
		return nil, nil
	}
	if dbLike == nil {
		return nil, ErrNilCon
	}

	isDelete := 0
	q := &model.GeneralQuery{Enterprise: &ruleGroup.EnterpriseID, IsDelete: &isDelete, UUID: []string{ruleGroup.UUID}}
	rules, err := ruleSpeedDao.GetByRuleGroup(dbLike.Conn(), q)
	if err != nil {
		logger.Error.Printf("get rule silence failed. %s\n", err)
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}

	rulesWithExcpetion := make([]SpeedRuleWithException, 0, len(rules))
	for _, r := range rules {
		var sr SpeedRuleWithException
		exceptions, sentences, err := GetRuleSpeedException(r)
		if err != nil {
			logger.Error.Printf("get speed excpetions failed. %s\n", err)
			return nil, err
		}
		sr.RuleGroupID = ruleGroup.ID
		sr.SpeedRule = *r
		sr.RuleSpeedException = *exceptions
		sr.sentences = sentences
		rulesWithExcpetion = append(rulesWithExcpetion, sr)
	}

	credits, err := checkSpeedRules(rulesWithExcpetion, tagMatchDat, segs, staffSpeed)
	if err != nil {
		logger.Error.Printf("speed rule check failed.%s\n", err)
		return nil, err
	}
	return credits, nil
}

func composeTagCredits(m *MatchedData, segID int64) []*TagCredit {
	resp := make([]*TagCredit, 0, len(m.Matched))
	for _, data := range m.Matched {
		var tagCredit TagCredit
		//TagID
		tagCredit.ID = data.Tag
		tagCredit.Score = data.Score
		//SentenceID is the cu term for segment Idx, which is 1 based index
		tagCredit.SegmentIdx = data.SentenceID
		tagCredit.Match = data.Match
		tagCredit.MatchTxt = data.MatchText
		tagCredit.SegmentID = segID
		resp = append(resp, &tagCredit)
	}
	return resp
}

func getExceptionMatched(criteria []model.SimpleSentence, segs []*SegmentWithSpeaker,
	senMatchDat map[uint64][]int, tagMatchDat []*MatchedData, typ levelType) ([]*ExceptionMatched, bool) {

	resp := make([]*ExceptionMatched, 0, len(criteria))
	happenException := false
	for _, c := range criteria {
		matched := &ExceptionMatched{SentenceID: int64(c.ID), Typ: typ}
		if segsIdx, ok := senMatchDat[c.ID]; ok {
			happenException = true
			matched.Valid = true
			for _, seg := range segsIdx {
				matchedIdx := seg - 1
				matchedTags := tagMatchDat[matchedIdx]
				tagCredits := composeTagCredits(matchedTags, segs[matchedIdx].ID)
				matched.Tags = append(matched.Tags, tagCredits...)
			}
		}
		resp = append(resp, matched)
	}
	return resp, happenException
}

//checkSpeedRules checks the speed rules
//arguments: rules records the speed rules that wants to check
//tagMatchDat is the tag matched for each segments
//use internally, this function would not check the nil ptr or len
func checkSpeedRules(rules []SpeedRuleWithException, tagMatchDat []*MatchedData,
	segs []*SegmentWithSpeaker, speed float64) ([]RulesException, error) {
	resp := make([]RulesException, 0, len(rules))
	segMatchedTag := extractTagMatchedData(tagMatchDat)
	for _, r := range rules {
		result := RulesException{RuleID: r.ID, Typ: levSpeedTyp, Whos: Speed, CallID: segs[0].CallID, Valid: true, RuleGroupID: r.RuleGroupID}

		if speed > float64(r.Max) {
			result.Valid = false
		} else if speed < float64(r.Min) {
			result.Valid = false
		}

		senMatchDat, _ := SentencesMatch(segMatchedTag, r.sentences)

		exceptions, happened := getExceptionMatched(r.Under.Customer, segs, senMatchDat, tagMatchDat, levLCustomerSenTyp)
		result.Exception = append(result.Exception, exceptions...)
		if happened {
			result.Valid = true
		}

		exceptions, happened = getExceptionMatched(r.Over.Customer, segs, senMatchDat, tagMatchDat, levUCustomerSenTyp)
		result.Exception = append(result.Exception, exceptions...)
		if happened {
			result.Valid = true
		}

		if result.Valid {
			if r.Score > 0 {
				result.Score = r.Score
			}
		} else {
			if r.Score < 0 {
				result.Score = r.Score
			}
		}
		resp = append(resp, result)
	}
	return resp, nil
}
