package qi

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

// ASRWorkFlow is the workflow of processing asr output.
// the return error is the no ack signal for RabbitMQ consumer.
// if error is nil, then the process is consider either done or unrecoverable error
// TODO: Add a special error type that can distinguish unrecoverable error.
func ASRWorkFlow(output []byte) error {
	var resp ASRResponse
	err := json.Unmarshal(output, &resp)
	if err != nil {
		logger.Error.Println("unrecoverable error: unmarshal asr response failed, ", err, " Body: ", string(output))
		return nil
	}
	c, err := Call(resp.CallID, "")
	if err == ErrNotFound {
		logger.Error.Printf("unrecoverable error: call '%d' no such call exist. \n", resp.CallID)
		return nil
	} else if err != nil {
		return fmt.Errorf("fetch call failed, %v", err)
	}

	if resp.Status != 0 {
		logger.Error.Println("unrecoverable error: asr response status is not ok, CallID: ", resp.CallID, ", Status: ", resp.Status)
		return nil
	}

	c.DurationMillSecond = int(resp.Length * 1000)

	err = UpdateCall(&c)
	if err != nil {
		return fmt.Errorf("update call duration failed, %v", err)
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return fmt.Errorf("can not begin a transaction")
	}
	// defer a clean up function.
	// If any error happened, tx will be revert and status will be marked as failed.
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

	segments := resp.Segments()
	segments, err = segmentDao.NewSegments(tx, segments)
	if err != nil {
		return fmt.Errorf("new segment failed, %v", err)
	}

	sort.SliceStable(segments, func(i, j int) bool {
		return segments[i].StartTime < segments[j].StartTime
	})

	var channelRoles = map[int8]int{
		1: int(c.LeftChanRole),
		2: int(c.RightChanRole),
	}
	segWithSp := make([]*SegmentWithSpeaker, len(segments))
	for _, s := range segments {
		ws := &SegmentWithSpeaker{
			RealSegment: s,
			Speaker:     channelRoles[s.Channel],
		}
		segWithSp = append(segWithSp, ws)
	}
	//TODO: 計算靜音比例跟規則
	isEnabled := true
	groups, err := serviceDAO.Group(tx, model.GroupQuery{
		IsEnable: &isEnabled,
	})
	if err != nil {
		return fmt.Errorf("get groups by call failed, %v", err)
	}
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
	err = UpdateCall(&c)
	logger.Info.Println("finish asr flow for ", resp.CallID)
	if err != nil {
		logger.Error.Printf("inconsistent status error: call '%d' ASR finished, but status update failed. %v", c.ID, err)
		//Dont bother trigger clean up function
		err = nil
	}
	return nil
}

// Segments transfer ASRResponse's sentence to []model.RealSegment
func (resp *ASRResponse) Segments() []model.RealSegment {

	var segments = []model.RealSegment{}
	//TODO: check sret & emotion = -1
	timestamp := time.Now().Unix()
	for _, sen := range resp.LeftChannel.Sentences {
		s := model.RealSegment{
			CallID:     resp.CallID,
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
			CallID:     resp.CallID,
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
	CallID       int64    `json:"call_id,string"`
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
