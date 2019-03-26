package model

import (
	"encoding/csv"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readMockSegments(t *testing.T) []RealSegment {
	f, err := os.Open("./testdata/seed/Segment.csv")
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	var segments = make([]RealSegment, 0, len(rows)-1)
	for _, cols := range rows[1:] {
		var s RealSegment
		Binding(&s, cols)
		segments = append(segments, s)
	}
	return segments
}

func TestITSegmentDao_Segments(t *testing.T) {
	skipIntergartion(t)
	dao := SegmentSQLDao{db: newIntegrationTestDB(t)}
	segments := readMockSegments(t)
	type args struct {
		query SegmentQuery
	}
	tests := []struct {
		name string
		args args
		want []RealSegment
	}{
		{
			name: "query by call",
			args: args{
				query: SegmentQuery{
					CallID: []int64{1},
				},
			},
			want: []RealSegment{segments[0]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.Segments(nil, tt.args.query)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
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
			s := &SegmentSQLDao{
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
