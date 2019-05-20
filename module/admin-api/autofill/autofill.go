package autofill

import (
	"errors"
	"fmt"
	"strings"

	qaData "emotibot.com/emotigo/module/admin-api/QADoc/data"
	qaServices "emotibot.com/emotigo/module/admin-api/QADoc/services"
	"emotibot.com/emotigo/module/admin-api/Service"
	"emotibot.com/emotigo/module/admin-api/autofill/dao"
	"emotibot.com/emotigo/module/admin-api/autofill/data"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/zhconverter"
	"emotibot.com/emotigo/pkg/logger"
)

var autofillDao dao.Dao

func Init() {
	db := util.GetMainDB()
	autofillDao = dao.NewDao(db)
}

// UpdateAutofills creates async task to update autofills stored in Solr
func UpdateAutofills(appID string, option *data.AutofillOption) {
	go func() {
		// Try to retrieve sync task lock
		result, resetFailed, err := autofillDao.TryGetSyncTaskLock(appID, option.TaskMode)
		if err != nil {
			logger.Error.Printf("Try to get autofills sync task lock failed, %s", err.Error())
			return
		}

		if result {
			defer func() {
				if err != nil {
					logger.Error.Printf(err.Error())
				}
				autofillDao.SyncTaskFinish(appID, err)
			}()

			logger.Info.Println("Got autofills sync task lock")
			logger.Info.Printf("Start autofills sync task, module: %s, task mode: %d",
				option.Module, option.TaskMode)

			switch option.Module {
			case data.AutofillModuleIntent:
				switch option.TaskMode {
				case data.SyncTaskModeReset:
					err = resetIntentAutofills(appID)
				case data.SyncTaskModeUpdate:
					// Previous sync task was a reset task and was failed,
					// redo the reset sync task
					// TODO: Multi-modules supports, current implementation
					//       cannot distinguish what previous failed sync task
					//		 belongs to which module
					if resetFailed {
						err = resetIntentAutofills(appID)
						if err != nil {
							return
						}

						err = autofillDao.UpdateSyncTaskMode(appID, data.SyncTaskModeUpdate)
						if err != nil {
							return
						}
					}

					err = updateIntentAutofills(appID)
				default:
					errMsg := fmt.Sprintf("Unknown autofills sync task mode, %d", option.TaskMode)
					err = errors.New(errMsg)
					return
				}
			}
		}
	}()
}

func createAutofills(docs []*qaData.QACoreDoc) error {
	_, err := qaServices.BulkCreateQADocs(docs)
	return err
}

func deleteAllAutofills(appID string, module string) error {
	_, err := qaServices.DeleteQADocs(appID, module)
	return err
}

func toggleAutofills(enabled bool, ids []interface{}) error {
	script := fmt.Sprintf("ctx._source.autofill_enabled=%t", enabled)
	_, err := qaServices.UpdateQADocsByQuery(script, ids...)
	return err
}

func createAutofillDocID(appID string, module string, moduleID int64,
	sentenceID int64) string {
	return fmt.Sprintf("%s_%s_%d_%d", appID, module, moduleID, sentenceID)
}

func complementSet(a []string, b []string) []string {
	m := map[string]bool{}

	for _, elementB := range b {
		m[elementB] = true
	}

	complement := make([]string, 0)

	for _, elementA := range a {
		if _, ok := m[elementA]; !ok {
			complement = append(complement, elementA)
		}
	}

	return complement
}

// Intent autofills
func resetIntentAutofills(appID string) error {
	// Clear Task Engine previous intents
	err := autofillDao.DeleteAllTEPrevIntents(appID)
	if err != nil {
		return err
	}

	// Sync task lock retrieved, start autofill sync task
	sentences := make([]string, 0)

	// Sentence: autofill bodies
	autofillBodies := map[string][]*data.QACoreDoc{}

	// Current intent IDs used by Task Engine
	teIntentIDs, err := autofillDao.GetTECurrentIntentIDs(appID)
	if err != nil {
		return err
	}

	teIntentIDsMap := make(map[int64]bool)
	for _, teIntentID := range teIntentIDs {
		teIntentIDsMap[teIntentID] = true
	}

	var lastSentenceID int64 = -1

	// Batch loading intent sentences
	for {
		intentSentences, err := autofillDao.GetIntentSentences(appID, nil, lastSentenceID)
		if err != nil {
			return err
		}

		if len(intentSentences) == 0 {
			// Nothing to do, goodbye!!
			break
		}

		// Update last sentence ID for next interation
		lastSentenceID = intentSentences[len(intentSentences)-1].SentenceID

		for _, intentSentence := range intentSentences {
			sentence := intentSentence.Sentence

			// Autofill body
			if _, ok := autofillBodies[sentence]; !ok {
				autofillBodies[sentence] = make([]*data.QACoreDoc, 0)
			}

			autofillBodies[sentence] = append(autofillBodies[sentence],
				&data.QACoreDoc{
					QACoreDoc: qaData.QACoreDoc{
						AppID:        appID,
						Module:       data.AutofillModuleIntent,
						SentenceOrig: strings.ToLower(zhconverter.T2S(sentence)),
					},
					ModuleID:   intentSentence.ModuleID,
					SentenceID: intentSentence.SentenceID,
					Sentence:   sentence,
				})

			// Sentences
			sentences = append(sentences, sentence)
		}

		nluResults, err := Service.BatchGetNLUResults(appID, sentences)
		if err != nil {
			return err
		}

		for sentence, nluResult := range nluResults {
			bodies := autofillBodies[sentence]

			for _, body := range bodies {
				body.QACoreDoc = qaData.QACoreDoc{
					DocID: createAutofillDocID(appID, data.AutofillModuleIntent,
						body.ModuleID, body.SentenceID),
					AppID:        appID,
					Module:       "autofill_intent",
					Domain:       "",
					Answers:      []*qaData.Answer{
						              &qaData.Answer{
							              Sentence: sentence,
						              },
					              },
					Sentence:     strings.ToLower(nluResult.Segment.ToString()),
					SentenceOrig: body.Sentence,
					SentenceType: strings.ToLower(nluResult.SentenceType),
					SentencePos:  strings.ToLower(nluResult.Segment.ToFullString()),
					Keywords:     strings.ToLower(nluResult.Keyword.ToString()),
					StdQID:       fmt.Sprintf("%d", body.SentenceID),
					StdQContent:  body.Sentence,
				}

				// Autofill enabled
				enabled := false
				if _, ok := teIntentIDsMap[body.ModuleID]; ok {
					enabled = true
				}
				body.QACoreDoc.AutofillEnabled = enabled
			}
		}

		// Double check if we should rerun the sync task before updating Solr
		rerun, err := autofillDao.ShouldRerunSyncTask(appID)
		if err != nil {
			return err
		} else if rerun {
			lastSentenceID = -1
			err = autofillDao.RerunSyncTask(appID)
			if err != nil {
				return err
			}
			continue
		}

		autofills := make([]*qaData.QACoreDoc, 0)
		for _, bodies := range autofillBodies {
			for _, body := range bodies {
				autofills = append(autofills, &body.QACoreDoc)
			}
		}

		// Recreate all intent autofills
		err = deleteAllAutofills(appID, data.AutofillModuleIntent)
		if err != nil {
			return err
		}

		err = createAutofills(autofills)
		if err != nil {
			return err
		}
	}

	// Double check if we should rerun the sync task
	rerun, err := autofillDao.ShouldRerunSyncTask(appID)
	if err != nil {
		return err
	} else if rerun {
		lastSentenceID = -1
		err = autofillDao.RerunSyncTask(appID)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateIntentAutofills(appID string) error {
	teIntents, err := autofillDao.GetTECurrentIntents(appID)
	if err != nil {
		return err
	}

	tePrevIntents, err := autofillDao.GetTEPrevIntents(appID)
	if err != nil {
		return err
	}

	var docIDs []interface{}
	var lastSentenceID int64 = -1

	// Enable autofills
	enableIntents := complementSet(teIntents, tePrevIntents)

	if len(enableIntents) > 0 {
		// Batch loading enabled intent sentences
		for {
			enableIntentSentences, err := autofillDao.GetIntentSentences(appID, enableIntents, lastSentenceID)
			if err != nil {
				return err
			}

			if len(enableIntentSentences) == 0 {
				break
			}

			// Update last sentence ID for next interation
			lastSentenceID = enableIntentSentences[len(enableIntentSentences)-1].SentenceID

			docIDs = make([]interface{}, len(enableIntentSentences))
			for i, intentSentence := range enableIntentSentences {
				docIDs[i] = createAutofillDocID(appID, data.AutofillModuleIntent,
					intentSentence.ModuleID, intentSentence.SentenceID)
			}

			err = toggleAutofills(true, docIDs)
			if err != nil {
				return err
			}
		}
	}

	// Disable autofills
	disableIntents := complementSet(tePrevIntents, teIntents)
	lastSentenceID = -1

	if len(disableIntents) > 0 {
		// Batch loading disabled intent sentences
		for {
			disableIntentSentences, err := autofillDao.GetIntentSentences(appID, disableIntents, lastSentenceID)
			if err != nil {
				return err
			}

			if len(disableIntentSentences) == 0 {
				break
			}

			// Update last sentence ID for next interation
			lastSentenceID = disableIntentSentences[len(disableIntentSentences)-1].SentenceID

			docIDs = make([]interface{}, len(disableIntentSentences))
			for i, intentSentence := range disableIntentSentences {
				docIDs[i] = createAutofillDocID(appID, data.AutofillModuleIntent,
					intentSentence.ModuleID, intentSentence.SentenceID)
			}

			err = toggleAutofills(false, docIDs)
			if err != nil {
				return err
			}
		}
	}

	// Update Task Engine's previous intents
	err = autofillDao.UpdateTEPrevIntents(appID, teIntents)
	return err
}
