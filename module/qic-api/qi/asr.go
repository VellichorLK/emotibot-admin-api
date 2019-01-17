package qi

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	var (
		staffChannels    []vChannel
		customerChannels []vChannel
	)
	switch c.LeftChanRole {
	case model.CallChanStaff:
		staffChannels = append(staffChannels, resp.LeftChannel)
	case model.CallChanCustomer:
		customerChannels = append(customerChannels, resp.LeftChannel)
	default:
		return fmt.Errorf("call's left channel role '%d' is unknown", c.LeftChanRole)
	}
	switch c.RightChanRole {
	case model.CallChanStaff:
		staffChannels = append(staffChannels, resp.RightChannel)
	case model.CallChanCustomer:
		customerChannels = append(customerChannels, resp.RightChannel)
	default:
		return fmt.Errorf("call's left channel role '%d' is unknown", c.LeftChanRole)
	}
	var segments = []model.RealSegment{}
	// for _, ch := range staffChannels {
	// 	for _, sen := range ch.Sentences {
	// 		s := model.RealSegment{
	// 			CallID:    callID,
	// 			StartTime: sen.Start,
	// 			EndTime:   sen.End,
	// 			// Channel:
	// 			Text: sen.ASR,
	// 			Emotions: []model.RealSegmentEmotion{
	// 				model.RealSegmentEmotion{
	// 					Typ:   model.ETypAngry,
	// 					Score: sen.Emotion,
	// 				},
	// 			},
	// 		}
	// 		segments = append(segments, s)
	// 	}
	// }

	tx, err := dbLike.Begin()
	if err != nil {
		return fmt.Errorf("can not begin a transaction")
	}
	segments, err = segmentDao.NewSegments(tx, segments)
	if err != nil {
		return fmt.Errorf("new segment failed, %v", err)
	}

	return nil
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
