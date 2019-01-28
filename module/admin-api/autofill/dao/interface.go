package dao

import (
	"database/sql"
	"errors"

	"emotibot.com/emotigo/module/admin-api/autofill/data"
)

type AutofillDao struct {
	db *sql.DB
}

var (
	// ErrDBNotInit is used to be returned if dao is not initialized
	ErrDBNotInit = errors.New("DB is not init")
)

type Dao interface {
	TryGetSyncTaskLock(appID string, taskMode int64) (result bool, resetFailed bool, err error)
	SyncTaskFinish(appID string, taskErr error) error
	ShouldRerunSyncTask(appID string) (bool, error)
	RerunSyncTask(appID string) error
	UpdateSyncTaskMode(appID string, newMode int64) error
	GetTECurrentIntents(appID string) ([]string, error)
	GetTEPrevIntents(appID string) ([]string, error)
	GetTECurrentIntentIDs(appID string) ([]int64, error)
	UpdateTEPrevIntents(appID string, intents []string) error
	DeleteAllTEPrevIntents(appID string) error
	GetIntentSentences(appID string, intents []string, fromID int64) (sentences []*data.Sentence, err error)
}
