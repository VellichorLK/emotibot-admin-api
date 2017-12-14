package FAQ
import (
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func updateSimilarQuestions(qid string, appid string, user string, sqs []SimilarQuestion) error {
	// delete old similar questions
	deleteSimilarQuestionsByQuestionId(qid, appid)

	// put new similar questions
	insertSimilarQuestions(qid, appid, user, sqs)

	// notify multicustomer TODO: update consul directly
	util.McManualBusiness(appid)

	// save audit log

	return nil
}

func deleteSimilarQuestions(qid string) error {
	return nil
}