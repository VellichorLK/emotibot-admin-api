package v2

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ErrReadOnlyIntent means trying to modify intent which is trained (version is not NULL)
	ErrReadOnlyIntent = errors.New("intent is readonly if it is trained")
)

type intentDaoInterface interface {
	GetIntents(appid string, version *int, keyword string) ([]*IntentV2, error)
	GetIntent(appid string, intentID int64, keyword string) (*IntentV2, error)
	AddIntent(appid, name string, positive, negative []string) (*IntentV2, error)
	ModifyIntent(appid string, intentID int64, name string, updateSentence []*SentenceV2WithType, deleteSentences []int64) error
	DeleteIntent(appid string, intentID int64) error
}

// intentDaoV2 implement interface of intentDaoInterface, which will store for service to use
type intentDaoV2 struct {
	db *sql.DB
}

const (
	typePositive = iota
	typeNegative
)

// GetIntents will get intents from db
// version can be NULL, will return data that version column in NULL
// each return object only has id, name and count of positive/negative sentences
func (dao intentDaoV2) GetIntents(appid string, version *int, keyword string) (ret []*IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	conditions := []string{"appid = ?"}
	params := []interface{}{appid}
	if version != nil {
		conditions = append(conditions, "version = ?")
		params = append(params, version)
	} else {
		conditions = append(conditions, "version is NULL")
	}

	// Get intents id and name
	queryStr := fmt.Sprintf("SELECT id, name FROM intents WHERE %s", strings.Join(conditions, " AND "))
	intentRows, err := tx.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer intentRows.Close()

	intentMap := map[int64]*IntentV2{}
	intents := []*IntentV2{}
	for intentRows.Next() {
		intent := &IntentV2{}
		err = intentRows.Scan(&intent.ID, &intent.Name)
		if err != nil {
			return
		}
		intents = append(intents, intent)
		intentMap[intent.ID] = intent
	}

	// If try to get specific version and get empty,
	if version != nil && len(intents) == 0 {
		return nil, sql.ErrNoRows
	}

	// Get train data count, type 0 is Positive, type 1 is Negative
	conditions = []string{"intents.appid = ?", "intents.id = s.intent"}
	params = []interface{}{appid}

	if version != nil {
		conditions = append(conditions, "intents.version = ?")
		params = append(params, version)
	} else {
		conditions = append(conditions, "intents.version is NULL")
	}

	if keyword != "" {
		conditions = append(conditions, "s.sentence like ?")
		params = append(params, getLikeValue(keyword))
	}

	queryStr = fmt.Sprintf(`
		SELECT s.intent, s.type, count(*)
		FROM intent_train_sets as s,intents
		WHERE %s
		GROUP BY s.intent, s.type`, strings.Join(conditions, " AND "))
	countRows, err := tx.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer countRows.Close()

	for countRows.Next() {
		var id int64
		contentType, count := 0, 0
		err = countRows.Scan(&id, &contentType, &count)
		if err != nil {
			return
		}
		if _, ok := intentMap[id]; !ok {
			continue
		}
		switch contentType {
		case 0:
			intentMap[id].PositiveCount = count
		case 1:
			intentMap[id].NegativeCount = count
		}
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	ret = []*IntentV2{}
	for idx := range intents {
		// intent has sentence with keyword
		if intents[idx].PositiveCount > 0 || intents[idx].NegativeCount > 0 {
			ret = append(ret, intents[idx])
			continue
		}

		// ignore case in english when checking name of intent
		intentName := strings.ToLower(intents[idx].Name)
		key := strings.ToLower(keyword)
		if strings.Index(intentName, key) >= 0 {
			ret = append(ret, intents[idx])
		}
	}
	return
}

// GetIntent will get intent from db with specific appid and id
// It will return full data of the intent, containing id, name, count and sentences
func (dao intentDaoV2) GetIntent(appid string, intentID int64, keyword string) (ret *IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	queryStr := "SELECT id, name FROM intents WHERE appid = ? AND id = ?"
	intentRow := dao.db.QueryRow(queryStr, appid, intentID)

	intent := IntentV2{}
	err = intentRow.Scan(&intent.ID, &intent.Name)
	if err != nil {
		return
	}

	negativeList := []*SentenceV2{}
	positiveList := []*SentenceV2{}
	// Get train data count, type 0 is Positive, type 1 is Negative
	var contentRows *sql.Rows
	if keyword != "" {
		queryStr = "SELECT id, sentence, type FROM intent_train_sets WHERE intent = ? AND sentence like ?"
		contentRows, err = dao.db.Query(queryStr, intentID, getLikeValue(keyword))
	} else {
		queryStr = "SELECT id, sentence, type FROM intent_train_sets WHERE intent = ?"
		contentRows, err = dao.db.Query(queryStr, intentID)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			intent.Negative = &[]*SentenceV2{}
			intent.Positive = &[]*SentenceV2{}
			ret = &intent
		}
		return
	}
	defer contentRows.Close()

	for contentRows.Next() {
		sentence := SentenceV2{}
		sentenceType := 0
		err = contentRows.Scan(&sentence.ID, &sentence.Content, &sentenceType)
		if sentenceType == typePositive {
			positiveList = append(positiveList, &sentence)
		} else if sentenceType == typeNegative {
			negativeList = append(negativeList, &sentence)
		}
	}
	intent.NegativeCount = len(negativeList)
	intent.PositiveCount = len(positiveList)
	intent.Negative = &negativeList
	intent.Positive = &positiveList
	ret = &intent
	return
}

// AddIntent will add new intent of appid. The version of new intent will be NULL, which
// means that intent hasn't been trained
func (dao intentDaoV2) AddIntent(appid, name string, positive, negative []string) (ret *IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	intent := IntentV2{}
	timestamp := time.Now().Unix()
	queryStr := "INSERT INTO intents (appid, name, version, updatetime) VALUES (?, ?, NULL, ?)"
	result, err := tx.Exec(queryStr, appid, name, timestamp)
	if err != nil {
		return nil, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	intent.ID = newID
	intent.Name = name

	positiveList := []*SentenceV2{}
	negativeList := []*SentenceV2{}

	queryStr = fmt.Sprintf("INSERT INTO intent_train_sets (sentence, intent, type) VALUES (?, %d, ?)", newID)
	sentenceMap := map[int][]string{
		typePositive: positive,
		typeNegative: negative,
	}

	for contentType := range sentenceMap {
		list := sentenceMap[contentType]
		if list != nil {
			for idx := range list {
				content := list[idx]
				result, err = tx.Exec(queryStr, content, contentType)
				if err != nil {
					return
				}
				var sentenceID int64
				sentenceID, err = result.LastInsertId()
				if err != nil {
					return
				}

				sentence := SentenceV2{}
				sentence.ID = sentenceID
				sentence.Content = content
				switch contentType {
				case typePositive:
					positiveList = append(positiveList, &sentence)
				case typeNegative:
					negativeList = append(negativeList, &sentence)
				}
			}
		}
	}
	intent.Negative = &negativeList
	intent.NegativeCount = len(negativeList)
	intent.Positive = &positiveList
	intent.PositiveCount = len(positiveList)

	err = tx.Commit()
	if err != nil {
		return
	}
	ret = &intent
	return
}

// ModifyIntent will update intent with diff information
// In updateSentence, if id is 0, means new sentence.
// If intent version is not NULL, return error because it is readonly
func (dao intentDaoV2) ModifyIntent(appid string, intentID int64, name string,
	updateSentence []*SentenceV2WithType, deleteSentences []int64) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	if name == "" {
		return util.ErrParameter
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	readOnly, err := checkIntent(tx, appid, intentID)
	if err != nil {
		return
	}
	if readOnly {
		return ErrReadOnlyIntent
	}

	timestamp := time.Now().Unix()

	queryStr := "UPDATE intents SET name = ?, updatetime = ? WHERE id = ?"
	_, err = tx.Exec(queryStr, name, timestamp, intentID)
	if err != nil {
		return
	}

	insertQuery := "INSERT INTO intent_train_sets (sentence, intent, type) VALUES (?, ?, ?)"
	updateQuery := "UPDATE intent_train_sets SET sentence = ? WHERE id = ?"
	for _, sentence := range updateSentence {
		if sentence.ID == 0 {
			_, err = tx.Exec(insertQuery, sentence.Content, intentID, sentence.Type)
		} else {
			_, err = tx.Exec(updateQuery, sentence.Content, sentence.ID)
		}
		if err != nil {
			return
		}
	}

	if len(deleteSentences) > 0 {
		params := make([]interface{}, len(deleteSentences))
		for idx := range deleteSentences {
			params[idx] = deleteSentences[idx]
		}
		queryStr := fmt.Sprintf("DELETE FROM intent_train_sets WHERE id in (?%s)",
			strings.Repeat(",?", len(deleteSentences)-1))
		_, err = tx.Exec(queryStr, params...)
		if err != nil {
			return
		}
	}

	err = tx.Commit()
	return
}

// DeleteIntent will delete intent with provided intentID only when intent is not read only
func (dao intentDaoV2) DeleteIntent(appid string, intentID int64) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	readOnly, err := checkIntent(tx, appid, intentID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	if readOnly {
		return ErrReadOnlyIntent
	}

	queryStr := "DELETE FROM intent_train_sets WHERE intent = ?"
	_, err = tx.Exec(queryStr, intentID)
	if err != nil {
		return err
	}

	queryStr = "DELETE FROM intents WHERE id = ? AND appid = ?"
	_, err = tx.Exec(queryStr, intentID, appid)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return
}

func (dao *intentDaoV2) checkDB() {
	if dao.db == nil {
		dao.db = util.GetMainDB()
	}
}

func checkIntent(tx db, appid string, intentID int64) (readOnly bool, err error) {
	// when error happen, return false
	if tx == nil {
		return false, errors.New("error parameter")
	}
	queryStr := "SELECT version FROM intents WHERE appid = ? AND id = ?"
	row := tx.QueryRow(queryStr, appid, intentID)

	var version *int
	err = row.Scan(&version)
	if err != nil {
		return false, err
	}

	return version != nil, nil
}

func getLikeValue(val string) string {
	return fmt.Sprintf("%%%s%%", strings.Replace(val, "%", "\\%", -1))
}

type db interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
