package qi

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
	"emotibot.com/emotigo/module/qic-api/util/redis"
	"emotibot.com/emotigo/pkg/api/rabbitmq/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo           util.ModuleInfo
	tagDao               model.TagDao
	callDao              model.CallDao = &model.CallSQLDao{}
	segmentDao           model.SegmentDao
	userValueDao         = &model.UserValueDao{}
	userKeyDao           = &model.UserKeySQLDao{}
	swDao                model.SensitiveWordDao
	condDao              = &model.GroupConditionDao{}
	serviceDAO           model.GroupDAO
	producer             *rabbitmq.Producer
	consumer             *rabbitmq.Consumer
	realtimeCallProducer *rabbitmq.Producer
	realtimeCallConsumer *rabbitmq.Consumer
	sqlConn              *sql.DB
	dbLike               model.DBLike
	volume               string
	creditDao            model.CreditDao = &model.CreditSQLDao{}
)
var (
	tags func(tx model.SqlLike, query model.TagQuery) ([]model.Tag, error)
)

var (
	newCondition    = condDao.NewCondition
	groups          func(delegatee model.SqlLike, query model.GroupQuery) ([]model.Group, error)
	newGroup        func(delegatee model.SqlLike, group model.Group) (model.Group, error)
	setGroupRule    func(delegatee model.SqlLike, groups ...model.Group) error
	groupRules      func(delegatee model.SqlLike, group model.Group) (conversationRules []int64, OtherGroupRules map[model.GroupRuleType][]string, err error)
	resetGroupRules func(delegatee model.SqlLike, groups ...model.Group) error
	setGroupBasic   func(delegatee model.SqlLike, group *model.Group) error
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "qi",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "groups", []string{}, handleCreateGroup),
			util.NewEntryPoint("GET", "groups", []string{}, handleGetGroups),
			util.NewEntryPoint("GET", "groups/filters", []string{}, handleGetGroupsByFilter),
			util.NewEntryPoint("GET", "groups/{group_id}", []string{}, simpleGroupRequest(handleGetGroup)),
			util.NewEntryPoint("PUT", "groups/{group_id}", []string{}, groupRequest(handleUpdateGroup)),
			util.NewEntryPoint("DELETE", "groups/{group_id}", []string{}, groupRequest(handleDeleteGroup)),

			util.NewEntryPoint("GET", "tags", []string{}, HandleGetTags),
			util.NewEntryPoint("POST", "tags", []string{}, HandlePostTags),
			util.NewEntryPoint("GET", "tags/{tag_id}", []string{}, HandleGetTag),
			util.NewEntryPoint("PUT", "tags/{tag_id}", []string{}, HandlePutTags),
			util.NewEntryPoint("DELETE", "tags/{tag_id}", []string{}, HandleDeleteTag),

			util.NewEntryPoint("GET", "sentences", []string{}, handleGetSentences),
			util.NewEntryPoint("POST", "sentences", []string{}, handleNewSentence),
			util.NewEntryPoint("GET", "sentences/{id}", []string{}, WithSenUUIDCheck(handleGetSentence)),
			util.NewEntryPoint("PUT", "sentences/{id}", []string{}, WithSenUUIDCheck(handleModifySentence)),
			util.NewEntryPoint("DELETE", "sentences/{id}", []string{}, WithSenUUIDCheck(handleDeleteSentence)),
			util.NewEntryPoint("PUT", "sentences/move/{id}", []string{}, handleMoveSentence),

			util.NewEntryPoint("POST", "sentence-groups", []string{}, handleCreateSentenceGroup),
			util.NewEntryPoint("GET", "sentence-groups", []string{}, handleGetSentenceGroups),
			util.NewEntryPoint("GET", "sentence-groups/{id}", []string{}, handleGetSentenceGroup),
			util.NewEntryPoint(http.MethodPut, "sentence-groups/{id}", []string{}, handleUpdateSentenceGroup),
			util.NewEntryPoint("DELETE", "sentence-groups/{id}", []string{}, handleDeleteSentenceGroup),

			util.NewEntryPoint("POST", "conversation-flow", []string{}, handleCreateConversationFlow),
			util.NewEntryPoint("GET", "conversation-flow", []string{}, handleGetConversationFlows),
			util.NewEntryPoint("GET", "conversation-flow/{id}", []string{}, handleGetConversationFlow),
			util.NewEntryPoint(http.MethodPut, "conversation-flow/{id}", []string{}, handleUpdateConversationFlow),
			util.NewEntryPoint("DELETE", "conversation-flow/{id}", []string{}, handleDeleteConversationFlow),

			util.NewEntryPoint("POST", "rules", []string{}, handleCreateConversationRule),
			util.NewEntryPoint("GET", "rules", []string{}, handleGetConversationRules),
			util.NewEntryPoint("GET", "rules/{id}", []string{}, handleGetConversationRule),
			util.NewEntryPoint("PUT", "rules/{id}", []string{}, handleUpdateConversationRule),
			util.NewEntryPoint("DELETE", "rules/{id}", []string{}, handleDeleteConversationRule),

			util.NewEntryPoint(http.MethodGet, "category", []string{}, handleGetCategoryies),
			util.NewEntryPoint(http.MethodPost, "category", []string{}, handleCreateCategory),
			util.NewEntryPoint(http.MethodGet, "category/{id}", []string{}, handleGetCategory),
			util.NewEntryPoint(http.MethodPut, "category/{id}", []string{}, handleUpdateCatgory),
			util.NewEntryPoint(http.MethodDelete, "category/{id}", []string{}, handleDeleteCategory),

			util.NewEntryPoint(http.MethodGet, "calls", []string{}, CallsHandler),
			util.NewEntryPoint(http.MethodPost, "calls", []string{}, NewCallsHandler),
			util.NewEntryPoint(http.MethodGet, "calls/{id}", []string{}, callRequest(CallsDetailHandler)),
			util.NewEntryPoint(http.MethodPost, "calls/{id}/file", []string{}, callRequest(UpdateCallsFileHandler)),
			util.NewEntryPoint(http.MethodGet, "calls/{id}/file", []string{}, callRequest(CallsFileHandler)),
			util.NewEntryPoint(http.MethodGet, "calls/{id}/credits", []string{}, WithCallIDCheck(handleGetCredit)),

			util.NewEntryPoint(http.MethodPost, "train/model", []string{}, handleTrainAllTags),
			util.NewEntryPoint(http.MethodGet, "train/model", []string{}, handleTrainStatus),
			util.NewEntryPoint(http.MethodGet, "train/model/training", []string{}, handleTrainingStatus),
			//util.NewEntryPoint(http.MethodDelete, "manual/use/all/tags", []string{}, handleUnload),

			util.NewEntryPoint(http.MethodGet, "call-in/navigation/{id}", []string{}, handleGetFlowSetting),
			util.NewEntryPoint(http.MethodPut, "call-in/navigation/{id}/name", []string{}, handleModifyFlow),
			util.NewEntryPoint(http.MethodDelete, "call-in/navigation/{id}", []string{}, handleDeleteFlow),
			util.NewEntryPoint(http.MethodPost, "call-in/navigation", []string{}, handleNewFlow),
			util.NewEntryPoint(http.MethodGet, "call-in/navigation", []string{}, handleFlowList),
			util.NewEntryPoint(http.MethodPut, "call-in/navigation/{id}/intent", []string{}, handleModifyIntent),
			util.NewEntryPoint(http.MethodPost, "call-in/navigation/{id}/node", []string{}, handleNewNode),
			util.NewEntryPoint(http.MethodPut, "call-in/navigation/{id}/node/order", []string{}, handleNodeOrder),

			util.NewEntryPoint(http.MethodPost, "call-in/conversation", []string{}, handleFlowCreate),
			util.NewEntryPoint(http.MethodPut, "call-in/{id}", []string{}, WithFlowCallIDEnterpriseCheck(handleFlowFinish)),
			util.NewEntryPoint(http.MethodPatch, "call-in/{id}", []string{}, callRequest(handleFlowUpdate)),
			util.NewEntryPoint(http.MethodPost, "call-in/{id}/append", []string{}, handleStreaming),
			//util.NewEntryPoint(http.MethodGet, "call-in/{id}", []string{}, handleGetCurCheck),

			util.NewEntryPoint(http.MethodGet, "backup/groups", []string{}, handleExportGroups),
			util.NewEntryPoint(http.MethodPost, "restore/groups", []string{}, handleImportGroups),
			util.NewEntryPoint(http.MethodGet, "export/calls", []string{}, handleExportCalls),
			util.NewEntryPoint(http.MethodPost, "import/tags", []string{}, handleImportTags),
			util.NewEntryPoint(http.MethodPost, "import/sentences", []string{}, handleImportSentences),
			util.NewEntryPoint(http.MethodPost, "import/rules", []string{}, handleImportRules),
			util.NewEntryPoint(http.MethodPost, "import/call-in", []string{}, handleImportCallIn),

			util.NewEntryPoint(http.MethodPost, "rule/silence", []string{}, handleNewRuleSilence),
			util.NewEntryPoint(http.MethodGet, "rule/silence", []string{}, handleGetRuleSilenceList),
			util.NewEntryPoint(http.MethodGet, "rule/silence/{id}", []string{}, handleGetRuleSilence),
			util.NewEntryPoint(http.MethodDelete, "rule/silence/{id}", []string{}, handleDeleteRuleSilence),
			util.NewEntryPoint(http.MethodPut, "rule/silence/{id}/name", []string{}, handleModifyRuleSilence),
			util.NewEntryPoint(http.MethodPut, "rule/silence/{id}/condition", []string{}, handleModifyRuleSilence),
			util.NewEntryPoint(http.MethodPut, "rule/silence/{id}/exception/before", []string{}, handleExceptionRuleSilenceBefore),
			util.NewEntryPoint(http.MethodPut, "rule/silence/{id}/exception/after", []string{}, handleExceptionRuleSilenceAfter),

			util.NewEntryPoint(http.MethodPost, "rule/speed", []string{}, handleNewRuleSpeed),
			util.NewEntryPoint(http.MethodGet, "rule/speed", []string{}, handleGetRuleSpeedList),
			util.NewEntryPoint(http.MethodGet, "rule/speed/{id}", []string{}, handleGetRuleSpeed),
			util.NewEntryPoint(http.MethodDelete, "rule/speed/{id}", []string{}, handleDeleteRuleSpeed),
			util.NewEntryPoint(http.MethodPut, "rule/speed/{id}/name", []string{}, handleModifyRuleSpeed),
			util.NewEntryPoint(http.MethodPut, "rule/speed/{id}/condition", []string{}, handleModifyRuleSpeed),
			util.NewEntryPoint(http.MethodPut, "rule/speed/{id}/exception/under", []string{}, handleExceptionRuleSpeedUnder),
			util.NewEntryPoint(http.MethodPut, "rule/speed/{id}/exception/over", []string{}, handleExceptionRuleSpeedOver),

			util.NewEntryPoint(http.MethodPost, "rule/interposal", []string{}, handleNewRuleInterposal),
			util.NewEntryPoint(http.MethodGet, "rule/interposal", []string{}, handleGetRuleInterposalList),
			util.NewEntryPoint(http.MethodGet, "rule/interposal/{id}", []string{}, handleGetRuleInterposal),
			util.NewEntryPoint(http.MethodDelete, "rule/interposal/{id}", []string{}, handleDeleteRuleInterposal),
			util.NewEntryPoint(http.MethodPut, "rule/interposal/{id}", []string{}, handleModifyRuleInterposal),

			util.NewEntryPoint(http.MethodGet, "testing/predict/sentences", []string{}, handlePredictSentences),
			util.NewEntryPoint(http.MethodGet, "testing/sentences/{id}", []string{}, handleGetTestSentences),
			util.NewEntryPoint(http.MethodPost, "testing/sentences", []string{}, handleNewTestSentence),
			util.NewEntryPoint(http.MethodDelete, "testing/sentences/{id}", []string{}, handleDeleteTestSentence),
			util.NewEntryPoint(http.MethodGet, "testing/overview/sentences", []string{}, handleGetSentenceTestOverview),
			util.NewEntryPoint(http.MethodGet, "testing/overview/sentences_info", []string{}, handleGetSentenceTestResult),
			util.NewEntryPoint(http.MethodGet, "testing/overview/sentences_detail/{id}", []string{}, handleGetSentenceTestDetail),
		},
		OneTimeFunc: map[string]func(){
			"init volume": func() {
				volume, _ = ModuleInfo.Environments["FILE_VOLUME"]
				if volume == "" {
					logger.Error.Println("module env \"FILE_VOLUME\" does not exist or empty, upload function will not work.")
					return
				}
				//path.Clean will treat empty as current dir, we dont want this result
				volume = path.Clean(volume)
				if info, err := os.Stat(volume); os.IsNotExist(err) {
					logger.Error.Println(volume + " is not exist.")
					volume = ""
				} else if !info.IsDir() {
					logger.Error.Println(volume + " is not a dir.")
					volume = ""
				}
				logger.Info.Println("volume: ", volume, "is recognized.")

			},
			"init db & rabbitmq": func() {
				envs := ModuleInfo.Environments

				url := envs["MYSQL_URL"]
				user := envs["MYSQL_USER"]
				pass := envs["MYSQL_PASS"]
				db := envs["MYSQL_DB"]

				newConn, err := util.InitDB(url, user, pass, db)
				sqlConn = newConn
				if err != nil {
					logger.Error.Printf("Cannot init qi db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
					return
				}

				dbLike = &model.DefaultDBLike{
					DB: sqlConn,
				}
				// init group dao
				groupSqlDao := model.NewGroupSQLDao(dbLike)
				serviceDAO = groupSqlDao
				groups = groupSqlDao.Group
				newGroup = groupSqlDao.NewGroup
				setGroupRule = groupSqlDao.SetGroupRules
				groupRules = groupSqlDao.GroupRules
				resetGroupRules = groupSqlDao.ResetGroupRules
				setGroupBasic = groupSqlDao.SetGroupBasic
				// init tag dao
				tagDao, err = model.NewTagSQLDao(sqlConn)
				if err != nil {
					logger.Error.Printf("init tag dao failed, %v", err)
					return
				}
				tags = tagDao.Tags
				// init sentence dao
				sentenceDao = model.NewSentenceSQLDao(sqlConn)
				// init call dao
				callDao = model.NewCallSQLDao(sqlConn)
				callCount = callDao.Count
				calls = callDao.Calls
				// init relation dao
				relationDao = &model.RelationSQLDao{}
				// init segment dao
				segmentDao = model.NewSegmentDao(dbLike)
				// init user value & keys dao
				userValueDao = model.NewUserValueDao(dbLike.Conn())
				valuesKey = userValueDao.ValuesKey
				userKeyDao = model.NewUserKeyDao(dbLike.Conn())
				userKeys = userKeyDao.UserKeys
				keyvalues = userKeyDao.KeyValues
				// init condition dao
				condDao = model.NewConditionDao(dbLike)
				newCondition = condDao.NewCondition
				// init sentenceTest dao
				sentenceTestDao = model.NewSentenceTestSQLDao(sqlConn)
				// init cu trainer & predictor
				cuURL := envs["LOGIC_PREDICT_URL"]
				predictor = &logicaccess.Client{URL: cuURL, Timeout: time.Duration(300 * time.Second)}
				trainer = &logicaccess.Client{URL: cuURL, Timeout: time.Duration(300 * time.Second)}
				// init RABBITMQ client
				host := envs["RABBITMQ_HOST"]
				if host == "" {
					logger.Error.Println("RABBITMQ_HOST is required!")
					return
				}
				port := envs["RABBITMQ_PORT"]
				if port == "" {
					logger.Error.Println("RABBITMQ_PORT is required!")
					return
				}
				_, err = strconv.Atoi(port)
				if err != nil {
					logger.Error.Println("RABBITMQ_PORT should be a valid int value, ", err)
					return
				}
				client, err := rabbitmq.Dial(fmt.Sprintf("amqp://guest:guest@%s:%s", host, port))
				if err != nil {
					logger.Error.Println("init rabbitmq client failed, ", err)
					return
				}
				producer = client.NewProducer(rabbitmq.ProducerConfig{
					QueueName:   "src_queue",
					ContentType: "application/json",
					MaxRetry:    10,
				})
				consumer = client.NewConsumer(rabbitmq.ConsumerConfig{
					QueueName: "dst_queue",
					MaxRetry:  10,
				})
				// require dao init first.
				err = consumer.Subscribe(ASRWorkFlow)
				if err != nil {
					logger.Error.Printf("Failed to subscribe queue: %s, error: %s",
						"dst_queue", err.Error())
					return
				}

				// Realtime call
				realtimeCallProducer = client.NewProducer(rabbitmq.ProducerConfig{
					QueueName:   "realtime_call_queue",
					ContentType: "application/json",
					MaxRetry:    10,
				})
				realtimeCallConsumer = client.NewConsumer(rabbitmq.ConsumerConfig{
					QueueName: "realtime_call_queue",
					MaxRetry:  10,
				})
				err = realtimeCallConsumer.Subscribe(RealtimeCallWorkflow)
				if err != nil {
					logger.Error.Printf("Failed to subscribe queue: %s, error: %s",
						"realtime_call_queue", err.Error())
					return
				}

				logger.Info.Println("init & subscribe to RabbitMQ success")

				// init swDao
				cluster, err := redis.NewClusterFromEnvs(envs)
				if err != nil {
					logger.Error.Printf("init redis cluster failed, err: %s", err.Error())
				}
				swDao = model.NewDefaultSensitiveWordDao(cluster)

			},
			"init nav cache": setUpNavCache,
		},
	}
}
