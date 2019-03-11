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

type mockSWVerificationDao struct{}

func (dao *mockSWVerificationDao) Create(sw *model.SensitiveWord, sqlLike model.SqlLike) (int64, error) {
	return 1, nil
}
func (dao *mockSWVerificationDao) CountBy(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) (int64, error) {
	return 1, nil
}

func (dao *mockSWVerificationDao) GetBy(filter *model.SensitiveWordFilter, sqlLike model.SqlLike) ([]model.SensitiveWord, error) {
	return []model.SensitiveWord{
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
	}, nil
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

func setupSensitiveWordVerificationTest() (model.DBLike, model.SensitiveWordDao) {
	mockDBLike := &test.MockDBLike{}
	mockDao := &mockSWVerificationDao{}

	originDBLike := dbLike
	originSWDao := swDao

	dbLike = mockDBLike
	swDao = mockDao
	sentenceMatchFunc = mockSentenceMatch

	return originDBLike, originSWDao
}

func TestSensitiveWordsVerification(t *testing.T) {
	setupSensitiveWordVerificationTest()

	segments := []SegmentWithSpeaker{
		violatedSegment,
		staffExceptionSegment,
		CustomerExceptionSegment,
		okSegment,
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
