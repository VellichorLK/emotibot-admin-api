package FAQ

import (
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func updateSimilarQuestions(qid int, appid string, user string, sqs []SimilarQuestion) error {
	var err error
	db := util.GetMainDB()
	t, err := db.Begin()
	if err != nil {
		return fmt.Errorf("can't aquire transaction lock, %s", err)
	}
	defer t.Commit()
	// delete old similar questions
	if err = deleteSimilarQuestionsByQuestionID(t, qid, appid); err != nil {
		t.Rollback()
		return fmt.Errorf("delete operation failed, %s", err)
	}

	// put new similar questions
	if err = insertSimilarQuestions(t, qid, appid, user, sqs); err != nil {
		t.Rollback()
		return fmt.Errorf("insert operation failed, %s", err)
	}
	t.Commit()

	// notify multicustomer TODO: update consul directly
	if _, err = util.McManualBusiness(appid); err != nil {
		return fmt.Errorf("error in requesting to MultiCustomer module: %s", err)
	}

	return nil
}

func deleteSimilarQuestions(qid string) error {
	return nil
}
