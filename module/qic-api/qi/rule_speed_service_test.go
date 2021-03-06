package qi

import (
	"testing"

	"emotibot.com/emotigo/module/qic-api/util/logicaccess"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

func TestSpeedRuleCheck(t *testing.T) {

	rules := []SpeedRuleWithException{
		SpeedRuleWithException{SpeedRule: model.SpeedRule{Min: 20, Max: 50},
			RuleSpeedException: RuleSpeedException{
				Under: RuleSpeedExceptionContent{Customer: []model.SimpleSentence{model.SimpleSentence{ID: 1}}},
				Over:  RuleSpeedExceptionContent{Customer: []model.SimpleSentence{model.SimpleSentence{ID: 2}}},
			},
			sentences: map[uint64][]uint64{
				1: []uint64{11},
				2: []uint64{12},
				3: []uint64{11, 13},
				4: []uint64{12, 13},
				5: []uint64{11, 12, 13},
			},
		},

		SpeedRuleWithException{SpeedRule: model.SpeedRule{Min: 10, Max: 30},
			RuleSpeedException: RuleSpeedException{
				Under: RuleSpeedExceptionContent{Customer: []model.SimpleSentence{model.SimpleSentence{ID: 3}}},
				Over:  RuleSpeedExceptionContent{Customer: []model.SimpleSentence{model.SimpleSentence{ID: 5}}},
			},
			sentences: map[uint64][]uint64{
				1: []uint64{11},
				2: []uint64{12},
				3: []uint64{11, 13},
				4: []uint64{12, 13},
				5: []uint64{11, 12, 13},
			},
		},

		SpeedRuleWithException{SpeedRule: model.SpeedRule{Min: 10, Max: 20},
			RuleSpeedException: RuleSpeedException{
				Under: RuleSpeedExceptionContent{Customer: []model.SimpleSentence{model.SimpleSentence{ID: 5}}},
				Over:  RuleSpeedExceptionContent{Customer: []model.SimpleSentence{model.SimpleSentence{ID: 4}}},
			},
			sentences: map[uint64][]uint64{
				1: []uint64{11},
				2: []uint64{12},
				3: []uint64{11, 13},
				4: []uint64{12, 13},
				5: []uint64{16},
			},
		},
	}

	expected := []bool{true, false, true}
	staffSpeaker := int(model.CallChanStaff)
	customerSpeaker := int(model.CallChanCustomer)

	tagMatchDat := []*MatchedData{
		&MatchedData{Matched: map[uint64]*logicaccess.AttrResult{11: &logicaccess.AttrResult{}}},
		&MatchedData{Matched: map[uint64]*logicaccess.AttrResult{12: &logicaccess.AttrResult{}}},
		&MatchedData{Matched: map[uint64]*logicaccess.AttrResult{13: &logicaccess.AttrResult{}}},
		&MatchedData{Matched: map[uint64]*logicaccess.AttrResult{14: &logicaccess.AttrResult{}}},
		&MatchedData{Matched: map[uint64]*logicaccess.AttrResult{15: &logicaccess.AttrResult{}}},
		&MatchedData{Matched: map[uint64]*logicaccess.AttrResult{16: &logicaccess.AttrResult{}}},
	}
	allSegs := []*SegmentWithSpeaker{
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 1, CallID: 1234, StartTime: 1.11, EndTime: 2.99}, Speaker: SilenceSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 2, CallID: 1234, StartTime: 2.99, EndTime: 18.7}, Speaker: staffSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 3, CallID: 1234, StartTime: 18.11, EndTime: 21.99}, Speaker: customerSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 4, CallID: 1234, StartTime: 21.99, EndTime: 35.7}, Speaker: SilenceSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 5, CallID: 1234, StartTime: 35.8, EndTime: 49.1}, Speaker: staffSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 6, CallID: 1234, StartTime: 49.2, EndTime: 50.11}, Speaker: SilenceSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 7, CallID: 1234, StartTime: 50.12, EndTime: 59.3}, Speaker: staffSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 8, CallID: 1234, StartTime: 60, EndTime: 61.2}, Speaker: customerSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 9, CallID: 1234, StartTime: 61.3, EndTime: 78.4}, Speaker: SilenceSpeaker},
		&SegmentWithSpeaker{RealSegment: model.RealSegment{ID: 10, CallID: 1234, StartTime: 78.5, EndTime: 89}, Speaker: staffSpeaker},
	}

	credits, err := checkSpeedRules(rules, tagMatchDat, allSegs, 40)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}

	expectedNum := len(rules)
	if expectedNum != len(credits) {
		t.Fatalf("expecting %d credit, but get %d\n", expectedNum, len(credits))
	}

	for k, v := range expected {
		if v != credits[k].Valid {
			t.Errorf("expecting get %t at %d credit, but get %t\n", v, k, credits[k].Valid)
		}
	}

}
