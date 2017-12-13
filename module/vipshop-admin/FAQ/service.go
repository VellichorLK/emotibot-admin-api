package FAQ

func updateSimilarQuestions(qid string, appid string, user string, sqs []SimilarQuestion) error {
	// delete old similar questions
	deleteSimilarQuestionsByQuestionId(qid, appid)

	// put new similar questions
	insertSimilarQuestions(qid, appid, user, sqs)

	// save audit log

	return nil
}

func deleteSimilarQuestions(qid string) error {
	return nil
}