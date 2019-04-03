package qi

import (
	"emotibot.com/emotigo/pkg/logger"
	"github.com/tealeg/xlsx"
	"fmt"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"strings"
	"time"
	"encoding/json"
	"bytes"
	"emotibot.com/emotigo/module/qic-api/util/general"
)

const (
	RuleSheetName       = "rule"
	SentenceSheetName   = "sentence"
	TagKeywordSheetName = "tag-keyword"
	TagIntentSheetName  = "tag-intent"

	PositiveCorpusType = "positive"
	NegativeCorpusType = "negative"

	NavigationFlowIntentType = "intent"

	FromRule = "rule"
	FromFlow = "flow"
)

func BatchAddTags(fileName string, enterpriseID string) error {
	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		logger.Error.Printf("can not open file %s \n", fileName)
		return err
	}

	err = batchAddTagsBy(xlFile, enterpriseID, TagKeywordSheetName)
	if err != nil {
		logger.Error.Printf("fail to batch add tags by sheet %s \n", "tag-keyword")
		return err
	}
	err = batchAddTagsBy(xlFile, enterpriseID, TagIntentSheetName)
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
		logger.Error.Printf("Failed to open file %s \n", fileName)
		return err
	}

	err = batchAddSentencesBy(xlFile, enterpriseID, SentenceSheetName)

	// TODO need a check function
	// tag_name不能重复 同一个tag的语料不能重复 需要is_delete=0的
	// sentence 所属的category_id必须是存在的，sentence_name必须是不能重复的，所包含的tag必须是存在的(借鉴NewSentence方法中的tag检查方法) 借鉴handleNewSentence中的check相关的方法
	// 导入的语料中如果有重复的sentenceName，目前会以第一条为准

	return err
}

func PrepareSentenceGroups(fileName string, enterpriseID string, from string) (map[string]*model.SentenceGroup, error) {
	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		logger.Error.Printf("Failed to open file %s \n", fileName)
		return nil, err
	}
	sheet, ok := xlFile.Sheet[SentenceSheetName]
	if !ok {
		logger.Error.Printf("Failed to get sheet %s \n", SentenceSheetName)
		return nil, fmt.Errorf("failed to get sheet %s \n", SentenceSheetName)
	}

	// key is sentenceGroup name
	sentenceGroupMap := make(map[string]ImportSentenceGroupItem)
	preparedSentenceGroupMap := make(map[string]*model.SentenceGroup)

	var groupName, sentenceName, tagName, nodeType, role string

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		// TODO if groupName, sentenceName is empty, report error
		groupName = row.Cells[0].String()
		sentenceName = row.Cells[1].String()
		tagName = row.Cells[2].String()
		if len(row.Cells) < 3 {
			return nil, fmt.Errorf("invalid column content")
		}
		if len(row.Cells) < 7 {
			nodeType = ""
			role = ""
		} else if len(row.Cells) < 8 {
			nodeType = row.Cells[6].String()
			role = ""
		} else {
			nodeType = row.Cells[6].String()
			role = row.Cells[7].String()
		}

		sentenceItem := ImportSentenceItem{
			SentenceName: sentenceName,
			TagName:      tagName,
		}
		if sentenceGroup, ok := sentenceGroupMap[groupName]; ok {
			sentenceGroup.SentenceItems = append(sentenceGroup.SentenceItems, &sentenceItem)
			sentenceGroupMap[groupName] = sentenceGroup
		} else {
			sentenceGroupMap[groupName] = ImportSentenceGroupItem{
				GroupName:     groupName,
				NodeType:      nodeType,
				Role:          role,
				SentenceItems: []*ImportSentenceItem{&sentenceItem},
			}
		}
	}

	// default category
	categoryID := uint64(0)
	deleted := int8(0)
	for groupName, sentenceGroupInfo := range sentenceGroupMap {
		// ------------------
		// create sentence
		// ------------------
		createdSentences := make([]*DataSentence, 0)
		for _, sentenceItem := range sentenceGroupInfo.SentenceItems {
			senQuery := &model.SentenceQuery{
				Enterprise: &enterpriseID,
				IsDelete:   &deleted,
				Name:       &sentenceItem.SentenceName,
			}
			count, err := sentenceDao.CountSentences(nil, senQuery)
			if err != nil {
				return nil, err
			}
			if count > 0 {
				logger.Trace.Printf("Found existing sentence %s \n", sentenceItem.SentenceName)
				// it is assumed that sentence name is unique
				data, err := getSentences(senQuery)
				if err != nil {
					return nil, err
				}
				createdSentences = append(createdSentences, data[0])
				continue
			}

			splits := strings.Split(sentenceItem.TagName, "+")
			var uuidArr []string
			// get uuid according to tag name
			for _, item := range splits {
				query := &model.TagQuery{
					Enterprise: &enterpriseID,
					Name:       &item,
				}
				resp, err := Tags(*query)
				if err != nil {
					logger.Error.Printf("Failed to query tag %s, %s \n", item, err.Error())
					return nil, err
				}
				if resp.Paging.Total == 0 {
					logger.Error.Printf("Failed to find tag %s for sentence %s \n", item, sentenceItem.SentenceName)
					return nil, fmt.Errorf("failed to find tag %s for sentence %s", item, sentenceItem.SentenceName)
				}
				uuidArr = append(uuidArr, resp.Data[0].TagUUID)
			}

			createdSentence, err := NewSentence(enterpriseID, categoryID, sentenceItem.SentenceName, uuidArr)
			if err != nil {
				logger.Error.Printf("Failed to create sentence %s \n", sentenceItem.SentenceName)
				return nil, err
			}
			createdSentences = append(createdSentences, createdSentence)
			logger.Trace.Printf("Create sentence %s \n", sentenceItem.SentenceName)
		}

		// ----------------------------
		// collect sentenceGroup info
		// ----------------------------

		filter := &model.SentenceGroupFilter{
			Name:       groupName,
			Enterprise: enterpriseID,
			Limit:      0,
			IsDelete:   &deleted,
		}
		total, groups, err := GetSentenceGroupsBy(filter)
		if err != nil {
			logger.Error.Printf("Failed to get sentenceGroup, %s \n", err.Error())
			return nil, err
		}
		if total > 1 {
			// TODO check if sentenceGroup exists, should skip, and report, if sentenceGroup contains deleted sentence
			// delete existing sentenceGroup ?
			logger.Trace.Printf("Found existing sentenceGroup %s, skip ... \n", groupName)
			// it is assumed that sentenceGroup name is unique
			//preparedSentenceGroupMap[groupName] = &groups[0]
			//continue
			if from == FromFlow {
				if err := DeleteSentenceGroup(groups[0].UUID); err != nil {
					return nil, err
				}
				logger.Trace.Printf("Delete existing sentenceGroup %s \n", groupName)
			}
		}

		senGroup := &model.SentenceGroup{
			Name:       groupName,
			Enterprise: enterpriseID,
		}
		// do not check if sentence exists
		sentences := make([]model.SimpleSentence, 0)
		for _, item := range createdSentences {
			sentences = append(sentences, model.SimpleSentence{
				ID:   item.ID,
				UUID: item.UUID,
				Name: item.Name,
			})
		}
		senGroup.Sentences = sentences

		if sentenceGroupInfo.Role != "" {
			if roleCode, ok := roleMapping[sentenceGroupInfo.Role]; ok {
				senGroup.Role = roleCode
			} else {
				// default role: any
				senGroup.Role = 2
			}
		}
		// default position: 2, positionMap
		senGroup.Position = 2
		// default type: call-in, typeMapping
		senGroup.Type = 1
		// TODO 中文
		if sentenceGroupInfo.NodeType == "可选" {
			senGroup.Optional = 1
		} else {
			// default optional: 0
			senGroup.Optional = 0
		}
		// default range: 0
		senGroup.Distance = 0

		preparedSentenceGroupMap[groupName] = senGroup
	}
	return preparedSentenceGroupMap, nil
}

type ImportSentenceGroupItem struct {
	GroupName     string
	NodeType      string
	Role          string
	SentenceItems []*ImportSentenceItem
}

type ImportSentenceItem struct {
	SentenceName string
	TagName      string
}

func BatchAddRules(fileName string, enterpriseID string) error {
	// TODO need a check function
	// [][WARN ] 2019/03/05 09:01:37 tags_services.go:472: more than one needed sentence: 回访-分红险责任告知-询问是否清楚
	// 如果系统原先错误操作导致more than one sentence会存在隐患

	// get sentenceGroup info
	preparedSentenceGroupMap, err := PrepareSentenceGroups(fileName, enterpriseID, FromRule)

	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		logger.Error.Printf("Failed to open file %s \n", fileName)
		return err
	}

	err = batchAddRules(xlFile, enterpriseID, preparedSentenceGroupMap, FromRule)
	return err
}

func batchAddRules(xlFile *xlsx.File, enterpriseID string, preparedSentenceGroupMap map[string]*model.SentenceGroup, from string) error {
	sheet, ok := xlFile.Sheet[RuleSheetName]
	if !ok {
		logger.Error.Printf("Failed to get sheet %s \n", RuleSheetName)
		return fmt.Errorf("failed to get sheet %s \n", RuleSheetName)
	}

	var ruleName, description, logicList, operatorStr string
	var method, score, dialogueRepeat int

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
		logicList = row.Cells[6].String()
		if len(row.Cells) < 8 {
			operatorStr = ""
		} else {
			operatorStr = row.Cells[7].String()
		}
		score, err = row.Cells[2].Int()
		if err != nil {
			logger.Error.Printf("Failed to get value from score column, row is %d, %s \n", i+1, err.Error())
			// default score
			score = 0
		}
		method, err = row.Cells[1].Int()
		if err != nil {
			logger.Error.Printf("Failed to get value from method column, row is %d, %s \n", i+1, err.Error())
			// default method is positive
			method = 1
		}
		dialogueRepeat, err = row.Cells[5].Int()
		if err != nil {
			logger.Error.Printf("Failed to get value from repeat column, row is %d, %s \n", i+1, err.Error())
			// default dialogue repeat
			dialogueRepeat = 1
		}

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
			logger.Trace.Printf("Found existing rule %s, skip ... \n", ruleName)
			continue
		}

		// TODO the length of operatorList should be equal to logicList
		var operatorList = make([]string, 0)
		if operatorStr != "" {
			operatorList = strings.Split(operatorStr, "|")
		}

		var createdSentenceGroups []*model.SentenceGroup

		splits := strings.Split(logicList, "|")
		for _, groupName := range splits {
			senGroupInfo, ok := preparedSentenceGroupMap[groupName]
			if !ok {
				logger.Error.Printf("Failed to find required sentenceGroup %s \n", groupName)
				return fmt.Errorf("failed to find required sentenceGroup info")
			}

			// ------------------
			// add sentence-group
			// ------------------
			sentenceGroup := &model.SentenceGroup{
				Name:       "",
				Enterprise: enterpriseID,
				Sentences:  senGroupInfo.Sentences,
				Role:       senGroupInfo.Role,
				Position:   senGroupInfo.Position,
				Type:       0,
				Optional:   senGroupInfo.Optional,
				Distance:   senGroupInfo.Distance,
			}

			createdSentenceGroup, err := CreateSentenceGroup(sentenceGroup)
			if err != nil {
				logger.Error.Printf("Failed to create sentenceGroup, sentence is %s, %s \n", groupName, err.Error())
				return err
			}
			createdSentenceGroups = append(createdSentenceGroups, createdSentenceGroup)
		}

		// ----------------------
		// add conversation-flow
		// ----------------------

		cfUUIDStr, err := general.UUID()
		if err != nil {
			return err
		}
		now := time.Now().Unix()

		conversationFlow := &model.ConversationFlow{
			UUID:       cfUUIDStr,
			Name:       ruleName + "-dialogue1",
			Enterprise: enterpriseID,
			Min:        dialogueRepeat,
			CreateTime: now,
			UpdateTime: now,
		}
		var cfExpression string
		numOfOP := len(operatorList)
		sentenceGroups := make([]model.SimpleSentenceGroup, len(createdSentenceGroups))
		for i, item := range createdSentenceGroups {
			sentenceGroups[i] = model.SimpleSentenceGroup{
				ID:   item.ID,
				UUID: item.UUID,
				Name: item.Name,
			}

			if from == FromFlow {
				if numOfOP != 0 {
					if i == 0 {
						cfExpression = operatorList[i] + " " + item.UUID
					} else {
						cfExpression = cfExpression + " " + operatorList[i] + " " + item.UUID
					}
				} else {
					// default logic format: must -> then
					if i == 0 {
						cfExpression = "if " + item.UUID
					} else {
						cfExpression = cfExpression + " then " + item.UUID
					}
				}
				continue
			}

			if numOfOP != 0 {
				if i == 0 {
					cfExpression = operatorList[i] + " " + item.UUID
				} else {
					cfExpression = cfExpression + " " + operatorList[i] + " " + item.UUID
				}
			} else {
				// default logic format: must -> then
				if i == 0 {
					cfExpression = "must " + item.UUID
				} else {
					cfExpression = cfExpression + " then " + item.UUID
				}
			}
		}

		conversationFlow.Expression = cfExpression
		conversationFlow.SentenceGroups = sentenceGroups

		createdConversationFlow, err := conversationFlowDao.Create(conversationFlow, tx)
		if err != nil {
			logger.Error.Printf("Failed to create conversationFlow %s, %s", ruleName, err.Error())
			return err
		}

		// ----------
		// add rule
		// ----------

		now = time.Now().Unix()
		ruleUUIDStr, err := general.UUID()
		if err != nil {
			return err
		}
		// default min, default max
		// default severity is normal
		rule := &model.ConversationRule{
			Name:        ruleName,
			Min:         1,
			Max:         0,
			Score:       score,
			Description: description,
			Enterprise:  enterpriseID,
			Severity:    int8(0),
			Method:      int8(method),
			CreateTime:  now,
			UpdateTime:  now,
			UUID:        ruleUUIDStr,
		}
		// create default conversation-flow
		flows := make([]model.SimpleConversationFlow, 1)
		flows[0] = model.SimpleConversationFlow{
			ID:   createdConversationFlow.ID,
			UUID: createdConversationFlow.UUID,
			Name: createdConversationFlow.Name,
		}
		rule.Flows = flows

		createdConversationRule, err := conversationRuleDao.Create(rule, tx)
		if err != nil {
			logger.Error.Printf("Failed to create conversationRule %s, %s \n", ruleName, err.Error())
			return err
		}
		logger.Trace.Printf("Create rule %s \n", createdConversationRule.Name)
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

		name = row.Cells[1].String()
		content = row.Cells[2].String()

		deleted := int8(0)
		query := &model.SentenceQuery{
			Enterprise: &enterpriseID,
			IsDelete:   &deleted,
			Name:       &name,
		}

		count, err := sentenceDao.CountSentences(nil, query)
		if err != nil {
			return err
		}
		if count > 0 {
			logger.Trace.Printf("Found existing sentence %s \n", name)
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
				logger.Error.Printf("Failed to query tag %s, %s \n", item, err.Error())
				return err
			}
			if resp.Paging.Total == 0 {
				// TODO should report error
				logger.Error.Printf("Failed to find tag %s for sentence %s \n", item, name)
				continue
			}
			uuidArr = append(uuidArr, resp.Data[0].TagUUID)
		}

		if len(uuidArr) == 0 {
			// TODO should report error
			logger.Error.Printf("Failed to find required tags for sentence %s \n", name)
			continue
		}

		_, err = NewSentence(enterpriseID, categoryID, name, uuidArr)
		if err != nil {
			logger.Error.Printf("Failed to create sentence %s, %s \n", name, err.Error())
			return err
		}
		logger.Trace.Printf("Create sentence %s \n", name)
	}
	return nil
}

func batchAddTagsBy(xlFile *xlsx.File, enterpriseID string, sheetName string) error {
	sheet, ok := xlFile.Sheet[sheetName]
	if !ok {
		logger.Error.Printf("Failed to load sheet %s \n", sheetName)
		return fmt.Errorf("failed to load sheet %s", sheetName)
	}

	var name, content, contentType string
	// key is tag name, value is positive corpus
	posCorpusMap := make(map[string][]string)
	// key is tag name, value is negative corpus
	negCorpusMap := make(map[string][]string)
	var tagType int8
	if sheetName == TagKeywordSheetName {
		tagType = 1
	} else {
		tagType = 2
	}

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		if len(row.Cells) < 2 || row.Cells[0] == nil || row.Cells[1] == nil {
			return fmt.Errorf("lack of required column")
		}
		name = row.Cells[0].String()
		content = row.Cells[1].String()
		if len(row.Cells) == 2 {
			contentType = ""
		} else {
			contentType = row.Cells[2].String()
		}

		switch contentType {
		case "":
			fallthrough
		case PositiveCorpusType:
			if corpus, ok := posCorpusMap[name]; ok {
				posCorpusMap[name] = append(corpus, content)
			} else {
				posCorpusMap[name] = []string{content}
			}
		case NegativeCorpusType:
			if corpus, ok := negCorpusMap[name]; ok {
				negCorpusMap[name] = append(corpus, content)
			} else {
				negCorpusMap[name] = []string{content}
			}
		}
	}

	for name, posCorpus := range posCorpusMap {
		query := model.TagQuery{
			Enterprise: &enterpriseID,
			Name:       &name,
		}

		resp, err := Tags(query)
		if err != nil {
			logger.Error.Printf("Failed to query tag %s, %s \n", name, err.Error())
			return err
		}
		if resp.Paging.Total > 0 {
			logger.Trace.Printf("Found existing tag %s \n", name)
			continue
		}

		posSentences, _ := json.Marshal(posCorpus)
		negStr := make([]string, 0)
		if negCorpus, ok := negCorpusMap[name]; ok {
			negStr = negCorpus
		}
		negSentences, _ := json.Marshal(negStr)

		current := time.Now().Unix()
		uuidStr, err := general.UUID()
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
			UUID:             uuidStr,
		}

		_, err = NewTag(tag)
		if err != nil {
			logger.Error.Printf("Failed to create tag %s, %s", name, err.Error())
			return err
		}
		logger.Trace.Printf("Create tag %s \n", name)
	}
	return nil
}

func BatchAddFlows(fileName string, enterpriseID string) error {
	var err error

	// add tag
	if err = BatchAddTags(fileName, enterpriseID); err != nil {
		logger.Error.Printf("Failed to call BatchAddTags, %s \n", err.Error())
		return err
	}

	// get sentenceGroup info
	preparedSentenceGroupMap, err := PrepareSentenceGroups(fileName, enterpriseID, FromFlow);
	if err != nil {
		logger.Error.Printf("Failed to call BatchAddSentenceGroups, %s \n", err.Error())
		return err
	}

	xlFile, err := xlsx.OpenFile(fileName)
	if err != nil {
		logger.Error.Printf("Failed to open file %s \n", fileName)
		return err
	}
	sheet, ok := xlFile.Sheet[RuleSheetName]
	if !ok {
		logger.Error.Printf("Failed to get sheet %s \n", RuleSheetName)
		return fmt.Errorf("failed to get sheet %s \n", RuleSheetName)
	}

	var flowName, logicList string

	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}

		// TODO check value or use constant variable
		flowName = row.Cells[3].String()
		logicList = row.Cells[6].String()

		flag := 0
		query := &model.NavQuery{Enterprise: &enterpriseID, IsDelete: &flag, Name: &flowName}
		flows, err := GetFlows(query, 1, 1)
		// if find existing flow, recreate it
		if len(flows) > 0 {
			logger.Trace.Printf("Found existing flow %s \n", flowName)
			// should use first
			_, err := DeleteFlow(flows[0].ID, enterpriseID)
			if err != nil {
				logger.Error.Printf("Failed to delete flow %s \n", flowName)
			}
			logger.Trace.Printf("delete existing flow: %s \n", flowName)
		}

		// default flow type
		flow := reqNewFlow{
			Name:       flowName,
			IntentName: flowName,
			Type:       NavigationFlowIntentType,
		}

		flowID, err := NewFlow(&flow, enterpriseID)
		logger.Trace.Printf("Create flow %s \n", flowName)
		if err != nil {
			logger.Error.Printf("Failed to create flow %s \n", err)
			return err
		}

		// ---------------------------------------------------
		// create intent node or normal node (SentenceGroup)
		// ---------------------------------------------------

		// get sentence according to logic_list
		splits := strings.Split(logicList, "|")
		for i, groupName := range splits {
			sentenceGroup, ok := preparedSentenceGroupMap[groupName]
			if !ok {
				// TODO need prepared sentenceGroup info ?
				logger.Error.Printf("Failed to find needed sentenceGroup %s \n", groupName)
				return fmt.Errorf("failed to find needed sentenceGroup info")
			}

			// the first sentence is intent node, the rest is normal node
			// add intent node
			if i == 0 {
				flow := &model.NavFlowUpdate{}
				ignore := 0
				flow.IgnoreIntent = &ignore

				createdSentenceGroup, err := CreateSentenceGroup(sentenceGroup)
				if err != nil {
					logger.Error.Printf("Failed to create sentenceGroup %s, %s \n", groupName, err.Error())
					return err
				}
				flow.IntentLinkID = &createdSentenceGroup.ID
				flow.IntentName = &flowName

				if _, err = UpdateFlow(flowID, enterpriseID, flow); err != nil {
					logger.Error.Println("Failed to update flow")
					return err
				}
				logger.Trace.Printf("Create intent node %s \n", groupName)
				continue
			}

			// add normal node
			err = NewNode(flowID, sentenceGroup)
			if err != nil {
				logger.Error.Println("Failed to create normal node")
				return err
			}
			logger.Trace.Printf("Create normal node %s \n", groupName)
		}
	}
	// add rules for flow
	err = batchAddRules(xlFile, enterpriseID, preparedSentenceGroupMap, FromFlow)
	return err
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
