package Robot

import (
	"database/sql"
	"fmt"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/zhconverter"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	qaServices "emotibot.com/emotigo/module/admin-api/QADoc/services"
	"emotibot.com/emotigo/module/admin-api/Service"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	syncTimeout = 5 * 60
)

func GetRobotQAListV3(appid string, locale string) ([]*QAInfoV3, int, error) {
	qainfos, err := getRobotQAListV3(appid, locale)
	if err == sql.ErrNoRows {
		return []*QAInfoV3{}, ApiError.SUCCESS, nil
	} else if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return qainfos, ApiError.SUCCESS, nil
}

func GetRobotQAV3(appid string, qid int) (*QAInfoV3, int, error) {
	qainfo, err := getRobotQAV3(appid, qid)
	if err == sql.ErrNoRows {
		return nil, ApiError.NOT_FOUND_ERROR, nil
	} else if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return qainfo, ApiError.SUCCESS, nil
}

func AddRobotQAAnswerV3(appid string, qid int, answer string) (int, int, error) {
	id, err := addRobotQAAnswerV3(appid, qid, answer)
	if err == sql.ErrNoRows {
		return 0, ApiError.NOT_FOUND_ERROR, util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
	} else if err == util.ErrDuplicated {
		return 0, ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["Answer"])
	} else if err != nil {
		return 0, ApiError.DB_ERROR, err
	}
	return id, ApiError.SUCCESS, nil
}

func GetBasicQusetionV3(qid int) (*InfoV3, error) {
	ret, err := getBasicQuestionV3(qid)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetRobotQAAnswerV3(appid string, qid, aid int) (*InfoV3, error) {
	ret, err := getRobotQAAnswerV3(appid, qid, aid)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdateRobotQAAnswerV3(appid string, qid int, aid int, answer string) (int, error) {
	err := updateRobotQAAnswerV3(appid, qid, aid, answer)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err == util.ErrDuplicated {
		return ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["Answer"])
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func DeleteRobotQAAnswerV3(appid string, qid, aid int) (int, error) {
	err := deleteRobotQAAnswerV3(appid, qid, aid)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func AddRobotQARQuestionV3(appid string, qid int, question string) (int, int, error) {
	id, err := addRobotQARQuestionV3(appid, qid, question)
	if err == sql.ErrNoRows {
		return 0, ApiError.NOT_FOUND_ERROR, util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
	} else if err == util.ErrDuplicated {
		return 0, ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["RelateQuestion"])
	} else if err != nil {
		return 0, ApiError.DB_ERROR, err
	}
	return id, ApiError.SUCCESS, nil
}

func GetRobotQARQuestionV3(appid string, qid, rQid int) (*InfoV3, error) {
	ret, err := getRobotQARQuestionV3(appid, qid, rQid)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdateRobotQARQuestionV3(appid string, qid, rQid int, relateQuestion string) (int, error) {
	err := updateRobotQARQuestionV3(appid, qid, rQid, relateQuestion)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err == util.ErrDuplicated {
		return ApiError.REQUEST_ERROR, util.GenDuplicatedError(
			util.Msg["Content"], util.Msg["RelateQuestion"])
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func DeleteRobotQARQuestionV3(appid string, qid, rQid int) (int, error) {
	err := deleteRobotQARQuestionV3(appid, qid, rQid)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func SyncOnce() {
	SyncRobotProfile()
}

func SyncRobotProfile() (err error) {
	return ForceSyncRobotProfile(false)
}

func ForceSyncRobotProfile(force bool) (err error) {
	restart := false
	var body []byte
	defer func() {
		if err != nil {
			logger.Error.Println("Error while syncing robot profile:", err.Error())
			return
		}
	}()

	if !force {
		var start bool
		var pid int
		start, pid, err = tryStartSyncProcess(syncTimeout)
		if err != nil {
			return
		}
		if !start {
			logger.Info.Println("Pass sync, there is still process running")
			return
		}
		defer func() {
			if r := recover(); r != nil {
				msg := ""
				switch r.(type) {
				case error:
					msg = (r.(error)).Error()
				case string:
					msg = r.(string)
				default:
					msg = fmt.Sprintf("%v", r)
				}
				finishSyncProcess(pid, false, msg)
			} else if err != nil {
				finishSyncProcess(pid, false, err.Error())
			} else {
				finishSyncProcess(pid, true, "")
			}

			restart, err = needProcessRobotData()
			if err != nil {
				logger.Error.Println("Check status fail: ", err.Error())
				return
			}
			if restart {
				logger.Trace.Println("Restart sync process")
				time.Sleep(time.Second)
				go SyncRobotProfile()
			}
		}()
	}

	rqIDs, ansIDs, delAnsIDs, tagInfos, appids, err := getProcessModifyRobotQA(force)
	if err != nil {
		return
	}

	deleteDocIDs, deleteRQIDs, err := getDeleteModifyRobotQA()
	if err != nil {
		return
	}
	logger.Trace.Printf("relate question: %+v", rqIDs)
	logger.Trace.Printf("update answers: %+v", ansIDs)
	logger.Trace.Printf("delete answers: %+v", delAnsIDs)
	logger.Trace.Printf("delete docid: %+v", deleteDocIDs)
	logger.Trace.Printf("delete relate question: %+v", deleteRQIDs)

	if len(tagInfos) == 0 && len(deleteRQIDs) == 0 && len(delAnsIDs) == 0 {
		logger.Trace.Println("No data need to update")
		return
	}

	tagInfos = convertTagInfoWithZhCn(tagInfos)

	validInfos := ManualTaggings{}
	for idx := range tagInfos {
		if tagInfos[idx].Answers != nil && len(tagInfos[idx].Answers) > 0 {
			validInfos = append(validInfos, tagInfos[idx])
		}
	}

	if len(validInfos) > 0 {
		err = fillNLUInfoInTaggingInfos(validInfos)
		if err != nil {
			logger.Error.Printf("Get NLUInfo fail: %s\n", err.Error())
			return
		}

		body, err = qaServices.BulkCreateOrUpdateQADocs(validInfos.convertToQACoreDocs())
		if err != nil {
			logger.Error.Printf("QA service fail, err: %s, response: %s",
				err.Error(), string(body))
			return
		}
	}

	deleteStdQIDs := map[string][]string{}
	for idx := range tagInfos {
		info := tagInfos[idx]
		if info.Answers == nil || len(info.Answers) <= 0 {
			if _, ok := deleteStdQIDs[info.AppID]; !ok {
				deleteStdQIDs[info.AppID] = []string{}
			}
			deleteStdQIDs[info.AppID] = append(deleteStdQIDs[info.AppID], info.DocID)
		}
	}

	if len(deleteStdQIDs) > 0 {
		deleteIDs := []interface{}{}
		for appID, ids := range deleteStdQIDs {
			for _, id := range ids {
				qDocID := createRobotQuestionDocID(appID, id)
				deleteIDs = append(deleteIDs, qDocID)
			}
		}

		// Delete questions
		body, err = qaServices.DeleteQADocsByIds(deleteIDs...)
		if err != nil {
			logger.Error.Printf("QA service fail, err: %s, response: %s, \n", err.Error(), string(body))
			return
		}
	}

	if len(deleteRQIDs) > 0 {
		deleteIDs := []interface{}{}
		for appID, ids := range deleteDocIDs {
			for _, id := range ids {
				qDocID := createRobotQuestionDocID(appID, id)
				deleteIDs = append(deleteIDs, qDocID)
			}
		}

		// Delete questions
		body, err = qaServices.DeleteQADocsByIds(deleteIDs...)
		if err != nil {
			logger.Error.Printf("QA service fail, err: %s, response: %s, \n", err.Error(), string(body))
			return
		}
	}

	err = resetRobotQAData(rqIDs, deleteRQIDs, ansIDs, delAnsIDs)
	if err != nil {
		logger.Error.Println("Reset status to 0 fail: ", err.Error())
		return
	}

	for _, appid := range appids {
		if appid == "" {
			continue
		}
		logger.Trace.Printf("Update profile consul of appid [%s]\n", appid)
		_, consulErr := util.ConsulUpdateProfile(appid)
		if err != nil {
			logger.Error.Println("Update consul error:", consulErr.Error())
		}
	}

	return
}

func convertTagInfoWithZhCn(tagInfos []*ManualTagging) []*ManualTagging {
	if tagInfos == nil {
		return tagInfos
	}
	for _, tagInfo := range tagInfos {
		if tagInfo == nil {
			continue
		}
		tagInfo.Question = zhconverter.T2S(tagInfo.Question)
		if tagInfo.Answers == nil {
			continue
		}
	}
	return tagInfos
}

func fillNLUInfoInTaggingInfos(tagInfos []*ManualTagging) error {
	sentences := make([]string, 0, len(tagInfos))
	for _, tagInfo := range tagInfos {
		sentences = append(sentences, tagInfo.Question)
	}

	sentenceResult, err := Service.BatchGetNLUResults("", sentences)
	if err != nil {
		return err
	}

	for _, tagInfo := range tagInfos {
		if _, ok := sentenceResult[tagInfo.Question]; !ok {
			continue
		}
		result := sentenceResult[tagInfo.Question]
		tagInfo.Segment = result.Segment.ToString()
		tagInfo.WordPos = result.Segment.ToFullString()
		tagInfo.Keyword = result.Keyword.ToString()
		tagInfo.SentenceType = result.SentenceType
	}

	return nil
}
