package qi

import (
	"fmt"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

// Segments retrive model's RealSegment
func Segments(query model.SegmentQuery) ([]model.RealSegment, error) {
	return segments(nil, query)
}

// getSegments get the responseSegment for GET calls api.
// It is not designed to be used with a more broadly usage. Use Segments instead.
// It only retrive the segments of channel 1 & 2.
func getSegments(call model.Call) ([]segment, error) {
	segments, err := segmentDao.Segments(nil, model.SegmentQuery{
		CallID:  []int64{call.ID},
		Channel: []int8{1, 2},
	})
	if err != nil {
		return nil, fmt.Errorf("get segments failed, %v", err)
	}
	var result = make([]segment, 0, len(segments))

	channelsRole := map[int8]string{
		1: callRoleTypStr(call.LeftChanRole),
		2: callRoleTypStr(call.RightChanRole),
	}
	for index, s := range segments {

		vr := segment{
			SentenceID: int64(index + 1),
			StartTime:  s.StartTime,
			EndTime:    s.EndTime,
			Speaker:    channelsRole[s.Channel],
			ASRText:    s.Text,
			Status:     s.Status,
			SegmentID:  s.ID,
		}
		result = append(result, vr)
	}

	return result, nil
}
