package data

import (
	"time"
)

const MaxLatestTests = 5

const (
	TestResultNotTested = iota
	TestResultCorrect
	TestResultWrong
)

const (
	TestStatusTesting = iota
	TestStatusTested
	TestStatusFailed
	TestStatusNeedTest
	TestStatusPending
)

var TestTaskExpiredDuration = 30 * time.Minute

const (
	MaxRecentTrained = 5
	MaxRecentTested  = 5
	MaxRecentSaved   = 10
)
