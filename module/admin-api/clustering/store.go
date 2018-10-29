package clustering

import (
	"database/sql"
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util"
)

type dbStore struct{}

func getTx() (*sql.Tx, error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}
	return db.Begin()
}
func (store *dbStore) Store(cr *clusteringResult) error {

	tx, err := getTx()
	if err != nil {
		return err
	}
	defer tx.Commit()

	resultSQL := "insert into " + TableProps.clusterResult.name +
		" (" + TableProps.clusterResult.feedbackID + "," + TableProps.clusterResult.reportID + "," + TableProps.clusterResult.clusterID + ") " +
		" values(?,?,?)"

	tagSQL := "insert into " + TableProps.clusterTag.name +
		" (" + TableProps.clusterTag.reportID + "," + TableProps.clusterTag.clusteringID + "," + TableProps.clusterTag.tag + ") values (?,?,?)"

	resultStmt, err := tx.Prepare(resultSQL)
	if err != nil {
		return err
	}
	defer resultStmt.Close()

	tagStmt, err := tx.Prepare(tagSQL)
	if err != nil {
		return err
	}
	defer tagStmt.Close()

	for idx, cluster := range cr.clusters {
		for _, qID := range cluster.feedbackID {
			_, err := resultStmt.Exec(qID, cr.reportID, idx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("insert clustering result %v, report_id:%v failed\n %v", qID, cr.reportID, err)
			}
		}

		for _, tag := range cluster.tags {
			_, err := tagStmt.Exec(cr.reportID, idx, tag)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("insert clustering tag %v, report_id:%v failed\n %v", tag, cr.reportID, err)
			}
		}

	}

	return nil
}

type clusterTable struct {
	feedback      feedbackProps
	clusterTag    tagProps
	clusterResult resultProps
	report        reportProps
}

type tagProps struct {
	name         string
	id           string
	reportID     string
	clusteringID string
	tag          string
}

type resultProps struct {
	name       string
	id         string
	reportID   string
	feedbackID string
	clusterID  string
}

type reportProps struct {
	name        string
	id          string
	createdTime string
	startTime   string
	endTime     string
	status      string
	appid       string
	rType       string
}

type feedbackProps struct {
	name        string
	id          string
	question    string
	stdQuestion string
	createdTime string
	updatedTime string
	appid       string
	qType       string
}

//table properties name in database
var TableProps = clusterTable{
	feedback:      feedbackProps{name: "user_feedback", id: "id", question: "question", stdQuestion: "std_question", createdTime: "created_time", updatedTime: "updated_time", appid: "appid", qType: "type"},
	clusterTag:    tagProps{name: "clustering_tag", id: "id", reportID: "report_id", clusteringID: "clustering_id", tag: "tag"},
	clusterResult: resultProps{name: "clustering_result", id: "id", reportID: "unresolved_report_id", feedbackID: "feedback_id", clusterID: "cluster_id"},
	report:        reportProps{name: "unresolved_report", id: "id", createdTime: "created_time", startTime: "start_time", endTime: "end_time", status: "status", appid: "appid", rType: "type"},
}

type clusteringResult struct {
	numClustered int
	clusters     []clustering
	reportID     uint64
}

type clustering struct {
	feedbackID []uint64
	tags       []string
}

type rankerElm struct {
	idx     int
	avgDist float64
}

type ranker []rankerElm

func (r ranker) Len() int {
	return len(r)
}

func (r ranker) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ranker) Less(i, j int) bool {
	return r[i].avgDist < r[j].avgDist
}

//StoreCluster store the cluster result
type StoreCluster interface {
	Store(cr *clusteringResult) error
}

//column name of <appid>_question
const (
	NQuestionID         = "Question_Id"
	NContent            = "Content"
	QuestionTableFormat = "%s_question"
)

//parameter name
const (
	PType = "type"
)

func isValidType(pType int) bool {

	switch pType {
	case 0:
		fallthrough
	case 1:
		return true
	default:
		return false
	}
}
