package qi

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

type levelType int

var (
	levRuleGrpTyp levelType = 1
	levRuleTyp    levelType = 10
	levCFTyp      levelType = 20
	levSenGrpTyp  levelType = 30
	levSenTyp     levelType = 40
	levSegTyp     levelType = 50
	levSWTyp      levelType = 60
)

var unactivate = -1
var matched = 1
var notMatched = 0

var validMap = map[int]bool{
	1: true,
}

//the const variable
const (
	BaseScore = 100
)

//StoreCredit stores the result of the quality
func StoreCredit(call uint64, credit *RuleGrpCredit) error {
	if credit == nil {
		return nil
	}
	var parentID uint64

	s := &model.SimpleCredit{}

	now := time.Now().Unix()

	s.CreateTime = now
	s.CallID = call
	s.OrgID = credit.ID
	s.ParentID = parentID
	s.Score = credit.Score
	s.Type = int(levRuleGrpTyp)
	s.Valid = unactivate
	s.Revise = unactivate

	tx, err := dbLike.Begin()
	if err != nil {
		logger.Error.Printf("get transaction failed. %s\n", err)
		return err
	}
	defer tx.Rollback()

	lastID, err := creditDao.InsertCredit(tx, s)
	if err != nil {
		logger.Error.Printf("insert credit %+v failed. %s\n", s, err)
		return err
	}

	parentID = uint64(lastID)
	for _, rule := range credit.Rules {

		s = &model.SimpleCredit{CallID: call, Type: int(levRuleTyp), ParentID: parentID,
			OrgID: rule.ID, Score: rule.Score, CreateTime: now, Revise: unactivate}
		if rule.Valid {
			s.Valid = matched
		}

		parentRule, err := creditDao.InsertCredit(tx, s)
		if err != nil {
			logger.Error.Printf("insert rule credit %+v failed. %s\n", s, err)
			return err
		}

		for _, cf := range rule.CFs {

			s = &model.SimpleCredit{CallID: call, Type: int(levCFTyp), ParentID: uint64(parentRule),
				OrgID: cf.ID, Score: 0, CreateTime: now, Revise: unactivate}
			if cf.Valid {
				s.Valid = matched
			}

			parentCF, err := creditDao.InsertCredit(tx, s)
			if err != nil {
				logger.Error.Printf("insert conversation flow credit %+v failed. %s\n", s, err)
				return err
			}

			for _, senGrp := range cf.SentenceGrps {

				s = &model.SimpleCredit{CallID: call, Type: int(levSenGrpTyp), ParentID: uint64(parentCF),
					OrgID: senGrp.ID, Score: 0, CreateTime: now, Revise: unactivate}
				if senGrp.Valid {
					s.Valid = matched
				}

				parentSenGrp, err := creditDao.InsertCredit(tx, s)
				if err != nil {
					logger.Error.Printf("insert sentence group credit %+v failed. %s\n", s, err)
					return err
				}

				for _, sen := range senGrp.Sentences {

					s = &model.SimpleCredit{CallID: call, Type: int(levSenTyp), ParentID: uint64(parentSenGrp),
						OrgID: sen.ID, Score: 0, CreateTime: now, Revise: unactivate}
					if sen.Valid {
						s.Valid = matched
					}

					parentSen, err := creditDao.InsertCredit(tx, s)
					if err != nil {
						logger.Error.Printf("insert sentence credit %+v failed. %s\n", s, err)
						return err
					}
					duplicateSegIDMap := make(map[uint64]bool)

					for _, tag := range sen.Tags {
						s := &model.SegmentMatch{SegID: uint64(tag.SegmentID), TagID: tag.ID, Score: tag.Score,
							Match: tag.Match, MatchedText: tag.MatchTxt, CreateTime: now}
						_, err = creditDao.InsertSegmentMatch(tx, s)
						if err != nil {
							logger.Error.Printf("insert matched tag segment  %+v failed. %s\n", s, err)
							return err
						}
						duplicateSegIDMap[uint64(tag.SegmentID)] = true
					}

					if sen.Valid {
						for segID := range duplicateSegIDMap {
							s := &model.SimpleCredit{CallID: call, Type: int(levSegTyp), ParentID: uint64(parentSen),
								OrgID: segID, Score: 0, CreateTime: now, Revise: unactivate, Valid: matched}

							_, err = creditDao.InsertCredit(tx, s)
							if err != nil {
								logger.Error.Printf("insert matched tag segment  %+v failed. %s\n", s, err)
								return err
							}
						}
					}
				}
			}
		}
	}
	return tx.Commit()
}

func errCannotFindParent(id, parent uint64) error {
	_, _, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf("Line[%d] .Cannot find %d's parent %d credit\n", line, id, parent)
	logger.Error.Printf("%s\n", msg)
	return errors.New(msg)
}

//HistoryCredit records the time and its credit
type HistoryCredit struct {
	CreateTime int64            `json:"create_time"`
	Score      int              `json:"score"`
	Credit     []*RuleGrpCredit `json:"credit"`
}

//RetrieveCredit gets the credit by call id
func RetrieveCredit(call uint64) ([]*HistoryCredit, error) {
	//!!MUST make sure the return credits in order from parent to child level
	//parent must be in the front of the child
	credits, err := creditDao.GetCallCredit(dbLike.Conn(), &model.CreditQuery{Calls: []uint64{call}})
	if err != nil {
		logger.Error.Printf("get credits failed\n")
		return nil, err
	}
	if len(credits) == 0 {
		return []*HistoryCredit{}, nil
	}
	var rgIDs, ruleIDs, cfIDs, senGrpIDs, senIDs, segIDs []uint64

	rgCreditsMap := make(map[uint64]*RuleGrpCredit)
	rgSetIDMap := make(map[uint64]*model.GroupWCond)
	rCreditsMap := make(map[uint64]*RuleCredit)
	rSetIDMap := make(map[uint64]*ConversationRuleInRes)
	cfCreditsMap := make(map[uint64]*ConversationFlowCredit)
	cfSetIDMap := make(map[uint64]*ConversationFlowInRes)
	senGrpCreditsMap := make(map[uint64]*SentenceGrpCredit)
	senGrpSetIDMap := make(map[uint64]*SentenceGroupInResponse)
	senCreditsMap := make(map[uint64]*SentenceCredit)
	senSetIDMap := make(map[uint64]*DataSentence)
	//segIDMap := make(map[uint64]*model.SegmentMatch)
	creditTimeMap := make(map[int64]*HistoryCredit)
	senSegMap := make(map[uint64][]uint64)
	senSegDup := make(map[uint64]map[uint64]bool)
	var resp []*HistoryCredit

	for _, v := range credits {
		//fmt.Printf("%d. id:%d org_id:%d parent_id:%d type:%d\n", k, v.ID, v.OrgID, v.ParentID, v.Type)
		switch levelType(v.Type) {
		case levRuleGrpTyp:
			var ok bool
			var history *HistoryCredit
			if history, ok = creditTimeMap[v.CreateTime]; !ok {
				history = &HistoryCredit{CreateTime: v.CreateTime, Score: v.Score}
				creditTimeMap[v.CreateTime] = history
				resp = append(resp, history)
			}
			credit := &RuleGrpCredit{ID: v.OrgID, Score: v.Score}
			history.Credit = append(history.Credit, credit)
			rgCreditsMap[v.ID] = credit
			if set, ok := rgSetIDMap[v.OrgID]; ok {
				credit.Setting = set
			} else {
				set := &model.GroupWCond{}
				rgSetIDMap[v.OrgID] = set
				credit.Setting = set
			}
			rgIDs = append(rgIDs, v.OrgID)
		case levRuleTyp:
			if parentCredit, ok := rgCreditsMap[v.ParentID]; ok {
				credit := &RuleCredit{ID: v.OrgID, Score: v.Score, Valid: validMap[v.Valid]}
				rCreditsMap[v.ID] = credit
				if set, ok := rSetIDMap[v.OrgID]; ok {
					credit.Setting = set
				} else {
					set := &ConversationRuleInRes{}
					rSetIDMap[v.OrgID] = set
					credit.Setting = set
				}
				parentCredit.Rules = append(parentCredit.Rules, credit)
				ruleIDs = append(ruleIDs, v.OrgID)
				//rgCreditMap[v.ParentID] = parentCredit
			} else {
				//return nil, errCannotFindParent(v.ID, v.ParentID)
			}

		case levCFTyp:
			if parentCredit, ok := rCreditsMap[v.ParentID]; ok {
				credit := &ConversationFlowCredit{ID: v.OrgID, Valid: validMap[v.Valid]}
				cfCreditsMap[v.ID] = credit
				if set, ok := cfSetIDMap[v.OrgID]; ok {
					credit.Setting = set
				} else {
					set := &ConversationFlowInRes{}
					cfSetIDMap[v.OrgID] = set
					credit.Setting = set
				}
				parentCredit.CFs = append(parentCredit.CFs, credit)
				cfIDs = append(cfIDs, v.OrgID)
				//rCreditsMap[v.ParentID] = parentCredit
			} else {
				//return nil, errCannotFindParent(v.ID, v.ParentID)
			}

		case levSenGrpTyp:
			if parentCredit, ok := cfCreditsMap[v.ParentID]; ok {
				credit := &SentenceGrpCredit{ID: v.OrgID, Valid: validMap[v.Valid]}
				senGrpCreditsMap[v.ID] = credit
				if set, ok := senGrpSetIDMap[v.OrgID]; ok {
					credit.Setting = set
				} else {
					set := &SentenceGroupInResponse{}
					senGrpSetIDMap[v.OrgID] = set
					credit.Setting = set
				}
				parentCredit.SentenceGrps = append(parentCredit.SentenceGrps, credit)
				senGrpIDs = append(senGrpIDs, v.OrgID)
				//cfCreditsMap[v.ParentID] = parentCredit
			} else {
				//return nil, errCannotFindParent(v.ID, v.ParentID)
			}

		case levSenTyp:
			if parentCredit, ok := senGrpCreditsMap[v.ParentID]; ok {
				credit := &SentenceCredit{ID: v.OrgID, Valid: validMap[v.Valid]}
				senCreditsMap[v.ID] = credit
				//senIDCreditsMap[v.OrgID] = credit
				if set, ok := senSetIDMap[v.OrgID]; ok {
					credit.Setting = set
				} else {
					set := &DataSentence{}
					senSetIDMap[v.OrgID] = set
					credit.Setting = set
				}
				parentCredit.Sentences = append(parentCredit.Sentences, credit)
				senIDs = append(senIDs, v.OrgID)
				//senGrpCreditsMap[v.ParentID] = parentCredit
			} else {
				//return nil, errCannotFindParent(v.ID, v.ParentID)
			}

		case levSegTyp:
			if parentCredit, ok := senCreditsMap[v.ParentID]; ok {
				sID := parentCredit.ID
				if _, ok := senSegDup[sID]; ok {
					if _, ok := senSegDup[sID][v.OrgID]; ok {
						continue
					}
				} else {
					senSegDup[sID] = make(map[uint64]bool)
				}
				senSegDup[sID][v.OrgID] = true
				senSegMap[sID] = append(senSegMap[sID], v.OrgID)
				segIDs = append(segIDs, v.OrgID)
			} else {
				//return nil, errCannotFindParent(v.ID, v.ParentID)
			}
		default:
			//logger.Error.Printf("credit result %d id has the unknown type %d\n", v.ID, v.Type)
			continue
		}
	}

	//get the rule group setting
	if len(rgIDs) > 0 {

		_, groupsSet, err := GetGroupsByFilter(&model.GroupFilter{ID: rgIDs})
		if err != nil {
			logger.Error.Printf("get rule group %+v failed. %s\n", rgIDs, err)
			return nil, err
		}
		for i := 0; i < len(groupsSet); i++ {
			if group, ok := rgSetIDMap[uint64(groupsSet[i].ID)]; ok {
				*group = groupsSet[i]
			} else {
				msg := fmt.Sprintf("return %d rule group doesn't meet request %+v\n", groupsSet[i].ID, rgIDs)
				logger.Error.Printf("%s\n", msg)
				return nil, errors.New(msg)
			}
		}
	}

	//get the rule setting
	if len(ruleIDs) > 0 {
		ruleSet, err := conversationRuleDao.GetBy(&model.ConversationRuleFilter{ID: ruleIDs, Severity: -1, IsDeleted: -1}, sqlConn)
		if err != nil {
			logger.Error.Printf("get rule %+v failed. %s\n", ruleIDs, err)
			return nil, err
		}
		for i := 0; i < len(ruleSet); i++ {
			if rule, ok := rSetIDMap[uint64(ruleSet[i].ID)]; ok {
				ruleInRes := conversationRuleToRuleInRes(&ruleSet[i])
				*rule = *ruleInRes
			} else {
				msg := fmt.Sprintf("return %d rule doesn't meet request %+v\n", ruleSet[i].ID, ruleIDs)
				logger.Error.Printf("%s\n", msg)
				return nil, errors.New(msg)
			}
		}
	}
	//get the conversation flow setting
	if len(cfIDs) > 0 {
		cfSet, err := conversationFlowDao.GetBy(&model.ConversationFlowFilter{ID: cfIDs}, sqlConn)
		if err != nil {
			logger.Error.Printf("get conversation flow %+v failed. %s\n", cfIDs, err)
			return nil, err
		}
		for i := 0; i < len(cfSet); i++ {
			if cf, ok := cfSetIDMap[uint64(cfSet[i].ID)]; ok {
				flowInRes := conversationfFlowToFlowInRes(&cfSet[i])
				*cf = flowInRes
			} else {
				msg := fmt.Sprintf("return %d conversation flow doesn't meet request %+v\n", cfSet[i].ID, cfIDs)
				logger.Error.Printf("%s\n", msg)
				return nil, errors.New(msg)
			}
		}
	}

	//get the sentence group setting
	if len(senGrpIDs) > 0 {
		senGrpSet, err := sentenceGroupDao.GetBy(&model.SentenceGroupFilter{ID: senGrpIDs}, sqlConn)
		if err != nil {
			logger.Error.Printf("get sentence group %+v failed. %s\n", senGrpIDs, err)
			return nil, err
		}
		for i := 0; i < len(senGrpSet); i++ {

			if senGrp, ok := senGrpSetIDMap[uint64(senGrpSet[i].ID)]; ok {
				groupInRes := sentenceGroupToSentenceGroupInResponse(&senGrpSet[i])
				*senGrp = groupInRes
			} else {
				msg := fmt.Sprintf("return %d sentence group doesn't meet request %+v\n", senGrpSet[i].ID, senGrpIDs)
				logger.Error.Printf("%s\n", msg)
				return nil, errors.New(msg)
			}
		}
	}
	//get the sentences setting
	if len(senIDs) > 0 {
		senSet, err := getSentences(&model.SentenceQuery{ID: senIDs})
		if err != nil {
			logger.Error.Printf("get sentence  %+v failed. %s\n", senIDs, err)
			return nil, err
		}
		for _, set := range senSet {
			if sen, ok := senSetIDMap[set.ID]; ok {
				*sen = *set
			} else {
				msg := fmt.Sprintf("return %d sentence  doesn't meet request %+v\n", set.ID, senIDs)
				logger.Error.Printf("%s\n", msg)
				return nil, errors.New(msg)
			}
		}
	}

	//get the matched segments
	if len(segIDs) > 0 {
		segsMatch, err := creditDao.GetSegmentMatch(dbLike.Conn(), &model.SegmentPredictQuery{Segs: segIDs})
		if err != nil {
			logger.Error.Printf("get matched segments  %+v failed. %s\n", segIDs, err)
			return nil, err
		}

		tagToSegMap := make(map[uint64]map[uint64]*model.SegmentMatch)
		for _, matched := range segsMatch {
			if _, ok := tagToSegMap[matched.TagID]; !ok {
				tagToSegMap[matched.TagID] = make(map[uint64]*model.SegmentMatch)
			}
			tagToSegMap[matched.TagID][matched.SegID] = matched
		}

		relation, _, err := GetLevelsRel(LevSentence, LevTag, senIDs, true)
		if err != nil {
			logger.Error.Printf("get relation failed\n")
			return nil, err
		}
		if len(relation) <= 0 {
			logger.Error.Printf("relation table less\n")
			return nil, errors.New("relation table less")
		}

		for _, credit := range senCreditsMap {
			senID := credit.ID
			if tagIDs, ok := relation[0][senID]; ok {
				if segIDs, ok := senSegMap[senID]; ok {
					for _, segID := range segIDs {
						for _, tagID := range tagIDs {
							if matched, ok := tagToSegMap[tagID][segID]; ok {
								credit.MatchedSegments = append(credit.MatchedSegments, matched)
							}
						}
					}
				}
			}
		}
	}
	//desc order
	sort.SliceStable(resp, func(i, j int) bool {
		return resp[i].CreateTime > resp[j].CreateTime
	})

	return resp, nil
}
