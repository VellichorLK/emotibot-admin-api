package qi

import (
	"encoding/json"
	"fmt"
	"strconv"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
)

var (
	trainer logicaccess.PredictClient
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
