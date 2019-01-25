package model

import (
	"reflect"
	"testing"
)

func TestITSegmentDaoNewSegments(t *testing.T) {
	skipIntergartion(t)
	type args struct {
		segments []RealSegment
	}
	segments := []RealSegment{
		RealSegment{
			ID:         1,
			CallID:     1,
			StartTime:  0,
			EndTime:    10,
			CreateTime: 1548225286,
			Channel:    1,
			Text:       "testing",
			Emotions: []RealSegmentEmotion{
				RealSegmentEmotion{
					ID:        1,
					SegmentID: 1,
					Typ:       ETypAngry,
					Score:     100,
				},
			},
		},
	}
	tests := []struct {
		name    string
		args    args
		want    []RealSegment
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				segments: segments,
			},
			want: segments,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SegmentDao{
				db: newIntegrationTestDB(t),
			}
			got, err := s.NewSegments(nil, tt.args.segments)
			if (err != nil) != tt.wantErr {
				t.Errorf("SegmentDao.NewSegments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SegmentDao.NewSegments() = %v, want %v", got, tt.want)
			}
		})
	}
}
