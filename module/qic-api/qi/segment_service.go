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

	channelsRole := map[int8]string{
		1: callRoleTypStr(call.LeftChanRole),
		2: callRoleTypStr(call.RightChanRole),
	}
	for index, s := range segments {

		vr := voiceResult{
			SentenceID: int64(index + 1),
			StartTime:  s.StartTime,
			EndTime:    s.EndTime,
			Speaker:    channelsRole[s.Channel],
			ASRText:    s.Text,
			Sret:       200,
			SegmentID:  s.ID,
		}
		//Since we dont have ster in db, we have to manually check it.
		if vr.ASRText == "" {
			vr.Sret = 500
		}
		result = append(result, vr)
	}

	return result, nil
}
