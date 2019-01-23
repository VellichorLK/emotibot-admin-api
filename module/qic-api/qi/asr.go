package qi

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

func ASRWorkFlow(output []byte) error {
	var resp ASRResponse
	err := json.Unmarshal(output, &resp)
	if err != nil {
		return fmt.Errorf("unmarshal asr response failed, %v", err)
	}
	if resp.Status != 0 {
		return fmt.Errorf("asr status is non-success %d", resp.Status)
	}
	callID, err := strconv.ParseInt(resp.CallID, 10, 64)
	if err != nil {
		return fmt.Errorf("asr result's call_id '%s' is not a valid int.", resp.CallID)
	}
	calls, err := callDao.Calls(nil, model.CallQuery{
		ID: []int64{callID},
	})
	if err != nil {
		return fmt.Errorf("fetch call failed, %v", err)
	}
	if len(calls) == 0 {
		return fmt.Errorf("call '%d' can not be found", callID)
	}
	c := calls[0]
	c.DurationMillSecond = int(resp.Length * 1000)
	err = UpdateCall(&c)
	if err != nil {
		return fmt.Errorf("call update status failed, %v", err)
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return fmt.Errorf("can not begin a transaction")
	}
	defer func() {
		if err != nil {
			//We need to release tx before update call, or it may be locked.
			tx.Rollback()
			c.Status = model.CallStatusFailed
			updateErr := UpdateCall(&c)
			if updateErr != nil {
				logger.Error.Println("update call critical failed, ", updateErr)
			}
		}
	}()
	var channelRoles = map[int8]int{
		1: int(c.LeftChanRole),
		2: int(c.RightChanRole),
	}
	var segments = []model.RealSegment{}
	//TODO: check sret & emotion = -1
	timestamp := time.Now().Unix()
	for _, sen := range resp.LeftChannel.Sentences {
		s := model.RealSegment{
			CallID:     callID,
			CreateTime: timestamp,
			StartTime:  sen.Start,
			EndTime:    sen.End,
			Channel:    1,
			Text:       sen.ASR,
			Emotions: []model.RealSegmentEmotion{
				model.RealSegmentEmotion{
					Typ:   model.ETypAngry,
					Score: sen.Emotion,
				},
			},
		}
		segments = append(segments, s)
	}

	for _, sen := range resp.RightChannel.Sentences {
		s := model.RealSegment{
			CallID:     callID,
			CreateTime: timestamp,
			StartTime:  sen.Start,
			EndTime:    sen.End,
			Channel:    2,
			Text:       sen.ASR,
			Emotions: []model.RealSegmentEmotion{
				model.RealSegmentEmotion{
					Typ:   model.ETypAngry,
					Score: sen.Emotion,
				},
			},
		}
		segments = append(segments, s)
	}

	segments, err = segmentDao.NewSegments(tx, segments)
	if err != nil {
		return fmt.Errorf("new segment failed, %v", err)
	}

	sort.SliceStable(segments, func(i, j int) bool {
		return segments[i].StartTime < segments[j].StartTime
	})
	segWithSp := make([]*SegmentWithSpeaker, 0, len(segments))
	for _, s := range segments {
		ws := &SegmentWithSpeaker{
			RealSegment: s,
			Speaker:     channelRoles[s.Channel],
		}
		segWithSp = append(segWithSp, ws)
	}
	//TODO: calculate the 語速 & 靜音比
	callGroups, err := serviceDAO.GroupsByCalls(tx, model.CallQuery{ID: []int64{c.ID}})
	if err != nil {
		return fmt.Errorf("get groups by call failed, %v", err)
	}
	groups := callGroups[c.ID]
	credits := []*RuleGrpCredit{}
	score := BaseScore
	for _, grp := range groups {
		var credit *RuleGrpCredit
		if !grp.IsEnable {
			continue
		}
		credit, err = RuleGroupCriteria(grp, segWithSp, time.Duration(30)*time.Minute)
		if err != nil {
			return fmt.Errorf("get rule group credit failed, %v", err)
		}
		credits = append(credits, credit)
		score += credit.Score
	}
	for _, credit := range credits {
		credit.Score = score
		err = StoreCredit(uint64(c.ID), credit)
		if err != nil {
			return fmt.Errorf("store credit failed, %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit sql failed, %v", err)
	}
	c.Status = model.CallStatusDone
	c.LeftSpeed = &resp.LeftChannel.Speed
	c.RightSpeed = &resp.RightChannel.Speed
	c.LeftSilenceTime = &resp.LeftChannel.Quiet
	c.RightSilenceTime = &resp.RightChannel.Quiet
	err = callDao.SetCall(nil, c)
	if err != nil {
		logger.Error.Println("ASR finished, but status update failed. It will cause an unsync status. error: ", err)
	}
	logger.Info.Println("finish asr flow for ", resp.CallID)
	return nil
}

type SegmentWithSpeaker struct {
	model.RealSegment
	Speaker int
}

// ASRResponse
type ASRResponse struct {
	Version      float64  `json:"version"`
	Status       int64    `json:"ret"`
	CallID       string   `json:"call_id"`
	Length       float64  `json:"length"`
	LeftChannel  vChannel `json:"left_channel"`
	RightChannel vChannel `json:"right_channel"`
}

// vChannel is the voice channel from ASR Result.
type vChannel struct {
	Speed     float64         `json:"speed"`
	Quiet     float64         `json:"quiet"`
	Emotion   float64         `json:"emotion"`
	Sentences []voiceSentence `json:"sentences"`
}

type voiceSentence struct {
	Status  int64   `json:"sret"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	ASR     string  `json:"asr"`
	Emotion float64 `json:"emotion"`
}
