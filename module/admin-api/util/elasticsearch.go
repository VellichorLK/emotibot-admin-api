package util

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

func InitElasticsearch(host string, port string) (err error) {
	esURL := fmt.Sprintf("http://%s:%s", host, port)

	// Turn-off sniffing
	client, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetSniff(false))
	if err != nil {
		return
	}

	esClient = client
	esCtx = context.Background()

	// Check existence of index
	exists, err := esClient.IndexExists(data.ESRecordsIndex).Do(esCtx)
	if err != nil {
		return
	}

	if !exists {
		// Create records index
		mapping, _err := ioutil.ReadFile(data.ESRecordsMappingFile)
		if _err != nil {
			err = _err
			return
		}

		service, _err := esClient.CreateIndex(data.ESRecordsIndex).BodyString(string(mapping)).Do(esCtx)
		if _err != nil {
			err = _err
			return
		}

		if !service.Acknowledged {
			err = data.ErrESNotAcknowledged
			return
		}
	}

	exists, err = esClient.IndexExists(data.ESSessionsIndex).Do(esCtx)
	if err != nil {
		return
	}

	if !exists {
		// Create sessions index
		mapping, _err := ioutil.ReadFile(data.ESSessionsMappingFile)
		if _err != nil {
			err = _err
			return
		}

		service, _err := esClient.CreateIndex(data.ESSessionsIndex).BodyString(string(mapping)).Do(esCtx)
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

func GetElasticsearch() (context.Context, *elastic.Client) {
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
	// Treat query times as local times,
	_startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.Local)
	_endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.Local)
	return
}
