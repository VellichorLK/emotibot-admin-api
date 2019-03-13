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
	MStatDeprecate
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

//UnloadModel unloads the given model
func UnloadModel(modelID int64) error {
	var appid logicaccess.TrainAPPID
	appid.ID = uint64(modelID)
	err := trainer.UnloadModel(&appid)
	if err != nil {
		logger.Error.Printf("unload model %d failed. %s\n", modelID, err)
	}
	return err
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
func TrainOneModelByEnterprise(tags []model.Tag, modelID int64, enterprise string) error {

	if dbLike == nil {
		return ErrNilCon
	}
	if len(tags) == 0 {
		return ErrNoArgument
	}

	now := time.Now().Unix()

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
			return fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", tag.ID, err)
		}

		err = json.Unmarshal([]byte(tag.NegativeSentence), &neg)
		if err != nil {
			return fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", tag.ID, err)
		}

		switch tag.Typ {
		case 1: //keyword
			if len(pos) == 0 {
				logger.Warn.Printf("tag %d has no keywords\n", tag.ID)
			} else {
				data := &logicaccess.TrainKeywordData{Tag: tag.ID, Words: pos}
				unit.Keyword.Data = append(unit.Keyword.Data, data)
			}
		case 2: //dialog_act
			if len(pos) == 0 && len(neg) == 0 {
				logger.Warn.Printf("tag %d has no pos and neg\n", tag.ID)
			} else {
				data := &logicaccess.TrainTagData{Tag: tag.ID, PosSentence: pos, NegSentence: neg}
				unit.Dialog.Data = append(unit.Dialog.Data, data)
			}
		case 3: //user_response
			if len(pos) == 0 && len(neg) == 0 {
				logger.Warn.Printf("tag %d has no pos and neg\n", tag.ID)
			} else {
				data := &logicaccess.TrainTagData{Tag: tag.ID, PosSentence: pos, NegSentence: neg}
				unit.UsrResponse.Data = append(unit.UsrResponse.Data, data)
			}
		default:
			logger.Warn.Printf("tag %d has unknown type %d, ignore it\n", tag.ID, tag.Typ)
			continue
		}
	}

	err := trainer.Train(unit)
	if err != nil {
		return err
	}

	err = waitTrainingModel(modelID)

	return err
}

func waitTrainingModel(modelID int64) error {
	//get the status of training model to make sure it is finished
	for {
		status, err := trainer.Status(&logicaccess.TrainAPPID{ID: uint64(modelID)})
		if err != nil {
			logger.Error.Printf("get status of training failed. %s,%s\n", status, err)
			return err
		}
		if status == "ready" {
			break
		} else if status == "error" {
			logger.Error.Printf("training model error\n")
			return errors.New("model training error")
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

//TrainModelByEnterprise trains model
func TrainModelByEnterprise(enterprise string) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}

	modelID, err := newModelWithCollisionDetection(enterprise)
	if err != nil {
		logger.Error.Printf("create a model failed. %s\n", err)
		return 0, err
	}
	go func() {
		query := model.TagQuery{Enterprise: &enterprise}
		tags, err := tagDao.Tags(nil, query)
		if err != nil {
			logger.Error.Printf("get tag failed\n")
			//return 0, err
			return
		}
		err = TrainOneModelByEnterprise(tags, modelID, enterprise)
		if err != nil {
			logger.Error.Printf("train model failed. %s\n", err)
			affected, _ := modelDao.UpdateModel(dbLike.Conn(), &model.TModel{ID: uint64(modelID), Status: MStatErr})
			if affected == 0 {
				logger.Warn.Printf("model %d may have wrong status\n", modelID)
			}

			//return 0, err
			return
		}

		//switch the using model, set the last using model to deprecate and new one to using
		tx, err := dbLike.Begin()
		if err != nil {
			//return 0, err
			return
		}
		defer tx.Rollback()

		//get the using models
		status := MStatUsing
		models, err := modelDao.TrainedModelInfo(tx, &model.TModelQuery{Enterprise: &enterprise, Status: &status})
		if err != nil {
			logger.Error.Printf("get using model failed. %s\n", err)
			//return 0, err
			return
		}
		//seet the current using model to deprecate status
		if len(models) > 0 {
			_, err = modelDao.UpdateModel(tx, &model.TModel{ID: models[0].ID, Status: MStatDeprecate})
			if err != nil {
				logger.Error.Printf("update model %d status failed. %s\n", models[0].ID, err)
				//return 0, err
				return
			}
		}
		//set the new model to using status
		_, err = modelDao.UpdateModel(tx, &model.TModel{ID: uint64(modelID), Status: MStatUsing})
		if err != nil {
			logger.Error.Printf("update model %d status failed. %s\n", models[0].ID, err)
			//return 0, err
			return
		}

		err = tx.Commit()
		if err != nil {
			logger.Error.Printf("commit failed. %s\n", err)
			//return 0, err
			return
		}
		//if exist deprecated models which somehow doesn't be cleaned
		cleanModel(enterprise)
		if len(models) > 0 {
			//unload the model after two hour
			setUnloadModelTimer(int64(models[0].ID), time.Duration(2*time.Hour))
		}
	}()
	return modelID, nil
}

func cleanModel(enterprise string) error {
	models, err := GetModelByEnterprise(enterprise, MStatDeprecate)
	if err != nil {
		logger.Error.Printf("get deprecate models failed. %s\n", err)
		return err
	}
	now := time.Now().Unix()
	for _, m := range models {
		if now > m.UpdateTime+60*60*2 {
			unloadModel(int64(m.ID))
		}

	}
	return nil
}

func unloadModel(id int64) error {

	err := UnloadModel(id)
	if err != nil {
		logger.Error.Printf("unload model %d failed. %s\n", id, err)
	} else {
		status := MStatDeletion
		_, err = modelDao.UpdateModel(dbLike.Conn(), &model.TModel{ID: uint64(id), Status: status})
		if err != nil {
			logger.Error.Printf("update model %d status to %d failed. %s\n", id, status, err)
		}
	}

	return err
}

func setUnloadModelTimer(id int64, d time.Duration) error {
	if dbLike == nil {
		return ErrNilCon
	}
	timer := time.NewTimer(d)
	<-timer.C

	err := unloadModel(id)
	return err
}

//GetUsingModelByEnterprise gets the using model id
func GetUsingModelByEnterprise(enterprise string) ([]*model.TModel, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	models, err := GetModelByEnterprise(enterprise, MStatUsing)
	if err != nil {
		logger.Error.Printf("get trained models failed.%s\n", err)
	}
	return models, err
}

//GetModelByEnterprise gets the model
func GetModelByEnterprise(enterprise string, status int) ([]*model.TModel, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	q := &model.TModelQuery{Enterprise: &enterprise, Status: &status}
	return modelDao.TrainedModelInfo(dbLike.Conn(), q)
}

//GetAllModelByEnterprise gets the all model
func GetAllModelByEnterprise(enterprise string) ([]*model.TModel, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	q := &model.TModelQuery{Enterprise: &enterprise}
	return modelDao.TrainedModelInfo(dbLike.Conn(), q)
}

var (
	ErrTrainingBusy = errors.New("training is going")
)

func newModelWithCollisionDetection(enterprise string) (int64, error) {

	if dbLike == nil {
		return 0, ErrNilCon
	}
	conn := dbLike.Conn()

	now := time.Now().Unix()
	status := MStatTraining

	//check whether someone is doing training
	models, err := modelDao.TrainedModelInfo(conn, &model.TModelQuery{Enterprise: &enterprise, Status: &status})
	if err != nil {
		logger.Error.Printf("get train model failed. %s\n", err)
		return 0, err
	}
	if len(models) != 0 {
		// training too long,may disconnected at the last time, auto recovery
		if (now - models[0].UpdateTime) > 60*60 {
			logger.Info.Printf("auto recovery model %d to err\n", models[0].ID)
			affected, _ := modelDao.UpdateModel(dbLike.Conn(), &model.TModel{ID: models[0].ID, Status: MStatErr})
			if affected == 0 {
				logger.Warn.Printf("model %d may have wrong status\n", models[0].ID)
			}
		} else {
			return 0, ErrTrainingBusy
		}
	}

	//insert a record to indicate the training is doing
	modelID, err := modelDao.NewModel(conn, &model.TModel{Enterprise: enterprise, CreateTime: now, UpdateTime: now, Status: status})
	if err != nil {
		logger.Error.Printf("insert train model failed. %s\n", err)
		return 0, err
	}

	//check whether only one is doing training
	models, err = modelDao.TrainedModelInfo(conn, &model.TModelQuery{Enterprise: &enterprise, Status: &status})
	if err != nil {
		logger.Error.Printf("get train model failed. %s\n", err)
		return 0, err
	}
	//if there is more than one that is doing training, delete itself
	if len(models) != 1 {
		modelDao.DeleteModel(conn, &model.TModelQuery{ID: []uint64{uint64(modelID)}})
		return 0, ErrTrainingBusy
	}

	return modelID, nil
}
