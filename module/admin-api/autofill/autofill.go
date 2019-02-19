package autofill

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/Service"
	"emotibot.com/emotigo/module/admin-api/autofill/dao"
	"emotibot.com/emotigo/module/admin-api/autofill/data"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/solr"
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

func createAutofills(appID string, docs []*data.AutofillBody) error {
	indexURL := fmt.Sprintf("%s/%s/update", solr.GetBaseURL(), data.THIRD_CORE)

	// NOTE:
	//	Solr accepts duplicate keys to add multiple documents in one request,
	//		ex: {
	//				"add": {
	//					"id": "ID1",
	//					"sentence": "Hello"
	//				},
	//				"add": {
	//					"id": "ID2",
	//					"sentence": "World"
	//				}
	//		 	}
	//  However, duplicated key JSON is not supported by "encoding/json" package in Golang.
	//  Thus, we have to compose JSON string body by our own selves.
	addCmds := make([]string, len(docs))

	for i, doc := range docs {
		addCmd := solr.AddCmd{
			Add: solr.AddCmdBody{
				Doc: doc,
			},
		}

		cmd, err := json.Marshal(&addCmd)
		if err != nil {
			return err
		}

		// Remove the leading '{' and trailing '}'
		addCmds[i] = string(cmd[1 : len(cmd)-1])
	}

	addCmdsJSON := fmt.Sprintf("{%s}", strings.Join(addCmds, ","))

	client := &http.Client{}
	req, err := http.NewRequest("POST", indexURL, strings.NewReader(addCmdsJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = client.Do(req)
	return err
}

func deleteAllAutofills(appID string, module string) error {
	deleteURL := fmt.Sprintf("%s/%s/update", solr.GetBaseURL(), data.THIRD_CORE)
	deleteCmd := solr.DeleteByQueryCmd{
		Delete: solr.DeleteByQueryCmdBody{
			Query: fmt.Sprintf("id:%s_autofill#%s_*", appID, module),
		},
	}

	reqBody, err := json.Marshal(&deleteCmd)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", deleteURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = client.Do(req)
	return err
}

func enableAutofills(ids []string) error {
	docs := make([]*data.AutofillToggleBody, len(ids))
	for i, id := range ids {
		docs[i] = data.NewAutofillToggleBody(id, true)
	}
	return toggleAutofills(docs)
}

func disableAutofills(ids []string) error {
	docs := make([]*data.AutofillToggleBody, len(ids))
	for i, id := range ids {
		docs[i] = data.NewAutofillToggleBody(id, false)
	}
	return toggleAutofills(docs)
}

func toggleAutofills(docs []*data.AutofillToggleBody) error {
	updateURL := fmt.Sprintf("%s/%s/update", solr.GetBaseURL(), data.THIRD_CORE)

	// NOTE:
	//	Solr accepts duplicate keys to in-place update multiple documents in one request,
	//		ex: {
	//				"add": {
	//					"id": "ID1",
	//					"sentence": { "set": "Hello" }
	//				},
	//				"add": {
	//					"id": "ID2",
	//					"sentence": { "set": "World" }
	//				}
	//		 	}
	//  However, duplicated key JSON is not supported by "encoding/json" package in Golang.
	//  Thus, we have to compose JSON string body by our own selves.
	updateCmds := make([]string, len(docs))

	for i, doc := range docs {
		updateCmd := solr.UpdateCmd{
			Add: solr.UpdateCmdBody{
				Doc: doc,
			},
		}

		cmd, err := json.Marshal(&updateCmd)
		if err != nil {
			return err
		}

		// Remove the leading '{' and trailing '}'
		updateCmds[i] = string(cmd[1 : len(cmd)-1])
	}

	updateCmdsJSON := fmt.Sprintf("{%s}", strings.Join(updateCmds, ","))

	client := &http.Client{}
	req, err := http.NewRequest("POST", updateURL, strings.NewReader(updateCmdsJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = client.Do(req)
	return err
}

func createAutofillDocID(appID string, module string, moduleID int64,
	sentenceID int64) string {
	return fmt.Sprintf("%s_autofill#%s_%d_%d", appID, module, moduleID, sentenceID)
}

func createAutofillDatabase(appID string, module string) string {
	return fmt.Sprintf("%s_autofill#%s", appID, module)
}

func complementSet(a []string, b []string) []string {
	complement := make([]string, 0)

	for _, elementA := range a {
		found := false

		for _, elementB := range b {
			if elementB == elementA {
				found = true
				break
			}
		}

		if !found {
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
	autofillBodies := make(map[string][]*data.AutofillBody)

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
				autofillBodies[sentence] = make([]*data.AutofillBody, 0)
			}

			autofillBodies[sentence] = append(autofillBodies[sentence], &data.AutofillBody{
				ModuleID:         intentSentence.ModuleID,
				SentenceID:       intentSentence.SentenceID,
				SentenceOriginal: sentence,
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
				body.ID = createAutofillDocID(appID,
					data.AutofillModuleIntent, body.ModuleID, body.SentenceID)
				body.Database = createAutofillDatabase(appID, data.AutofillModuleIntent)

				// Related sentences
				relatedSentence := data.NewRelatedSentence(sentence)
				relatedSentenceJSON, err := json.Marshal(&relatedSentence)
				if err != nil {
					return err
				}
				body.RelatedSentences = string(relatedSentenceJSON)

				body.Sentence = strings.ToLower(util.T2S(nluResult.Keyword.ToString()))
				body.SentenceCU = "{}"
				body.SentenceKeywords = strings.ToLower(util.T2S(nluResult.Keyword.ToString()))
				body.SentenceType = strings.ToLower(util.T2S(nluResult.SentenceType))
				body.SentencePos = strings.ToLower(util.T2S(nluResult.Segment.ToFullString()))
				body.Source = time.Now().Format("2006-01-02 15:04:05")

				// Autofill enabled
				enabled := false
				if _, ok := teIntentIDsMap[body.ModuleID]; ok {
					enabled = true
				}
				body.Enabled = enabled
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

		autofills := make([]*data.AutofillBody, 0)
		for _, bodies := range autofillBodies {
			for _, body := range bodies {
				autofills = append(autofills, body)
			}
		}

		// Recreate all intent autofills
		err = deleteAllAutofills(appID, data.AutofillModuleIntent)
		if err != nil {
			return err
		}

		err = createAutofills(appID, autofills)
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

	var docIDs []string
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

			docIDs = make([]string, len(enableIntentSentences))
			for i, intentSentence := range enableIntentSentences {
				docIDs[i] = createAutofillDocID(appID, data.AutofillModuleIntent,
					intentSentence.ModuleID, intentSentence.SentenceID)
			}

			err = enableAutofills(docIDs)
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

			docIDs = make([]string, len(disableIntentSentences))
			for i, intentSentence := range disableIntentSentences {
				docIDs[i] = createAutofillDocID(appID, data.AutofillModuleIntent,
					intentSentence.ModuleID, intentSentence.SentenceID)
			}

			err = disableAutofills(docIDs)
			if err != nil {
				return err
			}
		}
	}

	// Update Task Engine's previous intents
	err = autofillDao.UpdateTEPrevIntents(appID, teIntents)
	return err
}
