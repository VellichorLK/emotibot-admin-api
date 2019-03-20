package qi

import (
	"emotibot.com/emotigo/module/qic-api/util/general"
	"time"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"fmt"
	"emotibot.com/emotigo/pkg/logger"
	"sort"
)

var (
	sentenceTestDao model.SentenceTestDao
)

func PredictSentences(enterpriseID string) error {
	deleted := int8(0)
	query := &model.SentenceQuery{
		Enterprise: &enterpriseID,
		IsDelete:   &deleted,
	}
	count, err := sentenceDao.CountSentences(nil, query)
	if err != nil {
		logger.Error.Println(err.Error())
		return err
	}
	if count == 0 {
		return fmt.Errorf("the num of sentence is 0 \n")
	}

	// TODO use go routine
	dataSens, err := getSentences(query)

	for _, sen := range dataSens {
		testSenQuery := &model.TestSentenceQuery{
			Enterprise: &enterpriseID,
			IsDelete:   &deleted,
			TestedID:   []uint64{sen.ID},
		}
		totalTestSens, err := sentenceTestDao.CountTestSentences(nil, testSenQuery)

		if totalTestSens == 0 {
			continue
		}
		logger.Trace.Printf("predict %s \n", sen.Name)

		err = predictSentence(enterpriseID, sen.ID)
		if err != nil {
			return err
		}
	}

	logger.Trace.Println("predict done")
	return nil
}

func predictSentence(enterpriseID string, testedID uint64) error {
	deleted := int8(0)
	query := &model.SentenceQuery{
		ID:         []uint64{testedID},
		Enterprise: &enterpriseID,
		IsDelete:   &deleted,
	}
	sentences, err := sentenceDao.GetSentences(nil, query)
	if len(sentences) == 0 {
		logger.Error.Println("failed to find sentence")
		return fmt.Errorf("failed to find one sentence")
	}
	if len(sentences) > 1 {
		logger.Error.Println("find more than one sentence")
		return fmt.Errorf("find more than one sentence")
	}
	// only need one
	sentence := sentences[0]

	allTagIDs := make([]uint64, 0)
	allTagIDs = append(allTagIDs, sentence.TagIDs...)

	expectedTagID := make(map[uint64]bool)
	for _, tagID := range allTagIDs {
		expectedTagID[tagID] = true
	}

	models, err := GetUsingModelByEnterprise(enterpriseID)
	if err != nil {
		return err
	}
	if len(models) > 1 {
		logger.Warn.Printf("more than 1 model")
	}

	testSentences, err := GetTestSentences(enterpriseID, testedID)
	if err != nil {
		return err
	}
	if len(testSentences) == 0 {
		return fmt.Errorf("fail to find testSentence")
	}
	testSenName := make([]string, 0)
	for _, item := range testSentences {
		testSenName = append(testSenName, item.Name)
	}

	// use the first model
	matched, err := TagMatch([]uint64{models[0].ID}, testSenName, 3*time.Second)

	sentenceTestResult := model.SentenceTestResult{
		Name:       sentence.Name,
		Total:      len(allTagIDs),
		Enterprise: enterpriseID,
		CategoryID: sentence.CategoryID,
		OrgID:      testedID,
	}
	for _, matchedData := range matched {
		result := TestResult{
			Index:     matchedData.Index,
			Name:      testSenName[matchedData.Index-1],
			HitTags:   make([]uint64, 0),
			FailTags:  make([]uint64, 0),
			MatchText: make([]string, 0),
		}

		total := len(expectedTagID)
		hit := 0
		for tagID := range matchedData.Matched {
			// the key is tag id
			if _, ok := expectedTagID[tagID]; ok {
				hit++
				result.HitTags = append(result.HitTags, tagID)
				result.MatchText = append(result.MatchText, matchedData.Matched[tagID].MatchText)
			}
		}

		for k := range expectedTagID {
			isSelected := false
			for _, item := range result.HitTags {
				if k == item {
					isSelected = true
					break
				}
			}

			if !isSelected {
				result.FailTags = append(result.FailTags, k)
			}
		}

		result.Accuracy = float32(hit) / float32(total)

		testSentence := model.TestSentence{
			ID:        testSentences[matchedData.Index-1].ID,
			HitTags:   result.HitTags,
			FailTags:  result.FailTags,
			MatchText: result.MatchText,
		}
		err = updateSentenceTest(&testSentence)
		if err != nil {
			return err
		}

		sentenceTestResult.Hit = hit
		sentenceTestResult.Accuracy = float32(hit) / float32(total)
	}

	err = insertSentenceTestResult(&sentenceTestResult)
	if err != nil {
		return err
	}
	return nil
}

func GetTestSentences(enterpriseID string, testedID uint64) ([]*model.TestSentence, error) {
	deleted := int8(0)
	query := &model.TestSentenceQuery{
		Enterprise: &enterpriseID,
		TestedID:   []uint64{testedID},
		IsDelete:   &deleted,
	}
	testSentences, err := sentenceTestDao.GetTestSentences(nil, query)
	if err != nil {
		logger.Error.Printf("fail to get sentences. %s \n", err.Error())
		return nil, err
	}
	return testSentences, err
}

func NewTestSentence(enterpriseID string, testSenReq *TestSentenceReq) (*string, error) {
	uuidStr, err := general.UUID()
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()

	testSentence := model.TestSentence{
		Name:       testSenReq.Name,
		IsDelete:   0,
		Enterprise: enterpriseID,
		UUID:       uuidStr,
		CreateTime: now,
		UpdateTime: now,
		TestedID:   testSenReq.TestedID,
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err = sentenceTestDao.InsertTestSentence(tx, &testSentence); err != nil {
		return nil, err
	}

	dbLike.Commit(tx)
	return &uuidStr, nil
}

func SoftDeleteTestSentence(enterpriseID string, testSenUUID string) (int64, error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return 0, nil
	}
	defer dbLike.ClearTransition(tx)

	deleted := int8(0)
	query := &model.TestSentenceQuery{
		UUID:       []string{testSenUUID},
		Enterprise: &enterpriseID,
		IsDelete:   &deleted,
	}
	testSentences, err := sentenceTestDao.GetTestSentences(tx, query)
	if err != nil {
		logger.Error.Printf("fail to get sentences. %s \n", err.Error())
		return 0, err
	}
	if len(testSentences) == 0 {
		return 0, fmt.Errorf("fail to get sentnece")
	}

	affected, err := sentenceTestDao.SoftDeleteTestSentence(tx, query)
	if err != nil {
		return affected, err
	}
	if affected == 0 {
		return 0, nil
	}

	err = dbLike.Commit(tx)
	return affected, err
}

func GetSentenceTestResult(enterpriseID string, categoryID *uint64) ([]*model.SentenceTestResult, error) {
	query := &model.SentenceTestResultQuery{Enterprise: &enterpriseID, CategoryID: categoryID}
	testResults, err := sentenceTestDao.GetSentenceTestResult(nil, query)
	if err != nil {
		return nil, err
	}
	return testResults, nil
}

func insertSentenceTestResult(result *model.SentenceTestResult) error {
	uuidStr, err := general.UUID()
	if err != nil {
		return err
	}
	now := time.Now().Unix()

	result.UUID = uuidStr
	result.CreateTime = now
	result.UpdateTime = now

	tx, err := dbLike.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = sentenceTestDao.InsertOrUpdateSentenceTestResult(tx, result)
	if err != nil {
		return err
	}

	dbLike.Commit(tx)
	return err
}

func updateSentenceTest(testSentence *model.TestSentence) error {
	tx, err := dbLike.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = sentenceTestDao.UpdateSentenceTest(tx, testSentence)
	if err != nil {
		return err
	}
	dbLike.Commit(tx)
	return nil
}

func GetSentenceTestDetail(enterpriseID string, sentenceID uint64) (*SentenceTestDetail, error) {
	deleted := int8(0)

	senQuery := &model.SentenceQuery{ID: []uint64{sentenceID}, IsDelete: &deleted, Enterprise: &enterpriseID}

	sentences, err := sentenceDao.GetSentences(nil, senQuery)
	if len(sentences) == 0 {
		logger.Error.Println("failed to find sentence")
		return nil, fmt.Errorf("failed to find sentence")
	}
	if len(sentences) > 1 {
		logger.Error.Println("find more than one sentence")
		return nil, fmt.Errorf("find more than one sentence")
	}
	sentence := sentences[0]

	allTagIDs := make([]uint64, 0)
	allTagIDs = append(allTagIDs, sentence.TagIDs...)
	// get tags information
	tagQuery := model.TagQuery{ID: allTagIDs, Enterprise: &enterpriseID}

	tags, err := tagDao.Tags(nil, tagQuery)
	if err != nil {
		return nil, err
	}
	// transform tag data to map[tag_id] tag
	tagsIDMap := make(map[uint64]*model.Tag)
	// store index
	tagsIndexMap := make(map[uint64]int)
	for i := 0; i < len(tags); i++ {
		tagsIDMap[tags[i].ID] = &tags[i]
	}

	tagItems := make([]TagItem, 0)
	for i, tagID := range sentence.TagIDs {
		if tag, ok := tagsIDMap[tagID]; ok {
			item := TagItem{
				ID:   tag.ID,
				Name: tag.Name,
				Type: tagTypeDict[tag.Typ],
			}
			tagItems = append(tagItems, item)
		} else {
			return nil, fmt.Errorf("failed to convert tag info")
		}
		tagsIndexMap[tagID] = i + 1
	}

	sentenceItem := SentenceItem{
		ID:         sentence.ID,
		Name:       sentence.Name,
		CategoryID: sentence.CategoryID,
		UUID:       sentence.UUID,
		Tags:       tagItems,
	}

	query := &model.TestSentenceQuery{
		Enterprise: &enterpriseID,
		IsDelete:   &deleted,
		TestedID:   []uint64{sentenceID},
	}

	// get testSentence according to testedID
	testSentences, err := sentenceTestDao.GetTestSentences(nil, query)
	if err != nil {
		logger.Error.Printf("fail to get sentences. %s \n", err.Error())
		return nil, err
	}

	falseNum := 0
	testSentenceItems := make([]TestSentenceItem, 0)
	for _, testSentence := range testSentences {
		hitTagIndices := make([]int, 0)
		for _, id := range testSentence.HitTags {
			hitTagIndices = append(hitTagIndices, tagsIndexMap[id])
		}
		//sort.Slice(hitTagIndices, func(i, j int) bool {
		//	return hitTagIndices[i] < hitTagIndices[j]
		//})
		failTagIndices := make([]int, 0)
		for _, id := range testSentence.FailTags {
			failTagIndices = append(failTagIndices, tagsIndexMap[id])
		}
		//sort.Slice(failTagIndices, func(i, j int) bool {
		//	return failTagIndices[i] < failTagIndices[j]
		//})

		testSentenceItem := TestSentenceItem{
			ID:             testSentence.ID,
			Name:           testSentence.Name,
			HitTagIndices:  hitTagIndices,
			FailTagIndices: failTagIndices,
			MatchText:      testSentence.MatchText,
		}
		if len(failTagIndices) == 0 {
			testSentenceItem.IsCorrect = 1
		} else {
			testSentenceItem.IsCorrect = 0
			falseNum += 1
		}
		testSentenceItems = append(testSentenceItems, testSentenceItem)
	}

	detail := &SentenceTestDetail{
		Sentence:      sentenceItem,
		TrueNum:       len(testSentences) - falseNum,
		FalseNum:      falseNum,
		TestSentences: testSentenceItems,
	}

	return detail, err
}

func GetSentenceTestOverview(enterpriseID string) (*SentenceTestOverview, error) {
	deleted := int8(0)
	testSenQuery := &model.TestSentenceQuery{
		Enterprise: &enterpriseID,
		IsDelete:   &deleted,
	}
	totalTestSens, err := sentenceTestDao.CountTestSentences(nil, testSenQuery)
	if err != nil {
		return nil, err
	}

	senTesSummary := SenTestSummary{
		Accuracy: 0,
		Count:    totalTestSens,
		Account:  "hardcode_account",
		Date:     "2019/03/19",
	}

	senTestOverview := &SentenceTestOverview{
		Summary:    senTesSummary,
		Categories: make([]SenTestByCategory, 0),
	}

	// get all categories, type 0 indicates sentence
	isDelete := int8(0)
	ctype := int8(0)
	query := &model.CategoryQuery{
		Enterprise: &enterpriseID,
		IsDelete:   &isDelete,
		Type:       &ctype,
	}
	categories, err := GetCategories(query)
	if err != nil {
		logger.Error.Printf("error while get categories in handleGetCategories, reason: %s \n", err.Error())
		return nil, fmt.Errorf("error while get categories in handleGetCategories, reason: %s", err.Error())
	}

	// add "全部" category
	senTestOverview.Categories = append(senTestOverview.Categories, SenTestByCategory{
		ID:       -1,
		Name:     "全部",
		Accuracy: 0,
	})

	// default category id is 0
	categories = append(categories, &model.CategortInfo{
		ID:   0,
		Name: "未分类",
	})
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].ID < categories[j].ID
	})

	totalAcc := float32(0)
	totalNum := 0
	var updateTime int64
	for _, category := range categories {
		resultQuery := &model.SentenceTestResultQuery{Enterprise: &enterpriseID, CategoryID: &category.ID}
		testResults, err := sentenceTestDao.GetSentenceTestResult(nil, resultQuery)
		if err != nil {
			return nil, err
		}

		senTestByCategory := SenTestByCategory{
			ID:       int64(category.ID),
			Name:     category.Name,
			Accuracy: 0,
		}

		if len(testResults) == 0 {
			senTestOverview.Categories = append(senTestOverview.Categories, senTestByCategory)
			continue
		}

		// calculate accuracy with respect to each category
		var acc = float32(0)
		for _, testResult := range testResults {
			acc += testResult.Accuracy
			updateTime = testResult.UpdateTime
		}
		totalAcc += acc
		totalNum += len(testResults)

		acc = acc / float32(len(testResults))
		senTestByCategory.Accuracy = acc

		senTestOverview.Categories = append(senTestOverview.Categories, senTestByCategory)
	}

	totalAcc = totalAcc / float32(totalNum)
	senTestOverview.Summary.Accuracy = totalAcc
	senTestOverview.Summary.Date = time.Unix(updateTime, 0).Format("2006/01/02")
	// set accuracy for "全部" category
	senTestOverview.Categories[0].Accuracy = totalAcc

	return senTestOverview, nil
}

type SenTestSummary struct {
	Accuracy float32 `json:"accuracy"`
	Count    int64   `json:"count"`
	Account  string  `json:"account"`
	Date     string  `json:"date"`
}

type SenTestByCategory struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Accuracy float32 `json:"accuracy"`
}

type SentenceTestOverview struct {
	Summary    SenTestSummary      `json:"summary"`
	Categories []SenTestByCategory `json:"categories"`
}

type TagItem struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type SentenceItem struct {
	ID         uint64    `json:"id"`
	Name       string    `json:"name"`
	CategoryID uint64    `json:"category_id"`
	UUID       string    `json:"uuid"`
	Tags       []TagItem `json:"tags"`
}

type TestSentenceItem struct {
	ID             uint64   `json:"id"`
	Name           string   `json:"name"`
	IsCorrect      int      `json:"is_correct"`
	HitTagIndices  []int    `json:"hit_tag_indices"`
	FailTagIndices []int    `json:"fail_tag_indices"`
	MatchText      []string `json:"match_text"`
}

type SentenceTestDetail struct {
	Sentence      SentenceItem       `json:"sentence"`
	TrueNum       int                `json:"true_num"`
	FalseNum      int                `json:"false_num"`
	TestSentences []TestSentenceItem `json:"test_sentences"`
}
