package qi

import (
	"testing"

	"emotibot.com/emotigo/module/qic-api/util/logicaccess"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

func TestSilenceRuleCheck(t *testing.T) {

	rules := []SilenceRuleWithException{
		SilenceRuleWithException{SilenceRule: model.SilenceRule{Seconds: 10, Times: 1},
			RuleSilenceException: RuleSilenceException{
				Before: RuleException{Staff: []model.SimpleSentence{model.SimpleSentence{ID: 1}, model.SimpleSentence{ID: 2}}, Customer: []model.SimpleSentence{model.SimpleSentence{ID: 1}}},
				After:  OnlyStaffException{Staff: []model.SimpleSentence{model.SimpleSentence{ID: 4}}},
			},
			sentences: map[uint64][]uint64{
				1: []uint64{11},
				2: []uint64{12},
				3: []uint64{11, 13},
				4: []uint64{12, 13},
				5: []uint64{11, 12, 13},
			},
		},

		SilenceRuleWithException{SilenceRule: model.SilenceRule{Seconds: 10, Times: 1},
			RuleSilenceException: RuleSilenceException{
				Before: RuleException{Staff: []model.SimpleSentence{model.SimpleSentence{ID: 5}, model.SimpleSentence{ID: 2}}, Customer: []model.SimpleSentence{model.SimpleSentence{ID: 1}}},
				After:  OnlyStaffException{Staff: []model.SimpleSentence{model.SimpleSentence{ID: 4}}},
			},
			sentences: map[uint64][]uint64{
				1: []uint64{11},
				2: []uint64{12},
				3: []uint64{11, 13},
				4: []uint64{12, 13},
				5: []uint64{11, 12, 13},
			},
		},

		SilenceRuleWithException{SilenceRule: model.SilenceRule{Seconds: 10, Times: 1},
			RuleSilenceException: RuleSilenceException{
				Before: RuleException{Staff: []model.SimpleSentence{model.SimpleSentence{ID: 5}, model.SimpleSentence{ID: 2}}, Customer: []model.SimpleSentence{model.SimpleSentence{ID: 1}}},
				After:  OnlyStaffException{Staff: []model.SimpleSentence{model.SimpleSentence{ID: 5}}},
			},
			sentences: map[uint64][]uint64{
				1: []uint64{11},
				2: []uint64{12},
				3: []uint64{11, 13},
				4: []uint64{12, 13},
				5: []uint64{16},
			},
		},

		SilenceRuleWithException{SilenceRule: model.SilenceRule{Seconds: 1, Times: 1},
			RuleSilenceException: RuleSilenceException{
				Before: RuleException{Staff: []model.SimpleSentence{}, Customer: []model.SimpleSentence{}},
				After:  OnlyStaffException{Staff: []model.SimpleSentence{model.SimpleSentence{ID: 5}}},
			},
			sentences: map[uint64][]uint64{
				5: []uint64{},
			},
		},
	}

	expected := []bool{true, false, true, false}

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
	silenceSegs, segs := extractSegmentSpeaker(allSegs, SilenceSpeaker)

	credits, err := silenceRuleCheck(rules, tagMatchDat, allSegs, segs, silenceSegs)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}

	expectedNum := len(rules)
	if expectedNum != len(credits) {
		t.Fatalf("expecting %d credit, but get %d\n", expectedNum, len(credits))
	}

	for k, v := range expected {
		if v != credits[k].Valid {
			t.Errorf("expecting get %t at %d'th credit, but get %t\n", v, k+1, credits[k].Valid)
		}
	}

}

func TestInjectSilenceInterposalSegs(t *testing.T) {
	segs := []model.RealSegment{
		model.RealSegment{StartTime: 1.11, EndTime: 3.05},
		model.RealSegment{StartTime: 3.6, EndTime: 9},
		model.RealSegment{StartTime: 7, EndTime: 13},  //interposal
		model.RealSegment{StartTime: 15, EndTime: 17}, //silence
		model.RealSegment{StartTime: 17.2, EndTime: 19.2},
		model.RealSegment{StartTime: 24.3, EndTime: 26}, //silence
		model.RealSegment{StartTime: 29, EndTime: 31},   //silence
	}

	expect := []model.RealSegment{
		model.RealSegment{StartTime: 1.11, EndTime: 3.05},
		model.RealSegment{StartTime: 3.6, EndTime: 9},
		model.RealSegment{StartTime: 7, EndTime: 9, Channel: InterposalSpeaker}, //interposal
		model.RealSegment{StartTime: 7, EndTime: 13},
		model.RealSegment{StartTime: 13, EndTime: 15, Channel: SilenceSpeaker},
		model.RealSegment{StartTime: 15, EndTime: 17},
		model.RealSegment{StartTime: 17.2, EndTime: 19.2},
		model.RealSegment{StartTime: 19.2, EndTime: 24.3, Channel: SilenceSpeaker},
		model.RealSegment{StartTime: 24.3, EndTime: 26}, //silence
		model.RealSegment{StartTime: 26, EndTime: 29, Channel: SilenceSpeaker},
		model.RealSegment{StartTime: 29, EndTime: 31}, //silence
	}

	result := injectSilenceInterposalSegs(segs)

	if len(expect) != len(result) {
		t.Fatalf("expected get %d segments, but get %d\n", len(expect), len(result))
	}

	for idx, r := range result {
		e := expect[idx]
		if e.StartTime != r.StartTime || e.EndTime != r.EndTime || e.Channel != r.Channel {
			t.Errorf("expecting get start:%v, end:%v, channel:%d at %d'th credit, but get start:%v, end:%v, channel:%d",
				e.StartTime, e.EndTime, e.Channel, idx, r.StartTime, r.EndTime, r.Channel)
		}
	}
}
