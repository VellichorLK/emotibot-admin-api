package dao

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
)

// DB defines interface for different DAO modules
type DB interface {
	GetTags() (map[string]map[string][]data.Tag, error)
	GetFaqCategoryPathByID(categoryID string) (categoryPath string, err error)
	GetFaqCategoryPathsByIDs(categoryIDs []int64) (categoryPaths map[int64]*data.FaqCategoryPath, err error)
	GetAllFaqCategoryPaths() (categoryPaths map[int64]*data.FaqCategoryPath, err error)
	GetFaqRobotTagByID(robotTagID string) (robotTag string, err error)
	GetFaqRobotTagsByIDs(robotTagIDs []int64) (robotTags map[int64]*data.FaqRobotTag, err error)
	GetAllFaqRobotTags() (robotTags map[int64]*data.FaqRobotTag, err error)
	TryCreateExportTask(enterpriseID string) (exportTaskID string, err error)
	ExportTaskCompleted(taskID string, filePath string) (err error)
	ExportTaskFailed(taskID string, errReason string) error
	ExportTaskEmpty(taskID string) error
	DeleteExportTask(taskID string) error
	RemoveAllOutdatedExportTasks(timestamp string) error
	GetExportTaskStatus(taskID string) (status int, err error)
	GetExportRecordsFile(taskID string) (path string, err error)
	UnlockAllEnterprisesExportTask() error
	ExportTaskExists(taskID string) (bool, error)
}
