# 日誌管理

搜尋 `enterprise_id` 為 **`emotibot`**、`app_id` 為 **`csbot`** 且資料介於 **`2018-08-01 00:00:00`** 與 **`2018-08-31 23:59:59`**，`搜尋關鍵字` 為 **`APP无法付款`**，`問答類別` 為 **`業務類`** (`module` 為 **`faq`** 或 **`task_engine`**)，篩選條件為：`平台 (platform)` 為 **`android`** 或 **`ios`**，且 `性別 (sex)` 為 **`男`** 或 **`女`** 的對話日誌：

針對篩選條件為：`平台 (platform)` 為 **`android`** 或 **`ios`**，且 `性別 (sex)` 為 **`男`** 或 **`女`**，對話日誌需同時符合：

- `平台 (platform)` 為 **`android`** 及 `性別 (sex)` 為 **`男`**，或
- `平台 (platform)` 為 **`android`** 及 `性別 (sex)` 為 **`女`**，或
- `平台 (platform)` 為 **`ios`** 及 `性別 (sex)` 為 **`男`**，或
- `平台 (platform)` 為 **`ios`** 及 `性別 (sex)` 為 **`女`**

```
POST /emotibot-records-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "term": {
            "enterprise_id": "emotibot"
          }
        },
        {
          "term": {
            "app_id": "csbot"
          }
        },
        {
          "range": {
            "log_time": {
              "gte": "2018-08-01 00:00:00",
              "lte": "2018-08-31 23:59:59",
              "format": "yyyy-MM-dd HH:mm:ss",
              "time_zone": "+08:00"
            }
          }
        },
        {
          "multi_match": {
            "query": "APP无法付款",
            "fields": [
              "user_q",
              "answer.value"
            ]
          }
        },
        {
          "terms": {
            "module": [
              "faq",
              "task_engine"
            ]
          }
        },
        {
          "bool": {
            "should": [
              {
                "bool": {
                  "filter": [
                    {
                      "term": {
                        "custom_info.platform": "android"
                      }
                    },
                    {
                      "term": {
                        "custom_info.sex": "男"
                      }
                    }
                  ]
                }
              },
              {
                "bool": {
                  "filter": [
                    {
                      "term": {
                        "custom_info.platform": "android"
                      }
                    },
                    {
                      "term": {
                        "custom_info.sex": "女"
                      }
                    }
                  ]
                }
              },
              {
                "bool": {
                  "filter": [
                    {
                      "term": {
                        "custom_info.platform": "ios"
                      }
                    },
                    {
                      "term": {
                        "custom_info.sex": "男"
                      }
                    }
                  ]
                }
              },
              {
                "bool": {
                  "filter": [
                    {
                      "term": {
                        "custom_info.platform": "ios"
                      }
                    },
                    {
                      "term": {
                        "custom_info.sex": "女"
                      }
                    }
                  ]
                }
              }
            ]
          }
        }
      ]
    }
  }
}
```

範例結果：

```
{
    "took": 2,
    "timed_out": false,
    "_shards": {
        "total": 5,
        "successful": 5,
        "skipped": 0,
        "failed": 0
    },
    "hits": {
        "total": 1,
        "max_score": 0,
        "hits": [
            {
                "_index": "records",
                "_type": "doc",
                "_id": "T05782QBORawnYcLO5J4",
                "_score": 0,
                "_source": {
                    "session_id": "",
                    "intent_score": 100,
                    "host": "172.18.0.4",
                    "user_id": "4b21158a395311e88a710242ac110003",
                    "emotion_score": 100,
                    "score": 100,
                    "port": 26820,
                    "@version": "1",
                    "@timestamp": "2018-08-01T03:14:56.912Z",
                    "intent": "",
                    "custom_info": {
                        "sex": "男",
                        "platform": "android"
                    },
                    "unique_id": "20180801111456774623618",
                    "message": "{\"unique_id\":\"20180801111456774623618\",\"user_id\":\"4b21158a395311e88a710242ac110003\",\"app_id\":\"csbot\",\"session_id\":\"\",\"user_q\":\"app无法付款\",\"std_q\":\"APP无法付款怎么办\",\"answer\":[{\"type\":\"text\",\"subType\":\"text\",\"value\":\"您好，如果在线无法付款，可以联系机构门店协助刷卡付费。\",\"data\":[],\"extendData\":\"\"},{\"type\":\"text\",\"subType\":\"relatelist\",\"value\":\"指定相关\",\"data\":[\"e家保理赔\",\"一年期综合意外险保障时间\",\"一年期综合意外险可以预留被保险人信息吗\",\"一年期综合意外险对于被保险人行动及职业要求\"],\"extendData\":\"\"}],\"module\":\"faq\",\"intent\":\"\",\"intent_score\":100,\"emotion\":\"中性\",\"emotion_score\":100,\"score\":100,\"host\":\"172.18.0.4\",\"log_time\":\"2018-08-01T03:14:56.000Z\",\"custom_info\":{\"platform\":\"android\",\"sex\":\"男\"},\"note\":\"\"}",
                    "module": "faq",
                    "app_id": "csbot",
                    "user_q": "app无法付款",
                    "emotion": "中性",
                    "answer": [
                        {
                            "value": "您好，如果在线无法付款，可以联系机构门店协助刷卡付费。",
                            "extendData": "",
                            "data": [],
                            "subType": "text",
                            "type": "text"
                        },
                        {
                            "value": "指定相关",
                            "extendData": "",
                            "data": [
                                "e家保理赔",
                                "一年期综合意外险保障时间",
                                "一年期综合意外险可以预留被保险人信息吗",
                                "一年期综合意外险对于被保险人行动及职业要求"
                            ],
                            "subType": "relatelist",
                            "type": "text"
                        }
                    ],
                    "log_time": "2018-08-01T03:14:56.000Z",
                    "std_q": "APP无法付款怎么办",
                    "note": ""
                }
            }
        ]
    }
}
```
