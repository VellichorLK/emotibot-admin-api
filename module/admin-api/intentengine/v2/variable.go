package intentenginev2

import (
	"errors"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/util"
)

const (
	typePositive = iota
	typeNegative
)

const (
	trainResultInit    = 0
	trainResultSuccess = 1
	trainResultFail    = -1
)

// used in return
const (
	statusNeedTrain = "NEED_TRAIN"
	statusTraining  = "TRAINING"
	statusFinish    = "TRAINED"
	statusFail      = "TRAIN_FAILED"
)

// used when check return from intent-engine
const (
	statusIETraining   = "training"
	statusIETrainReady = "ready"
	statusIETrainError = "error"
)

const (
	defaultIntentEngineURL = "http://127.0.0.1:15001"
	defaultRuleEngineURL   = "http://127.0.0.1:15002"
)

const (
	typeBFOP = iota
	typeBF2
	typeBF2New
)

var (
	// moduleName used to get correct environment name
	moduleName = "intents"
	// EntryList will be merged in the module controller
	EntryList = []util.EntryPoint{
		util.NewEntryPointWithVer("GET", "intents", []string{"view"}, handleGetIntentsV2, 2),
		util.NewEntryPointWithVer("POST", "intent", []string{"create"}, handleAddIntentV2, 2),
		util.NewEntryPointWithVer("GET", "intent/{intentID}", []string{"view"}, handleGetIntentV2, 2),
		util.NewEntryPointWithVer("PATCH", "intent/{intentID}", []string{"edit"}, handleUpdateIntentV2, 2),
		util.NewEntryPointWithVer("DELETE", "intent/{intentID}", []string{"delete"}, handleDeleteIntentV2, 2),
		util.NewEntryPointWithVer("POST", "delete", []string{"delete"}, handleDeleteMultiIntentV2, 2),

		util.NewEntryPointWithVer("GET", "intent/{intentID}/sentence/search", []string{"view"}, handleSearchSentence, 2),
		util.NewEntryPointWithVer("GET", "sentence/search", []string{"view"}, handleSearchSentence, 2),

		util.NewEntryPointWithVer("POST", "train", []string{"view"}, handleStartTrain, 2),
		util.NewEntryPointWithVer("GET", "status", []string{"view"}, handleGetIntentStatusV2, 2),

		util.NewEntryPointWithVer("GET", "getData", []string{}, handleGetTrainDataV2, 2),
		util.NewEntryPointWithVer("POST", "import", []string{"view"}, handleImportIntentV2, 2),
		util.NewEntryPointWithVer("GET", "export", []string{}, handleExportIntentV2, 2),
	}
)

var (
	// ErrReadOnlyIntent means trying to modify intent which is trained (version is not NULL)
	ErrReadOnlyIntent   = errors.New("intent is readonly if it is trained")
	dao                 intentDaoInterface
	defaultTrainTimeout = 30 * 60 // 30 minutes
)

func getTrainTimeout() int {
	str := getEnvironment("INTENT_TRAIN_TIMEOUT")
	timeout, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultTrainTimeout
	}
	return int(timeout)
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(moduleName)
}

func getEnvironment(key string) string {
	envs := util.GetEnvOf(moduleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}
