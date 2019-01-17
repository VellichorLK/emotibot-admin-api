package qi

import (
	"encoding/json"

	"emotibot.com/emotigo/pkg/logger"
)

func ASRWorkFlow(output []byte) error {
	var resp ASRResponse
	err := json.Unmarshal(output, &resp)
	if err != nil {
		logger.Error.Println("unmarshal asr response failed, ", err)
		return err
	}
	// model.
	return nil
}

// ASRResponse
type ASRResponse struct {
	Version      float64  `json:"version"`
	Ret          int64    `json:"ret"`
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
