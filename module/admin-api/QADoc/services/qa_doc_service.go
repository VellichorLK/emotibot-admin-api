package services

import (
	"encoding/json"

	"emotibot.com/emotigo/module/admin-api/QADoc/data"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	esData "emotibot.com/emotigo/module/admin-api/util/elasticsearch/data"

	"github.com/olivere/elastic"
)

const maxNumBulkDocs = 5000

func CreateQADoc(doc *data.QACoreDoc) ([]byte, error) {
	if doc == nil {
		return []byte{}, nil
	}

	ctx, client := elasticsearch.GetClient()
	service := elastic.NewIndexService(client)

	resp, err := service.
		Index(esData.ESQACoreIndex).
		Type(esData.ESQACoreType).
		BodyJson(doc).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp)
}

func BulkCreateOrUpdateQADocs(docs []*data.QACoreDoc) ([]byte, error) {
	if len(docs) == 0 {
		return []byte{}, nil
	}

	ctx, client := elasticsearch.GetClient()
	service := elastic.NewBulkService(client)

	var resp *elastic.BulkResponse
	var err error
	start := 0
	end := 0

	for {
		end += maxNumBulkDocs
		if end > len(docs) {
			end = len(docs)
		}

		for i := start; i < end; i++ {
			indexRequest := elastic.NewBulkUpdateRequest()
			indexRequest.Index(esData.ESQACoreIndex).Type(esData.ESQACoreType).Id(docs[i].DocID).Doc(docs[i]).DocAsUpsert(true)
			service.Add(indexRequest)
		}

		resp, err = service.Do(ctx)
		if err != nil {
			return nil, err
		}

		if end >= len(docs) {
			break
		}

		start = end
	}

	return json.Marshal(resp)
}
func BulkCreateQADocs(docs []*data.QACoreDoc) ([]byte, error) {
	if len(docs) == 0 {
		return []byte{}, nil
	}

	ctx, client := elasticsearch.GetClient()
	service := elastic.NewBulkService(client)

	var resp *elastic.BulkResponse
	var err error
	start := 0
	end := 0

	for {
		end += maxNumBulkDocs
		if end > len(docs) {
			end = len(docs)
		}

		for i := start; i < end; i++ {
			indexRequest := elastic.NewBulkIndexRequest()
			indexRequest.Index(esData.ESQACoreIndex).Type(esData.ESQACoreType).Doc(docs[i])
			service.Add(indexRequest)
		}

		resp, err = service.Do(ctx)
		if err != nil {
			return nil, err
		}

		if end >= len(docs) {
			break
		}

		start = end
	}

	return json.Marshal(resp)
}

func UpdateQADocsByQuery(script string, ids ...interface{}) ([]byte, error) {
	if len(ids) == 0 {
		return []byte{}, nil
	}

	ctx, client := elasticsearch.GetClient()
	termsQuery := elastic.NewTermsQuery("doc_id", ids)
	elastic.NewScript(script)

	service := elastic.NewUpdateByQueryService(client)
	resp, err := service.
		Index(esData.ESQACoreIndex).
		Type(esData.ESQACoreType).
		Query(termsQuery).
		Do(ctx)
	if err != nil {
		// Ignore index not found error
		if elastic.IsNotFound(err) {
			return []byte{}, nil
		}
		return nil, err
	}

	return json.Marshal(resp)
}

func DeleteQADocs(appID string, module string) ([]byte, error) {
	ctx, client := elasticsearch.GetClient()

	boolQuery := elastic.NewBoolQuery()
	termQuery := elastic.NewTermQuery("app_id", appID)
	boolQuery.Filter(termQuery)
	termQuery = elastic.NewTermQuery("module", module)
	boolQuery.Filter(termQuery)

	service := elastic.NewDeleteByQueryService(client)
	resp, err := service.
		Index(esData.ESQACoreIndex).
		Type(esData.ESQACoreType).
		Query(boolQuery).
		Do(ctx)
	if err != nil {
		// Ignore index not found error
		if elastic.IsNotFound(err) {
			return []byte{}, nil
		}
		return nil, err
	}

	return json.Marshal(resp)
}

func DeleteQADocsByIds(ids []interface{}) ([]byte, error) {
	ctx, client := elasticsearch.GetClient()
	termsQuery := elastic.NewTermsQuery("doc_id", ids...)

	service := elastic.NewDeleteByQueryService(client)
	qResp, err := service.
		Index(esData.ESQACoreIndex).
		Type(esData.ESQACoreType).
		Query(termsQuery).
		Do(ctx)
	if err != nil {
		// Ignore index not found error
		if elastic.IsNotFound(err) {
			return []byte{}, nil
		}
		return nil, err
	}

	return json.Marshal(qResp)
}

func DeleteQADocsByRegex(queries []*data.RegexQuery) ([]byte, error) {
	if len(queries) == 0 {
		return []byte{}, nil
	}

	ctx, client := elasticsearch.GetClient()
	boolQuery := elastic.NewBoolQuery()

	for _, query := range queries {
		regexQuery := elastic.NewRegexpQuery(query.Field, query.Expression)
		boolQuery.Should(regexQuery)
	}

	service := elastic.NewDeleteByQueryService(client)
	resp, err := service.
		Index(esData.ESQACoreIndex).
		Type(esData.ESQACoreType).
		Query(boolQuery).
		Do(ctx)
	if err != nil {
		// Ignore index not found error
		if elastic.IsNotFound(err) {
			return []byte{}, nil
		}
		return nil, err
	}

	return json.Marshal(resp)
}
