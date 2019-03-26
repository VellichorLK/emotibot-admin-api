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
	levSWTyp        levelType = 60
	levSWSegTyp     levelType = 61
	levSWUserValTyp levelType = 62
	levSWSenTyp     levelType = 63
	levSWSenSegTyp  levelType = 64
)

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
				if _, ok := passedMap[sw.ID]; ok {
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
								Type:       int(levSWSenTyp),
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
								Type:       int(levSWSenTyp),
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
