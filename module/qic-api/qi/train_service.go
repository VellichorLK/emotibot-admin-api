package qi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
)

var (
	trainer  logicaccess.PredictClient
	modelDao model.TrainedModelDao = &model.TrainedModelSQLDao{}
)

//the status code
const (
	MStatTraining = iota
	MStatReady
	MStatErr
	MStatUsing
	MStatDeletion
)

//TrainAllTags trains all tag, only for demo usage, not for production
func TrainAllTags() error {
	query := model.TagQuery{}
	tags, err := tagDao.Tags(nil, query)
	if err != nil {
		return err
	}
	return TrainTags(tags)
}

//TrainTags trains tags
func TrainTags(tags []model.Tag) error {
	for _, tag := range tags {
		logic := &logicaccess.TrainLogic{}
		trainLogic := &logicaccess.TrainLogicData{Operator: "+", Name: tag.Name}
		trainLogic.Tags = append(trainLogic.Tags, strconv.FormatUint(tag.ID, 10))
		logic.Data = append(logic.Data, trainLogic)

		unit := &logicaccess.TrainUnit{Logic: logic}

		var pos, neg []string
		err := json.Unmarshal([]byte(tag.PositiveSentence), &pos)
		if err != nil {
			return fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", tag.ID, err)
		}

		err = json.Unmarshal([]byte(tag.NegativeSentence), &neg)
		if err != nil {
			return fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", tag.ID, err)
		}

		logic.ID = tag.ID

		switch tag.Typ {
		case 1: //keyword
			keyword := &logicaccess.TrainKeyword{ID: tag.ID}
			data := &logicaccess.TrainKeywordData{}
			data.Tag = tag.ID
			data.Words = pos
			keyword.Data = append(keyword.Data, data)
			unit.Keyword = keyword

		case 2: //dialog_act
			common := &logicaccess.CommonTrainData{ID: tag.ID}
			data := &logicaccess.TrainTagData{}
			data.Tag = tag.ID
			data.PosSentence = pos
			data.NegSentence = neg

			common.Data = append(common.Data, data)
			unit.Dialog = common
		case 3: //user_response
			common := &logicaccess.CommonTrainData{ID: tag.ID}
			data := &logicaccess.TrainTagData{}
			data.Tag = tag.ID
			data.PosSentence = pos
			data.NegSentence = neg

			common.Data = append(common.Data, data)
			unit.UsrResponse = common
		default:
			continue
		}

		err = trainer.Train(unit)
		if err != nil {
			return err
		}
	}
	return nil
}

//UnloadAllTags unloads the all tags
func UnloadAllTags() error {
	query := model.TagQuery{}
	tags, err := tagDao.Tags(nil, query)
	if err != nil {
		return err
	}
	return UnloadTags(tags)
}

//UnloadTags unloads tags
func UnloadTags(tags []model.Tag) error {

	for _, tag := range tags {
		var appid logicaccess.TrainAPPID
		appid.ID = tag.ID
		err := trainer.UnloadModel(&appid)
		if err != nil {
			//currently not break, continue to try left
			logger.Error.Printf("unload model %d failed. %s\n", tag.ID, err)
		}
	}
	return nil
}

//errorr message
var (
	ErrNilCon = errors.New("Nil db connection")
)

//TrainOneModelByEnterprise trains all tags as one model
func TrainOneModelByEnterprise(tags []model.Tag, enterprise string) (int64, error) {

	if dbLike == nil {
		return 0, ErrNilCon
	}
	if len(tags) == 0 {
		return 0, ErrNoArgument
	}

	now := time.Now().Unix()
	modelID, err := modelDao.NewModel(dbLike.Conn(), &model.TModel{Enterprise: enterprise, CreateTime: now, UpdateTime: now, Status: MStatTraining})
	if err != nil {
		logger.Error.Printf("Insert new model failed. %s\n", err)
		return 0, err
	}

	logic := &logicaccess.TrainLogic{}
	trainLogic := &logicaccess.TrainLogicData{Operator: "+", Name: enterprise + "-" + strconv.FormatInt(now, 10)}
	//only assigned one tag, but actually we don't need the logic
	trainLogic.Tags = append(trainLogic.Tags, strconv.FormatUint(tags[0].ID, 10))
	logic.Data = append(logic.Data, trainLogic)
	logic.ID = uint64(modelID)
	unit := &logicaccess.TrainUnit{Logic: logic}
	unit.Dialog = &logicaccess.CommonTrainData{ID: logic.ID, Data: []*logicaccess.TrainTagData{}}
	unit.Keyword = &logicaccess.TrainKeyword{ID: logic.ID, Data: []*logicaccess.TrainKeywordData{}}
	unit.UsrResponse = &logicaccess.CommonTrainData{ID: logic.ID, Data: []*logicaccess.TrainTagData{}}
	for _, tag := range tags {
		var pos, neg []string
		err := json.Unmarshal([]byte(tag.PositiveSentence), &pos)
		if err != nil {
			return 0, fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", tag.ID, err)
		}

		err = json.Unmarshal([]byte(tag.NegativeSentence), &neg)
		if err != nil {
			return 0, fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", tag.ID, err)
		}

		switch tag.Typ {
		case 1: //keyword
			data := &logicaccess.TrainKeywordData{Tag: tag.ID, Words: pos}
			unit.Keyword.Data = append(unit.Keyword.Data, data)
		case 2: //dialog_act
			data := &logicaccess.TrainTagData{Tag: tag.ID, PosSentence: pos, NegSentence: neg}
			unit.Dialog.Data = append(unit.Dialog.Data, data)
		case 3: //user_response
			data := &logicaccess.TrainTagData{Tag: tag.ID, PosSentence: pos, NegSentence: neg}
			unit.UsrResponse.Data = append(unit.UsrResponse.Data, data)
		default:
			logger.Warn.Printf("tag %d has unknown type %d, ignore it\n", tag.ID, tag.Typ)
			continue
		}
	}
	err = trainer.Train(unit)
	if err != nil {
		return 0, err
	}
	return modelID, nil
}

//TrainModelByEnterprise trains model
func TrainModelByEnterprise(enterprise string) (int64, error) {
	query := model.TagQuery{Enterprise: &enterprise}
	tags, err := tagDao.Tags(nil, query)
	if err != nil {
		logger.Error.Printf("get tag failed\n")
		return 0, err
	}
	modelID, err := TrainOneModelByEnterprise(tags, enterprise)
	return modelID, err
}

//GetUsingModelByEnterprise gets the using model id
func GetUsingModelByEnterprise(enterprise string) ([]*model.TModel, error) {
	models, err := GetModelByEnterprise(enterprise, MStatUsing)
	if err != nil {
		logger.Error.Printf("get trained models failed.%s\n", err)
	}
	return models, err
}

//GetModelByEnterprise gets the model
func GetModelByEnterprise(enterprise string, status int) ([]*model.TModel, error) {
	q := &model.TModelQuery{Enterprise: &enterprise, Status: &status}
	return modelDao.TrainedModelInfo(dbLike.Conn(), q)
}
