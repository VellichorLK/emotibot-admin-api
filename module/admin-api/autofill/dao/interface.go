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
	TryStartSyncProcess(appID string) (result bool, err error)
	SyncTaskFinish(appID string, errMsg *string) error
	ShouldRerunSyncTask(appID string) (bool, error)
	RerunSyncTask(appID string) error
	UpdateSyncTaskMode(appID string, newMode int64) error
	GetTECurrentIntents(appID string) ([]string, error)
	GetTEPrevIntents(appID string) ([]string, error)
	UpdateTEPrevIntents(appID string, intents []string) error
	DeleteAllTEPrevIntents(appID string) error
	GetIntentSentences(appID string, intents []string) (sentences []*data.Sentence, err error)
}
