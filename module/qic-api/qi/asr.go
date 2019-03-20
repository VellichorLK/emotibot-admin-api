package qi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

// ASRWorkFlow is the workflow of processing asr output.
// the return error is the no ack signal for RabbitMQ consumer.
// if error is nil, then the process is consider either done or unrecoverable error
// TODO: Add a special error type that can distinguish unrecoverable error.
func ASRWorkFlow(output []byte) error {
	logger.Trace.Println("ASR workflow started")

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

	if resp.Status != 0 {
		logger.Error.Println("unrecoverable error: asr response status is not ok, CallID: ", resp.CallID, ", Status: ", resp.Status)
		return nil
	}

	c.DurationMillSecond = int(resp.Length * 1000)

	if volume == "" {
		return fmt.Errorf("volume is not exist, please contact ops and check init log for volume init error.")
	}

	if resp.Mp3 != nil {
		match, _ := regexp.MatchString("\\S+.mp3", *resp.Mp3)
		if match {
			s := strings.Split(*resp.Mp3, "/")
			if len(s) > 0 {
				fp := fmt.Sprint(s[len(s)-1])
				c.DemoFilePath = &fp
			}
		}
	}

	err = UpdateCall(&c)
	if err != nil {
		return fmt.Errorf("update call duration failed, %v", err)
	}

	segments := resp.Segments()

	switch c.Type {
	case model.CallTypeWholeFile:
		logger.Trace.Println("Create segments returned from ASR.")
		segments, err = segmentDao.NewSegments(tx, segments)
		if err != nil {
			return fmt.Errorf("new segment failed, %v", err)
		}
	case model.CallTypeRealTime:
		logger.Trace.Println("Create segment emotions returned from ASR.")
		emotions := make([]model.RealSegmentEmotion, 0)
		for _, segment := range segments {
			for _, emotion := range segment.Emotions {
				emotions = append(emotions, emotion)
			}
		}

		err = segmentDao.NewEmotions(tx, emotions)
		if err != nil {
			return fmt.Errorf("new emotions failed, %v", err)
		}
	}

	sort.SliceStable(segments, func(i, j int) bool {
		return segments[i].StartTime < segments[j].StartTime
	})

	var channelRoles = map[int8]int{
		1: int(c.LeftChanRole),
		2: int(c.RightChanRole),
	}
	segWithSp := make([]*SegmentWithSpeaker, 0, len(segments))
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

	score := BaseScore
	credits, err := RuleGroupCriteria(groups, segWithSp, time.Duration(30)*time.Minute)
	if err != nil {
		return fmt.Errorf("get rule group credit failed, %v", err)
	}
	for _, credit := range credits {
		score += credit.Score
	}

	err = StoreCredit(uint64(c.ID), credits)
	if err != nil {
		return fmt.Errorf("store credit failed, %v", err)
	}

	swCredits, err := SensitiveWordsVerification(resp.CallID, segWithSp, c.EnterpriseID)
	if err != nil {
		return err
	}
	err = creditDao.InsertCredits(tx, swCredits)

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
					SegmentID: sen.SegmentID,
					Typ:       model.ETypAngry,
					Score:     sen.Emotion,
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
					SegmentID: sen.SegmentID,
					Typ:       model.ETypAngry,
					Score:     sen.Emotion,
				},
			},
		}
		segments = append(segments, s)
	}
	return segments
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
	Mp3          *string  `json:"mp3"`
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
	Status    int64   `json:"sret"`
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
	ASR       string  `json:"asr"`
	Emotion   float64 `json:"emotion"`
	SegmentID int64   `json:"segment_id"`
}
