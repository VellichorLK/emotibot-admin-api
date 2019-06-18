package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/util"

	"github.com/satori/go.uuid"
)

const (
	TagTypeTable           = "tag_type"
	TagsTable              = "tags"
	FaqCategoryTable       = "tbl_sq_category"
	FaqCategoryTableV3     = "faq3.tbl_folder"
	FaqRobotTagTable       = "tbl_robot_tag"
	RecordsExportTaskTable = "records_export_tasks"
	RecordsExportTable     = "records_export"
)

func GetTags() (map[string]map[string][]data.Tag, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(`
		SELECT p.code, t.code, t.name, t.app_id
		FROM %s AS p
		INNER JOIN %s AS t
		WHERE t.type = p.id`, TagTypeTable, TagsTable)

	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]map[string][]data.Tag)

	for rows.Next() {
		var tagType string
		var tagAppID string
		tag := data.Tag{}
		err = rows.Scan(&tagType, &tag.Code, &tag.Name, &tagAppID)
		if err != nil {
			return nil, err
		}

		_, ok := tags[tagAppID]
		if !ok {
			tags[tagAppID] = make(map[string][]data.Tag)
		}

		_, ok = tags[tagAppID][tagType]
		if !ok {
			tags[tagAppID][tagType] = make([]data.Tag, 0)
		}

		tags[tagAppID][tagType] = append(tags[tagAppID][tagType], tag)
	}

	return tags, nil
}

func GetFaqCategoryPathByID(categoryID int64) (categoryPath string, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`SELECT path FROM %s WHERE id = ?`, FaqCategoryTable)
	queryParams := []interface{}{categoryID}

	row := db.QueryRow(queryStr, queryParams...)

	err = row.Scan(&categoryPath)
	if err != nil && err == sql.ErrNoRows {
		err = data.ErrFaqCategoryPathNotFound
	}

	return
}

func GetFaqCategoryPathsByIDs(categoryIDs []int64) (categoryPaths map[int64]*data.FaqCategoryPath,
	err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	if len(categoryIDs) > 0 {
		ids := strings.Repeat(", ?", len(categoryIDs)-1)
		queryStr := fmt.Sprintf(`SELECT id, path FROM %s WHERE id IN (?%s`, FaqCategoryTable, ids)
		queryParams := make([]interface{}, len(categoryIDs))

		for i, categoryID := range categoryIDs {
			queryParams[i] = categoryID
		}

		rows, err := db.Query(queryStr, queryParams...)
		if err != nil {
			return nil, err
		}

		categoryPaths = make(map[int64]*data.FaqCategoryPath)

		for rows.Next() {
			categoryPath := data.FaqCategoryPath{}
			err = rows.Scan(&categoryPath.ID, &categoryPath.Path)
			if err != nil {
				return nil, err
			}

			categoryPaths[categoryPath.ID] = &categoryPath
		}
	}

	return
}

func GetAllFaqCategoryPaths() (categoryPaths map[int64]*data.FaqCategoryPath, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`SELECT id, path FROM %s`, FaqCategoryTable)
	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	categoryPaths = make(map[int64]*data.FaqCategoryPath)

	for rows.Next() {
		categoryPath := data.FaqCategoryPath{}
		err = rows.Scan(&categoryPath.ID, &categoryPath.Path)
		if err != nil {
			return nil, err
		}

		categoryPaths[categoryPath.ID] = &categoryPath
	}

	return
}

func GetAllFaqCategoryPathsV3() (categoryPaths map[int64]*data.FaqCategoryPath, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`SELECT id, substring_index(fullname,',', 1) as fullname FROM %s`, FaqCategoryTableV3)
	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	categoryPaths = make(map[int64]*data.FaqCategoryPath)

	for rows.Next() {
		categoryPath := data.FaqCategoryPath{}
		err = rows.Scan(&categoryPath.ID, &categoryPath.Path)
		if err != nil {
			return nil, err
		}

		categoryPaths[categoryPath.ID] = &categoryPath
	}

	return
}

func GetFaqRobotTagByID(robotTagID int64) (robotTag string, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`SELECT name FROM %s WHERE id = ?`, FaqRobotTagTable)
	queryParams := []interface{}{robotTagID}

	row := db.QueryRow(queryStr, queryParams...)

	err = row.Scan(&robotTag)
	if err != nil && err == sql.ErrNoRows {
		err = data.ErrFaqRobotTagNotFound
	}

	return
}

func GetFaqRobotTagsByIDs(robotTagIDs []int64) (robotTags map[int64]*data.FaqRobotTag, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	if len(robotTagIDs) > 0 {
		ids := strings.Repeat(", ?", len(robotTagIDs)-1)
		queryStr := fmt.Sprintf(`SELECT id, name FROM %s WHERE id IN (?%s`, FaqRobotTagTable, ids)
		queryParams := make([]interface{}, len(robotTagIDs))

		for i, robotTagID := range robotTagIDs {
			queryParams[i] = robotTagID
		}

		rows, err := db.Query(queryStr, queryParams...)
		if err != nil {
			return nil, err
		}

		robotTags = make(map[int64]*data.FaqRobotTag)

		for rows.Next() {
			robotTag := data.FaqRobotTag{}
			err = rows.Scan(&robotTag.ID, &robotTag.Tag)
			if err != nil {
				return nil, err
			}

			robotTags[robotTag.ID] = &robotTag
		}
	}

	return
}

func GetAllFaqRobotTags() (robotTags map[int64]*data.FaqRobotTag, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`SELECT id, name FROM %s`, FaqRobotTagTable)
	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	robotTags = make(map[int64]*data.FaqRobotTag)

	for rows.Next() {
		robotTag := data.FaqRobotTag{}
		err = rows.Scan(&robotTag.ID, &robotTag.Tag)
		if err != nil {
			return nil, err
		}

		robotTags[robotTag.ID] = &robotTag
	}

	return
}

func TryCreateExportTask(enterpriseID string) (exportTaskID string, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	tx, _ := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	/* Try to create export task for given enterprise
	 * We only allow a single exporting task in process
	 * if there's already an exporting task in process, return error
	 */
	queryStr := fmt.Sprintf(`
		INSERT INTO %s (enterprise_id)
		VALUES (?)`, RecordsExportTaskTable)
	queryParams := []interface{}{enterpriseID}

	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		if strings.Contains(err.Error(), "1062") {
			/* Error number: 1062; Symbol: ER_DUP_ENTRY; SQLSTATE: 23000
			 * enterprise_id already exists:
			 * Currently there's already an exporting task in process
			 */
			err = data.ErrExportTaskInProcess
		}

		return
	}

	// Create exporting task
	taskUUID, err := uuid.NewV4()
	if err != nil {
		return
	}

	exportTaskID = taskUUID.String()

	queryStr = fmt.Sprintf(`
		INSERT INTO %s (uuid, status, created_at)
		VALUEs (?, ?, ?)`, RecordsExportTable)
	queryParams = []interface{}{exportTaskID, data.CodeExportTaskRunning, time.Now().Unix()}

	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return
	}

	queryStr = fmt.Sprintf(`
		UPDATE %s
		SET task_id = ?
		WHERE enterprise_id = ?`, RecordsExportTaskTable)
	queryParams = []interface{}{exportTaskID, enterpriseID}

	_, err = tx.Exec(queryStr, queryParams...)
	return
}

func ExportTaskCompleted(taskID string, filePath string) (err error) {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	/* Update RecordsExportTaskTableand and delete the correspond record in RecordsExportTable
	 * to allow other records exporting tasks to be proccessed
	 */
	tx, _ := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET path = ?, status = ?
		WHERE uuid = ?`, RecordsExportTable)
	queryParams := []interface{}{filePath, data.CodeExportTaskCompleted, taskID}

	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return
	}

	err = deleteExportTaskWithTx(tx, taskID)
	return
}

func ExportTaskFailed(taskID string, errReason string) (err error) {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	/* Update RecordsExportTaskTableand and delete the correspond record in RecordsExportTable
	 * to allow other records exporting tasks to be proccessed
	 */
	tx, _ := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET status = ?, err_reason = ?
		WHERE uuid = ?`, RecordsExportTable)
	queryParams := []interface{}{data.CodeExportTaskFailed, errReason, taskID}

	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return
	}

	err = deleteExportTaskWithTx(tx, taskID)
	return err
}

func ExportTaskEmpty(taskID string) (err error) {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	/* Update RecordsExportTaskTableand and delete the correspond record in RecordsExportTable
	 * to allow other records exporting tasks to be proccessed
	 */
	tx, _ := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET status = ?
		WHERE uuid = ?`, RecordsExportTable)
	queryParams := []interface{}{data.CodeExportTaskEmpty, taskID}

	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return
	}

	err = deleteExportTaskWithTx(tx, taskID)
	return err
}

func DeleteExportTask(taskID string) (err error) {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	tx, _ := db.Begin()
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
		WHERE uuid = ?`, RecordsExportTable)
	queryParams := []interface{}{taskID}

	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE task_id = ?`, RecordsExportTaskTable)

	_, err = tx.Exec(queryStr, queryParams...)
	return err
}

func deleteExportTaskWithTx(tx *sql.Tx, taskID string) error {
	queryStr := fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE task_id = ?`, RecordsExportTaskTable)
	queryParams := []interface{}{taskID}

	_, err := tx.Exec(queryStr, queryParams...)
	return err
}

// RemoveAllOutdatedExportTasks removes all records before timestamp
func RemoveAllOutdatedExportTasks(timestamp int64) error {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	// Get all outdated tasks
	queryStr := fmt.Sprintf(`
		SELECT uuid
		FROM %s
		WHERE created_at < ?`, RecordsExportTable)
	queryParams := []interface{}{timestamp}

	rows, err := db.Query(queryStr, queryParams...)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	deleteTaskIDs := make([]interface{}, 0)

	for rows.Next() {
		var taskID string

		err = rows.Scan(&taskID)
		if err != nil {
			return err
		}

		deleteTaskIDs = append(deleteTaskIDs, taskID)
	}

	if len(deleteTaskIDs) > 0 {
		// Delete all outdated tasks
		tx, _ := db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		queryStr = fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE uuid IN (?%s)`, RecordsExportTable, strings.Repeat(",?", len(deleteTaskIDs)-1))

		_, err = tx.Exec(queryStr, deleteTaskIDs...)
		if err != nil {
			return err
		}

		queryStr = fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE task_id IN (?%s)`, RecordsExportTaskTable, strings.Repeat(",?", len(deleteTaskIDs)-1))

		_, err = tx.Exec(queryStr, deleteTaskIDs...)
		return err
	}

	return nil
}

func GetExportTaskStatus(taskID string) (status int, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`
		SELECT status
		FROM %s
		WHERE uuid = ?`, RecordsExportTable)
	queryParams := []interface{}{taskID}

	row := db.QueryRow(queryStr, queryParams...)

	err = row.Scan(&status)
	if err != nil && err == sql.ErrNoRows {
		err = data.ErrExportTaskNotFound
	}

	return
}

func GetExportRecordsFile(taskID string) (path string, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`
		SELECT path
		FROM %s
		WHERE uuid = ?`, RecordsExportTable)
	queryParams := []interface{}{taskID}

	row := db.QueryRow(queryStr, queryParams...)

	var pathNullable sql.NullString
	err = row.Scan(&pathNullable)
	if err != nil && err == sql.ErrNoRows {
		err = data.ErrExportTaskNotFound
	}

	path = pathNullable.String
	return
}

// UnlockAllEnterprisesExportTask removes all records in RecordsExportTaskTable
// to prevent any enterprise locked due to the crash of admin-api
// This function should only be called at init() stage
// Note: This scenario is not well-designed for distributed system,
//       if any of admin-api node crash and restart, it will unlock all enterprises
//		 We should replace this mechanism with distributed lock (e.g. Redis, Zookeeper)
func UnlockAllEnterprisesExportTask() error {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("TRUNCATE TABLE %s", RecordsExportTaskTable)
	_, err := db.Exec(queryStr)
	return err
}

func ExportRecordsExists(taskID string) (bool, error) {
	queryStr := fmt.Sprintf("SELECT 1 FROM %s WHERE uuid = ?",
		RecordsExportTable)
	return rowExists(queryStr, taskID)
}

func rowExists(query string, args ...interface{}) (bool, error) {
	db := util.GetMainDB()
	if db == nil {
		return false, errors.New("DB not init")
	}

	var exists bool
	queryStr := fmt.Sprintf("SELECT EXISTS (%s)", query)
	err := db.QueryRow(queryStr, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}
