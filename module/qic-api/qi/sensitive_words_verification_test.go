package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
	"testing"
)

var violatedPattern = "segment"

var okSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		ID:   605,
		Text: "ok",
	},
	Speaker: int(model.CallChanStaff),
}
var violatedSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		ID:   603,
		Text: "violated",
	},
	Speaker: int(model.CallChanStaff),
}
var cvSegment = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		ID:   604,
		Text: violatedPattern,
	},
	Speaker: int(model.CallChanStaff),
}
var staffExceptionSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		ID:   601,
		Text: "staff" + violatedPattern,
	},
	Speaker: int(model.CallChanStaff),
}
var CustomerExceptionSegment SegmentWithSpeaker = SegmentWithSpeaker{
	RealSegment: model.RealSegment{
		ID:   602,
		Text: "customer",
	},
	Speaker: int(model.CallChanStaff),
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

func (dao *mockSWVerificationDao) GetSentences(tx model.SqlLike, q *model.SentenceQuery) ([]*model.Sentence, error) {
	return []*model.Sentence{
		&model.Sentence{
			ID: 1,
			TagIDs: []uint64{
				501,
				502,
			},
		},
		&model.Sentence{
			ID: 2,
			TagIDs: []uint64{
				503,
				504,
			},
		},
	}, nil
}
func (dao *mockSWVerificationDao) InsertSentence(tx model.SqlLike, s *model.Sentence) (int64, error) {
	return 1, nil
}
func (dao *mockSWVerificationDao) SoftDeleteSentence(tx model.SqlLike, q *model.SentenceQuery) (int64, error) {
	return 1, nil
}
func (dao *mockSWVerificationDao) CountSentences(tx model.SqlLike, q *model.SentenceQuery) (uint64, error) {
	return 1, nil
}

func (dao *mockSWVerificationDao) InsertSenTagRelation(tx model.SqlLike, s *model.Sentence) error {
	return nil
}
func (dao *mockSWVerificationDao) GetRelSentenceIDByTagIDs(tx model.SqlLike, tagIDs []uint64) (map[uint64][]uint64, error) {
	return map[uint64][]uint64{}, nil
}

func (dao *mockSWVerificationDao) MoveCategories(x model.SqlLike, q *model.SentenceQuery, category uint64) (int64, error) {
	return 1, nil
}
func (dao *mockSWVerificationDao) InsertSentences(x model.SqlLike, ss []model.Sentence) error {
	return nil
}

func mockSentenceMatch(segs []string, ids []uint64, enterprise string) (map[uint64][]int, error) {
	return map[uint64][]int{
		1: []int{
			1,
		},
		2: []int{
			0,
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
				ID:     401,
				Type:   model.UserValueTypSensitiveWord,
				Value:  "201",
				LinkID: 303,
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

	if len(passedMap[301]) != 0 || len(passedMap[302]) != 0 || len(passedMap[303]) == 0 {
		t.Errorf("get map failed, map: %+v", passedMap)
		return
	}
}

func TestSensitiveWordsVerification(t *testing.T) {
	sws := []model.SensitiveWord{
		model.SensitiveWord{
			ID:   301,
			Name: violatedPattern,
			StaffException: []model.SimpleSentence{
				model.SimpleSentence{
					ID: 1,
				},
			},
		},
		model.SensitiveWord{
			ID:   302,
			Name: "customer",
			CustomerException: []model.SimpleSentence{
				model.SimpleSentence{
					ID: 2,
				},
			},
		},
		model.SensitiveWord{
			ID:   303,
			Name: "violated",
		},
		model.SensitiveWord{
			ID:   304,
			Name: "will not violate",
		},
	}
	setupSensitiveWordVerificationTest(sws)

	customerSegment := SegmentWithSpeaker{
		RealSegment: model.RealSegment{
			ID:   606,
			Text: "ok",
		},
		Speaker: int(model.CallChanCustomer),
	}

	segments := []*SegmentWithSpeaker{
		&customerSegment,
		&staffExceptionSegment,
		&CustomerExceptionSegment,
		&violatedSegment,
		&cvSegment,
		&okSegment,
	}

	callID := int64(5)
	credits, err := SensitiveWordsVerification(callID, segments, "enterprise")
	if err != nil {
		t.Errorf("something happened in verification, err: %s", err.Error())
		return
	}

	if len(credits) != 13 {
		t.Errorf("verification failed, credits: %+v", credits)
		return
	}
}
