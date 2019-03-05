package qi

import (
	"encoding/json"
	"fmt"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/pkg/logger"

	"github.com/satori/go.uuid"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"github.com/tealeg/xlsx"
	"encoding/hex"
	"strings"
)

var tagTypeDict = map[int8]string{
	0: "default",
	1: "keyword",
	2: "dialogue_act",
	3: "user_response",
}

// Tags is the service for getting the tags json response.
// If the query.Paging is nil, response.paging.Limit & paging will be 0, 0.
// If the query.Enterprise is nil, an error will be returned.
func Tags(query model.TagQuery) (resp *TagResponse, err error) {
	if query.Enterprise == nil {
		return nil, fmt.Errorf("query must contain enterpriseID")
	}
	counts, err := tagDao.CountTags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tag count from dao failed, %v", err)
	}
	tags, err := TagsByQuery(query)
	if err != nil {
		return nil, fmt.Errorf("call tags by query failed, %v", err)
	}
	resp = &TagResponse{
		Paging: general.Paging{
			Total: int64(counts),
		},
		Data: tags,
	}
	if query.Paging != nil {
		resp.Paging.Limit = query.Paging.Limit
		resp.Paging.Page = query.Paging.Page
	}
	return
}

// toTag transform model.Tag to the presentive tag for tag controller
func toTag(result ...model.Tag) ([]tag, error) {
	var err error
	var tags = make([]tag, 0, len(result))
	for _, t := range result {
		typ, found := tagTypeDict[t.Typ]
		if !found {
			typ = "default"
		}
		var posSentences, negSentences []string
		err = json.Unmarshal([]byte(t.PositiveSentence), &posSentences)
		if err != nil {
			return nil, fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", t.ID, err)
		}
		err = json.Unmarshal([]byte(t.NegativeSentence), &negSentences)
		if err != nil {
			return nil, fmt.Errorf("tag %d negative sentence payload is not a valid string array, %v", t.ID, err)
		}

		tags = append(tags, tag{
			TagUUID:      t.UUID,
			TagName:      t.Name,
			TagType:      typ,
			PosSentences: posSentences,
			NegSentences: negSentences,
		})
	}
	return tags, nil
}
func TagsByQuery(query model.TagQuery) ([]tag, error) {

	result, err := tagDao.Tags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tags from dao failed, %v", err)
	}

	return toTag(result...)
}

// NewTag create a tag from t.
// incremental id will be returned, if the dao supported it.
// If t is not valid(etc uuid or positive sentence is empty...) then an adminError will returned.
func NewTag(t model.Tag) (id uint64, err error) {

	if _, err := uuid.FromString(t.UUID); t.UUID == "" && err != nil {
		return 0, &controllerError{
			errNo: AdminErrors.ErrnoRequestError,
			error: fmt.Errorf("tag UUID '%s' format is not correct", t.UUID),
		}
	}
	var ps []string
	json.Unmarshal([]byte(t.PositiveSentence), &ps)
	// If positive sentence is empty, cu model can not be trained.
	if len(ps) < 1 {
		return 0, &controllerError{
			errNo: AdminErrors.ErrnoRequestError,
			error: fmt.Errorf("must have at least one positive tag"),
		}
	}
	input := []model.Tag{t}
	createdTags, err := tagDao.NewTags(nil, input)
	if err != nil {
		return 0, fmt.Errorf("create tag from dao failed, %v", err)
	}
	if len(createdTags) != len(input) {
		return 0, fmt.Errorf("unexpected dao internal error, %d have been returned instead of input %d", len(createdTags), len(input))
	}
	return createdTags[0].ID, nil
}

// UpdateTag update the origin t, since tag need to keep the origin value.
// update will try to delete the old one with the same uuid, and create new one.
// multiple update called on the same t will try it best to resolve to one state, but not guarantee success.
// if conflicted can not be resolved, id will be 0 and err will be nil.
func UpdateTag(t model.Tag) (id uint64, err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return 0, fmt.Errorf("dao init transaction failed, %v", err)
	}
	defer func() {
		tx.Rollback()
	}()

	tagsSentences, err := sentenceDao.GetRelSentenceIDByTagIDs(tx, []uint64{t.ID})
	if err != nil {
		return 0, fmt.Errorf("dao get sentence id failed, %v", err)
	}
	rowsCount, err := tagDao.DeleteTags(tx, model.TagQuery{
		ID: []uint64{t.ID},
	})
	if err != nil {
		return 0, fmt.Errorf("dao delete failed, %v", err)
	}
	if rowsCount != 1 {
		return 0, fmt.Errorf("delete tag failed, no affected rows")
	}
	t.UpdateTime = time.Now().Unix()
	tags, err := tagDao.NewTags(tx, []model.Tag{t})
	// TODO: support elegant handle for sql driver not support return incremental id.
	// if err == model.ErrAutoIDDisabled {
	//	tagDao.Tags()
	// 	tx.Commit()
	// }
	if err != nil {
		return 0, fmt.Errorf("dao create new tags failed, %v", err)
	}
	newTag := tags[0]

	sentences := tagsSentences[t.ID]
	// to avoid nil panic if tag have no sentences
	if sentences == nil {
		return newTag.ID, nil
	}

	sentenceGrp, err := sentenceDao.GetSentences(tx, &model.SentenceQuery{
		ID: sentences,
	})

	err = propagateUpdateFromSentence(sentenceGrp, newTag.ID, t.ID, t.Enterprise, tx)
	if err != nil {
		logger.Error.Printf("propage update  failed. %s", err)
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		logger.Error.Printf("commit transcation failed. %s", err)
		return 0, err
	}

	return newTag.ID, nil
}

// DeleteTag delete the tag by dao.
// all operation will be included in one transaction.
// Which ensure each uuid is delete or nothing is deleted.
func DeleteTag(uuid ...string) error {
	if len(uuid) == 0 {
		return nil
	}
	tx, err := dbLike.Begin()
	if err != nil {
		return fmt.Errorf("dao init transaction failed, %v", err)
	}
	defer func() {
		tx.Rollback()
	}()

	query := model.TagQuery{
		UUID:             uuid,
		IgnoreSoftDelete: false,
	}
	tags, err := tagDao.Tags(tx, query)
	if err != nil {
		return fmt.Errorf("dao get tags failed. %v", err)
	}

	affectedrows, err := tagDao.DeleteTags(tx, query)
	if err == model.ErrAutoIDDisabled {
		logger.Warn.Println("tag table does not support affectedrow, we will continue to work, but we can not detect conflict now.")
		tx.Commit()
		return nil
	}
	if err != nil {
		return fmt.Errorf("dao delete failed, %v", err)
	}
	if err != model.ErrAutoIDDisabled && int(affectedrows) != len(uuid) {
		return fmt.Errorf("dao delete should delete %d of rows, but only %d. possible conflict operation at the same time", len(uuid), affectedrows)
	}

	tagMap := map[uint64]bool{}
	tagID := []uint64{}
	var enterprise string
	for _, tag := range tags {
		tagMap[tag.ID] = true
		tagID = append(tagID, tag.ID)
		enterprise = tag.Enterprise
	}

	sentences, err := sentenceDao.GetRelSentenceIDByTagIDs(tx, tagID)
	if err != nil {
		return fmt.Errorf("get sentences failed. %v", err)
	}

	sentenceID := []uint64{}
	for _, v := range sentences {
		sentenceID = append(sentenceID, v...)
	}

	sq := &model.SentenceQuery{
		ID:         sentenceID,
		Enterprise: &enterprise,
	}

	sentenceGrp, err := sentenceDao.GetSentences(tx, sq)
	if err != nil {
		return err
	}

	for i := range sentenceGrp {
		s := sentenceGrp[i]
		if len(s.TagIDs) == 1 {
			s.TagIDs = []uint64{}
			continue
		}

		for j, tag := range s.TagIDs {
			if _, ok := tagMap[tag]; ok {
				if j == len(s.TagIDs)-1 {
					s.TagIDs = s.TagIDs[:j]
				} else {
					s.TagIDs = append(s.TagIDs[:j], s.TagIDs[j+1:]...)
				}
			}
		}
	}

	err = propagateUpdateFromSentence(sentenceGrp, 0, 0, enterprise, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func propagateUpdateFromSentence(sentences []*model.Sentence, newTag, oldTag uint64, enterprise string, tx model.SQLTx) (err error) {
	logger.Info.Printf("tags %+v\n", newTag)
	logger.Info.Printf("sentences: %+v\n", sentences)
	if len(sentences) == 0 {
		return
	}

	sUUID := []string{}
	sID := []int64{}
	activeSentences := []model.Sentence{}
	for i := range sentences {
		s := sentences[i]
		if s.IsDelete == 1 {
			continue
		}

		for j, tagID := range s.TagIDs {
			if tagID == oldTag {
				s.TagIDs[j] = newTag
			}
		}
		sID = append(sID, int64(s.ID))
		sUUID = append(sUUID, s.UUID)
		activeSentences = append(activeSentences, *s)
	}

	// delete old sentences
	var deleted int8
	sentenceQuery := &model.SentenceQuery{
		UUID:       sUUID,
		IsDelete:   &deleted,
		Enterprise: &enterprise,
	}

	logger.Info.Printf("sq: %+v\n", sentenceQuery)

	_, err = sentenceDao.SoftDeleteSentence(tx, sentenceQuery)
	if err != nil {
		logger.Error.Printf("delete sentence failed. %s", err)
		return
	}

	err = sentenceDao.InsertSentences(tx, activeSentences)
	if err != nil {
		logger.Error.Printf("insert sentences failed. %s", err)
		return
	}

	sgs, err := sentenceGroupDao.GetBySentenceID(sID, tx)
	if err != nil {
		logger.Error.Printf("get sentence groups failed. %s", err)
		return
	}

	return propagateUpdateFromSentenceGroup(sgs, activeSentences, tx)
}

func BatchAddTags(fileName string, enterpriseID string) error {

	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		logger.Error.Printf("can not open file %s \n", fileName)
		return err
	}

	err = batchAddTagsBy(xlFile, enterpriseID, "tag-keyword")
	if err != nil {
		logger.Error.Printf("fail to batch add tags by sheet %s \n", "tag-keyword")
		return err
	}
	err = batchAddTagsBy(xlFile, enterpriseID, "tag-intent")
	if err != nil {
		logger.Error.Printf("fail to batch add tag by sheet %s \n", "tag-intent")
		return err
	}

	// TODO need a check function
	// tag_name不能重复 同一个tag的语料不能重复 需要is_delete=0的
	// sentence 所属的category_id必须是存在的，sentence_name必须是不能重复的，所包含的tag必须是存在的(借鉴NewSentence方法中的tag检查方法) 借鉴handleNewSentence中的check相关的方法

	return nil
}

func BatchAddSentences(fileName string, enterpriseID string) error {

	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		logger.Error.Printf("can not open file %s \n", fileName)
		return err
	}

	err = batchAddSentencesBy(xlFile, enterpriseID, "sentence")

	// TODO need a check function
	// tag_name不能重复 同一个tag的语料不能重复 需要is_delete=0的
	// sentence 所属的category_id必须是存在的，sentence_name必须是不能重复的，所包含的tag必须是存在的(借鉴NewSentence方法中的tag检查方法) 借鉴handleNewSentence中的check相关的方法
	// 导入的语料中如果有重复的sentenceName，目前会以第一条为准

	return err
}

func BatchAddRules(fileName string, enterpriseID string) error {

	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		logger.Error.Printf("can not open file %s \n", fileName)
		return err
	}

	err = batchAddRulesBy(xlFile, enterpriseID, "rule")

	// TODO need a check function
	// [][WARN ] 2019/03/05 09:01:37 tags_services.go:472: more than one needed sentence: 回访-分红险责任告知-询问是否清楚
	// 如果系统原先错误操作导致more than one sentence会存在隐患

	return err
}

func batchAddRulesBy(xlFile *xlsx.File, enterpriseID string, sheetName string) error {

	sheet, ok := xlFile.Sheet[sheetName]

	if !ok {
		logger.Error.Printf("can not get sheet %s \n", sheetName)
		return fmt.Errorf("can not get sheet %s \n", sheetName)
	}

	var ruleName, description, logicList string

	tx, err := dbLike.Begin()
	if err != nil {
		return err
	}

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		// TODO check value
		ruleName = row.Cells[3].String()
		description = row.Cells[4].String()
		logicList = row.Cells[5].String()

		// check if rule exists
		filter := &model.ConversationRuleFilter{
			Enterprise: enterpriseID,
			Severity:   -1,
			Name:       ruleName,
			IsDeleted:  int8(0),
		}

		total, err := conversationRuleDao.CountBy(filter, tx)
		if err != nil {
			return err
		}

		if total > 0 {
			logger.Trace.Printf("found exist rule: %s \n", ruleName)
			continue
		}

		var createdSentenceGroups []*model.SentenceGroup

		splits := strings.Split(logicList, "|")
		flag := int8(0)
		for _, split := range splits {
			// find sentence uuid according to sentence
			query := &model.SentenceQuery{
				Enterprise: &enterpriseID,
				IsDelete:   &flag,
				Name:       &split,
			}

			count, err := sentenceDao.CountSentences(nil, query)
			if err != nil {
				return err
			}

			if count == 0 {
				logger.Trace.Printf("invalid sentence: %s \n", split)
				continue
			}
			if count > 1 {
				logger.Warn.Printf("more than one needed sentence: %s \n", split)
			}

			data, err := getSentences(query)
			if err != nil {
				return err
			}
			sentence := data[0]
			sentenceUUID := sentence.UUID

			// ------------------
			// add sentence-group
			// ------------------
			sentenceGroup := &model.SentenceGroup{
				Name:       "",
				Enterprise: enterpriseID,
			}
			// add sentence to sentenceGroup
			sentences := make([]model.SimpleSentence, 0)
			sentences = append(sentences, model.SimpleSentence{
				ID:   sentence.ID,
				UUID: sentenceUUID,
				Name: sentence.Name,
			})
			sentenceGroup.Sentences = sentences

			// default role is staff
			if roleCode, ok := roleMapping["staff"]; ok {
				sentenceGroup.Role = roleCode
			} else {
				sentenceGroup.Role = -1
			}

			// default position
			if positionCode, ok := positionMap[""]; ok {
				sentenceGroup.Position = positionCode
			} else {
				sentenceGroup.Position = -1
			}

			// default type
			sentenceGroup.Type = 0
			// default optional
			sentenceGroup.Optional = 0
			// default positionSentence
			sentenceGroup.Distance = 0

			createdSentenceGroup, err := CreateSentenceGroup(sentenceGroup)

			createdSentenceGroups = append(createdSentenceGroups, createdSentenceGroup)
		}

		// ------------------
		// add conversation-flow
		// ------------------
		cfUUID, err := uuid.NewV4()
		if err != nil {
			return err
		}
		cfUUIDStr := strings.Replace(cfUUID.String(), "-", "", -1)

		conversationFlow := &model.ConversationFlow{
			UUID:       cfUUIDStr,
			Name:       ruleName + "-dialog1",
			Enterprise: enterpriseID,
			Min:        1,
		}
		var cfExpression string
		sentenceGroups := make([]model.SimpleSentenceGroup, len(createdSentenceGroups))
		for i, item := range createdSentenceGroups {
			sentenceGroups[i] = model.SimpleSentenceGroup{
				ID:   item.ID,
				UUID: item.UUID,
				Name: item.Name,
			}

			if i == 0 {
				cfExpression = "must " + item.UUID
			} else {
				// default: must and
				cfExpression = cfExpression + " and " + item.UUID
			}
		}

		conversationFlow.Expression = cfExpression
		conversationFlow.SentenceGroups = sentenceGroups
		now := time.Now().Unix()
		conversationFlow.CreateTime = now
		conversationFlow.UpdateTime = now

		createdConversationFlow, err := conversationFlowDao.Create(conversationFlow, tx)
		if err != nil {

		}

		// ------------------
		// add rule
		// ------------------

		// default min, default max, default score, default enterprise
		rule := &model.ConversationRule{
			Name:        ruleName,
			Min:         1,
			Max:         0,
			Score:       -5,
			Description: description,
			Enterprise:  enterpriseID,
		}

		// default severity is normal
		serverity := int8(0)

		// default method is positive
		method := int8(1)

		rule.Severity = serverity
		rule.Method = method

		// a default conversation-flow
		flows := make([]model.SimpleConversationFlow, 1)
		flows[0] = model.SimpleConversationFlow{
			ID:   createdConversationFlow.ID,
			UUID: createdConversationFlow.UUID,
			Name: createdConversationFlow.Name,
		}

		rule.Flows = flows

		ruleUUID, err := uuid.NewV4()
		if err != nil {
			return err
		}

		rule.UUID = ruleUUID.String()
		rule.UUID = strings.Replace(rule.UUID, "-", "", -1)

		now = time.Now().Unix()
		rule.CreateTime = now
		rule.UpdateTime = now

		createdConversationRule, err := conversationRuleDao.Create(rule, tx)
		if err != nil {
			logger.Error.Printf("error occurred, when create rule %s \n", ruleName)
			return err
		}

		logger.Trace.Printf("create rule: %s \n", createdConversationRule.Name)
	}

	dbLike.Commit(tx)
	defer dbLike.ClearTransition(tx)

	return nil
}

func batchAddSentencesBy(xlFile *xlsx.File, enterpriseID string, sheetName string) error {

	sheet, ok := xlFile.Sheet[sheetName]
	if !ok {
		logger.Error.Printf("can not get sheet %s \n", sheetName)
		return fmt.Errorf("can not get sheet %s \n", sheetName)
	}

	var name, content string

	// default category
	var categoryID uint64
	categoryID = 0

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		name = row.Cells[0].String()
		content = row.Cells[1].String()

		flag := int8(0)
		query := &model.SentenceQuery{
			Enterprise: &enterpriseID,
			IsDelete:   &flag,
			Name:       &name,
		}

		count, err := sentenceDao.CountSentences(nil, query)
		if err != nil {
			return err
		}

		if count > 0 {
			logger.Trace.Printf("found exist sentence: %s \n", name)
			continue
		}

		splits := strings.Split(content, "+")

		var uuidArr []string

		// get uuid according to tag name
		for _, item := range splits {

			query := &model.TagQuery{
				Enterprise: &enterpriseID,
				Name:       &item,
			}

			resp, err := Tags(*query)
			if err != nil {
				logger.Error.Printf("error occurred, when query tag %s \n", name)
				return err
			}

			if resp.Paging.Total == 0 {
				// TODO should report error
				logger.Error.Printf("the sentence %s need tag %s \n", name, item)
				continue
			}
			uuidArr = append(uuidArr, resp.Data[0].TagUUID)
		}

		if len(uuidArr) == 0 {
			// TODO should report error
			logger.Error.Printf("the sentence %s not found needed tag \n", name)
			continue
		}

		_, err = NewSentence(enterpriseID, categoryID, name, uuidArr)

		if err != nil {
			logger.Error.Printf("fail to insert sentence %s \n", name)
			return err
		}

		logger.Trace.Printf("create sentence: %s \n", name)
	}

	return nil
}

func batchAddTagsBy(xlFile *xlsx.File, enterpriseID string, sheetName string) error {

	sheet, ok := xlFile.Sheet[sheetName]

	if !ok {
		logger.Error.Printf("can not get sheet %s \n", sheetName)
		return fmt.Errorf("can not get sheet %s \n", sheetName)
	}

	var name, content string
	tagMap := make(map[string][]string)
	var tagType int8
	if sheetName == "tag-keyword" {
		tagType = 1
	} else {
		tagType = 2
	}

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		// TODO check value
		name = row.Cells[0].String()
		content = row.Cells[1].String()

		if corpus, ok := tagMap[name]; ok {
			tagMap[name] = append(corpus, content)
		} else {
			tagMap[name] = []string{content}
		}
	}

	for name, corpus := range tagMap {
		query := model.TagQuery{
			Enterprise: &enterpriseID,
			Name:       &name,
		}

		resp, err := Tags(query)
		if err != nil {
			logger.Error.Printf("error occurred, when query tag %s \n", name)
			return err
		}

		if resp.Paging.Total > 0 {
			logger.Trace.Printf("found exist tag: %s \n", name)
			continue
		}

		posSentences, _ := json.Marshal(corpus)
		negSentences, _ := json.Marshal([]string{})

		current := time.Now().Unix()
		tagUUID, err := uuid.NewV4()
		if err != nil {
			return err
		}

		tag := model.Tag{
			Enterprise:       enterpriseID,
			Name:             name,
			Typ:              tagType,
			PositiveSentence: string(posSentences),
			NegativeSentence: string(negSentences),
			CreateTime:       current,
			UpdateTime:       current,
			UUID:             hex.EncodeToString(tagUUID[:]),
		}

		_, err = NewTag(tag)

		if err != nil {
			return fmt.Errorf("db error")
		}
		logger.Trace.Printf("create tag: %s \n", name)
	}
	return nil
}

func BatchAddFlows(fileName string, enterpriseID string) error {

	xlFile, err := xlsx.OpenFile(fileName)

	if err != nil {
		logger.Error.Printf("can not open file %s \n", fileName)
		return err
	}

	// TODO intent_name is not unique
	//flag := 0
	//query := &model.NavQuery{Enterprise: &enterpriseID, IsDelete: &flag}

	// add tag
	if err = BatchAddTags(fileName, enterpriseID); err != nil {
		logger.Error.Println("fail to add tag when execute BatchAddFlows")
		return err
	}

	// add sentence
	if err = BatchAddSentences(fileName, enterpriseID); err != nil {
		logger.Error.Println("fail to add sentence when execute BatchAddFlows")
		return err
	}

	flowSheetName := "rule"
	sheet, ok := xlFile.Sheet[flowSheetName]

	if !ok {
		logger.Error.Printf("can not get sheet %s \n", flowSheetName)
		return fmt.Errorf("can not get sheet %s \n", flowSheetName)
	}

	var name, logicList string

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		// TODO check value
		name = row.Cells[3].String()
		logicList = row.Cells[5].String()

		// default flow type
		flow := reqNewFlow{
			Name:       name,
			IntentName: name,
			Type:       "intent",
		}

		flowID, err := NewFlow(&flow, enterpriseID)
		logger.Trace.Printf("create flow: %s \n", name)

		if err != nil {
			logger.Error.Printf("fail to create flow: %s\n", err)
			return err
		}

		// ------------------
		// add node
		// ------------------

		// TODO only one logic in logicList ?
		sentenceGroup := &model.SentenceGroup{
			Name:       name + "-node1",
			Enterprise: enterpriseID,
		}

		var dataSentences []*DataSentence
		flag := int8(0)
		// get sentence according to logic_list, a node may contain more than one sentence
		splits := strings.Split(logicList, "|")
		for _, item := range splits {
			query := &model.SentenceQuery{
				Enterprise: &enterpriseID,
				IsDelete:   &flag,
				Name:       &item,
			}

			count, err := sentenceDao.CountSentences(nil, query)
			if err != nil {
				return err
			}

			if count == 0 {
				logger.Trace.Printf("invalid sentence: %s \n", item)
				continue
			}
			if count > 1 {
				logger.Warn.Printf("more than one needed sentence: %s \n", item)
			}

			data, err := getSentences(query)
			if err != nil {
				return err
			}
			// default get the first one
			sentence := data[0]

			dataSentences = append(dataSentences, sentence)
		}

		sentences := make([]model.SimpleSentence, 0)
		for _, item := range dataSentences {
			sentences = append(sentences, model.SimpleSentence{
				ID:   item.ID,
				UUID: item.UUID,
				Name: item.Name,
			})
		}

		sentenceGroup.Sentences = sentences

		// default role: any
		sentenceGroup.Role = 2
		if roleCode, ok := roleMapping["any"]; ok {
			sentenceGroup.Role = roleCode
		} else {
			sentenceGroup.Position = -1
		}

		// default position: 2
		if positionCode, ok := positionMap[""]; ok {
			sentenceGroup.Position = positionCode
		} else {
			sentenceGroup.Position = -1
		}

		// default type: call-in
		if typeCode, ok := typeMapping["call_in"]; ok {
			sentenceGroup.Type = typeCode
		} else {
			sentenceGroup.Type = -1
		}

		// default optional: 0
		sentenceGroup.Optional = 0

		// default range: 0
		sentenceGroup.Distance = 0

		err = NewNode(flowID, sentenceGroup)

		if err != nil {
			logger.Error.Println("fail to create node")
			return err
		}
		logger.Trace.Printf("create node: %s \n", name+"-node1")
	}
	return nil
}
