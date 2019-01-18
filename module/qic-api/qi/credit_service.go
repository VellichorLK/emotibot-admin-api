package qi

import (
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
						s := &model.SegmentMatch{SegID: tag.SegmentID, TagID: tag.ID, Score: tag.Score,
							Match: tag.Match, MatchedText: tag.MatchTxt, CreateTime: now}
						_, err = creditDao.InsertSegmentMatch(tx, s)
						if err != nil {
							logger.Error.Printf("insert matched tag segment  %+v failed. %s\n", s, err)
							return err
						}
						duplicateSegIDMap[tag.SegmentID] = true
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

/*
//RetrieveCredit gets the credit by call id
func RetrieveCredit(call uint64) (*RuleGrpCredit, error) {

}
*/
