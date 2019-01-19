package qi

import (
	"fmt"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

func getSegments(call model.Call) ([]voiceResult, error) {
	segments, err := segmentDao.Segments(nil, model.SegmentQuery{
		CallID: []int64{call.ID},
	})
	if err != nil {
		return nil, fmt.Errorf("get segments failed, %v", err)
	}
	var result = make([]voiceResult, 0, len(segments))
	for index, s := range segments {
		vr := voiceResult{
			SentenceID: int64(index + 1),
			StartTime:  s.StartTime,
			EndTime:    s.EndTime,
			ASRText:    s.Text,
			Sret:       0,
		}
		result = append(result, vr)
	}

	return result, nil
}
