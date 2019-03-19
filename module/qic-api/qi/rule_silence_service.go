package qi

import (
	"encoding/json"
	"errors"
	"sort"
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
type OnlyStaffException struct {
	Staff []model.SimpleSentence `json:"staff"`
}

//RuleSilenceException is used to return the exception
type RuleSilenceException struct {
	Before RuleException      `json:"before"`
	After  OnlyStaffException `json:"after"`
}

type SilenceRuleWithException struct {
	RuleGroupID int64
	model.SilenceRule
	RuleSilenceException
	sentences map[uint64][]uint64
}

//GetRuleSilenceException gets the rule silence exception
func GetRuleSilenceException(r *model.SilenceRule) (*RuleSilenceException, map[uint64][]uint64, error) {

	if r == nil {
		return &RuleSilenceException{
			Before: RuleException{Staff: make([]model.SimpleSentence, 0), Customer: make([]model.SimpleSentence, 0)},
			After:  OnlyStaffException{Staff: make([]model.SimpleSentence, 0)},
		}, nil, nil
	}

	if dbLike == nil {
		return nil, nil, ErrNilCon
	}

	var exception []string

	var lowerExcpt, upperExcpt RuleExceptionInteral
	if r.ExceptionBefore != "" {
		err := json.Unmarshal([]byte(r.ExceptionBefore), &lowerExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d exception before failed\n", r.ID)
			return nil, nil, err
		}
	} else {
		lowerExcpt.Customer = make([]string, 0)
		lowerExcpt.Staff = make([]string, 0)
	}
	if r.ExceptionAfter != "" {
		err := json.Unmarshal([]byte(r.ExceptionAfter), &upperExcpt)
		if err != nil {
			logger.Error.Printf("unmarshal %d exception after failed\n", r.ID)
			return nil, nil, err
		}
	} else {
		upperExcpt.Staff = make([]string, 0)
	}

	exception = append(exception, lowerExcpt.Customer...)
	exception = append(exception, lowerExcpt.Staff...)
	exception = append(exception, upperExcpt.Staff...)

	var isDelete int8
	sentences, err := sentenceDao.GetSentences(dbLike.Conn(),
		&model.SentenceQuery{UUID: exception, Enterprise: &r.Enterprise, IsDelete: &isDelete})

	sentencesCriteria := make(map[uint64][]uint64)
	if err != nil {
		logger.Error.Printf("get sentence %+v failed. %s\n", exception, err)
		return nil, nil, err
	}

	uuidSentenceMap := make(map[string]*model.Sentence, len(sentences))
	for _, v := range sentences {
		uuidSentenceMap[v.UUID] = v
		sentencesCriteria[v.ID] = append(sentencesCriteria[v.ID], v.TagIDs...)
	}

	resp := RuleSilenceException{Before: RuleException{Staff: make([]model.SimpleSentence, 0), Customer: make([]model.SimpleSentence, 0)},
		After: OnlyStaffException{Staff: make([]model.SimpleSentence, 0)}}
	for _, v := range lowerExcpt.Staff {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID, ID: s.ID}
			resp.Before.Staff = append(resp.Before.Staff, ss)
		} else {
			logger.Warn.Printf("Cannot find %s sentence, but it exists in %d before staff exception\n", v, r.ID)
		}
	}

	for _, v := range lowerExcpt.Customer {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID, ID: s.ID}
			resp.Before.Customer = append(resp.Before.Customer, ss)
		} else {
			logger.Warn.Printf("Cannot find %s sentence, but it exists in %d before customer exception\n", v, r.ID)
		}
	}

	for _, v := range upperExcpt.Staff {
		if s, ok := uuidSentenceMap[v]; ok {
			ss := model.SimpleSentence{UUID: v, Name: s.Name, CategoryID: s.CategoryID, ID: s.ID}
			resp.After.Staff = append(resp.After.Staff, ss)
		} else {
			logger.Warn.Printf("Cannot find %s sentence, but it exists in %d after staff exception\n", v, r.ID)
		}
	}

	return &resp, sentencesCriteria, nil

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

const SilenceSpeaker = -1

type segDuration struct {
	index    int
	duration float64
}

//RuleSilenceCheck checks the silence rules
func RuleSilenceCheck(ruleGroup model.Group, allSegs []*SegmentWithSpeaker, matched []*MatchedData) ([]RulesException, error) {

	if len(allSegs) == 0 {
		return nil, nil
	}

	//TODO: fix it, get rules by rule group id and sets the rulegroup id to the attribute RuleGroup in RulesException
	isDelete := 0
	q := &model.GeneralQuery{Enterprise: &ruleGroup.EnterpriseID, IsDelete: &isDelete}
	sRules, err := GetRuleSilences(q, nil)
	if err != nil {
		logger.Error.Printf("get rule silence failed. %s\n", err)
		return nil, err
	}
	if len(sRules) == 0 {
		return nil, nil
	}

	silenceSegs, segs := extractSegmentSpeaker(allSegs, SilenceSpeaker)

	silenceNum := len(silenceSegs)

	totalSeg := len(allSegs)
	matchedSeg := len(matched)
	if (totalSeg - silenceNum) != matchedSeg {
		return nil, errors.New("total seg without silence not equal to matched seg")
	}

	//get rules with exception data
	rules := make([]SilenceRuleWithException, 0, len(sRules))
	for _, r := range sRules {
		var exceptionRule SilenceRuleWithException
		exceptionRule.SilenceRule = *r
		exception, senCriteria, err := GetRuleSilenceException(r)
		if err != nil {
			logger.Error.Printf("get silence exception failed. %s\n", err)
			return nil, err
		}
		exceptionRule.RuleGroupID = ruleGroup.ID
		exceptionRule.RuleSilenceException = *exception
		exceptionRule.sentences = senCriteria
		rules = append(rules, exceptionRule)
	}

	credits, err := silenceRuleCheck(rules, matched, allSegs, segs, silenceSegs)
	if err != nil {
		logger.Error.Printf("silence rule check failed.%s\n", err)
		return nil, err
	}

	return credits, nil
}

//extract the segment with given speaker and sorts it in descending order by segment's duration
//return  segment with only given speaker and semgent without segment with given speaker
func extractSegmentSpeaker(segments []*SegmentWithSpeaker, speaker int) ([]segDuration, []*SegmentWithSpeaker) {
	silenceSegs := make([]segDuration, 0, 16)
	segs := make([]*SegmentWithSpeaker, 0, len(segments))
	for k, v := range segments {
		if v.Speaker == speaker {
			silenceDur := v.EndTime - v.StartTime
			s := segDuration{index: k, duration: silenceDur}
			silenceSegs = append(silenceSegs, s)
		} else {
			segs = append(segs, v)
		}
	}
	sort.SliceStable(silenceSegs, func(i, j int) bool {
		return silenceSegs[i].duration > silenceSegs[j].duration
	})
	return silenceSegs, segs
}

func getDExceptionAndBuildMap(exception []model.SimpleSentence, typ levelType, m map[senTypeKey]*ExceptionMatched) []*ExceptionMatched {
	resp := make([]*ExceptionMatched, 0, len(exception))
	for _, v := range exception {
		e := &ExceptionMatched{SentenceID: int64(v.ID), Typ: typ}
		resp = append(resp, e)
	}
	addExceptionMap(m, resp)
	return resp
}

func getDefaultException(exception []model.SimpleSentence, typ levelType) []*ExceptionMatched {
	resp := make([]*ExceptionMatched, 0, len(exception))
	for _, v := range exception {
		e := &ExceptionMatched{SentenceID: int64(v.ID), Typ: typ}
		resp = append(resp, e)
	}
	return resp
}
func addExceptionMap(m map[senTypeKey]*ExceptionMatched, es []*ExceptionMatched) {
	for _, v := range es {
		k := senTypeKey{typ: v.Typ, sentenceID: v.SentenceID}
		m[k] = v
	}
}

//findFirstSegBeforeIndex finds the first index of segment, for both staff and customer, which is not silence before the given index
//return the index of the first index for staff and customer  before the given idx
func findFirstSegBeforeIndex(idx int, allSegs []*SegmentWithSpeaker) (int, int) {
	var isCheckStaff, isCheckCustomer bool
	var staffIdx, customerIdx int
	//check before exception
	for j := idx - 1; j >= 0; j-- {
		if isCheckStaff && isCheckCustomer {
			break
		}
		if allSegs[j].Speaker == SilenceSpeaker {
			continue
		} else if allSegs[j].Speaker == int(model.CallChanStaff) && !isCheckStaff {
			staffIdx = j
			isCheckStaff = true
		} else if allSegs[j].Speaker == int(model.CallChanCustomer) && !isCheckCustomer {
			customerIdx = j
			isCheckCustomer = true
		}
	}
	return staffIdx, customerIdx
}

//findFirstSegAfterIndex finds the first index of segment, for both staff and customer, which is not silence after the given index
//return the index of the first index for staff and customer  before the given idx
func findFirstSegAfterIndex(idx int, allSegs []*SegmentWithSpeaker) (int, int) {
	var isCheckStaff, isCheckCustomer bool
	var staffIdx, customerIdx int

	for j := idx + 1; j < len(allSegs); j++ {
		if isCheckStaff && isCheckCustomer {
			break
		}
		if allSegs[j].Speaker == SilenceSpeaker {
			continue
		} else if allSegs[j].Speaker == int(model.CallChanStaff) && !isCheckStaff {
			staffIdx = j
			isCheckStaff = true
		} else if allSegs[j].Speaker == int(model.CallChanCustomer) && !isCheckCustomer {
			customerIdx = j
			isCheckCustomer = true
		}
	}
	return staffIdx, customerIdx
}

//counts number of other segments, not from staff and customer
func countNumOfOtherSegsBeforeIdx(idx int, allSegs []*SegmentWithSpeaker) int {
	var numOfOtherSegs int
	for k, v := range allSegs {
		if k >= idx {
			break
		}
		if v.Speaker == SilenceSpeaker {
			numOfOtherSegs++
		}
	}

	return numOfOtherSegs
}

/*
func countNumOfSilenceBeforeIdxs(idxs []int, allSegs []*SegmentWithSpeaker) []int {
	numOfIdx := len(idxs)
	if numOfIdx == 0 {
		return nil
	}

	sortIdx := make([]int, 0, numOfIdx)
	copy(sortIdx, idxs)
	sort.Ints(sortIdx)
	max := sortIdx[numOfIdx-1]

	if max >= len(allSegs) {
		return nil
	}
	numOfSilences := make([]int, numOfIdx)

	var cur int

	var numOfSilence int
	for i := 0; i <= max; i++ {
		if allSegs[i].Speaker == SilenceSpeaker {
			numOfSilence++
		}
		if sortIdx[cur] == i {
			numOfSilences[cur] = numOfSilence
			cur++
		}
	}

	resp := make([]int, numOfIdx)

	for k, v := range idxs {
		for k2, v2 := range sortIdx {
			if v == v2 {
				resp[k] = numOfSilences[k2]
			}
		}
	}
	return resp
}
*/

type senTypeKey struct {
	typ        levelType
	sentenceID int64
}

//NOTICE !!! allSegs includes the silence segments,
//but matched only contains the segments with words, excludes the silence segment,
//which means length of the slice of them are different
//this function doesn't check the len of data, the input data must not be empty
//silenceSegs must sort by duration in descending order
func silenceRuleCheck(sRules []SilenceRuleWithException, tagMatchDat []*MatchedData,
	allSegs []*SegmentWithSpeaker, segs []*SegmentWithSpeaker, silenceSegs []segDuration) ([]RulesException, error) {

	callID := allSegs[0].CallID
	segMatchedTag := extractTagMatchedData(tagMatchDat)

	resp := make([]RulesException, 0, len(sRules))
	for _, rule := range sRules {
		var numOfBreak int

		var violateSegs []int64
		//find which segments break the silence rule
		for _, seg := range silenceSegs {
			if seg.duration <= float64(rule.Seconds) {
				break
			}
			violateSegs = append(violateSegs, allSegs[seg.index].RealSegment.ID)
			numOfBreak++
		}
		var defaultVaild bool
		if numOfBreak <= rule.Times {
			defaultVaild = true
		}

		result := RulesException{RuleID: rule.ID, Typ: levSilenceTyp,
			Whos: Silence, CallID: callID, Valid: defaultVaild, SilenceSegs: violateSegs,
			RuleGroupID: rule.RuleGroupID}

		//generates the all exception sentences result structure
		exceptionMap := make(map[senTypeKey]*ExceptionMatched) //the map with key sentence id and value ExceptionMatch structure
		staffBeforeExceptions := getDExceptionAndBuildMap(rule.Before.Staff, levLStaffSenTyp, exceptionMap)
		result.Exception = append(result.Exception, staffBeforeExceptions...)
		customerBeforeExceptions := getDExceptionAndBuildMap(rule.Before.Customer, levLCustomerSenTyp, exceptionMap)
		result.Exception = append(result.Exception, customerBeforeExceptions...)
		staffAfterExceptions := getDExceptionAndBuildMap(rule.After.Staff, levUStaffSenTyp, exceptionMap)
		result.Exception = append(result.Exception, staffAfterExceptions...)

		//some segments break the rule
		if numOfBreak > 0 {
			var exceptionTimes int
			//check whether the exception happened before the silence segment
			for i := 0; i < numOfBreak; i++ {
				silenceIdx := silenceSegs[i].index

				//find the first sentence index before the silence index
				staffIdx, customerIdx := findFirstSegBeforeIndex(silenceIdx, allSegs)
				numOfSilence := countNumOfOtherSegsBeforeIdx(staffIdx, allSegs)
				staffMatchIdx := staffIdx - numOfSilence
				numOfSilence = countNumOfOtherSegsBeforeIdx(customerIdx, allSegs)
				customerMatchIdx := customerIdx - numOfSilence

				//find the first sentence index after the silence index
				aStaffIdx, _ := findFirstSegAfterIndex(silenceIdx, allSegs)
				numOfSilence = countNumOfOtherSegsBeforeIdx(aStaffIdx, allSegs)
				aStaffMatchIdx := aStaffIdx - numOfSilence

				//do the checking, sentence match
				senMatchDat, _ := SentencesMatch(segMatchedTag, rule.sentences)

				//go through each exception
				for _, exception := range result.Exception {
					//check whether the exception sentence is matched
					if segIdxs, ok := senMatchDat[uint64(exception.SentenceID)]; ok {

						//check each matched segment, whether the segment is at the right posotion
						if matched, ok := exceptionMap[senTypeKey{sentenceID: exception.SentenceID, typ: exception.Typ}]; ok {
							for _, segIdx := range segIdxs {
								matchedIdx := segIdx - 1
								switch matched.Typ {
								case levLStaffSenTyp:
									if staffMatchIdx == matchedIdx {
										matched.Valid = true
										result.Valid = true
									} else {
										continue
									}
								case levLCustomerSenTyp:
									if customerMatchIdx == matchedIdx {
										matched.Valid = true
										result.Valid = true
									} else {
										continue
									}
								case levUStaffSenTyp:
									if aStaffMatchIdx == matchedIdx {
										matched.Valid = true
										result.Valid = true
									} else {
										continue
									}
								default:
									continue
								}

								matchedTags := tagMatchDat[matchedIdx]
								for _, data := range matchedTags.Matched {
									var tagCredit TagCredit
									//TagID
									tagCredit.ID = data.Tag
									tagCredit.Score = data.Score
									//SentenceID is the cu term for segment Idx, which is 1 based index
									tagCredit.SegmentIdx = data.SentenceID
									tagCredit.Match = data.Match
									tagCredit.MatchTxt = data.MatchText
									tagCredit.SegmentID = segs[matchedIdx].ID
									matched.Tags = append(matched.Tags, &tagCredit)
								}
							}
						}
					}

				}
				if result.Valid {
					exceptionTimes++
				}
			}

			if (numOfBreak - exceptionTimes) <= rule.Times {
				result.Valid = true
			}
			if result.Valid {
				if rule.Score > 0 {
					result.Score = rule.Score
				}
			} else {
				if rule.Score < 0 {
					result.Score = rule.Score
				}
			}

		}
		resp = append(resp, result)
	}

	return resp, nil
}
