package qi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

func TestASRResponse_Segments(t *testing.T) {
	defer BackupPointers(unix)
	unix = func() int64 {
		return 0
	}
	type fields struct {
		LeftChannel  vChannel
		RightChannel vChannel
	}
	tests := []struct {
		name   string
		fields fields
		want   []model.RealSegment
	}{
		{
			name: "simple sentences",
			fields: fields{
				LeftChannel: vChannel{
					Sentences: []voiceSentence{
						voiceSentence{
							Start:   1.0,
							End:     2.0,
							ASR:     "hello",
							Emotion: 10,
							Status:  200,
						},
						voiceSentence{
							Start:   2.1,
							End:     3.0,
							ASR:     "world",
							Emotion: 20,
							Status:  200,
						},
					},
				},
				RightChannel: vChannel{
					Sentences: []voiceSentence{},
				},
			},
			want: []model.RealSegment{
				{
					StartTime: 1.0,
					EndTime:   3.0,
					Text:      "hello world",
					Status:    200,
					Channel:   1,
					Emotions: []model.RealSegmentEmotion{
						model.RealSegmentEmotion{
							Typ:   model.ETypAngry,
							Score: 20,
						},
					},
				},
			},
		},
		{
			name: "error status should not be merged",
			fields: fields{
				LeftChannel: vChannel{
					Sentences: []voiceSentence{
						voiceSentence{
							Start:   1.0,
							End:     2.0,
							ASR:     "hello",
							Emotion: 10,
							Status:  200,
						},
						voiceSentence{
							Start:  2.0,
							End:    2.1,
							Status: 500,
						},
						voiceSentence{
							Start:   2.1,
							End:     3.0,
							ASR:     "world",
							Emotion: 20,
							Status:  200,
						},
					},
				},
			},
			want: []model.RealSegment{
				{
					StartTime: 1.0,
					EndTime:   2.0,
					Text:      "hello",
					Status:    200,
					Channel:   1,
					Emotions: []model.RealSegmentEmotion{
						model.RealSegmentEmotion{
							Typ:   model.ETypAngry,
							Score: 10,
						},
					},
				},
				{
					StartTime: 2.0,
					EndTime:   2.1,
					Status:    500,
					Channel:   1,
					Emotions: []model.RealSegmentEmotion{
						{
							Typ:   model.ETypAngry,
							Score: 0,
						},
					},
				},
				{
					StartTime: 2.1,
					EndTime:   3.0,
					Text:      "world",
					Status:    200,
					Channel:   1,
					Emotions: []model.RealSegmentEmotion{
						{
							Typ:   model.ETypAngry,
							Score: 20,
						},
					},
				},
			},
		},
		{
			name: "two channel merging & ordering",
			fields: fields{
				LeftChannel: vChannel{
					Sentences: []voiceSentence{
						{
							Start:   2,
							End:     4,
							ASR:     "Bonjour",
							Emotion: 20,
							Status:  200,
						},
					},
				},
				RightChannel: vChannel{
					Sentences: []voiceSentence{
						voiceSentence{
							Start:   2,
							End:     3,
							ASR:     "世界",
							Emotion: 0,
							Status:  200,
						},
						voiceSentence{
							Start:   0,
							End:     2,
							ASR:     "你好,",
							Emotion: 100,
							Status:  200,
						},
					},
				},
			},
			want: []model.RealSegment{
				{
					StartTime: 0,
					EndTime:   3,
					Text:      "你好, 世界",
					Status:    200,
					Channel:   2,
					Emotions: []model.RealSegmentEmotion{
						model.RealSegmentEmotion{
							Typ:   model.ETypAngry,
							Score: 100,
						},
					},
				},
				{
					StartTime: 2,
					EndTime:   4,
					Text:      "Bonjour",
					Status:    200,
					Channel:   1,
					Emotions: []model.RealSegmentEmotion{
						model.RealSegmentEmotion{
							Typ:   model.ETypAngry,
							Score: 20,
						},
					},
				},
			},
		},
		{
			name: "maximum length check",
			fields: fields{
				LeftChannel: vChannel{
					Sentences: []voiceSentence{
						{
							Start:   2,
							End:     4,
							ASR:     strings.Repeat("a", 250),
							Emotion: 20,
							Status:  200,
						},
						{
							Start:  3,
							End:    5,
							ASR:    strings.Repeat("b", 250),
							Status: 200,
						},
					},
				},
			},
			want: []model.RealSegment{
				{
					StartTime: 2,
					EndTime:   4,
					Text:      strings.Repeat("a", 250),
					Emotions: []model.RealSegmentEmotion{
						model.RealSegmentEmotion{
							Typ:   model.ETypAngry,
							Score: 20,
						},
					},
					Status:  200,
					Channel: 1,
				},
				{
					StartTime: 3,
					EndTime:   5,
					Text:      strings.Repeat("b", 250),
					Emotions: []model.RealSegmentEmotion{
						model.RealSegmentEmotion{
							Typ:   model.ETypAngry,
							Score: 0,
						},
					},
					Status:  200,
					Channel: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &ASRResponse{
				LeftChannel:  tt.fields.LeftChannel,
				RightChannel: tt.fields.RightChannel,
			}
			got := resp.Segments()
			assert.Equal(t, tt.want, got)
		})
	}
}
