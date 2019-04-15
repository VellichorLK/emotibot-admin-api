package data

import (
	"time"
)

const (
	SyncTaskModeReset = iota
	SyncTaskModeUpdate
)

const (
	SyncTaskStatusRunning = iota
	SyncTaskStatusFinished
	SyncTaskStatusExpired
	SyncTaskStatusError
)

const (
	AutofillModuleIntent = "autofill_intent"
)

const ExpiredDuration = 15 * time.Minute
