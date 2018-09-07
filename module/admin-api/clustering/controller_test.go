package clustering

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"emotibot.com/emotigo/pkg/api/faqcluster/v1"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

//We need a way to mock es service, skip
func TestNewDoReportHandler(t *testing.T) {
	t.Skip("do not have a way to mock elastic search for now, so skip the test.")
	db, writer, _ := sqlmock.New()
	ss := &sqlService{db: db}
	writer.ExpectQuery("SELECT .+ FROM `reports`").WillReturnRows(sqlmock.NewRows([]string{}))
	writer.ExpectQuery("")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockContent := ``
		fmt.Fprint(w, mockContent)
	}))
	defer server.Close()
	addr, _ := url.Parse(server.URL)
	faqClient := faqcluster.NewClientWithHTTPClient(addr, server.Client())
	handler := NewDoReportHandler(ss, ss, ss, ss, faqClient)
	worker = newMockWorker()
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodPut, "/test", nil))
	//TODO: NOT VERIFY YET

	if err := writer.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func newMockWorker() cluster {
	return func(reportID uint64, paramas map[string]interface{}, inputs []interface{}) error {
		return nil
	}
}
func TestNewClusteringWorkWithSQLService(t *testing.T) {
	db, writer, _ := sqlmock.New()
	ss := &sqlService{db: db}
	// mockedResult := faqcluster.Result{
	// 	Clusters: []faqcluster.Cluster{
	// 		faqcluster.Cluster{
	// 			Data: []faqcluster.Data{faqcluster.Data{}, faqcluster.Data[]},
	// 		},
	// 		faqcluster.Cluster{
	// 			Data: []faqcluster.Data{faqcluster.Data{}},
	// 		},
	// 	},
	// 	Filtered: []faqcluster.Data{faqcluster.Data{}, faqcluster.Data{}},
	// }
	testCase := map[string]int{
		"clusterSize": 2,
		"sentence":    1,
		"RemovedSize": 3,
	}

	for turn := 1; turn <= testCase["clusterSize"]; turn++ {
		writer.ExpectExec("INSERT INTO `report_clusters`").WillReturnResult(sqlmock.NewResult(1, 1))
		writer.ExpectBegin()
		prepare := writer.ExpectPrepare("INSERT INTO `report_records`")
		for i := 0; i < testCase["sentence"]; i++ {
			prepare.ExpectExec().WillReturnResult(sqlmock.NewResult(int64(turn+i), 1))
		}

	}
	writer.ExpectBegin()
	prepare := writer.ExpectPrepare("INSERT INTO `report_records`")
	for i := 1; i <= testCase["RemovedSize"]; i++ {
		prepare.ExpectExec().WillReturnResult(sqlmock.NewResult(int64(i), 1))
	}
	writer.ExpectExec("UPDATE `reports` SET `status` = ?").WillReturnResult(sqlmock.NewResult(1, 1))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//mockContent is a simple clustering result with 2 cluster and 3 removed data
		mockContent := `{"errno":"success","error_message":"","para":{"model_version":"unknown_20180830143445","deduplicate":false},"result":{"data":[{"centerQuestion":["1.6高，体重120穿多大号"],"clusterTag":["測試"],"cluster":[{"id":"9","value":"1.6高，体重120穿多大号"}]},{"centerQuestion":["測試B"],"clusterTag":["測試"],"cluster":[{"id":"1","value":"測試B"}]}],"removed":[{"id":"1","value":"1"},{"id":"2","value":"1,76"},{"id":"12","value":"1.8床"}]}}`
		fmt.Fprint(w, mockContent)
	}))
	defer server.Close()
	addr, _ := url.Parse(server.URL)
	faqClient := faqcluster.NewClientWithHTTPClient(addr, server.Client())

	w := newClusteringWork(ss, ss, ss, faqClient)
	err := w(1, map[string]interface{}{}, []interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if err = writer.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
