package imagesManager

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestDaemonFindImages(t *testing.T) {
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
			var d = NewDaemon(5, picDB, questionDB)
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

			answers, err := d.FindImages()
			if err != nil {
				t.Fatal(err)
			}

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

// func TestFindImageIDByContent(t *testing.T) {
// 	type testCase struct {
// 		input    []string
// 		expected map[string]int
// 	}

// 	testCases := map[string]testCase{
// 		"Find 3": testCase{
// 			input: []string{"1", "2", "3"},
// 			expected: map[string]int{
// 				"1": 1,
// 				"2": 2,
// 				"3": 3,
// 			},
// 		},
// 	}

// 	for name, tt := range testCases {
// 		t.Run(name, func(t *testing.T) {
// 			var mocker sqlmock.Sqlmock
// 			db, mocker, _ = sqlmock.New()
// 			stmt := mocker.ExpectPrepare("SELECT id FROM images")
// 			for _, str := range tt.input {
// 				row := sqlmock.NewRows([]string{"id"}).AddRow(tt.expected[str])
// 				stmt.ExpectQuery().WithArgs(str).WillReturnRows(row)
// 			}
// 			imgs, err := FindImageIDByContent(tt.input)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			for content, id := range tt.expected {
// 				imageID, ok := imgs[content]
// 				if !ok {
// 					t.Fatalf("expect map contains %s, but cant find", content)
// 				}
// 				if imageID != id {
// 					t.Fatalf("expect image ID be %d, but got %d", id, imageID)
// 				}
// 			}
// 		})
// 	}
// }
