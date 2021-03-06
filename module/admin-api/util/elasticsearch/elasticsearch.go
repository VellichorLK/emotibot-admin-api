package elasticsearch

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/elasticsearch/data"
	"emotibot.com/emotigo/pkg/logger"

	elastic "gopkg.in/olivere/elastic.v6"
)

var (
	esClient          *elastic.Client
	esCtx             context.Context
	host              string
	port              string
	basicAuthUsername string
	basicAuthPassword string
)

func Setup(envHost string, envPort string,
	envBasicAuthUsername string, envBasicAuthPasword string) error {
	host = envHost
	port = envPort
	basicAuthUsername = envBasicAuthUsername
	basicAuthPassword = envBasicAuthPasword

	ctx, client, err := initClient()
	if err != nil {
		return err
	}

	esClient = client
	esCtx = ctx

	err = createRecordsIndexTemplateIfNeed()
	if err != nil {
		return err
	}

	err = createSessionsIndexTemplateIfNeed()
	if err != nil {
		return err
	}

	err = createUsersIndexTemplateIfNeed()
	if err != nil {
		return err
	}

	err = createQACoreIndextemplateIfNeed()
	if err != nil {
		return err
	}

	return nil
}

func initClient() (ctx context.Context, client *elastic.Client, err error) {
	defer func() {
		if err != nil {
			ctx = nil
			client = nil
		}
	}()
	esURL := fmt.Sprintf("http://%s:%s", host, port)

	// Turn-off sniffing
	client, err = elastic.NewClient(elastic.SetURL(esURL),
		elastic.SetErrorLog(logger.Error),
		elastic.SetBasicAuth(basicAuthUsername, basicAuthPassword),
		elastic.SetSniff(false))
	if err != nil {
		return
	}
	ctx = context.Background()

	return
}

func GetClient() (context.Context, *elastic.Client) {
	if esClient == nil {
		ctx, client, err := initClient()
		if err != nil {
			logger.Error.Println("Init es client fail:", err.Error())
			return nil, nil
		}
		esClient = client
		esCtx = ctx
		return ctx, client
	}
	return esCtx, esClient
}

func createRecordsIndexTemplateIfNeed() error {
	ctx, client := GetClient()
	if ctx == nil || client == nil {
		return nil
	}

	// Check existence of records index template
	exists, err := client.IndexTemplateExists(data.ESRecordsTemplate).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		// Create records index template
		template, err := ioutil.ReadFile(data.ESRecordsTemplateFile)
		if err != nil {
			return err
		}

		service, err := client.IndexPutTemplate(data.ESRecordsTemplate).BodyString(string(template)).Do(ctx)
		if err != nil {
			return err
		}

		if !service.Acknowledged {
			return data.ErrESNotAcknowledged
		}
	}

	return nil
}

func createSessionsIndexTemplateIfNeed() error {
	ctx, client := GetClient()
	if ctx == nil || client == nil {
		return nil
	}

	// Check existence of sessions index template
	exists, err := client.IndexTemplateExists(data.ESSessionsTemplate).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		// Create sessions index template
		template, err := ioutil.ReadFile(data.ESSessionsTemplateFile)
		if err != nil {
			return err
		}

		service, err := client.IndexPutTemplate(data.ESSessionsTemplate).BodyString(string(template)).Do(ctx)
		if err != nil {
			return err
		}

		if !service.Acknowledged {
			return data.ErrESNotAcknowledged
		}
	}

	return nil
}

func createUsersIndexTemplateIfNeed() error {
	ctx, client := GetClient()
	if ctx == nil || client == nil {
		return nil
	}

	// Check existence of users index template
	exists, err := client.IndexTemplateExists(data.ESUsersTemplate).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		// Create users index template
		template, err := ioutil.ReadFile(data.ESUsersTemplateFile)
		if err != nil {
			return err
		}

		service, err := client.IndexPutTemplate(data.ESUsersTemplate).BodyString(string(template)).Do(ctx)
		if err != nil {
			return err
		}

		if !service.Acknowledged {
			return data.ErrESNotAcknowledged
		}
	}

	return nil
}

func createQACoreIndextemplateIfNeed() error {
	ctx, client := GetClient()
	if ctx == nil || client == nil {
		return nil
	}

	// Check existence of records index template.
	exists, err := client.IndexTemplateExists(data.ESQACoreTemplate).Do(ctx)
	if err != nil {
		return err
	}

	if !exists {
		// Create records index template.
		template, err := ioutil.ReadFile(data.ESQACoreTemplateFile)
		if err != nil {
			return err
		}

		service, err := client.IndexPutTemplate(data.ESQACoreTemplate).BodyString(string(template)).Do(ctx)
		if err != nil {
			return err
		}

		if !service.Acknowledged {
			return data.ErrESNotAcknowledged
		}
	}

	return nil
}

func CreateTimeRangeFromString(startDate string,
	endDate string, timeFormat string) (startTime time.Time, endTime time.Time, err error) {
	startTime, err = time.Parse(timeFormat, startDate)
	if err != nil {
		return
	}

	endTime, err = time.Parse(timeFormat, endDate)
	if err != nil {
		return
	}

	startTime, endTime = createTimeRange(startTime, endTime)
	return
}

func CreateTimeRangeFromTimestamp(startTimestamp int64,
	endTimestamp int64) (startTime time.Time, endTime time.Time) {
	startTime = time.Unix(startTimestamp, 0)
	endTime = time.Unix(endTimestamp, 0)

	startTime, endTime = createTimeRange(startTime, endTime)
	return
}

func createTimeRange(startTime time.Time,
	endTime time.Time) (_startTime time.Time, _endTime time.Time) {
	// Treat query times as local times
	_startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.Local)
	_endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.Local)
	return
}

func ExtractElasticsearchRootCauseErrors(err interface{}) ([]string, bool) {
	if esErr, ok := err.(*elastic.Error); ok {
		rootCause := esErr.Details.RootCause
		reasons := make([]string, len(rootCause))
		for i, cause := range rootCause {
			reasons[i] = cause.Reason
		}

		return reasons, true
	}

	// Not instance of elastic.Error return false
	return nil, false
}
