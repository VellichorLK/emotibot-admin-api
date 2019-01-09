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

var mockMatchIdx = []uint64{1, 7}
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

	senCriteria := make(map[uint64][]uint64)
	senCriteria[1] = append(senCriteria[1], 1)
	senCriteria[1] = append(senCriteria[1], 6)

	senCriteria[9] = append(senCriteria[9], 2)
	senCriteria[9] = append(senCriteria[9], 7)

	senMatch, err := SentencesMatch(matched, senCriteria)
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

	senMatch, err = SentencesMatch(matched, senCriteria)
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
