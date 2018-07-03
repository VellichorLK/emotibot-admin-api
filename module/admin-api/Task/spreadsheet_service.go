package Task

import (
	"encoding/json"
	"errors"

	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/tealeg/xlsx"
)

// ParseUploadSpreadsheet will parse and verify uploaded spreadsheet
func ParseUploadSpreadsheet(scenarioString string, fileBuf []byte) ([]string, *Scenario, error) {
	var scenario Scenario
	var triggerPhrases []string
	json.Unmarshal([]byte(scenarioString), &scenario)
	xlFile, err := xlsx.OpenBinary(fileBuf)
	if err != nil {
		return nil, nil, err
	}

	// crreate an default intent named by scenario_name
	trigger := createDefaultTrigger(&scenario)
	triggerList := &scenario.EditingContent.Skills["mainSkill"].TriggerList
	*triggerList = append(*triggerList, trigger)

	var sheet *xlsx.Sheet
	// parse triggier phrases to return
	sheet = xlFile.Sheet[SheetName["triggerPhrase"]]
	if sheet != nil {
		triggerPhrases, err = parseTriggerPhrase(sheet)
		if err != nil {
			return nil, nil, err
		}
	}

	// parser entity and assign to scenario
	sheet = xlFile.Sheet[SheetName["entityCollecting"]]
	if sheet != nil {
		entityList, err := parseEntity(sheet)
		if err != nil {
			return nil, nil, err
		}
		scenario.EditingContent.Skills["mainSkill"].EntityCollectorList = entityList
	}

	sheet = xlFile.Sheet[SheetName["responseMessage"]]
	if sheet != nil {
		actionGroupList, err := parseMsgAction(sheet)
		if err != nil {
			return nil, nil, err
		}
		scenario.EditingContent.Skills["mainSkill"].ActionGroupList = actionGroupList
	}

	return triggerPhrases, &scenario, err
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

func createDefaultTrigger(scenario *Scenario) *Trigger {
	intentName := scenario.EditingContent.Metadata["scenario_name"]
	trigger := Trigger{
		Type:       "intent_engine",
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
