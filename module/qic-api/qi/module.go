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
	"emotibot.com/emotigo/pkg/api/rabbitmq/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	tagDao     model.TagDao
	callDao    model.CallDao
	taskDao    model.TaskDao
	segmentDao model.SegmentDao
	producer   *rabbitmq.Producer
	consumer   *rabbitmq.Consumer
	sqlConn    *sql.DB
	dbLike     model.DBLike
	volume     string
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "qi",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "groups", []string{}, handleCreateGroup),
			util.NewEntryPoint("GET", "groups", []string{}, handleGetGroups),
			util.NewEntryPoint("GET", "groups/filters", []string{}, handleGetGroupsByFilter),
			util.NewEntryPoint("GET", "groups/{id}", []string{}, handleGetGroup),
			util.NewEntryPoint("PUT", "groups/{id}", []string{}, handleUpdateGroup),
			util.NewEntryPoint("DELETE", "groups/{id}", []string{}, handleDeleteGroup),

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
			util.NewEntryPoint(http.MethodGet, "calls/{call_id}", []string{}, CallsDetailHandler),
			util.NewEntryPoint(http.MethodPost, "calls/{call_id}/file", []string{}, UpdateCallsFileHandler),
			util.NewEntryPoint(http.MethodGet, "calls/{call_id}/file", []string{}, CallsFileHandler),
			util.NewEntryPoint(http.MethodGet, "calls/{call_id}/credits", []string{}, WithCallIDCheck(handleGetCredit)),

			util.NewEntryPoint(http.MethodPost, "manual/use/all/tags", []string{}, handleTrainAllTags),
			util.NewEntryPoint(http.MethodDelete, "manual/use/all/tags", []string{}, handleUnload),

			util.NewEntryPoint(http.MethodGet, "call-in/navigation/{id}", []string{}, handleGetFlowSetting),
			util.NewEntryPoint(http.MethodPut, "call-in/navigation/{id}/name", []string{}, handleModifyFlow),
			util.NewEntryPoint(http.MethodDelete, "call-in/navigation/{id}", []string{}, handleDeleteFlow),
			util.NewEntryPoint(http.MethodPost, "call-in/navigation", []string{}, handleNewFlow),
			util.NewEntryPoint(http.MethodGet, "call-in/navigation", []string{}, handleFlowList),
			util.NewEntryPoint(http.MethodPut, "call-in/navigation/{id}/intent", []string{}, handleModifyIntent),
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
			"init db": func() {
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
				serviceDAO = model.NewGroupSQLDao(sqlConn)
				tagDao, err = model.NewTagSQLDao(sqlConn)
				if err != nil {
					logger.Error.Printf("init tag dao failed, %v", err)
					return
				}
				sentenceDao = model.NewSentenceSQLDao(sqlConn)

				cuURL := envs["LOGIC_PREDICT_URL"]
				predictor = &logicaccess.Client{URL: cuURL, Timeout: time.Duration(300 * time.Second)}
				callDao = model.NewCallSQLDao(sqlConn)
				taskDao = model.NewTaskDao(sqlConn)
				relationDao = &model.RelationSQLDao{}
				trainer = &logicaccess.Client{URL: cuURL, Timeout: time.Duration(300 * time.Second)}
				segmentDao = model.NewSegmentDao(dbLike)

			},
			"init RabbitMQ": func() {
				envs := ModuleInfo.Environments
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
				_, err := strconv.Atoi(port)
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
				consumer.Subscribe(ASRWorkFlow)
				logger.Info.Println("init & subscribe to RabbitMQ success")

			},
		},
	}
}
