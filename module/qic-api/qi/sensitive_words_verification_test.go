package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
	"testing"
)

var violatedPattern = "segment"

var okSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		Text: "ok",
	},
	Speaker: int(model.CallChanCustomer),
}
var violatedSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		Text: "violated" + violatedPattern,
	},
	Speaker: int(model.CallChanStaff),
}
var staffExceptionSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		Text: "staff" + violatedPattern,
	},
	Speaker: int(model.CallChanStaff),
}
var CustomerExceptionSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		Text: "customer" + violatedPattern,
	},
	Speaker: int(model.CallChanCustomer),
}

type mockSWVerificationDao struct {
	sws []model.SensitiveWord
}

func (dao *mockSWVerificationDao) Create(sw *model.SensitiveWord, sqlLike model.SqlLike) (int64, error) {
	return 1, nil
}
func (dao *mockSWVerificationDao) CountBy(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) (int64, error) {
	return 1, nil
}

func (dao *mockSWVerificationDao) GetBy(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) ([]model.SensitiveWord, error) {
	return dao.sws, nil
}

func (dao *mockSWVerificationDao) GetRel(int64, model.SqlLike) (map[int8][]uint64, error) {
	return map[int8][]uint64{}, nil
}

func (dao *mockSWVerificationDao) GetRels([]int64, model.SqlLike) (map[int64][]uint64, map[int64][]uint64, error) {
	staffExecptionMaps := map[int64][]uint64{
		55688: []uint64{1},
	}

	customerExceptions := map[int64][]uint64{
		55688: []uint64{2},
	}

	return staffExecptionMaps, customerExceptions, nil
}

func (dao *mockSWVerificationDao) Delete(*model.SensitiveWordFilter, model.SqlLike) (int64, error) {
	return 1, nil
}
func (dao *mockSWVerificationDao) Move(*model.SensitiveWordFilter, int64, model.SqlLike) (int64, error) {
	return 1, nil
}

func (dao *mockSWVerificationDao) Names(model.SqlLike, bool) ([]string, error) {
	return []string{}, nil
}

func mockSentenceMatch(segs []string, ids []uint64, enterprise string) (map[uint64][]int, error) {
	return map[uint64][]int{
		1: []int{
			0,
		},
		2: []int{
			1,
		},
	}, nil
}

func mockUserValues(delegatee model.SqlLike, query model.UserValueQuery) ([]model.UserValue, error) {
	if query.Type[0] == model.UserValueTypCall {
		return []model.UserValue{
			model.UserValue{
				ID:    201,
				Type:  model.UserValueTypCall,
				Value: "201",
			},
			model.UserValue{
				ID:    202,
				Type:  model.UserValueTypCall,
				Value: "202",
			},
			model.UserValue{
				ID:    203,
				Type:  model.UserValueTypCall,
				Value: "203",
			},
			model.UserValue{
				ID:    204,
				Type:  model.UserValueTypCall,
				Value: "204",
			},
		}, nil

	} else if query.Type[0] == model.UserValueTypSensitiveWord {
		return []model.UserValue{
			model.UserValue{
				ID:     301,
				Type:   model.UserValueTypSensitiveWord,
				Value:  "201",
				LinkID: 301,
			},
		}, nil
	} else {
		return []model.UserValue{}, nil
	}
}

func setupSensitiveWordVerificationTest(sws []model.SensitiveWord) (model.DBLike, model.SensitiveWordDao) {
	mockDBLike := &test.MockDBLike{}
	mockDao := &mockSWVerificationDao{sws}

	originDBLike := dbLike
	originSWDao := swDao

	dbLike = mockDBLike
	swDao = mockDao
	sentenceMatchFunc = mockSentenceMatch
	userValues = mockUserValues

	return originDBLike, originSWDao
}

func TestSensitiveWordsVerification(t *testing.T) {
	sws := []model.SensitiveWord{
		model.SensitiveWord{
			ID:   55688,
			Name: violatedPattern,
			StaffException: []model.SimpleSentence{
				model.SimpleSentence{
					ID: 1,
				},
			},
			CustomerException: []model.SimpleSentence{
				model.SimpleSentence{
					ID: 2,
				},
			},
		},
		model.SensitiveWord{
			ID:   55699,
			Name: violatedPattern[0 : len(violatedPattern)-2],
		},
	}
	setupSensitiveWordVerificationTest(sws)

	segments := []*SegmentWithSpeaker{
		&violatedSegment,
		&staffExceptionSegment,
		&CustomerExceptionSegment,
		&okSegment,
	}

	callID := int64(5)
	credits, err := SensitiveWordsVerification(callID, segments, "enterprise")
	if err != nil {
		t.Errorf("something happened in verification, err: %s", err.Error())
		return
	}

	if len(credits) != 4 {
		t.Errorf("verification failed, credits: %+v", credits)
		return
	}
}

func TestSensitiveWordsCustomValues(t *testing.T) {
	sws := []model.SensitiveWord{
		model.SensitiveWord{
			ID:   301,
			Name: "301",
		},
		model.SensitiveWord{
			ID:   302,
			Name: "302",
		},
	}
	setupSensitiveWordVerificationTest(sws)

	segments := []*SegmentWithSpeaker{
		&SegmentWithSpeaker{
			RealSegment: model.RealSegment{
				Text: "301",
			},
			Speaker: int(model.CallChanStaff),
		},
		&SegmentWithSpeaker{
			RealSegment: model.RealSegment{
				Text: "302",
			},
			Speaker: int(model.CallChanStaff),
		},
	}

	callID := int64(5)
	credits, err := SensitiveWordsVerification(callID, segments, "enterprise")
	if err != nil {
		t.Errorf("something happened in verification, err: %s", err.Error())
		return
	}

	if len(credits) != 1 {
		t.Errorf("verification failed, credits: %+v", credits)
		return
	}
}

func TestCallToSWUserKeyValues(t *testing.T) {
	setupSensitiveWordVerificationTest([]model.SensitiveWord{})

	sws := []int64{
		301,
		302,
		303,
	}

	passedMap, err := callToSWUserKeyValues(1, sws, nil)
	if err != nil {
		t.Errorf("errro while get passed map, err: %s", err.Error())
		return
	}

	if !passedMap[301] || passedMap[302] || passedMap[303] {
		t.Errorf("get map failed, map: %+v", passedMap)
		return
	}
}
