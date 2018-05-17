package imagesManager

import (
	"fmt"
	"reflect"
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

func TestDoFetchImagesJob(t *testing.T) {
	db, mocker, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	var j = FetchDBJob{
		PicDB: db,
	}
	var expectedRows = map[string]uint64{
		"c4ca4238a0b923820dcc509a6f75849b.jpg": 1,
		"c81e728d9d4c2f636f067f89cc14862c.png": 2,
	}

	rows := sqlmock.NewRows([]string{"id", "raw_file_name"})
	for fileName, id := range expectedRows {
		rows.AddRow(id, fileName)
	}
	mocker.ExpectQuery("SELECT id, raw_file_name FROM images").WillReturnRows(rows)
	err = j.Do(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedRows, j.Output) {
		t.Error("expect output is equal to input.")
	}
}
func TestFindImagesJob(t *testing.T) {
	var images = map[string]uint64{
		"c4ca4238a0b923820dcc509a6f75849b.jpg": 1,
		"c81e728d9d4c2f636f067f89cc14862c.png": 2,
		"9bce2d0f8473453d8a3847bbd807ca55.jpg": 3,
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
			"Test<img src=\"http://vip.api.com/c4ca4238a0b923820dcc509a6f75849b.jpg\"/>",
			[]int{1},
		},
		"EmptySrc": testCase{
			"Test<img src=\"\" formatted wrong",
			[]int{},
		},
		"WithoutExt": testCase{
			"Test<img src=\"http://vip.api.com/12345678\"/>",
			[]int{},
		},
		"MultipleSrc": testCase{
			"Test1<img src=\"http://google.com/c4ca4238a0b923820dcc509a6f75849b.jpg\"/>Test2<img src=\"c81e728d9d4c2f636f067f89cc14862c.png\"/>",
			[]int{3, 2},
		},
		"Test": testCase{
			`<p><img src="https://vca.vip.com/img/9bce2d0f8473453d8a3847bbd807ca55.jpg"/><img src="https://vca.vip.com/img/98e12b535c5b43da9d8618ffb0c1851a.png"/></p><p><br/></p><p><img src="https://vca.vip.com/img/9bce2d0f8473453d8a3847bbd807ca55.jpg"/></p>`,
			[]int{4},
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
				LookUpImages: images,
				questionDB:   questionDB,
				picDB:        picDB,
			}
			stmt := picMock.ExpectQuery("SELECT id FROM images ")
			rows := sqlmock.NewRows([]string{"id"})
			for id := range images {
				rows.AddRow(id)
			}
			stmt.WillReturnRows(rows)
			var mockedAnswerID uint64 = 1
			rows = sqlmock.NewRows([]string{"Answer_id", "Content"}).AddRow(mockedAnswerID, tt.input)
			questionMock.ExpectQuery("SELECT Answer_Id, Content ").WillReturnRows(rows)

			err := j.Do(nil)
			if err != nil {
				t.Fatal(err)
			}
			answers := j.Result
			fmt.Printf("%v\n", answers)

			if len(tt.expected) == 0 {
				if len(answers) != 0 {
					t.Fatalf("expect")
				}
				return
			}

			expectedImgIDGroup := answers[mockedAnswerID]
			if len(expectedImgIDGroup) == 0 {
				t.Fatal("expect got at least one images id, but no images in result.")
			}
			for expectedImgID := range expectedImgIDGroup {
				var found bool
				for _, id := range images {
					if id == expectedImgID {
						found = true
					}
				}
				if !found {
					t.Fatalf("expect id %d extracted from answer %v", expectedImgID, tt.input)
				}

			}
		})
	}

}

func TestLinkImageJob(t *testing.T) {
	type testCase struct {
		input    map[uint64]imageSet
		expected bool
	}

	testCases := map[string]testCase{
		"Normal": testCase{
			input: map[uint64]imageSet{
				1: imageSet{1: 1, 2: 1},
				2: imageSet{1: 1, 3: 1},
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
				for id := range values {
					stmt.ExpectExec().WithArgs(key, id).WillReturnResult(sqlmock.NewResult(int64(key), 1))
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
