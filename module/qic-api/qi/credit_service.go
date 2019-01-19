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

var creditDao model.CreditDao

type levelType int

var (
	levRuleGrpTyp levelType = 1
	levRuleTyp    levelType = 10
	levCFTyp      levelType = 20
	levSenGrpTyp  levelType = 30
	levSenTyp     levelType = 40
	levSegTyp     levelType = 50
)

var unactivate = -1
var matched = 1
var notMatched = 0

var validMap = map[int]bool{
	1: true,
}

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

	tx, err := dbLike.Conn().Begin()
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
	Credit     []*RuleGrpCredit `json:"credit"`
}

//RetrieveCredit gets the credit by call id
func RetrieveCredit(call uint64) ([]*HistoryCredit, error) {
	//!!MUST make sure the return credits in order from parent to child level
	//parent must be in the front of the child
	credits, err := creditDao.GetCallCredit(dbLike.Conn(), call)
	if err != nil {
		logger.Error.Printf("get credits failed\n")
		return nil, err
	}

	var rgIDs, ruleIDs, cfIDs, senGrpIDs, senIDs, segIDs []uint64

	rgCreditsMap := make(map[uint64]*RuleGrpCredit)
	rgCreditsIDMap := make(map[uint64]*RuleGrpCredit)
	rCreditsMap := make(map[uint64]*RuleCredit)
	rCreditsIDMap := make(map[uint64]*RuleCredit)
	cfCreditsMap := make(map[uint64]*ConversationFlowCredit)
	cfCreditsIDMap := make(map[uint64]*ConversationFlowCredit)
	senGrpCreditsMap := make(map[uint64]*SentenceGrpCredit)
	senGrpCreditsIDMap := make(map[uint64]*SentenceGrpCredit)
	senCreditsMap := make(map[uint64]*SentenceCredit)
	senCreditsIDMap := make(map[uint64]*SentenceCredit)
	segIDMap := make(map[uint64]*model.SegmentMatch)
	creditTimeMap := make(map[int64]*HistoryCredit)

	var resp []*HistoryCredit

	for _, v := range credits {
		//fmt.Printf("%d. id:%d org_id:%d parent_id:%d type:%d\n", k, v.ID, v.OrgID, v.ParentID, v.Type)
		switch levelType(v.Type) {
		case levRuleGrpTyp:
			var ok bool
			var history *HistoryCredit
			if history, ok = creditTimeMap[v.CreateTime]; !ok {
				history = &HistoryCredit{CreateTime: v.CreateTime}
				creditTimeMap[v.CreateTime] = history
			}
			credit := &RuleGrpCredit{ID: v.OrgID, Score: v.Score}
			history.Credit = append(history.Credit, credit)
			rgCreditsMap[v.ID] = credit
			rgCreditsIDMap[v.OrgID] = credit
			rgIDs = append(rgIDs, v.OrgID)
		case levRuleTyp:
			if parentCredit, ok := rgCreditsMap[v.ParentID]; ok {
				credit := &RuleCredit{ID: v.OrgID, Score: v.Score, Valid: validMap[v.Valid]}
				rCreditsMap[v.ID] = credit
				rCreditsIDMap[v.OrgID] = credit
				parentCredit.Rules = append(parentCredit.Rules, credit)
				//rgCreditMap[v.ParentID] = parentCredit
			} else {
				return nil, errCannotFindParent(v.ID, v.ParentID)
			}
			ruleIDs = append(ruleIDs, v.OrgID)
		case levCFTyp:
			if parentCredit, ok := rCreditsMap[v.ParentID]; ok {
				credit := &ConversationFlowCredit{ID: v.OrgID, Valid: validMap[v.Valid]}
				cfCreditsMap[v.ID] = credit
				cfCreditsIDMap[v.OrgID] = credit
				parentCredit.CFs = append(parentCredit.CFs, credit)
				//rCreditsMap[v.ParentID] = parentCredit
			} else {
				return nil, errCannotFindParent(v.ID, v.ParentID)
			}
			cfIDs = append(cfIDs, v.OrgID)
		case levSenGrpTyp:
			if parentCredit, ok := cfCreditsMap[v.ParentID]; ok {
				credit := &SentenceGrpCredit{ID: v.OrgID, Valid: validMap[v.Valid]}
				senGrpCreditsMap[v.ID] = credit
				senGrpCreditsIDMap[v.OrgID] = credit
				parentCredit.SentenceGrps = append(parentCredit.SentenceGrps, credit)
				//cfCreditsMap[v.ParentID] = parentCredit
			} else {
				return nil, errCannotFindParent(v.ID, v.ParentID)
			}
			senGrpIDs = append(senGrpIDs, v.OrgID)
		case levSenTyp:
			if parentCredit, ok := senGrpCreditsMap[v.ParentID]; ok {
				credit := &SentenceCredit{ID: v.OrgID, Valid: validMap[v.Valid]}
				senCreditsMap[v.ID] = credit
				senCreditsIDMap[v.OrgID] = credit
				parentCredit.Sentences = append(parentCredit.Sentences, credit)
				//senGrpCreditsMap[v.ParentID] = parentCredit
			} else {
				return nil, errCannotFindParent(v.ID, v.ParentID)
			}
			senIDs = append(senIDs, v.OrgID)
		case levSegTyp:
			if parentCredit, ok := senCreditsMap[v.ParentID]; ok {
				seg := &model.SegmentMatch{ID: v.OrgID}
				parentCredit.MatchedSegments = append(parentCredit.MatchedSegments, seg)
				segIDMap[v.OrgID] = seg
			} else {
				return nil, errCannotFindParent(v.ID, v.ParentID)
			}

			segIDs = append(segIDs, v.OrgID)
		default:
			logger.Error.Printf("credit result %d id has the unknow type %d\n", v.ID, v.Type)
			continue
		}
	}

	//get the rule group setting
	_, groupsSet, err := GetGroupsByFilter(&model.GroupFilter{Deal: -1, ID: rgIDs})
	if err != nil {
		logger.Error.Printf("get rule group %+v failed. %s\n", rgIDs, err)
		return nil, err
	}
	for i := 0; i < len(groupsSet); i++ {
		if group, ok := rgCreditsIDMap[uint64(groupsSet[i].ID)]; ok {
			group.Setting = &groupsSet[i]
		} else {
			msg := fmt.Sprintf("return %d rule group doesn't meet request %+v\n", groupsSet[i].ID, rgIDs)
			logger.Error.Printf("%s\n", msg)
			return nil, errors.New(msg)
		}
	}

	//get the rule setting
	ruleSet, err := conversationRuleDao.GetBy(&model.ConversationRuleFilter{ID: ruleIDs, Severity: -1, IsDeleted: -1}, sqlConn)
	if err != nil {
		logger.Error.Printf("get rule %+v failed. %s\n", ruleIDs, err)
		return nil, err
	}
	for i := 0; i < len(ruleSet); i++ {
		if rule, ok := rCreditsIDMap[uint64(ruleSet[i].ID)]; ok {
			ruleInRes := conversationRuleToRuleInRes(&ruleSet[i])
			rule.Setting = ruleInRes
		} else {
			msg := fmt.Sprintf("return %d rule doesn't meet request %+v\n", ruleSet[i].ID, ruleIDs)
			logger.Error.Printf("%s\n", msg)
			return nil, errors.New(msg)
		}
	}

	//get the conversation flow setting
	cfSet, err := conversationFlowDao.GetBy(&model.ConversationFlowFilter{ID: cfIDs, IsDelete: -1}, sqlConn)
	if err != nil {
		logger.Error.Printf("get conversation flow %+v failed. %s\n", cfIDs, err)
		return nil, err
	}
	for i := 0; i < len(cfSet); i++ {
		if cf, ok := cfCreditsIDMap[uint64(cfSet[i].ID)]; ok {
			flowInRes := conversationfFlowToFlowInRes(&cfSet[i])
			cf.Setting = &flowInRes
		} else {
			msg := fmt.Sprintf("return %d conversation flow doesn't meet request %+v\n", cfSet[i].ID, cfIDs)
			logger.Error.Printf("%s\n", msg)
			return nil, errors.New(msg)
		}
	}

	//get the sentence group setting
	senGrpSet, err := sentenceGroupDao.GetBy(&model.SentenceGroupFilter{Role: -1, Position: -1, ID: senGrpIDs, IsDelete: -1}, sqlConn)
	if err != nil {
		logger.Error.Printf("get sentence group %+v failed. %s\n", senGrpIDs, err)
		return nil, err
	}
	for i := 0; i < len(senGrpSet); i++ {
		if senGrp, ok := senGrpCreditsIDMap[uint64(senGrpSet[i].ID)]; ok {
			groupInRes := sentenceGroupToSentenceGroupInResponse(&senGrpSet[i])
			senGrp.Setting = &groupInRes
		} else {
			msg := fmt.Sprintf("return %d sentence group doesn't meet request %+v\n", senGrpSet[i].ID, senGrpIDs)
			logger.Error.Printf("%s\n", msg)
			return nil, errors.New(msg)
		}
	}

	//get the sentences setting
	senSet, err := getSentences(&model.SentenceQuery{ID: senIDs})
	if err != nil {
		logger.Error.Printf("get sentence  %+v failed. %s\n", senIDs, err)
		return nil, err
	}
	for _, set := range senSet {
		if sen, ok := senCreditsIDMap[set.ID]; ok {
			sen.Setting = set
		} else {
			msg := fmt.Sprintf("return %d sentence  doesn't meet request %+v\n", set.ID, senIDs)
			logger.Error.Printf("%s\n", msg)
			return nil, errors.New(msg)
		}
	}

	//get the matched segments
	segsMatch, err := creditDao.GetSegmentMatch(dbLike.Conn(), segIDs)
	if err != nil {
		logger.Error.Printf("get matched segments  %+v failed. %s\n", segIDs, err)
		return nil, err
	}

	for _, matched := range segsMatch {
		if seg, ok := segIDMap[matched.ID]; ok {
			*seg = *matched
		} else {
			msg := fmt.Sprintf("return %d segment matched  doesn't meet request %+v\n", matched.ID, segIDs)
			logger.Error.Printf("%s\n", msg)
			return nil, errors.New(msg)
		}
	}

	//desc order
	sort.SliceStable(resp, func(i, j int) bool {
		return resp[i].CreateTime > resp[j].CreateTime
	})

	return resp, nil
}
