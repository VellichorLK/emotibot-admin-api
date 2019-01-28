package data

import (
    "time"
)

const THIRD_CORE = "3rd_core"

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
    AutofillModuleIntent = "intent"
)

const ExpiredDuration = 15 * time.Minute
