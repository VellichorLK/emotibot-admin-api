package qi

import (
	"database/sql"
	"testing"
	"time"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
)

var mockCtxSentences = []string{
	"所以在女童一下梅子我们的对话都会全程录影可以吗",
	"好的",
	"我姓中国信讬忍受的张泰阶",
	"大陆酱汁好事涟漪领事一二三四五料",
	"为妇女的权益在你",
	"同一下漠视我们的对话都会全程录音可以吗要挑做什么",
	"奥勒的权益尽管同样存在进一步说明在案的内容",
	"那到底可以要做什么啊",
	"奥运我们是我们是与花旗银行合作再附加新保险专案",
	"妇女的权益在女童一下漠视我们的对话会全程录音可以吗",
	"好因为你是花旗银行的贵宾卡右脑资料部分有银行提高"}

var mockMatchIdx = []int{1, 7}
var mockMatchTagID = []uint64{1, 6}
var mockAllTagID = []uint64{1, 3, 5, 6, 9}

type mockPredictClient struct {
}

func (m *mockPredictClient) Train(d *logicaccess.TrainUnit) error {
	return nil
}
func (m *mockPredictClient) Status(d *logicaccess.TrainAPPID) (string, error) {
	return "", nil
}
func (m *mockPredictClient) PredictAndUnMarshal(d *logicaccess.PredictRequest) (*logicaccess.PredictResult, error) {
	return nil, nil
}
func (m *mockPredictClient) BatchPredictAndUnMarshal(d *logicaccess.BatchPredictRequest) (*logicaccess.PredictResult, error) {

	var resp logicaccess.PredictResult

	var attr logicaccess.AttrResult

	if d.ID == mockMatchTagID[0] {
		attr.Tag = mockMatchTagID[0]
		attr.SentenceID = int(mockMatchIdx[0])
		attr.Sentence = mockCtxSentences[attr.SentenceID-1]
		attr.Score = 87
		resp.Dialogue = append(resp.Dialogue, attr)
	} else if d.ID == mockMatchTagID[1] {
		attr.Tag = mockMatchTagID[1]
		attr.SentenceID = int(mockMatchIdx[0])
		attr.Sentence = mockCtxSentences[attr.SentenceID-1]
		attr.Score = 87
		resp.Dialogue = append(resp.Dialogue, attr)

		attr.Tag = mockMatchTagID[1]
		attr.SentenceID = int(mockMatchIdx[1])
		attr.Sentence = mockCtxSentences[attr.SentenceID-1]
		attr.Score = 94
		resp.Dialogue = append(resp.Dialogue, attr)
	}

	return &resp, nil
}
func (m *mockPredictClient) SessionCreate(d *logicaccess.SessionRequest) error {
	return nil
}
func (m *mockPredictClient) SessionDelete(d *logicaccess.SessionRequest) error {
	return nil
}
func (m *mockPredictClient) UnloadModel(d *logicaccess.TrainAPPID) error {
	return nil
}

func TestTagMatch(t *testing.T) {

	predictor = &mockPredictClient{}

	tags := mockAllTagID
	matched, err := TagMatch(tags, mockCtxSentences, 3*time.Second)
	if err != nil {
		t.Errorf("Expecting no error, but get %s\n", err)
	} else {

		//	fmt.Printf("%s\n", matched)

		if len(matched) != len(mockCtxSentences) {
			t.Errorf("Expecting %d return, but get %d\n", len(mockCtxSentences), len(matched))
		} else {
			if matched[0].Index != mockMatchIdx[0] {
				t.Errorf("Expecting index to be 1, but get %d\n", matched[0].Index)
			}
			if matched[0].Matched[mockMatchTagID[0]] == nil {
				t.Fatalf("Expecting has match data at tag id %d, but get nothing\n", mockMatchTagID[0])
			}

			for matchTag, v := range matched[0].Matched {
				switch matchTag {
				case mockMatchTagID[0]:
					if v.Sentence != mockCtxSentences[mockMatchIdx[0]-1] {
						t.Fatalf("Expecting match sentecne %s, but get %s\n",
							mockCtxSentences[mockMatchIdx[0]-1], v.Sentence)
					}
					if v.Tag != mockMatchTagID[0] {
						t.Fatalf("Expecting match tag %d, but get %d\n",
							mockMatchTagID[0], v.Tag)
					}
				case mockMatchTagID[1]:
					if v.Sentence != mockCtxSentences[mockMatchIdx[0]-1] {
						t.Fatalf("Expecting match sentecne %s, but get %s\n",
							mockCtxSentences[mockMatchIdx[0]-1], v.Sentence)
					}
					if v.Tag != mockMatchTagID[1] {
						t.Fatalf("Expecting match tag %d, but get %d\n",
							mockMatchTagID[1], v.Tag)
					}
				default:
					t.Fatalf("Expecting no other matched tag id, but get %d\n", matchTag)
				}
			}

		}
	}
}

func TestSentenceMatch(t *testing.T) {
	predictor = &mockPredictClient{}
	tags := mockAllTagID
	matched, err := TagMatch(tags, mockCtxSentences, 3*time.Second)
	if err != nil {
		t.Fatalf("Expecting no error, but get %s\n", err)
	}

	segMatchedTag := extractTagMatchedData(matched)

	senCriteria := make(map[uint64][]uint64)
	senCriteria[1] = append(senCriteria[1], 1)
	senCriteria[1] = append(senCriteria[1], 6)

	senCriteria[9] = append(senCriteria[9], 2)
	senCriteria[9] = append(senCriteria[9], 7)

	senMatch, err := SentencesMatch(segMatchedTag, senCriteria)
	if err != nil {
		t.Fatalf("Expecting no error, but get %s\n", err)
	}

	if len(senMatch) != 1 {
		t.Errorf("Expecting has %d matched sentence, but get %d\n", 1, len(senMatch))
	}
	if matchedSeg, ok := senMatch[1]; ok {
		if len(matchedSeg) != 1 {
			t.Errorf("Expecting has %d matched segment, but get %d\n", 1, len(matchedSeg))
		} else {
			if matchedSeg[0] != 1 {
				t.Errorf("Expecting %d segment meet criteria 1, but get segment %d meet\n", 1, matchedSeg[0])
			}
		}
	} else {
		t.Errorf("Expecting %d sentence crieria meet, but get none\n", 1)
	}

	senCriteria[10] = append(senCriteria[10], 6)
	senCriteria[10] = append(senCriteria[10], 6)

	senMatch, err = SentencesMatch(segMatchedTag, senCriteria)
	if err != nil {
		t.Fatalf("Expecting no error, but get %s\n", err)
	}

	if len(senMatch) != 2 {
		t.Errorf("Expecting has %d matched sentence, but get %d\n", 2, len(senMatch))
	}
	if matchedSeg, ok := senMatch[1]; ok {
		if len(matchedSeg) != 1 {
			t.Errorf("Expecting has %d matched segment, but get %d\n", 1, len(matchedSeg))
		} else {
			if matchedSeg[0] != 1 {
				t.Errorf("Expecting %d segment meet criteria 1, but get segment %d meet\n", 1, matchedSeg[0])
			}
		}
	} else {
		t.Errorf("Expecting %d sentence crieria meet, but get none\n", 1)
	}

	if matchedSeg, ok := senMatch[10]; ok {
		if len(matchedSeg) != 2 {
			t.Errorf("Expecting has %d matched segment, but get %d\n", 2, len(matchedSeg))
		} else {
			for _, matchIdx := range matchedSeg {
				switch matchIdx {
				case 1:
				case 7:
				default:
					t.Errorf("Expecting %d,%d segment meet criteria 10, but get segment %d meet\n", 1, 7, matchedSeg[0])
				}
			}

		}
	} else {
		t.Errorf("Expecting %d,%d sentence crieria meet, but get none\n", 1, 7)
	}
}

func TestSentenceGroupMatch(t *testing.T) {

	//	matchedSenID := []uint64{1, 5, 11}
	//matchedSen := make(map[uint64][]int)

	numOfSegs := 20
	segments := make([]*SegmentWithSpeaker, numOfSegs, numOfSegs)
	for i := 0; i < numOfSegs; i++ {
		segments[i] = new(SegmentWithSpeaker)
	}

	matchedSenIDAndSegmentIdx := map[uint64][]int{1: {3, 5, 6}, 5: {3, 7}, 9: {11, 8}}
	criteria := map[uint64]*SenGroupCriteria{11: &SenGroupCriteria{ID: 11, SentenceID: []uint64{5}},
		21: &SenGroupCriteria{ID: 21, SentenceID: []uint64{9, 5}, Role: 1},
		31: &SenGroupCriteria{ID: 31, SentenceID: []uint64{138, 139}},
		41: &SenGroupCriteria{ID: 41, SentenceID: []uint64{1, 9}, Role: 0}}

	result, err := SentenceGroupMatch(matchedSenIDAndSegmentIdx, criteria, segments)
	if err != nil {
		t.Fatalf("Expecting no error, but get %s\n", err)
	}

	//fmt.Printf("%v\n", result)
	if len(result) != 2 {
		t.Fatalf("Expecting %d sentence group is meet, but get %d\n", 2, len(result))
	}
	for sgID := range result {
		switch sgID {
		case 11:
		case 41:
		default:
			t.Fatalf("Expecting %d,%d sentence group is meet, but get %d\n", 11, 41, sgID)
		}
	}

}

func TestFCDigest(t *testing.T) {
	c := &ConFlowCriteria{}
	expression := "if A and B"
	c.Expression = expression
	err := c.FlowExpressionToNode()
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	}

	expecting := []ExprNode{{uuid: "A"}, {uuid: "B"}}
	if len(c.nodes) != len(expecting) {
		t.Errorf("Expecting %d node, but get %d\n", len(expecting), len(c.nodes))
		for i := 0; i < len(c.nodes); i++ {
			t.Errorf("node %d uuid %s withNot %v isThen %v\n", i, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
		}
	} else {

		for i := 0; i < len(expecting); i++ {
			if expecting[i].uuid != c.nodes[i].uuid ||
				expecting[i].withNot != c.nodes[i].withNot ||
				expecting[i].isThen != c.nodes[i].isThen {
				t.Errorf("Expecting %d node uuid %s withNot %v isThen %v, but get uuid %s withNot %v isThen %v\n",
					i, expecting[i].uuid, expecting[i].withNot, expecting[i].isThen, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
			}
		}
	}

	expression = "if A and B and C then D and E"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err != nil {
		t.Errorf("Expecting has error,but get none\n")
	}
	expecting = []ExprNode{{uuid: "A"}, {uuid: "B"}, {uuid: "C"}, {uuid: "D", isThen: true}, {uuid: "E"}}
	if len(c.nodes) != len(expecting) {
		t.Errorf("Expecting %d node, but get %d\n", len(expecting), len(c.nodes))
		for i := 0; i < len(c.nodes); i++ {
			t.Errorf("node %d uuid %s withNot %v isThen %v\n", i, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
		}
	} else {

		for i := 0; i < len(expecting); i++ {
			if expecting[i].uuid != c.nodes[i].uuid ||
				expecting[i].withNot != c.nodes[i].withNot ||
				expecting[i].isThen != c.nodes[i].isThen {
				t.Errorf("Expecting %d node uuid %s withNot %v isThen %v, but get uuid %s withNot %v isThen %v\n",
					i, expecting[i].uuid, expecting[i].withNot, expecting[i].isThen, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
			}
		}
	}

	expression = "if A and not B and not C then D and not E"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err != nil {
		t.Errorf("Expecting has error,but get none\n")
	}
	expecting = []ExprNode{{uuid: "A"}, {uuid: "B", withNot: true}, {uuid: "C", withNot: true}, {uuid: "D", isThen: true}, {uuid: "E", withNot: true}}
	if len(c.nodes) != len(expecting) {
		t.Errorf("Expecting %d node, but get %d\n", len(expecting), len(c.nodes))
		for i := 0; i < len(c.nodes); i++ {
			t.Errorf("node %d uuid %s withNot %v isThen %v\n", i, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
		}
	} else {

		for i := 0; i < len(expecting); i++ {
			if expecting[i].uuid != c.nodes[i].uuid ||
				expecting[i].withNot != c.nodes[i].withNot ||
				expecting[i].isThen != c.nodes[i].isThen {
				t.Errorf("Expecting %d node uuid %s withNot %v isThen %v, but get uuid %s withNot %v isThen %v\n",
					i, expecting[i].uuid, expecting[i].withNot, expecting[i].isThen, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
			}
		}
	}

	expression = "if A and not B and not C then D and not not E"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err != nil {
		t.Errorf("Expecting has error,but get none\n")
	}
	expecting = []ExprNode{{uuid: "A"}, {uuid: "B", withNot: true}, {uuid: "C", withNot: true}, {uuid: "D", isThen: true}, {uuid: "E"}}
	if len(c.nodes) != len(expecting) {
		t.Errorf("Expecting %d node, but get %d\n", len(expecting), len(c.nodes))
		for i := 0; i < len(c.nodes); i++ {
			t.Errorf("node %d uuid %s withNot %v isThen %v\n", i, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
		}
	} else {

		for i := 0; i < len(expecting); i++ {
			if expecting[i].uuid != c.nodes[i].uuid ||
				expecting[i].withNot != c.nodes[i].withNot ||
				expecting[i].isThen != c.nodes[i].isThen {
				t.Errorf("Expecting %d node uuid %s withNot %v isThen %v, but get uuid %s withNot %v isThen %v\n",
					i, expecting[i].uuid, expecting[i].withNot, expecting[i].isThen, c.nodes[i].uuid, c.nodes[i].withNot, c.nodes[i].isThen)
			}
		}
	}

	expression = "if A and B and"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err == nil {
		t.Errorf("Expecting has error,but get none\n")
	}

	expression = "if A and B and then"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err == nil {
		t.Errorf("Expecting has error,but get none\n")
	}

	expression = "if A and B if"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err == nil {
		t.Errorf("Expecting has error,but get none\n")
	}

	expression = "if A and B and not not not c d"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err == nil {
		t.Errorf("Expecting has error,but get none\n")
	}

	expression = "if A and B and not and C"
	c.Expression = expression
	c.nodes = []*ExprNode{}
	err = c.FlowExpressionToNode()
	if err == nil {
		t.Errorf("Expecting has error,but get none\n")
	}
}

func TestConversationFlowMatch(t *testing.T) {

	senGrpUUIDMapID := map[string]uint64{"A": 1, "B": 2, "C": 3, "D": 4, "E": 5, "F": 6}
	numOfSeg := 20

	//case 1, must starts from A in the 3 lines and B
	expression := "must A and B"
	senGrpCriteria := map[uint64]*SenGroupCriteria{
		1: &SenGroupCriteria{ID: 1, Position: 0, Range: 3},
		2: &SenGroupCriteria{ID: 2},
	}

	cfCriteria := &ConFlowCriteria{Expression: expression, Repeat: 1}
	matchSgID := map[uint64][]int{1: []int{3}, 2: []int{5, 7, 9}, 3: []int{4, 8}, 4: []int{11, 14}}

	matched, err := ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}

	matchSgID = map[uint64][]int{1: []int{3}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

	//case 2, if starts from A then B in the next 5 lines
	expression = "if A then B"
	senGrpCriteria = map[uint64]*SenGroupCriteria{
		1: &SenGroupCriteria{ID: 1, Position: 0, Range: 3},
		2: &SenGroupCriteria{ID: 2, Range: 5},
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{5, 7, 9}, 3: []int{4, 8}, 4: []int{11, 14}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}
	//matched B but out of assigned range
	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{11, 13, 19}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}
	//not matched A, but with if
	matchSgID = map[uint64][]int{2: []int{5, 7, 9}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}

	//case 3, must ends from A in the 10 lines and B
	expression = "must A and B"
	senGrpCriteria = map[uint64]*SenGroupCriteria{
		1: &SenGroupCriteria{ID: 1, Position: 1, Range: 10},
		2: &SenGroupCriteria{ID: 2},
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{12}, 2: []int{14, 15, 19}, 3: []int{4, 8}, 4: []int{11, 14}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}

	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{4, 19}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

	//case 4 must A and B then C and D then E

	expression = "Must A and B then C and D then E"
	senGrpCriteria = map[uint64]*SenGroupCriteria{
		1: &SenGroupCriteria{ID: 1, Position: 0, Range: 3},
		2: &SenGroupCriteria{ID: 2},
		3: &SenGroupCriteria{ID: 3, Range: 3},
		4: &SenGroupCriteria{ID: 4},
		5: &SenGroupCriteria{ID: 5},
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{4, 7}, 3: []int{7, 11}, 4: []int{12, 14}, 5: []int{15}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{4, 7}, 3: []int{7, 11}, 4: []int{12, 14}, 5: []int{11}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{4, 19}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

	//case 5

	expression = "Must A and not B then C and not D"
	senGrpCriteria = map[uint64]*SenGroupCriteria{
		1: &SenGroupCriteria{ID: 1, Position: 0, Range: 3},
		2: &SenGroupCriteria{ID: 2},
		3: &SenGroupCriteria{ID: 3, Range: 3},
		4: &SenGroupCriteria{ID: 4},
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{3}, 3: []int{6, 11}, 5: []int{15}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{4, 7}, 3: []int{7, 11}, 4: []int{12, 14}, 5: []int{11}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{4, 19}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

	//repeat use case
	cfCriteria.Repeat = 2

	expression = "Must A then B"
	senGrpCriteria = map[uint64]*SenGroupCriteria{
		1: &SenGroupCriteria{ID: 1, Position: 0, Range: 5},
		2: &SenGroupCriteria{ID: 2, Range: 5},
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{3, 4}, 2: []int{5, 7, 9}, 3: []int{4, 8}, 4: []int{11, 14}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}
	//not matched repeat 2
	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{11, 13, 19}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

	//case 2 repeat
	expression = "Must A"
	cfCriteria.Repeat = 3
	senGrpCriteria = map[uint64]*SenGroupCriteria{
		1: &SenGroupCriteria{ID: 1, Position: -1},
	}

	cfCriteria.Expression = expression
	matchSgID = map[uint64][]int{1: []int{3, 4, 7}, 2: []int{5, 7, 9}, 3: []int{4, 8}, 4: []int{11, 14}}

	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if !matched {
			t.Errorf("Expecting matched the flow conversation,but no\n")
		}
	}
	//not matched repeat 2
	matchSgID = map[uint64][]int{1: []int{3}, 2: []int{11, 13, 19}, 3: []int{4, 8}, 4: []int{11, 14}}
	matched, err = ConversationFlowMatch(matchSgID, senGrpCriteria, cfCriteria, senGrpUUIDMapID, numOfSeg)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		if matched {
			t.Errorf("Expecting not matched the flow conversation,but get matched\n")
		}
	}

}

func TestRuleMatch(t *testing.T) {

	var criteria map[uint64]*RuleCriteria
	var expecting map[uint64]*RuleMatchedResult
	cfMatchID := map[uint64]bool{
		1:  true,
		2:  false,
		3:  true,
		4:  true,
		5:  true,
		6:  false,
		7:  true,
		8:  true,
		9:  true,
		10: false,
	}

	criteria = map[uint64]*RuleCriteria{
		1: &RuleCriteria{ID: 1, Min: 3, Score: -2, Method: 1, CFIDs: []uint64{1, 2, 3, 4}},
	}
	expecting = map[uint64]*RuleMatchedResult{
		1: &RuleMatchedResult{Valid: true, Score: 0},
	}
	result, total, err := RuleMatch(cfMatchID, criteria)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		var expectingTotal int
		for id, c := range expecting {
			if val, ok := result[id]; ok {
				if c.Valid != val.Valid || c.Score != val.Score {
					t.Errorf("Expecting valid %t score %d, but get valid %t score %d\n", c.Valid, c.Score, val.Valid, val.Score)
				}
			} else {
				t.Errorf("Expecting get %d result,but get no\n", id)
			}
			expectingTotal = expectingTotal + c.Score
		}
		if expectingTotal != total {
			t.Errorf("Expecting total score %d, but get %d\n", expectingTotal, total)
		}
	}

	criteria = map[uint64]*RuleCriteria{
		1: &RuleCriteria{ID: 1, Min: 5, Score: -2, Method: 1, CFIDs: []uint64{1, 2, 3, 4}},
	}
	expecting = map[uint64]*RuleMatchedResult{
		1: &RuleMatchedResult{Valid: false, Score: -2},
	}
	result, total, err = RuleMatch(cfMatchID, criteria)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		var expectingTotal int
		for id, c := range expecting {
			if val, ok := result[id]; ok {
				if c.Valid != val.Valid || c.Score != val.Score {
					t.Errorf("Expecting valid %t score %d, but get valid %t score %d\n", c.Valid, c.Score, val.Valid, val.Score)
				}
			} else {
				t.Errorf("Expecting get %d result,but get no\n", id)
			}
			expectingTotal = expectingTotal + c.Score
		}
		if expectingTotal != total {
			t.Errorf("Expecting total score %d, but get %d\n", expectingTotal, total)
		}
	}

	criteria = map[uint64]*RuleCriteria{
		1: &RuleCriteria{ID: 1, Min: 5, Score: 4, Method: 1, CFIDs: []uint64{1, 2, 3, 4}},
		2: &RuleCriteria{ID: 2, Min: 5, Score: -2, Method: -1, CFIDs: []uint64{1, 2, 3, 4}},
		3: &RuleCriteria{ID: 3, Min: 3, Score: 5, Method: 1, CFIDs: []uint64{1, 2, 3, 4}},
	}
	expecting = map[uint64]*RuleMatchedResult{
		1: &RuleMatchedResult{Valid: false, Score: 0},
		2: &RuleMatchedResult{Valid: true, Score: 0},
		3: &RuleMatchedResult{Valid: true, Score: 5},
	}
	result, total, err = RuleMatch(cfMatchID, criteria)
	if err != nil {
		t.Errorf("Expecting no error,but get %s\n", err)
	} else {
		var expectingTotal int
		for id, c := range expecting {
			if val, ok := result[id]; ok {
				if c.Valid != val.Valid || c.Score != val.Score {
					t.Errorf("Expecting valid %t score %d, but get valid %t score %d\n", c.Valid, c.Score, val.Valid, val.Score)
				}
			} else {
				t.Errorf("Expecting get %d result,but get no\n", id)
			}
			expectingTotal = expectingTotal + c.Score
		}
		if expectingTotal != total {
			t.Errorf("Expecting total score %d, but get %d\n", expectingTotal, total)
		}
	}
}

var groupsToRules = map[uint64][]uint64{
	1: []uint64{11, 12, 13},
}
var rulesToCFs = map[uint64][]uint64{
	11: []uint64{21, 22, 23}, //true
	12: []uint64{24},         //false
	13: []uint64{25},         //true
}

var expectRule = map[uint64]struct {
	valid bool
	score int
}{
	11: {valid: true, score: 5},
	12: {valid: false, score: 0},
	13: {valid: true, score: 5},
}

var cfsToSenGrps = map[uint64][]uint64{
	21: []uint64{31, 32}, //if A then B
	22: []uint64{33, 34}, //if C then D
	23: []uint64{35, 36}, //if E then F
	24: []uint64{31, 32}, //if A then B
	25: []uint64{34, 35}, //if D then E
}

var expectCFCredit = map[uint64]bool{
	21: false,
	22: true,
	23: true,
	24: false,
	25: true,
}

var cfsExpression = map[uint64]string{
	21: "if A then B", //false
	22: "if C then D", //true
	23: "if E then F", //true
	24: "if A then B", //false
	25: "if D then E", //true
}

var sentenceGrpUUID = map[uint64]string{
	31: "A",
	32: "B",
	33: "C",
	34: "D",
	35: "E",
	36: "F",
}

//31->32
//33->34
//35->36
//34->35

//31,33,34,35,36
var senGrpsToSens = map[uint64][]uint64{
	31: []uint64{41},
	32: []uint64{42},
	33: []uint64{43},
	34: []uint64{43, 44},
	35: []uint64{45, 46},
	36: []uint64{46},
}
var expectSenGrpdCredit = map[uint64]bool{
	31: true,
	32: false,
	33: true,
	34: true,
	35: true,
	36: true,
}

//41,43,45,46
var sensToTags = map[uint64][]uint64{
	41: []uint64{51},
	42: []uint64{52, 51},
	43: []uint64{53},
	44: []uint64{54, 55},
	45: []uint64{55},
	46: []uint64{56},
}

var expectSensCredit = map[uint64]bool{
	41: true,
	42: false,
	43: true,
	44: false,
	45: true,
	46: true,
}

var expectTagSeg = map[uint64][]int{
	51: []int{1, 7},
	52: []int{2, 8},
	53: []int{3, 9},
	54: []int{4, 10},
	55: []int{5},
	56: []int{6},
}

var segToTags = map[int][]uint64{
	1: []uint64{51},
	2: []uint64{52},
	3: []uint64{53},
	4: []uint64{54},
	5: []uint64{55},
	6: []uint64{56},

	7:  []uint64{51},
	8:  []uint64{52},
	9:  []uint64{53},
	10: []uint64{54},
}

type mockRelationDao struct {
}

func (m *mockRelationDao) GetLevelRelationID(sql model.SqlLike, from int, to int, id []uint64) ([]map[uint64][]uint64, [][]uint64, error) {
	resp := make([]map[uint64][]uint64, 0)
	resp = append(resp, groupsToRules)
	resp = append(resp, rulesToCFs)
	resp = append(resp, cfsToSenGrps)
	resp = append(resp, senGrpsToSens)
	resp = append(resp, sensToTags)
	return resp, nil, nil

}

type mockPredictClient2 struct {
}

func (m *mockPredictClient2) Train(d *logicaccess.TrainUnit) error {
	return nil
}
func (m *mockPredictClient2) Status(d *logicaccess.TrainAPPID) (string, error) {
	return "", nil
}
func (m *mockPredictClient2) PredictAndUnMarshal(d *logicaccess.PredictRequest) (*logicaccess.PredictResult, error) {
	return nil, nil
}

func (m *mockPredictClient2) BatchPredictAndUnMarshal(d *logicaccess.BatchPredictRequest) (*logicaccess.PredictResult, error) {

	var resp logicaccess.PredictResult
	/*
		tagsToSegs := make(map[uint64][]int)
		for seg, tags := range segToTags {
			for _, tag := range tags {
				tagsToSegs[tag] = append(tagsToSegs[tag], seg)
			}
		}

		matched := len(tagsToSegs[d.ID])

		for i := 0; i < matched; i++ {
			var attr logicaccess.AttrResult

			attr.Tag = d.ID
			attr.SentenceID = tagsToSegs[d.ID][i]
			attr.Score = 87 + i
			resp.Dialogue = append(resp.Dialogue, attr)
		}
	*/
	for seg, tags := range segToTags {

		for _, tag := range tags {

			var attr logicaccess.AttrResult

			attr.Tag = tag
			attr.SentenceID = seg
			attr.Score = 87
			resp.Dialogue = append(resp.Dialogue, attr)
		}
	}
	return &resp, nil
}
func (m *mockPredictClient2) SessionCreate(d *logicaccess.SessionRequest) error {
	return nil
}
func (m *mockPredictClient2) SessionDelete(d *logicaccess.SessionRequest) error {
	return nil
}
func (m *mockPredictClient2) UnloadModel(d *logicaccess.TrainAPPID) error {
	return nil
}

type mockdbLike struct {
}

func (m *mockdbLike) Begin() (*sql.Tx, error) {
	return nil, nil
}
func (m *mockdbLike) ClearTransition(tx *sql.Tx) {

}
func (m *mockdbLike) Commit(tx *sql.Tx) error {
	return nil
}
func (m *mockdbLike) Conn() *sql.DB {
	return nil
}

type mockSentenceGroupsDao struct {
}

func (m *mockSentenceGroupsDao) Create(group *model.SentenceGroup, sql model.SqlLike) (*model.SentenceGroup, error) {
	return nil, nil
}
func (m *mockSentenceGroupsDao) CountBy(filter *model.SentenceGroupFilter, sql model.SqlLike) (int64, error) {
	return int64(len(senGrpsToSens)), nil
}
func (m *mockSentenceGroupsDao) GetBy(filter *model.SentenceGroupFilter, sql model.SqlLike) ([]model.SentenceGroup, error) {
	var resp []model.SentenceGroup

	for sgID, sIDs := range senGrpsToSens {
		var s model.SentenceGroup
		s.ID = int64(sgID)
		s.Distance = -1
		s.Position = -1
		s.Role = -1
		s.UUID = sentenceGrpUUID[uint64(s.ID)]
		for _, sID := range sIDs {
			var sentence model.SimpleSentence
			sentence.ID = sID
			s.Sentences = append(s.Sentences, sentence)
		}
		resp = append(resp, s)
	}
	return resp, nil
}
func (m *mockSentenceGroupsDao) Update(id string, group *model.SentenceGroup, sql model.SqlLike) (*model.SentenceGroup, error) {
	return nil, nil
}
func (m *mockSentenceGroupsDao) Delete(id string, sqllike model.SqlLike) error {
	return nil
}

type mockCfDao struct {
}

func (m *mockCfDao) Create(flow *model.ConversationFlow, sql model.SqlLike) (*model.ConversationFlow, error) {
	return nil, nil
}
func (m *mockCfDao) CountBy(filter *model.ConversationFlowFilter, sql model.SqlLike) (int64, error) {
	return int64(len(cfsToSenGrps)), nil
}
func (m *mockCfDao) GetBy(filter *model.ConversationFlowFilter, sql model.SqlLike) ([]model.ConversationFlow, error) {
	var resp []model.ConversationFlow

	for cfID, sgIDs := range cfsToSenGrps {
		var s model.ConversationFlow
		s.ID = int64(cfID)
		s.Expression = cfsExpression[cfID]
		s.Min = 1
		for _, sgID := range sgIDs {
			var senGrp model.SimpleSentenceGroup
			senGrp.UUID = sentenceGrpUUID[sgID]
			senGrp.ID = int64(sgID)
			s.SentenceGroups = append(s.SentenceGroups, senGrp)
		}
		resp = append(resp, s)
	}
	return resp, nil
}
func (m *mockCfDao) Update(id string, flow *model.ConversationFlow, sql model.SqlLike) (*model.ConversationFlow, error) {
	return nil, nil
}
func (m *mockCfDao) Delete(id string, sql model.SqlLike) error {
	return nil
}

type mockRuleDao struct {
}

func (m *mockRuleDao) Create(rule *model.ConversationRule, sql model.SqlLike) (*model.ConversationRule, error) {
	return nil, nil
}
func (m *mockRuleDao) CountBy(filter *model.ConversationRuleFilter, sql model.SqlLike) (int64, error) {
	return int64(len(rulesToCFs)), nil
}
func (m *mockRuleDao) GetBy(filter *model.ConversationRuleFilter, sql model.SqlLike) ([]model.ConversationRule, error) {
	var resp []model.ConversationRule

	for rID, cfIDs := range rulesToCFs {
		var s model.ConversationRule
		s.ID = int64(rID)
		s.Min = 1
		s.Method = 1
		s.Score = 5
		for _, cfID := range cfIDs {
			var cf model.SimpleConversationFlow

			cf.ID = int64(cfID)
			s.Flows = append(s.Flows, cf)
		}
		resp = append(resp, s)
	}
	return resp, nil
}
func (m *mockRuleDao) Delete(id string, sql model.SqlLike) error {
	return nil
}

type mockGroupDaoMatch struct {
}

func (m *mockGroupDaoMatch) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockGroupDaoMatch) Commit(tx *sql.Tx) error {
	return nil
}

func (m *mockGroupDaoMatch) ClearTranscation(tx *sql.Tx) {}

func (m *mockGroupDaoMatch) CreateGroup(group *model.GroupWCond, tx *sql.Tx) (*model.GroupWCond, error) {
	return nil, nil
}

func (m *mockGroupDaoMatch) GetGroupBy(id string) (*model.GroupWCond, error) {
	return nil, nil
}

func (m *mockGroupDaoMatch) CountGroupsBy(filter *model.GroupFilter) (int64, error) {
	return 2, nil
}

func (m *mockGroupDaoMatch) GetGroupsBy(filter *model.GroupFilter) ([]model.GroupWCond, error) {
	return []model.GroupWCond{
		model.GroupWCond{Enterprise: "abcdeg"},
	}, nil
}

func (m *mockGroupDaoMatch) DeleteGroup(id string, tx *sql.Tx) (err error) {
	return
}

func (m *mockGroupDaoMatch) Group(delegatee model.SqlLike, query model.GroupQuery) ([]model.Group, error) {
	return nil, nil
}

func (m *mockGroupDaoMatch) GroupsByCalls(delegatee model.SqlLike, query model.CallQuery) (map[int64][]model.Group, error) {
	return nil, nil
}

type mockTrainedModelDao struct {
}

func (m *mockTrainedModelDao) TrainedModelInfo(conn model.SqlLike, q *model.TModelQuery) ([]*model.TModel, error) {
	return []*model.TModel{&model.TModel{ID: 4444}}, nil
}

func (m *mockTrainedModelDao) DeleteModel(conn model.SqlLike, q *model.TModelQuery) (int64, error) {
	return 0, nil
}

func (m *mockTrainedModelDao) NewModel(conn model.SqlLike, q *model.TModel) (int64, error) {
	return 0, nil
}

func (m *mockTrainedModelDao) UpdateModel(conn model.SqlLike, q *model.TModel) (int64, error) {
	return 0, nil
}
func TestRuleGroupCriteria(t *testing.T) {
	mockRelation := &mockRelationDao{}
	relationDao = mockRelation
	predictor = &mockPredictClient2{}

	modelDao = &mockTrainedModelDao{}
	dbLike = &mockdbLike{}

	sentenceGroupDao = &mockSentenceGroupsDao{}
	conversationFlowDao = &mockCfDao{}
	conversationRuleDao = &mockRuleDao{}

	serviceDAO = &mockGroupDaoMatch{}

	segments := make([]*SegmentWithSpeaker, 0, len(segToTags))
	for i := 0; i < len(segToTags); i++ {
		s := &SegmentWithSpeaker{Speaker: i % 2}
		segments = append(segments, s)
	}

	timeout := time.Duration(3 * time.Second)

	c, err := RuleGroupCriteria(model.Group{ID: 1}, segments, timeout)
	if err != nil {
		t.Errorf("expecting no error, but get %s\n", err)
	} else {
		for groupID, ruleIDs := range groupsToRules {
			var expect RuleGrpCredit

			//generat the expecting result
			expect.ID = groupID
			for _, ruleID := range ruleIDs {
				rule := &RuleCredit{ID: ruleID, Valid: expectRule[ruleID].valid, Score: expectRule[ruleID].score}
				expect.Rules = append(expect.Rules, rule)
				expect.Plus += rule.Score

				cfs := rulesToCFs[ruleID]
				for _, cf := range cfs {
					cfresult := &ConversationFlowCredit{ID: cf, Valid: expectCFCredit[cf]}
					rule.CFs = append(rule.CFs, cfresult)

					senGrps := cfsToSenGrps[cf]
					for _, senGrp := range senGrps {
						senGrpResult := &SentenceGrpCredit{ID: senGrp, Valid: expectSenGrpdCredit[senGrp]}
						cfresult.SentenceGrps = append(cfresult.SentenceGrps, senGrpResult)

						sentences := senGrpsToSens[senGrp]
						for _, sentence := range sentences {
							senResult := &SentenceCredit{ID: sentence, Valid: expectSensCredit[sentence]}
							senGrpResult.Sentences = append(senGrpResult.Sentences, senResult)
							if senResult.Valid {
								tags := sensToTags[sentence]
								for _, tag := range tags {

									segs := expectTagSeg[tag]

									for _, seg := range segs {
										tagResult := &TagCredit{ID: tag, SegmentIdx: seg}
										senResult.Tags = append(senResult.Tags, tagResult)
									}
								}
							}
						}
					}
				}
			}
			/*
				b, _ := json.Marshal(expect)
				fmt.Printf("%s\n", b)
			*/
			//check the result
			if expect.ID != c.ID {
				t.Fatalf("expecting rule group id %d, but get %d\n", expect.ID, c.ID)
			}
			if len(expect.Rules) != len(c.Rules) {
				t.Fatalf("expecting get %d rule result, but get %d\n", len(expect.Rules), len(c.Rules))
			}

			for ridx, rrules := range c.Rules {
				expectR := expect.Rules[ridx]
				if expectR.ID != rrules.ID {
					t.Fatalf("expecting get %d'th rule %d, but get %d\n", ridx, expectR.ID, rrules.ID)
				}

				if expectR.Score != rrules.Score {
					t.Fatalf("expecting %d'th rule %d get score %d, but get %d\n", ridx, expectR.ID, expectR.Score, rrules.Score)
				}

				if expectR.Valid != rrules.Valid {
					t.Fatalf("expecting %d'th rule %d get valid %t, but get %t\n", ridx, expectR.ID, expectR.Valid, rrules.Valid)
				}

				if len(expectR.CFs) != len(rrules.CFs) {
					t.Fatalf("expecting %d'th rule %d get %d conversation flow, but get %d\n", ridx, expectR.ID, len(expectR.CFs), len(rrules.CFs))
				}

				for cfIdx, cfs := range rrules.CFs {
					expectCF := expectR.CFs[cfIdx]
					if expectCF.ID != cfs.ID {
						t.Fatalf("expecting %d'th rule %d %d'th conversation flow get id %d, but get %d\n",
							ridx, expectR.ID, cfIdx, expectCF.ID, cfs.ID)
					}
					if expectCF.Valid != cfs.Valid {
						t.Fatalf("expecting %d'th rule %d %d'th conversation flow id %d get valid %t, but get %t\n",
							ridx, expectR.ID, cfIdx, expectCF.ID, expectCF.Valid, cfs.Valid)
					}
					if len(expectCF.SentenceGrps) != len(cfs.SentenceGrps) {
						t.Fatalf("expecting %d'th rule %d %d'th conversation flow id %d get %d sentence group, but get %d\n",
							ridx, expectR.ID, cfIdx, expectCF.ID, len(expectCF.SentenceGrps), len(cfs.SentenceGrps))
					}

					for senGrpIdx, senGrps := range cfs.SentenceGrps {
						expectSenGrps := expectCF.SentenceGrps[senGrpIdx]
						if expectSenGrps.ID != senGrps.ID {
							t.Fatalf("expecting %d'th rule %d %d'th conversation flow %d %d'th sentence group get %d, but get %d\n",
								ridx, expectR.ID, cfIdx, expectCF.ID, senGrpIdx, expectSenGrps.ID, senGrps.ID)
						}
						if expectSenGrps.Valid != senGrps.Valid {
							t.Fatalf("expecting %d'th rule %d %d'th conversation flow %d %d'th sentence group %d get valid %t, but get %t\n",
								ridx, expectR.ID, cfIdx, expectCF.ID, senGrpIdx, expectSenGrps.ID, expectSenGrps.Valid, senGrps.Valid)
						}

						if len(expectSenGrps.Sentences) != len(senGrps.Sentences) {
							t.Fatalf("expecting %d'th rule %d %d'th conversation flow %d %d'th sentence group %d get %d sentence, but get %d\n",
								ridx, expectR.ID, cfIdx, expectCF.ID, senGrpIdx, expectSenGrps.ID, len(expectSenGrps.Sentences), len(senGrps.Sentences))
						}

						for senIdx, sentence := range senGrps.Sentences {
							expectSen := expectSenGrps.Sentences[senIdx]
							if expectSen.ID != sentence.ID {
								t.Fatalf("expecting %d'th rule %d, %d'th conversation flow %d, %d'th sentence group %d,%d'th sentence get id %d, but get %d\n",
									ridx, expectR.ID, cfIdx, expectCF.ID, senGrpIdx, expectSenGrps.ID, senIdx, expectSen.ID, sentence.ID)
							}

							if expectSen.Valid != sentence.Valid {
								t.Fatalf("expecting %d'th rule %d, %d'th conversation flow %d, %d'th sentence group %d,%d'th sentence id %d get valid %t, but get %t\n",
									ridx, expectR.ID, cfIdx, expectCF.ID, senGrpIdx, expectSenGrps.ID, senIdx, expectSen.ID, expectSen.Valid, sentence.Valid)
							}

							if len(expectSen.Tags) != len(sentence.Tags) {
								t.Fatalf("expecting %d'th rule %d, %d'th conversation flow %d, %d'th sentence group %d,%d'th sentence id %d get %d tag, but get %d\n",
									ridx, expectR.ID, cfIdx, expectCF.ID, senGrpIdx, expectSenGrps.ID, senIdx, expectSen.ID, len(expectSen.Tags), len(sentence.Tags))
							}

							for tIdx, tag := range sentence.Tags {
								expectTag := expectSen.Tags[tIdx]

								if expectTag.ID != tag.ID {
									t.Fatalf("expecting %d'th rule %d, %d'th conversation flow %d, %d'th sentence group %d,%d'th sentence id %d, %d'th tag get %d, but get %d\n",
										ridx, expectR.ID, cfIdx, expectCF.ID, senGrpIdx, expectSenGrps.ID, senIdx, expectSen.ID, tIdx, expectTag.ID, tag.ID)
								}
							}
						}
					}
				}

			}
			break
		}

	}

}
