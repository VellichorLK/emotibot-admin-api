# 日誌管理

搜尋 `app_id` 為 **`csbot`** 且資料介於 **`2018-08-01 00:00:00`** 與 **`2018-08-31 23:59:59`**，`搜尋關鍵字` 為 **`请推荐我保险`**，`問答類別` 為 **`聊天類`** (`module` 為 **`chat`)，篩選條件為：`平台 (platform)` 為 **`android`** 或 **`ios`**，且 `性別 (sex)` 為 **`男`** 或 **`女`** 的對話日誌：

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
            "app_id": "csbot"
          }
        },
        {
          "range": {
            "log_time": {
              "gte": "2018-09-01 00:00:00",
              "lte": "2018-09-30 23:59:59",
              "format": "yyyy-MM-dd HH:mm:ss",
              "time_zone": "+08:00"
            }
          }
        },
        {
          "bool": {
            "should": [
              {
                "match": {
                  "user_q": "请推荐我保险"
                }
              },
              {
                "nested": {
                  "path": "answer",
                  "query": {
                    "match": {
                      "answer.value": "请推荐我保险"
                    }
                  }
                }
              }
            ]
          }
        },
        {
          "terms": {
            "module": [
              "chat"
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
    "took": 14,
    "timed_out": false,
    "_shards": {
        "total": 180,
        "successful": 180,
        "skipped": 130,
        "failed": 0
    },
    "hits": {
        "total": 1,
        "max_score": 0,
        "hits": [
            {
                "_index": "emotibot-records-csbot-2018-09-12",
                "_type": "doc",
                "_id": "DNUbzGUBwrBYEqaKAhrx",
                "_score": 0,
                "_source": {
                    "user_q": "请推荐我保险",
                    "intent": "",
                    "app_id": "csbot",
                    "@timestamp": "2018-09-12T04:47:26.000Z",
                    "log_time": "2018-09-12T04:47:26.000Z",
                    "note": "",
                    "user_id": "4b21158a395311e88a710242ac110003",
                    "unique_id": "20180912124726469611233",
                    "host": "172.19.0.4",
                    "message": "{\"unique_id\":\"20180912124726469611233\",\"user_id\":\"4b21158a395311e88a710242ac110003\",\"app_id\":\"csbot\",\"session_id\":\"59d38550-feb8-44d4-99f6-6caa87c4b260\",\"user_q\":\"请推荐我保险\",\"std_q\":\"有人向我推荐买财产保险。\",\"answer\":[{\"type\":\"text\",\"subType\":\"text\",\"value\":\"你买了吗？\",\"data\":[],\"extendData\":\"\"}],\"module\":\"chat\",\"intent\":\"\",\"intent_score\":56,\"emotion\":\"中性\",\"emotion_score\":56,\"score\":56,\"host\":\"172.19.0.4\",\"log_time\":\"2018-09-12T04:47:26.000Z\",\"custom_info\":{\"platform\":\"ios\",\"sex\":\"男\"},\"note\":\"\"}",
                    "module": "chat",
                    "emotion_score": 56,
                    "score": 56,
                    "port": 30840,
                    "type": "records",
                    "session_id": "59d38550-feb8-44d4-99f6-6caa87c4b260",
                    "intent_score": 56,
                    "@version": "1",
                    "custom_info": {
                        "platform": "ios",
                        "sex": "男"
                    },
                    "emotion": "中性",
                    "std_q": "有人向我推荐买财产保险。",
                    "answer": [
                        {
                            "extendData": "",
                            "type": "text",
                            "subType": "text",
                            "value": "你买了吗？",
                            "data": []
                        }
                    ]
                }
            }
        ]
    }
}
```
