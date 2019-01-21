package dao

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/autofill/data"
	uuid "github.com/satori/go.uuid"
)

const (
	SyncTable                  = "autofill_sync"
	SyncTasksTable             = "autofill_sync_tasks"
	TaskEngineIntentsTable     = "taskengine_intents"
	TaskEngineIntentsPrevTable = "taskengine_intents_prev"
	IntentsTable               = "intents"
	IntentTrainSetsTable       = "intent_train_sets"
)

func NewDao(db *sql.DB) *AutofillDao {
	return &AutofillDao{
		db: db,
	}
}

// TryGetSyncTaskLock tries to get the lock for autofill sync task,
// 'result' flag is set to true if retrieves the lock successfully
// and the caller is responsible for creating the sync task
// 'resetFailed' flag is set to true if previous sync task was a reset task
// and was failed
func (dao *AutofillDao) TryGetSyncTaskLock(appID string, taskMode int64) (result bool,
	resetFailed bool, err error) {
	if dao.db == nil {
		return false, false, ErrDBNotInit
	}

	createNewSyncTask := func(tx *sql.Tx, _startTime int64) (_result bool, _err error) {
		if tx == nil {
			tx, _ = dao.db.Begin()
			defer func() {
				if _err != nil {
					tx.Rollback()
				} else {
					tx.Commit()
				}
			}()
		}

		insertQuery := fmt.Sprintf(`
        	INSERT INTO %s (app_id, start_time, mode)
			VALUES (?, ?, ?)`, SyncTable)

		_, _err = tx.Exec(insertQuery, appID, _startTime, taskMode)
		if _err != nil {
			if strings.Contains(err.Error(), "1062") {
				// Error number: 1062; Symbol: ER_DUP_ENTRY; SQLSTATE: 23000
				// app_id already exists:
				// Currently there's already a newer sync task started by other process,
				// do nothing
				return false, nil
			}

			return false, _err
		}

		// Sync task lock retrieved successfully
		// Create sync task
		syncTaskID, _err := createSyncTaskWithTx(tx, _startTime)
		if _err != nil {
			return false, _err
		}

		updateQuery := fmt.Sprintf(`
			UPDATE %s
			SET task_id = ?
			WHERE app_id = ?`, SyncTable)

		_, _err = tx.Exec(updateQuery, syncTaskID, appID)
		if _err != nil {
			return false, _err
		}

		return true, nil
	}

	updateSyncTaskStatus := func(_startTime int64, _rerun bool) (_result bool, _err error) {
		tx, _ := dao.db.Begin()
		defer func() {
			if _err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		updateQuery := fmt.Sprintf(`
        	UPDATE %s
        	SET start_time = ?, rerun = ?
			WHERE app_id = ?`, SyncTable)

		result, _err := tx.Exec(updateQuery, _startTime, _rerun, appID)
		if _err != nil {
			return false, _err
		}

		affects, _err := result.RowsAffected()
		if _err != nil {
			return false, _err
		}

		if affects == 0 {
			// Previous sync task was finished (thus the row was deleted),
			// create a new sync task
			return createNewSyncTask(tx, _startTime)
		}

		if !_rerun {
			// Sync task lock retrieved successfully
			// Create sync task
			syncTaskID, _err := createSyncTaskWithTx(tx, _startTime)
			if _err != nil {
				return false, _err
			}

			updateQuery := fmt.Sprintf(`
				UPDATE %s
				SET task_id = ?
				WHERE app_id = ?`, SyncTable)

			_, _err = tx.Exec(updateQuery, syncTaskID, appID)
			if _err != nil {
				return false, _err
			}

			return true, nil
		}

		// Update rerun flag successfully
		return false, nil
	}

	var startTime, mode, status int64
	var rerun bool
	var taskNullableID sql.NullInt64

	queryStr := fmt.Sprintf(`
        SELECT t.start_time, t.rerun, t.task_id, t.mode, s.status
		FROM %s AS t
		INNER JOIN %s AS s
		ON t.task_id = s.id
        WHERE t.app_id = ?`, SyncTable, SyncTasksTable)
	err = dao.db.QueryRow(queryStr, appID).Scan(&startTime, &rerun, &taskNullableID,
		&mode, &status)
	if err != nil && err != sql.ErrNoRows {
		return false, false, err
	}

	now := time.Now().Unix()
	expiredTime := startTime + int64(data.ExpiredDuration.Seconds())

	if err == sql.ErrNoRows {
		// No running sync task, create a new one
		result, err = createNewSyncTask(nil, now)
		return result, false, err
	} else if startTime > now {
		// Newer sync task is running, do nothing
		return false, false, nil
	} else if status == data.SyncTaskStatusError {
		// Previous sync task was failed,
		// create a new sync task
		// and set 'resetFailed' flag to true if previous sync task was a reset task
		result, err = updateSyncTaskStatus(now, false)
		if mode == data.SyncTaskModeReset {
			resetFailed = true
		}
		return
	} else if expiredTime > now && !rerun {
		// Previous sync task is still running,
		// set the rerun flag to true to let previous sync task
		// rerun the sync task after it finishes
		result, err = updateSyncTaskStatus(startTime, true)
		return result, false, err
	} else if expiredTime < now {
		// Previous sync task has expired,
		// mark previous sync task as expired
		// and create a new sync task
		if taskNullableID.Valid {
			taskID := taskNullableID.Int64
			err = dao.markTaskExpired(taskID)
			if err != nil {
				return false, false, err
			}
		}
		result, err = updateSyncTaskStatus(now, false)
		return
	} else {
		// Rerun flag is set to true already
		// and previous sync task has not expired yet,
		// do nothing
		return false, false, nil
	}
}

// SyncTaskFinish delete the correspond row in task status table
// which indicated the current sync task has completed or failed
func (dao *AutofillDao) SyncTaskFinish(appID string, taskErr error) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	var queryStr string
	var queryParams []interface{}
	var taskID int64

	queryStr = fmt.Sprintf(`
		SELECT task_id
		FROM %s
		WHERE app_id = ?`, SyncTable)

	err := dao.db.QueryRow(queryStr, appID).Scan(&taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if taskErr != nil {
		queryStr = fmt.Sprintf(`
			UPDATE %s
			SET status = ?, message = ?
			WHERE id = ?`, SyncTasksTable)
		queryParams = []interface{}{data.TaskStatusError, taskErr.Error(), taskID}
		_, err := tx.Exec(queryStr, queryParams...)
		if err != nil {
			return err
		}
	} else {
		queryStr = fmt.Sprintf(`
			UPDATE %s
			SET status = ?
			WHERE id = ?`, SyncTasksTable)
		queryParams = []interface{}{data.TaskStatusFinished, taskID}
		_, err := tx.Exec(queryStr, queryParams...)
		if err != nil {
			return err
		}
	}

	queryStr = fmt.Sprintf(`
        DELETE
        FROM %s
        WHERE app_id = ?`, SyncTable)
	_, err = tx.Exec(queryStr, appID)
	return err
}

func createSyncTaskWithTx(tx *sql.Tx, _startTime int64) (taskID int64, _err error) {
	// Create sync task
	taskUUID, _err := uuid.NewV4()
	if _err != nil {
		return 0, _err
	}

	syncTaskID := taskUUID.String()

	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (uuid, start_time)
		VALUES (?, ?)`, SyncTasksTable)

	result, _err := tx.Exec(insertQuery, syncTaskID, _startTime)
	if _err != nil {
		return 0, _err
	}

	taskID, _err = result.LastInsertId()
	return
}

func (dao *AutofillDao) markTaskExpired(taskID int64) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET status = ?
		WHERE id = ?`, SyncTasksTable)
	_, err := dao.db.Exec(queryStr, data.TaskStatusExpired, taskID)
	return err
}

// ShouldRerunSyncTask checks whether should rerun the sync task again or not
func (dao *AutofillDao) ShouldRerunSyncTask(appID string) (bool, error) {
	if dao.db == nil {
		return false, ErrDBNotInit
	}

	var rerun bool

	queryStr := fmt.Sprintf("SELECT rerun FROM %s WHERE app_id = ?", SyncTable)
	err := dao.db.QueryRow(queryStr, appID).Scan(&rerun)
	if err != nil {
		return false, err
	}

	return rerun, nil
}

// RerunSyncTask marks rerun flag to false, update start time and create new sync task
func (dao *AutofillDao) RerunSyncTask(appID string) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	now := time.Now().Unix()

	syncTaskID, err := createSyncTaskWithTx(tx, now)
	if err != nil {
		return err
	}

	updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET rerun = ?, start_time = ?, task_id = ?
		WHERE app_id = ?`, SyncTable)

	_, err = tx.Exec(updateQuery, false, now, syncTaskID, appID)
	return err
}

// UpdateSyncTaskMode updates the mode of sync task
func (dao *AutofillDao) UpdateSyncTaskMode(appID string, newMode int64) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET mode = ?
		WHERE app_id = ?`, SyncTable)
	_, err := dao.db.Exec(queryStr, newMode, appID)
	return err
}

// GetTECurrentIntents returns current intent names used by Task Engine
func (dao *AutofillDao) GetTECurrentIntents(appID string) ([]string, error) {
	return dao.getIntents(appID, TaskEngineIntentsTable)
}

// GetTEPrevIntents returns previous intent names used by Task Engine
func (dao *AutofillDao) GetTEPrevIntents(appID string) ([]string, error) {
	return dao.getIntents(appID, TaskEngineIntentsPrevTable)
}

func (dao *AutofillDao) getIntents(appID string, table string) ([]string, error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
	  	 SELECT name
	  	 FROM %s
	  	 WHERE app_id = ?`, table)

	rows, err := dao.db.Query(queryStr, appID)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

// GetTECurrentIntentIDs returns current intent IDs used by Task Engine
func (dao *AutofillDao) GetTECurrentIntentIDs(appID string) ([]int64, error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT i.id
		FROM %s as i
		INNER JOIN %s as t
		ON i.name = t.name
		WHERE app_id = ? AND version = (
			SELECT MAX(version)
			FROM intents
			WHERE appid = ?)`, IntentsTable, TaskEngineIntentsTable)

	rows, err := dao.db.Query(queryStr, appID, appID)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0)

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// UpdateTEPrevIntents replace all intents in taskengine_intents_prev with given intents
func (dao *AutofillDao) UpdateTEPrevIntents(appID string, intents []string) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	queryStr := fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE app_id = ?`, TaskEngineIntentsPrevTable)
	_, err = tx.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	if len(intents) == 0 {
		return nil
	}

	valuesClause := fmt.Sprintf("%s%s", strings.Repeat("(?, ?), ", len(intents)-1), "(?, ?)")
	queryParams := make([]interface{}, len(intents)*2)

	for i, intent := range intents {
		queryParams[i*2] = appID
		queryParams[i*2+1] = intent
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s (app_id, name)
		VALUES %s`, TaskEngineIntentsPrevTable, valuesClause)
	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAllTEPrevIntents deletes all intents in taskengine_intents_prev
func (dao *AutofillDao) DeleteAllTEPrevIntents(appID string) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE app_id = ?`, TaskEngineIntentsPrevTable)
	_, err := dao.db.Exec(queryStr, appID)
	return err
}

// GetIntentSentences returns the intent train set sentences
// of given app ID and intent names
// If @intents is nil, all intent train set sentences will be returned
func (dao *AutofillDao) GetIntentSentences(appID string, intents []string,
	fromID int64) (sentences []*data.Sentence, err error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	const MaxNumOfSentences = 10000

	queryStr := fmt.Sprintf(`
		SELECT i.id, s.id, sentence
		FROM %s AS i
		INNER JOIN %s AS s
		ON i.id = s.intent`, IntentsTable, IntentTrainSetsTable)
	orderClause := "ORDER by s.id ASC"
	limitClause := fmt.Sprintf("LIMIT %d", MaxNumOfSentences)

	var whereClause string
	var queryParams []interface{}

	if intents == nil {
		whereClause = fmt.Sprintf(`
			WHERE i.appid = ? AND s.id > ? AND version = (
				SELECT MAX(version)
				FROM %s
				WHERE appid = ?)`, IntentsTable)
		queryParams = []interface{}{appID, fromID, appID}
	} else {
		if len(intents) == 0 {
			return
		}

		whereClause = fmt.Sprintf(`
			WHERE i.appid = ? AND i.name IN (?%s) AND s.id > ? AND version = (
				SELECT MAX(version)
				FROM %s
				WHERE appid = ?)`,
			strings.Repeat(", ?", len(intents)-1), IntentsTable)

		queryParams = make([]interface{}, len(intents)+3)
		queryParams[0] = appID
		for i, intent := range intents {
			queryParams[i+1] = intent
		}
		queryParams[len(queryParams)-2] = fromID
		queryParams[len(queryParams)-1] = appID
	}

	queryStr = fmt.Sprintf("%s %s %s %s", queryStr, whereClause, orderClause, limitClause)

	rows, err := dao.db.Query(queryStr, queryParams...)
	if err != nil {
		return nil, err
	}

	sentences = make([]*data.Sentence, 0)

	for rows.Next() {
		sentence := data.Sentence{}
		err = rows.Scan(&sentence.ModuleID, &sentence.SentenceID,
			&sentence.Sentence)
		if err != nil {
			return nil, err
		}

		sentences = append(sentences, &sentence)
	}

	return sentences, nil
}
