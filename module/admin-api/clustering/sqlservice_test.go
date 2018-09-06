package clustering

import (
	"flag"
	"os"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var isIntegrate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegrate, "it", false, "only run on integration test env")
	flag.Parse()
	os.Exit(m.Run())
}

var exampleSSMConfig = `{
    "items": [{
        "name": "candidate",
        "value": 5
    }, {
        "name": "seg_url",
        "value": "http:\/\/172.17.0.1:13901"
    }, {
        "name": "rank",
        "value": [{
            "dependency": [{
                "name": "ml",
                "weight": 1
            }],
            "order": 1,
            "source": "ml",
            "threadholds": 97
        }, {
            "dependency": [{
                "name": "se",
                "weight": 0.3
            }, {
                "name": "w2v",
                "weight": 0.6
            }, {
                "name": "solr",
                "weight": 0.1
            }],
            "order": 2,
            "source": "qq",
            "threadholds": 85
        }]
    }, {
        "name": "dependency",
        "value": [{
            "candidate": 30,
            "clazz": "SSMDependencySolr",
            "enabled": true,
            "method": "get",
            "name": "solr",
            "order": 1,
            "parameters": "p",
            "result": [{
                "answer": "@[d]\/answer_list[0]",
                "keywords": "@[d]\/keywords",
                "matchQuestion": "@[d]\/question",
                "question_id": "@[d]\/question_id:int",
                "score": "@[d]\/score:~toPercent",
                "source": "solr",
                "stop_words": "@[d]\/stop_words",
                "tokens": "@[d]\/tokens"
            }],
            "timeout": 60000,
            "url": "http:\/\/172.17.0.1:8081\/solr"
        }, {
            "candidate": 1,
            "enabled": true,
            "method": "POST",
            "name": "ml",
            "order": 2,
            "parameters": {
                "candidate": "@candidate",
                "data": "@Text",
                "model": "unknown_20180903222359"
            },
            "result": [{
                "answer": "@result\/predict",
                "matchQuestion": "@result\/predict",
                "score": "@result\/score",
                "source": "ML"
            }],
            "timeout": 60000,
            "url": "http:\/\/172.17.0.1:8895\/ft_predict"
        }, {
            "enabled": true,
            "formData": false,
            "method": "POST",
            "name": "se",
            "order": 2,
            "parameters": {
                "match_q": ["@solr[d]\/matchQuestion"],
                "user_q": "@Text"
            },
            "result": [{
                "answer": "@solr[d]\/answer",
                "matchQuestion": "@solr[d]\/matchQuestion",
                "score": "@Msg\/{d}:~toPercent",
                "source": "SE"
            }],
            "timeout": 60000,
            "url": "http:\/\/172.17.0.1:15601\/similar"
        }, {
            "enabled": true,
            "formData": false,
            "method": "POST",
            "name": "w2v",
            "order": 2,
            "parameters": [{
                "src": {
                    "src_kw": "@nlp\/key_words:~joinByComma",
                    "src_seg": "@nlp\/tokens:~joinByComma",
                    "stoplist": "@nlp\/stop_words:~joinByComma"
                },
                "tar": [{
                    "id": "{d}",
                    "tar_kw": "@solr[d]\/keywords:~joinByComma",
                    "tar_seg": "@solr[d]\/tokens:~joinByComma"
                }]
            }],
            "result": [{
                "answer": "@solr[d]\/answer",
                "matchQuestion": "@solr[d]\/matchQuestion",
                "score": "@scores\/{d}:~toPercent",
                "source": "w2v"
            }],
            "timeout": 60000,
            "url": "http:\/\/172.17.0.1:11501\/partial_w2v"
        }]
    }]
}`

func TestGetFTModel(t *testing.T) {
	appID := "csbot"
	db, writer, _ := sqlmock.New()
	writer.ExpectQuery("SELECT").WithArgs(appID).WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(exampleSSMConfig))
	var service = sqlService{
		db: db,
	}
	modelName, err := service.GetFTModel(appID)
	if err != nil {
		t.Fatal(err)
	}
	if modelName != "unknown_20180903222359" {
		t.Fatalf("expect model name to be unknown_20180903222359 but got %s", modelName)
	}
}
