package handlers

import "testing"

func Test_getEmotionBaseSQL(t *testing.T) {
	selectColumns := [2][]string{{NID, NTAG, NTAG2}, {NCHANNEL, NEMOTYPE, NSCORE}}
	whereStates := [2][]WhereStates{
		{
			{name: NANARES, compare: "="},
			{name: NUPT, compare: ">="},
			{name: NUPT, compare: "<="},
		},
		{
			{name: NEMOTYPE, compare: "="},
		},
	}

	orderStates := [2][]string{{NUPT, NID}, {NCHANNEL}}

	correctAnswer := "select a.id,a.tag1,a.tag2,b.channel,b.emotion_type,b.score from fileInformation as a inner join channelScore as b on a.id=b.id where a.analysis_result=? and a.upload_time>=? and a.upload_time<=? and b.emotion_type=? order by a.upload_time,a.id,b.channel"

	sql := getEmotionBaseSQL(selectColumns, whereStates, orderStates)

	if sql != correctAnswer {
		t.Error(sql)
	}

}
