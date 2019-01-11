package qi

import (
	"testing"
	"time"

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
	segments := make([]*ASRSegment, numOfSegs, numOfSegs)
	for i := 0; i < numOfSegs; i++ {
		segments[i] = new(ASRSegment)
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
