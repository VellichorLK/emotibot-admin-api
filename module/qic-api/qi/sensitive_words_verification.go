package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/anknown/ahocorasick"
	"time"
)

var sentenceMatchFunc func([]string, []uint64, string) (map[uint64][]int, error) = SimpleSentenceMatch

// DoSensitiveWordsVerification　takes callID and segments as input and do sensitive word verification
// it return slice of credits and err if err is happened
func SensitiveWordsVerification(callID int64, segments []SegmentWithSpeaker, enterprise string) (credits []model.SimpleCredit, err error) {
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

	swID := make([]int64, len(sws))
	swMap := map[string]model.SensitiveWord{}
	swNames := make([]string, len(sws))
	for idx, sw := range sws {
		swID[idx] = sw.ID
		swMap[sw.Name] = sw
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
	appendSentenceID(&sids, staffExceptions)
	appendSentenceID(&sids, customerExceptions)

	// get sentence to segment match
	senToSegments, err := sentenceMatchFunc(segContents, sids, enterprise)
	if err != nil {
		return
	}

	// create matching machine
	rnames := general.StringsToRunes(swNames)
	m := new(goahocorasick.Machine)
	if err = m.Build(rnames); err != nil {
		return
	}

	// for each segment check if violate sensitive word
	for idx, seg := range segments {
		if violates := m.MultiPatternSearch([]rune(seg.Text), false); len(violates) > 0 {
			// violate some sensitive words
			for _, term := range violates {
				violated := true
				sw, ok := swMap[string(term.Word)]
				if !ok {
					logger.Warn.Printf("should get sensitive words, but do exist")
					continue
				}

				// verify if pass staff exception condition
				sentences := staffExceptions[sw.ID]
				for _, sid := range sentences {
					if segIndxes, ok := senToSegments[sid]; ok {
						for _, segIdx := range segIndxes {
							if seg.Speaker == int(model.CallChanStaff) && segIdx < idx {
								logger.Info.Printf("segIdx: %d", segIdx)
								logger.Info.Printf("idx: %d", idx)
								logger.Info.Printf("seg: %s", seg.Text)
								violated = false
								break
							}
						}
					}

					if !violated {
						break
					}
				}

				if !violated {
					break
				}

				// verify if pass customer exception condition
				sentences = customerExceptions[sw.ID]
				for _, sid := range sentences {
					if segIndxes, ok := senToSegments[sid]; ok {
						for _, segIdx := range segIndxes {
							if seg.Speaker == int(model.CallChanCustomer) && segIdx < idx {
								violated = false
								break
							}
						}

						if !violated {
							break
						}
					}
				}

				// the segment violates this sensitive word
				if violated {
					now := time.Now().Unix()
					credit := model.SimpleCredit{
						CallID:     uint64(callID),
						Type:       int(levSWTyp),
						OrgID:      uint64(sw.ID),
						Revise:     -1,
						Score:      sw.Score,
						CreateTime: now,
						UpdateTime: now,
					}
					credits = append(credits, credit)
				}
			}
		}
	}

	return

}

func appendSentenceID(ids *[]uint64, sentences map[int64][]uint64) {
	newids := *ids
	for sid := range sentences {
		newids = append(newids, uint64(sid))
	}
	ids = &newids
}
