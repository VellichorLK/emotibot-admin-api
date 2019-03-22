package qi

import (
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	ruleInterposalDao model.InterposalRuleDao = &model.InterposalRuleSQLDao{}
)

//NewRuleInterposal creates the new interposal rule
func NewRuleInterposal(r *model.InterposalRule, enterprise string) (string, error) {
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
	_, err = ruleInterposalDao.Add(dbLike.Conn(), r)
	return r.UUID, err
}

//GetRuleInterposals gets the list of interposal rule
func GetRuleInterposals(q *model.GeneralQuery, p *model.Pagination) ([]*model.InterposalRule, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	return ruleInterposalDao.Get(dbLike.Conn(), q, p)
}

//CountRuleInterposal counts the total number of interposal rule
func CountRuleInterposal(q *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	return ruleInterposalDao.Count(dbLike.Conn(), q)
}

//DeleteRuleInterposal deletes the interposal rule
func DeleteRuleInterposal(q *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNoID
	}
	return ruleInterposalDao.SoftDelete(dbLike.Conn(), q)
}

//UpdateRuleInterposal updates the interposal rule
func UpdateRuleInterposal(q *model.GeneralQuery, d *model.InterposalUpdateSet) (int64, error) {
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

	current, err := ruleInterposalDao.Get(tx, q, &model.Pagination{Limit: 10})
	if err != nil {
		return 0, err
	}
	if len(current) == 0 {
		return 0, ErrNoSuchID
	}

	q.ID = append(q.ID, current[0].ID)
	newID, err := ruleInterposalDao.Copy(tx, q)
	if err != nil {
		logger.Error.Printf("copy new record %v failed. %s\n", *q, err)
		return 0, err
	}

	affected, err := ruleInterposalDao.SoftDelete(tx, q)
	if err != nil {
		logger.Error.Printf("delete failed. %s\n", err)
		return 0, err
	}
	if affected == 0 {
		return 0, ErrNoSuchID
	}

	newQuery := &model.GeneralQuery{ID: []int64{newID}}
	affected, err = ruleInterposalDao.Update(tx, newQuery, d)
	if err != nil {
		logger.Error.Printf("update failed. %s\n", err)
		return 0, err
	}
	tx.Commit()
	return affected, err

}

//InterposalSpeaker indicates the speak in SegmentWithSpeaker strcuture for interposal segments
const InterposalSpeaker = -2

//RuleInterposalCheck checks the interposal rules
func RuleInterposalCheck(ruleGroup model.Group, segs []*SegmentWithSpeaker) ([]RulesException, error) {
	if len(segs) == 0 {
		return nil, nil
	}
	if dbLike == nil {
		return nil, ErrNilCon
	}

	callID := segs[0].CallID

	isDelete := 0
	q := &model.GeneralQuery{Enterprise: &ruleGroup.EnterpriseID, IsDelete: &isDelete, UUID: []string{ruleGroup.UUID}}
	rules, err := ruleInterposalDao.GetByRuleGroup(dbLike.Conn(), q)
	if err != nil {
		logger.Error.Printf("get rule silence failed. %s\n", err)
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}

	interposalSegs, _ := extractOtherSegment(segs, InterposalSpeaker)

	resp := make([]RulesException, 0, len(rules))
	for _, r := range rules {

		var numOfBreak int
		var violateSegs []int64
		for _, seg := range interposalSegs {
			if seg.duration <= float64(r.Seconds) {
				break
			}
			violateSegs = append(violateSegs, segs[seg.index].RealSegment.ID)
			numOfBreak++
		}
		var defaultVaild bool
		if numOfBreak <= r.Times {
			defaultVaild = true
		}

		result := RulesException{RuleID: r.ID, Typ: levInterposalTyp,
			Whos: Interposal, CallID: callID, Valid: defaultVaild, InterposalSegs: violateSegs,
			RuleGroupID: ruleGroup.ID}

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
