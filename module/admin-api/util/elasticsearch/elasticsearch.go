package elasticsearch

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"

	"github.com/olivere/elastic"
)

var (
	esClient *elastic.Client
	esCtx    context.Context
)

func Init(host string, port string) (err error) {
	esURL := fmt.Sprintf("http://%s:%s", host, port)

	// Turn-off sniffing
	client, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetSniff(false))
	if err != nil {
		return
	}

	esClient = client
	esCtx = context.Background()

	// Check existence of records index template
	exists, err := client.IndexTemplateExists(data.ESRecordsTemplate).Do(esCtx)
	if err != nil {
		return
	}

	if !exists {
		// Create records index template
		template, _err := ioutil.ReadFile(data.ESRecordsTemplateFile)
		if _err != nil {
			err = _err
			return
		}

		service, _err := client.IndexPutTemplate(data.ESRecordsTemplate).BodyString(string(template)).Do(esCtx)
		if _err != nil {
			err = _err
			return
		}

		if !service.Acknowledged {
			err = data.ErrESNotAcknowledged
			return
		}
	}

	// Check existence of sessions index template
	exists, err = client.IndexTemplateExists(data.ESSessionsTemplate).Do(esCtx)
	if err != nil {
		return
	}

	if !exists {
		// Create sessions index template
		template, _err := ioutil.ReadFile(data.ESSessionsTemplateFile)
		if _err != nil {
			err = _err
			return
		}

		service, _err := client.IndexPutTemplate(data.ESSessionsTemplate).BodyString(string(template)).Do(esCtx)
		if _err != nil {
			err = _err
			return
		}

		if !service.Acknowledged {
			err = data.ErrESNotAcknowledged
			return
		}
	}

	return
}

func GetClient() (context.Context, *elastic.Client) {
	return esCtx, esClient
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
