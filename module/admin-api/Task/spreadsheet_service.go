package Task

import (
	"encoding/json"
	"errors"
	"strings"

	"emotibot.com/emotigo/module/admin-api/Dictionary"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/tealeg/xlsx"
)

// ParseUploadSpreadsheet will parse and verify uploaded spreadsheet
// Use intent_engine_2.0 intent as trigger
func ParseUploadSpreadsheet(appID string, scenarioString string, fileBuf []byte) (*Scenario, error) {
	var scenario Scenario
	json.Unmarshal([]byte(scenarioString), &scenario)
	xlFile, err := xlsx.OpenBinary(fileBuf)
	if err != nil {
		return nil, err
	}

	var sheet *xlsx.Sheet
	// parse triggier phrases and register an intent to intent engine 1.0
	sheet = xlFile.Sheet[SheetName["triggerPhrase"]]
	if sheet != nil {
		// create an intent with intent_name = scenario_name
		scenarioName := scenario.EditingContent.Metadata["scenario_name"]
		trigger := createDefaultTrigger(scenarioName, "intent_engine")
		triggerList := &scenario.EditingContent.Skills["mainSkill"].TriggerList
		*triggerList = append(*triggerList, trigger)

		// register an intent to intent engine 1.0
		triggerPhrases, err := parseTriggerPhrase(sheet)
		if err != nil {
			return nil, err
		}
		err = UpdateIntentV1(appID, scenarioName, triggerPhrases)
		if err != nil {
			return nil, err
		}
	}

	// parse triggier intents 2.0 and assign to scenario
	sheet = xlFile.Sheet[SheetName["triggerIntent"]]
	if sheet != nil {
		triggerList, err := parseTriggerIntent(sheet)
		if err != nil {
			return nil, err
		}
		scenario.EditingContent.Skills["mainSkill"].TriggerList = triggerList
	}

	// parser entity and assign to scenario
	sheet = xlFile.Sheet[SheetName["entityCollecting"]]
	if sheet != nil {
		entityList, err := parseEntity(sheet)
		if err != nil {
			return nil, err
		}
		scenario.EditingContent.Skills["mainSkill"].EntityCollectorList = entityList
	}

	// parse action and assign to scenario
	sheet = xlFile.Sheet[SheetName["responseMessage"]]
	if sheet != nil {
		actionGroupList, err := parseMsgAction(sheet)
		if err != nil {
			return nil, err
		}
		scenario.EditingContent.Skills["mainSkill"].ActionGroupList = actionGroupList
	}

	// parse action and assign to scenario
	sheet = xlFile.Sheet[SheetName["nerMap"]]
	if sheet != nil {
		nerMap, err := parseNerMap(sheet)
		if err != nil {
			return nil, err
		}
		scenario.EditingContent.IDToNerMap = nerMap
	}

	return &scenario, err
}

func parseTriggerPhrase(sheet *xlsx.Sheet) ([]string, error) {
	phrases := make([]string, 0)
	sheetPhrase := new(SpreadsheetTrigger)
	if sheet.MaxRow == 0 {
		return phrases, errors.New("Missing trigger phrases")
	}
	for i := 0; i < sheet.MaxRow; i++ {
		err := sheet.Rows[i].ReadStruct(sheetPhrase)
		if err != nil {
			return nil, err
		}
		phrases = append(phrases, sheetPhrase.Phrase)
	}
	return phrases, nil
}

func parseTriggerIntent(sheet *xlsx.Sheet) ([]*Trigger, error) {
	triggerList := make([]*Trigger, 0)
	sheetIntent := new(SpreadsheetTriggerIntent)
	if sheet.MaxRow == 0 {
		return triggerList, errors.New("Missing trigger intents")
	}
	for i := 0; i < sheet.MaxRow; i++ {
		err := sheet.Rows[i].ReadStruct(sheetIntent)
		if err != nil {
			return nil, err
		}
		trigger := createDefaultTrigger(sheetIntent.Intent, "intent_engine_2.0")
		triggerList = append(triggerList, trigger)
	}
	return triggerList, nil
}

func createDefaultTrigger(intentName string, intentType string) *Trigger {
	trigger := Trigger{
		Type:       intentType,
		IntentName: intentName,
		Editable:   true,
	}
	return &trigger
}

func parseEntity(sheet *xlsx.Sheet) ([]*Entity, error) {
	entities := make([]*Entity, 0)
	sheetEntity := new(SpreadsheetEntity)
	if sheet.MaxRow <= 1 {
		// skip entity collecting
		return nil, nil
	}
	for i := 1; i < sheet.MaxRow; i++ {
		err := sheet.Rows[i].ReadStruct(sheetEntity)
		if err != nil {
			return nil, err
		}
		entity := sheetEntity.ToEntity()
		entities = append(entities, &entity)
	}
	return entities, nil
}

func parseMsgAction(sheet *xlsx.Sheet) ([]*ActionGroup, error) {
	actionGroupList := make([]*ActionGroup, 0)
	sheetMsgAction := new(SpreadsheetMsgAction)
	if sheet.MaxRow == 0 {
		return nil, errors.New("Missing response message")
	}
	for i := 0; i < sheet.MaxRow; i++ {
		err := sheet.Rows[i].ReadStruct(sheetMsgAction)
		if err != nil {
			return nil, err
		}
		action := Action{
			Type: "msg",
			Msg:  sheetMsgAction.Msg,
		}
		actionGroup := ActionGroup{
			ActionGroupID: util.GenRandomUUIDSameAsOpenAPI(),
			ActionList:    []*Action{&action},
			ConditionList: make([]*interface{}, 0),
		}

		actionGroupList = append(actionGroupList, &actionGroup)
	}
	return actionGroupList, nil
}

func parseNerMap(sheet *xlsx.Sheet) (map[string]*CustomNer, error) {
	nerMap := map[string]*CustomNer{}
	wordBankRow := new(Dictionary.WordBankRow)
	lastWordBankRow := new(Dictionary.WordBankRow)
	pathToEntitySynonymsList := make(map[string][]*EntitySynonyms)
	if sheet.MaxRow <= 1 {
		// skip nerMap parsing
		return nil, nil
	}

	for i := 1; i < sheet.MaxRow; i++ {
		if empty := isEmptyRow(sheet.Rows[i]); empty == true {
			// skip empty row
			continue
		}

		err := sheet.Rows[i].ReadStruct(wordBankRow)
		if err != nil {
			return nil, err
		}
		entitySynonyms := &EntitySynonyms{
			Entity:   wordBankRow.Name,
			Synonyms: wordBankRow.SimilarWords,
		}

		newWordBankRow := wordBankRow.FillLevel(*lastWordBankRow)
		lastWordBankRow = &newWordBankRow
		path := newWordBankRow.GetPath()

		if _, ok := pathToEntitySynonymsList[path]; !ok {
			pathToEntitySynonymsList[path] = []*EntitySynonyms{}
		}
		pathToEntitySynonymsList[path] = append(pathToEntitySynonymsList[path], entitySynonyms)
	}

	for k, v := range pathToEntitySynonymsList {
		customNer := newCustomNer()
		customNer.EntityType = k
		customNer.EntitySynonymsList = v
		nerMap[customNer.ID] = &customNer
	}
	return nerMap, nil
}

func isEmptyRow(row *xlsx.Row) bool {
	cells := row.Cells
	rowCellStr := make([]string, len(cells))
	for cellIdx, cell := range cells {
		rowCellStr[cellIdx] = strings.TrimSpace(cell.Value)
	}
	if strings.TrimSpace(strings.Join(rowCellStr, "")) == "" {
		return true
	}
	return false
}
