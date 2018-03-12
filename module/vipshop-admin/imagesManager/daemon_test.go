package imagesManager

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func BenchmarkFindImagesJOb(b *testing.B) {
	var questionMock, picMock sqlmock.Sqlmock
	var j FindImageJob
	j.questionDB, questionMock, _ = sqlmock.New()
	j.picDB, picMock, _ = sqlmock.New()
	for i := 0; i <= b.N; i++ {
		rows := sqlmock.NewRows([]string{"Answer_id", "Content"}).AddRow(1, "Test1<img src=\"http://google.com/sdfs.jpg\"/>Test2<img src=\"gopher.png\"/>")
		questionMock.ExpectQuery("SELECT Answer_Id, Content ").WillReturnRows(rows)
		stmt := picMock.ExpectPrepare("SELECT id FROM images ")
		stmt.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		stmt.ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		err := j.Do(nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func TestFindImagesJob(t *testing.T) {
	var images = map[int]string{
		1: "sdfs.jpg",
		2: "gopher.png",
	}
	type testCase struct {
		input    string
		expected []int
	}
	testCases := map[string]testCase{
		"NotFound": testCase{
			"无http无端口号8",
			[]int{},
		},
		"Normal": testCase{
			"Test<img src=\"emotibot.com\"/>",
			[]int{0},
		},
		"EmptySrc": testCase{
			"Test<img src=\"\" formatted wrong",
			[]int{0},
		},
		"MultipleSrc": testCase{
			"Test1<img src=\"http://google.com/sdfs.jpg\"/>Test2<img src=\"gopher.png\"/>",
			[]int{1, 2},
		},
	}
	questionDB, questionMock, err := sqlmock.New()
	picDB, picMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			j := FindImageJob{
				questionDB: questionDB,
				picDB:      picDB,
			}
			rows := sqlmock.NewRows([]string{"Answer_id", "Content"}).AddRow(1, tt.input)
			questionMock.ExpectQuery("SELECT Answer_Id, Content ").WillReturnRows(rows)
			stmt := picMock.ExpectPrepare("SELECT id FROM images ")
			for _, expectImgID := range tt.expected {
				rows = sqlmock.NewRows([]string{"id"})
				img, ok := images[expectImgID]
				if ok {
					rows.AddRow(expectImgID)
					stmt.ExpectQuery().WithArgs(img).WillReturnRows(rows)
				} else {
					stmt.ExpectQuery().WillReturnRows(rows)
				}
			}

			err := j.Do(nil)
			if err != nil {
				t.Fatal(err)
			}
			answers := j.Result
			for i, id := range tt.expected {
				if id == 0 {
					continue
				}
				gotImgID := answers[1][i]
				_, ok := images[gotImgID]
				if !ok {
					t.Fatal("expect id %d in images", gotImgID)
				}
				if gotImgID != id {
					t.Fatalf("expect extract got image No.%d but got %d", id, gotImgID)
				}
			}
		})
	}

}

func TestLinkImageJob(t *testing.T) {
	type testCase struct {
		input    map[int][]int
		expected bool
	}

	testCases := map[string]testCase{
		"Normal": testCase{
			input: map[int][]int{
				1: []int{1, 2},
				2: []int{1, 3},
			},
			expected: true,
		},
	}
	var picMocker sqlmock.Sqlmock
	db, picMocker, _ = sqlmock.New()
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			job := LinkImageJob{
				input: tt.input,
			}
			var size int
			picMocker.ExpectBegin()
			picMocker.ExpectExec("DELETE FROM image_answer").WillReturnResult(sqlmock.NewResult(int64(0), 0))
			stmt := picMocker.ExpectPrepare("INSERT INTO image_answer")
			picMocker.MatchExpectationsInOrder(false)
			for key, values := range tt.input {
				size += len(values)
				for _, v := range values {
					stmt.ExpectExec().WithArgs(key, v).WillReturnResult(sqlmock.NewResult(int64(key), 1))
				}
			}
			err := job.Do(nil)
			if err != nil {
				t.Fatal(err)
			}
			if err = picMocker.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}

			if job.AffecttedRows != size {
				t.Errorf("expect job insert %d of rows, but got %v", size, job.AffecttedRows)
			}
			if job.IsDone != tt.expected {
				t.Errorf("expect job IsDone is %v, but got %v", tt.expected, job.IsDone)
			}
		})
	}
}
