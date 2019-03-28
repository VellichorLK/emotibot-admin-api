package qi

import (
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
	goahocorasick "github.com/anknown/ahocorasick"
)

var sentenceMatchFunc func([]string, []uint64, string) (map[uint64][]int, error) = SimpleSentenceMatch

var (
	levSWTyp            levelType = 60
	levSWSegTyp         levelType = 61
	levSWUserValTyp     levelType = 62
	levSWStaffSenTyp    levelType = 63
	levSWCustomerSenTyp levelType = 64
	levSWSenSegTyp      levelType = 65
)

//SensitiveWordCredit stores the sensitive word result
type SensitiveWordCredit struct {
	sensitiveWord                model.SimpleCredit
	customerExceptions           []model.SimpleCredit
	usrVals                      []model.SimpleCredit
	staffExceptions              []model.SimpleCredit
	invalidSegments              []model.SimpleCredit
	staffMatchedExceptionSegs    map[int64][]model.SimpleCredit
	customerMatchedExceptionSegs map[int64][]model.SimpleCredit
	usrValsMatched               bool
}

func insertCreditsWithParentID(tx model.SqlLike, credits []model.SimpleCredit, parent int64) error {
	for _, v := range credits {
		v.ParentID = uint64(parent)
		_, err := creditDao.InsertCredit(tx, &v)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertCreditsWthParentIDAndChildCredits(tx model.SqlLike, credits []model.SimpleCredit,
	parent int64, children map[int64][]model.SimpleCredit) error {
	for _, v := range credits {
		v.ParentID = uint64(parent)
		parent, err := creditDao.InsertCredit(tx, &v)
		if err != nil {
			return err
		}
		if descendant, ok := children[int64(v.OrgID)]; ok {
			for _, c := range descendant {
				c.ParentID = uint64(parent)
				_, err = creditDao.InsertCredit(tx, &c)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil

}

//StoreSensitiveCredit stores the sensitive credits
func StoreSensitiveCredit(credits []*SensitiveWordCredit, root int64) error {
	if dbLike == nil {
		return ErrNilCon
	}

	tx, err := dbLike.Begin()
	if err != nil {
		logger.Error.Printf("get transaction failed. %s\n", err)
		return err
	}
	defer tx.Rollback()

	for _, c := range credits {

		if c == nil {
			continue
		}

		c.sensitiveWord.ParentID = uint64(root)
		lastID, err := creditDao.InsertCredit(tx, &c.sensitiveWord)
		if err != nil {
			logger.Error.Printf("insert credit %v failed. %s\n", c.sensitiveWord, err)
			return err
		}

		err = insertCreditsWthParentIDAndChildCredits(tx, c.customerExceptions, lastID, c.customerMatchedExceptionSegs)
		if err != nil {
			logger.Error.Printf("insert credit %+v failed. %s\n", c.customerExceptions, err)
			return err
		}

		err = insertCreditsWthParentIDAndChildCredits(tx, c.staffExceptions, lastID, c.staffMatchedExceptionSegs)
		if err != nil {
			logger.Error.Printf("insert credit %+v failed. %s\n", c.staffExceptions, err)
			return err
		}

		err = insertCreditsWithParentID(tx, c.invalidSegments, lastID)
		if err != nil {
			logger.Error.Printf("insert credit %+v failed. %s\n", c.invalidSegments, err)
			return err
		}

		err = insertCreditsWithParentID(tx, c.usrVals, lastID)
		if err != nil {
			logger.Error.Printf("insert credit %+v failed. %s\n", c.usrVals, err)
			return err
		}
	}
	return tx.Commit()

}

//SensitiveWordsVerificationWithPacked packages the sensitive words result into the structure
func SensitiveWordsVerificationWithPacked(callID int64, segments []*SegmentWithSpeaker, enterprise string) ([]*SensitiveWordCredit, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	sqlConn := dbLike.Conn()
	var deleted int8
	filter := &model.SensitiveWordFilter{
		Enterprise: &enterprise,
		Deleted:    &deleted,
	}

	sws, err := swDao.GetBy(filter, sqlConn)
	if err != nil {
		logger.Error.Printf("get sensitive words failed\n")
		return nil, err
	}

	if len(sws) == 0 {
		return nil, nil
	}

	now := time.Now().Unix()
	swID := make([]int64, len(sws))
	swMap := map[string]model.SensitiveWord{}
	swViolated := map[string]bool{} // records sensitive words which are violated
	swNames := make([]string, len(sws))
	resp := make([]*SensitiveWordCredit, 0, len(sws))
	swCredits := make(map[int64]*SensitiveWordCredit)
	for idx, sw := range sws {
		swID[idx] = sw.ID
		swMap[sw.Name] = sw
		swViolated[sw.Name] = false
		swNames[idx] = sw.Name

		//create the sensitive credits and its exception setting
		c := &SensitiveWordCredit{sensitiveWord: model.SimpleCredit{
			OrgID: uint64(sw.ID), CallID: uint64(callID), Type: int(levSWTyp), Valid: 1, CreateTime: now, UpdateTime: now,
		}}
		for _, e := range sw.CustomerException {
			c.customerExceptions = append(c.customerExceptions, model.SimpleCredit{
				OrgID: e.ID, CallID: uint64(callID), Type: int(levSWCustomerSenTyp), CreateTime: now, UpdateTime: now})
		}
		for _, e := range sw.StaffException {
			c.staffExceptions = append(c.staffExceptions, model.SimpleCredit{
				OrgID: e.ID, CallID: uint64(callID), Type: int(levSWStaffSenTyp), CreateTime: now, UpdateTime: now})
		}
		resp = append(resp, c)
		swCredits[sw.ID] = c
	}

	staffExceptions, customerExceptions, err := swDao.GetRels(swID, sqlConn)
	if err != nil {
		logger.Error.Printf("get the relations failed. %s\n", err)
		return nil, err
	}

	segContents := make([]string, len(segments))
	for idx, seg := range segments {
		segContents[idx] = seg.Text
	}

	sids := []uint64{} // sentence ids
	sids = appendSentenceID(sids, staffExceptions)
	sids = appendSentenceID(sids, customerExceptions)

	// get sentence to segment match
	senToSegments := map[uint64][]int{}
	if len(sids) > 0 {
		senToSegments, err = sentenceMatchFunc(segContents, sids, enterprise)
		if err != nil {
			logger.Error.Printf("call sentence matched function failed. %s\n", err)
			return nil, err
		}
	}

	// create matching machine
	rnames := general.StringsToRunes(swNames)
	m := new(goahocorasick.Machine)
	if err = m.Build(rnames); err != nil {
		return nil, err
	}

	// sensitive word passed maps
	passedMap, err := callToSWUserKeyValues(callID, swID, sqlConn)
	if err != nil {
		return nil, err
	}

	//sets the usr val exception
	for swid, userValueID := range passedMap {
		for _, vid := range userValueID {
			credit := model.SimpleCredit{
				CallID:     uint64(callID),
				Type:       int(levSWUserValTyp),
				OrgID:      uint64(vid),
				ParentID:   uint64(swid),
				Revise:     unactivate,
				Valid:      0,
				Score:      0,
				CreateTime: now,
				UpdateTime: now,
			}
			if c, ok := swCredits[swid]; ok {
				c.usrVals = append(c.usrVals, credit)
			}
		}
	}

	// for each segment check if violate sensitive word
	for segIdx, seg := range segments {
		if seg.Speaker == int(model.CallChanCustomer) {
			// ignore what customer said
			continue
		}
		if violates := m.MultiPatternSearch([]rune(seg.Text), false); len(violates) > 0 {
			for _, term := range violates {
				sw, ok := swMap[string(term.Word)]
				if !ok {
					logger.Warn.Printf("should get sensitive words, but doesn't exist")
					continue
				}
				if swCredit, ok := swCredits[sw.ID]; ok {

					//add the invalid segment which means that segment contains the sensitive words
					credit := model.SimpleCredit{
						CallID:     uint64(callID),
						Type:       int(levSWSegTyp),
						OrgID:      uint64(seg.ID),
						ParentID:   uint64(sw.ID),
						Revise:     unactivate,
						Valid:      unactivate,
						Score:      0,
						CreateTime: now,
						UpdateTime: now,
					}
					swCredit.invalidSegments = append(swCredit.invalidSegments, credit)

					passed := false

					// check if the call passed user value
					if vals, ok := passedMap[sw.ID]; ok && len(vals) != 0 {
						passed = true
						swCredit.usrValsMatched = true
					}

					// check if the segment is a staff exception sentence
					exceptions := sw.StaffException
					for _, sen := range exceptions {
						matchedSegIdxes := senToSegments[sen.ID]
						for _, idx := range matchedSegIdxes {
							if segIdx == idx && seg.Speaker == int(model.CallChanStaff) {
								passed = true

								credit := model.SimpleCredit{
									CallID:     uint64(callID),
									Type:       int(levSWSenSegTyp),
									OrgID:      uint64(seg.ID),
									ParentID:   uint64(sen.ID),
									Revise:     unactivate,
									Valid:      unactivate,
									Score:      0,
									CreateTime: now,
									UpdateTime: now,
								}
								swCredit.staffMatchedExceptionSegs[int64(sen.ID)] = append(swCredit.staffMatchedExceptionSegs[int64(sen.ID)], credit)
								break
							}
						}
					}

					// check if customer has say exception sentences
					exceptions = sw.CustomerException
					for _, sen := range exceptions {
						matchedSegIdxes := senToSegments[sen.ID]
						logger.Info.Printf("matchedSegIdxes: %+v\n", matchedSegIdxes)
						for _, idx := range matchedSegIdxes {
							matchedSeg := segments[idx]
							if idx < segIdx && matchedSeg.Speaker == int(model.CallChanCustomer) {
								passed = true

								credit = model.SimpleCredit{
									CallID:     uint64(callID),
									Type:       int(levSWSenSegTyp),
									OrgID:      uint64(matchedSeg.ID),
									ParentID:   uint64(sen.ID),
									Revise:     unactivate,
									Valid:      unactivate,
									Score:      0,
									CreateTime: now,
									UpdateTime: now,
								}
								swCredit.customerMatchedExceptionSegs[int64(sen.ID)] = append(swCredit.customerMatchedExceptionSegs[int64(sen.ID)], credit)
								break
							}
						}
					}

					if !passed {
						swCredit.sensitiveWord.Valid = 0
						swCredit.sensitiveWord.Score = -sw.Score
					}
				}
			}
		}
	}

	for _, v := range resp {
		//sets the sentence with exception to be true
		for _, e := range v.customerExceptions {
			if len(v.customerMatchedExceptionSegs[int64(e.OrgID)]) != 0 {
				e.Valid = 1
			}
		}
		for _, e := range v.staffExceptions {
			if len(v.staffMatchedExceptionSegs[int64(e.OrgID)]) != 0 {
				e.Valid = 1
			}
		}
		//if usr val is matched, set the valid
		if v.usrValsMatched {
			for _, usrVal := range v.usrVals {
				usrVal.Valid = 1
			}
		}

	}

	return resp, nil
}

// SensitiveWordsVerification　takes callID and segments as input and do sensitive word verification
// it return slice of credits and err if err is happened
func SensitiveWordsVerification(callID int64, segments []*SegmentWithSpeaker, enterprise string) (credits []model.SimpleCredit, err error) {
	// get sensitive words and its settings
	sqlConn := dbLike.Conn()
	var deleted int8
	filter := &model.SensitiveWordFilter{
		Enterprise: &enterprise,
		Deleted:    &deleted,
	}
	sws, err := swDao.GetBy(filter, sqlConn)
	if err != nil {
		return
	}

	if len(sws) == 0 {
		return
	}

	swID := make([]int64, len(sws))
	swMap := map[string]model.SensitiveWord{}
	swViolated := map[string]bool{} // records sensitive words which are violated
	swNames := make([]string, len(sws))
	for idx, sw := range sws {
		swID[idx] = sw.ID
		swMap[sw.Name] = sw
		swViolated[sw.Name] = false
		swNames[idx] = sw.Name
	}

	staffExceptions, customerExceptions, err := swDao.GetRels(swID, sqlConn)
	if err != nil {
		return
	}

	segContents := make([]string, len(segments))
	for idx, seg := range segments {
		segContents[idx] = seg.Text
	}

	sids := []uint64{} // sentence ids
	sids = appendSentenceID(sids, staffExceptions)
	sids = appendSentenceID(sids, customerExceptions)

	// get sentence to segment match
	senToSegments := map[uint64][]int{}
	if len(sids) > 0 {
		senToSegments, err = sentenceMatchFunc(segContents, sids, enterprise)
		if err != nil {
			return
		}
	}

	// create matching machine
	rnames := general.StringsToRunes(swNames)
	m := new(goahocorasick.Machine)
	if err = m.Build(rnames); err != nil {
		return
	}

	// sensitive word passed maps
	passedMap, err := callToSWUserKeyValues(callID, swID, sqlConn)
	if err != nil {
		return
	}

	now := time.Now().Unix()
	for swid, userValueID := range passedMap {
		for _, vid := range userValueID {
			credit := model.SimpleCredit{
				CallID:     uint64(callID),
				Type:       int(levSWUserValTyp),
				OrgID:      uint64(vid),
				ParentID:   uint64(swid),
				Revise:     unactivate,
				Valid:      unactivate,
				Score:      0,
				CreateTime: now,
				UpdateTime: now,
			}
			credits = append(credits, credit)
		}
	}

	// for each segment check if violate sensitive word
	for segIdx, seg := range segments {
		if seg.Speaker == int(model.CallChanCustomer) {
			// ignore what customer said
			continue
		}
		if violates := m.MultiPatternSearch([]rune(seg.Text), false); len(violates) > 0 {
			for _, term := range violates {
				sw := swMap[string(term.Word)]
				credit := model.SimpleCredit{
					CallID:     uint64(callID),
					Type:       int(levSWSegTyp),
					OrgID:      uint64(seg.ID),
					ParentID:   uint64(sw.ID),
					Revise:     unactivate,
					Valid:      unactivate,
					Score:      0,
					CreateTime: now,
					UpdateTime: now,
				}
				credits = append(credits, credit)

				passed := false
				sw, ok := swMap[string(term.Word)]
				if !ok {
					logger.Warn.Printf("should get sensitive words, but do exist")
					continue
				}

				// check if the call passed user value
				if vals, ok := passedMap[sw.ID]; ok && len(vals) != 0 {
					passed = true

				}

				// check if the segment is a staff exception sentence
				exceptions := sw.StaffException
				for _, sen := range exceptions {
					matchedSegIdxes := senToSegments[sen.ID]
					for _, idx := range matchedSegIdxes {
						if segIdx == idx && seg.Speaker == int(model.CallChanStaff) {
							passed = true
							credit = model.SimpleCredit{
								CallID:     uint64(callID),
								Type:       int(levSWStaffSenTyp),
								OrgID:      uint64(sen.ID),
								ParentID:   uint64(sw.ID),
								Revise:     unactivate,
								Score:      0,
								Valid:      1,
								CreateTime: now,
								UpdateTime: now,
							}
							credits = append(credits, credit)

							credit = model.SimpleCredit{
								CallID:     uint64(callID),
								Type:       int(levSWSenSegTyp),
								OrgID:      uint64(seg.ID),
								ParentID:   uint64(sen.ID),
								Revise:     unactivate,
								Valid:      unactivate,
								Score:      0,
								CreateTime: now,
								UpdateTime: now,
							}
							credits = append(credits, credit)
							break
						}
					}
				}

				// check if customer has say exception sentences
				exceptions = sw.CustomerException
				for _, sen := range exceptions {
					matchedSegIdxes := senToSegments[sen.ID]
					logger.Info.Printf("matchedSegIdxes: %+v\n", matchedSegIdxes)
					for _, idx := range matchedSegIdxes {
						matchedSeg := segments[idx]
						if idx < segIdx && matchedSeg.Speaker == int(model.CallChanCustomer) {
							passed = true
							credit = model.SimpleCredit{
								CallID:     uint64(callID),
								Type:       int(levSWCustomerSenTyp),
								OrgID:      uint64(sen.ID),
								ParentID:   uint64(sw.ID),
								Revise:     unactivate,
								Valid:      1,
								Score:      0,
								CreateTime: now,
								UpdateTime: now,
							}
							credits = append(credits, credit)

							credit = model.SimpleCredit{
								CallID:     uint64(callID),
								Type:       int(levSWSenSegTyp),
								OrgID:      uint64(matchedSeg.ID),
								ParentID:   uint64(sen.ID),
								Revise:     unactivate,
								Valid:      unactivate,
								Score:      0,
								CreateTime: now,
								UpdateTime: now,
							}
							credits = append(credits, credit)
							break
						}
					}
				}

				swViolated[sw.Name] = !passed
			}

		}
	}

	// for each sensitive words, check if any violated
	for name, violated := range swViolated {
		sw := swMap[name]
		credit := model.SimpleCredit{
			CallID:     uint64(callID),
			Type:       int(levSWTyp),
			OrgID:      uint64(sw.ID),
			Revise:     unactivate,
			Valid:      1,
			CreateTime: now,
			UpdateTime: now,
		}

		if violated {
			credit.Valid = 0
			//notice,currently sensitive words score use the positive number as the violated score
			credit.Score = -sw.Score
		}
		credits = append(credits, credit)
	}
	return
}

// callToSWUserKeyValues takes callID, slice of sensitive word id, and sqlLike as input
// and returns a map which indicates if the call passes a sensitive word or not
// if some error happened, it will returns the error
func callToSWUserKeyValues(callID int64, sws []int64, sqlLike model.SqlLike) (passedMap map[int64][]int64, err error) {
	// init map
	passedMap = map[int64][]int64{}
	for _, swid := range sws {
		passedMap[swid] = []int64{}
	}

	// get custom values of the call
	query := model.UserValueQuery{
		Type:             []int8{model.UserValueTypCall},
		ParentID:         []int64{callID},
		IgnoreSoftDelete: true,
	}
	callValues, err := userValues(sqlLike, query)
	if err != nil {
		return
	}

	if len(callValues) == 0 {
		return
	}

	usrCallVals := make([]string, 0, len(callValues))
	for _, v := range callValues {
		usrCallVals = append(usrCallVals, v.Value)
	}

	// get custom values of all sensitive words
	query = model.UserValueQuery{
		Type:             []int8{model.UserValueTypSensitiveWord},
		Values:           usrCallVals,
		IgnoreSoftDelete: true,
	}
	swValues, err := userValues(sqlLike, query)
	if err != nil {
		return
	}

	// set true to the sensitive word if custom values of the call exist in custom values of a sensitive word
	for _, cv := range swValues {
		if _, ok := passedMap[cv.LinkID]; !ok {
			passedMap[cv.LinkID] = []int64{}
		}
		passedMap[cv.LinkID] = append(passedMap[cv.LinkID], cv.ID)
	}
	return
}

func appendSentenceID(ids []uint64, sentences map[int64][]uint64) []uint64 {
	for sid := range sentences {
		ids = append(ids, uint64(sid))
	}
	return ids
}
