package imagesManager

import (
	"os"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestMain(m *testing.M) {
	os.Setenv("SYNC_PERIOD_BY_SECONDS", "100")
	retCode := m.Run()
	os.Exit(retCode)
}

func TestDaemonFindImages(t *testing.T) {
	type testCase struct {
		input      string
		imagesDict map[string]int
		expected   []string
	}
	testCases := map[string]testCase{
		"NotFound": testCase{
			"无http无端口号8",
			[]string{},
		},
		"Normal": testCase{
			"Test<img src=\"emotibot.com\"/>",
			[]string{"emotibot.com"},
		},
		"EmptySrc": testCase{
			"Test<img src=\"\" formatted wrong",

			[]string{""},
		},
		"MultipleSrc": testCase{
			"Test1<img src=\"http://google.com/sdfs.jpg\"/>Test2<img src=\"gopher.png\"/>",
			map[string]int{
				"sdfs.jpg":   1,
				"gopher.png": 2,
			},
			[]string{"sdfs.jpg", "gopher.png"},
		},
	}
	questionDB, questionMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var d = NewDaemon(5, nil, questionDB)
			rows := sqlmock.NewRows([]string{"Answer_id", "Content"}).AddRow(1, tt.input)
			questionMock.ExpectQuery("SELECT Answer_Id, Content ").WillReturnRows(rows)
			stmt := questionMock.ExpectPrepare("SELECT id FROM images ")
			stmt.ExpectQuery()
			answers, err := d.FindImages()
			if err != nil {
				t.Fatal(err)
			}
			if len(tt.expected) != len(answers[1]) {
				t.Fatalf("expect to find %d of img, but got %d", len(tt.expected), len(answers[1]))
			}
			for i, src := range tt.expected {

				if imgSrc := answers[1][i]; imgSrc != src {
					t.Fatalf("expect extractFileName %s but got %s", tt.expected[i], imgSrc)
				}
			}
		})
	}

}

func TestFindImageIDByContent(t *testing.T) {
	type testCase struct {
		input    []string
		expected map[string]int
	}

	testCases := map[string]testCase{
		"Find 3": testCase{
			input: []string{"1", "2", "3"},
			expected: map[string]int{
				"1": 1,
				"2": 2,
				"3": 3,
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var mocker sqlmock.Sqlmock
			db, mocker, _ = sqlmock.New()
			stmt := mocker.ExpectPrepare("SELECT id FROM images")
			for _, str := range tt.input {
				row := sqlmock.NewRows([]string{"id"}).AddRow(tt.expected[str])
				stmt.ExpectQuery().WithArgs(str).WillReturnRows(row)
			}
			imgs, err := FindImageIDByContent(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			for content, id := range tt.expected {
				imageID, ok := imgs[content]
				if !ok {
					t.Fatalf("expect map contains %s, but cant find", content)
				}
				if imageID != id {
					t.Fatalf("expect image ID be %d, but got %d", id, imageID)
				}
			}
		})
	}
}
