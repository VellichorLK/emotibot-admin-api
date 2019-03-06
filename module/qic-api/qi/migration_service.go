package qi

import (
	"emotibot.com/emotigo/pkg/logger"
	"github.com/tealeg/xlsx"
	"fmt"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"strings"
	"github.com/satori/go.uuid"
	"time"
	"encoding/json"
	"encoding/hex"
	"bytes"
)

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

	var flowName, logicList string

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		// TODO check value
		flowName = row.Cells[3].String()
		logicList = row.Cells[5].String()

		// default flow type
		flow := reqNewFlow{
			Name:       flowName,
			IntentName: flowName,
			Type:       "intent",
		}

		flowID, err := NewFlow(&flow, enterpriseID)
		logger.Trace.Printf("create flow: %s \n", flowName)

		if err != nil {
			logger.Error.Printf("fail to create flow: %s\n", err)
			return err
		}

		// ------------------
		// add intent or node
		// ------------------

		//var dataSentences []*DataSentence
		flag := int8(0)
		// get sentence according to logic_list, a node may contain more than one sentence
		splits := strings.Split(logicList, "|")
		for i, item := range splits {

			sentenceGroup := &model.SentenceGroup{
				Name:       item,
				Enterprise: enterpriseID,
			}

			// TODO check sentenceGroup ?

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

			// add sentence to sentenceGroup
			sentences := make([]model.SimpleSentence, 0)
			sentences = append(sentences, model.SimpleSentence{
				ID:   sentence.ID,
				UUID: sentence.UUID,
				Name: sentence.Name,
			})
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

			// the first sentence is intent, the rest is node
			// add intent SentenceGroup
			if i == 0 {
				flow := &model.NavFlowUpdate{}
				ignore := 0
				flow.IgnoreIntent = &ignore

				createdSentenceGroup, err := CreateSentenceGroup(sentenceGroup)
				if err != nil {
					logger.Error.Printf("error while create sentence in handleCreateSentenceGroup, reason: %s\n", err.Error())
					return err
				}
				flow.IntentLinkID = &createdSentenceGroup.ID
				flow.IntentName = &flowName

				if _, err = UpdateFlow(flowID, enterpriseID, flow); err != nil {
					logger.Error.Println("fail to update flow")
					return err
				}
				logger.Trace.Printf("create intent node %s \n", item)
				continue
			}

			// add node SentenceGroup
			err = NewNode(flowID, sentenceGroup)

			if err != nil {
				logger.Error.Println("fail to create node")
				return err
			}
			logger.Trace.Printf("create node: %s \n", item)
		}
	}
	return nil
}

func ExportGroups() (*bytes.Buffer, error) {
	sqlConn := dbLike.Conn()
	return serviceDAO.ExportGroups(sqlConn)
}

func ImportGroups(fileName string) (error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return err
	}
	defer dbLike.ClearTransition(tx)

	err = serviceDAO.ImportGroups(tx, fileName)
	if err != nil {
		return err
	}

	err = dbLike.Commit(tx)
	return err
}

func ExportCalls() (*bytes.Buffer, error) {
	sqlConn := dbLike.Conn()
	return callDao.ExportCalls(sqlConn)
}
