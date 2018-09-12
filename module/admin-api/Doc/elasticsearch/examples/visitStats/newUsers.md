# New Users (新增用戶)

### 依時間統計：

#### 各個時間區段的新增用戶數

此數據無法直接透過 ElasticSearch 計算，因此先統計 `app_id` 為 **`csbot`**，結果依各使用者 **`(new_users)`** 分群，並取得各使用者的第一筆資料。最後再透過程式統計各使用者的第一筆資料是否落在所統計的各個時間區段中而得到新增用戶數：

```
POST /emotibot-records-csbot-*/_search
{
  "aggs": {
    "new_users": {
      "terms": {
        "field": "user_id",
        "size": 3000000
      },
      "aggs": {
        "user_first_chat": {
          "top_hits": {
            "size": 1,
            "sort": [
              {
                "log_time": {
                  "order": "asc"
                }
              }
            ]
          }
        }
      }
    }
  },
  "size": 0
}
```

範例結果：

```
{
  "took": 4,
  "timed_out": false,
  "_shards": {
    "total": 5,
    "successful": 5,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 4,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "new_users": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "9876",
          "doc_count": 4,
          "user_first_chat": {
            "hits": {
              "total": 4,
              "max_score": null,
              "hits": [
                {
                  "_index": "records",
                  "_type": "doc",
                  "_id": "cGgtjmQBuCoMdkfhVdMs",
                  "_score": null,
                  "_source": {
                    "user_id": "9876",
                    "message": """{"unique_id":"0188326a891a4e1c97c856ceebe6e857","user_id":"9876","app_id":"csbot","session_id":"test_session","user_q":"西雅图天气如何","std_q":"","module":"backfill","intent":"查，天气","intent_score":0,"emotion":"疑惑","emotion_score":0,"score":0,"host":"172.17.0.2","log_time":"2018-07-12T11:08:12.000Z","custom_info":{"platform":"android"},"note":""}""",
                    "intent": "查，天气",
                    "custom_info": {
                      "platform": "android"
                    },
                    "emotion_score": 0,
                    "session_id": "test_session",
                    "app_id": "csbot",
                    "emotion": "疑惑",
                    "note": "",
                    "std_q": "",
                    "user_q": "西雅图天气如何",
                    "log_time": "2018-07-12T11:08:12.000Z",
                    "@timestamp": "2018-07-12T11:08:12.861Z",
                    "unique_id": "0188326a891a4e1c97c856ceebe6e857",
                    "module": "backfill",
                    "host": "172.17.0.2",
                    "intent_score": 0,
                    "score": 0,
                    "port": 33930,
                    "@version": "1"
                  },
                  "sort": [
                    1531393692000
                  ]
                }
              ]
            }
          }
        }
      ]
    }
  }
}
```

### 依維度統計：

#### 在所篩選的時間範圍內，各個維度的新增用戶數
##### (以平台 (platform) 維度為例)

此數據無法直接透過 ElasticSearch 計算，因此先統計 `app_id` 為 **`csbot`**，且 `platform` 欄位不為 **`空字串`**，結果依各使用者 **`(new_users)`** 分群後，再依各平台 **`(group_by_platform)`** 分群，並取得各使用者在各平台下的第一筆資料。最後再透過程式統計各使用者在各平台下的第一筆資料是否落在所篩選的時間範圍內而得到新增用戶數：

```
POST /emotibot-records-csbot-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "exists": {
            "field": "custom_info.platform.keyword"
          }
        }
      ]
    }
  },
  "aggs": {
    "new_users": {
      "terms": {
        "field": "user_id",
        "size": 3000000
      },
      "aggs": {
        "group_by_platform": {
          "terms": {
            "field": "custom_info.platform.keyword",
            "size": 3000000
          },
          "aggs": {
            "user_first_chat": {
              "top_hits": {
                "size": 1,
                "sort": [
                  {
                    "log_time": {
                      "order": "asc"
                    }
                  }
                ]
              }
            }
          }
        }
      }
    }
  },
  "size": 0
}
```

範例結果：

```
{
  "took": 1,
  "timed_out": false,
  "_shards": {
    "total": 5,
    "successful": 5,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 3,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "new_users": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "9876",
          "doc_count": 3,
          "group_by_platform": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "android",
                "doc_count": 3,
                "user_first_chat": {
                  "hits": {
                    "total": 3,
                    "max_score": null,
                    "hits": [
                      {
                        "_index": "records",
                        "_type": "doc",
                        "_id": "cGgtjmQBuCoMdkfhVdMs",
                        "_score": null,
                        "_source": {
                          "user_id": "9876",
                          "message": """{"unique_id":"0188326a891a4e1c97c856ceebe6e857","user_id":"9876","app_id":"csbot","session_id":"test_session","user_q":"西雅图天气如何","std_q":"","module":"backfill","intent":"查，天气","intent_score":0,"emotion":"疑惑","emotion_score":0,"score":0,"host":"172.17.0.2","log_time":"2018-07-12T11:08:12.000Z","custom_info":{"platform":"android"},"note":""}""",
                          "intent": "查，天气",
                          "custom_info": {
                            "platform": "android"
                          },
                          "emotion_score": 0,
                          "session_id": "test_session",
                          "app_id": "csbot",
                          "emotion": "疑惑",
                          "note": "",
                          "std_q": "",
                          "user_q": "西雅图天气如何",
                          "log_time": "2018-07-12T11:08:12.000Z",
                          "@timestamp": "2018-07-12T11:08:12.861Z",
                          "unique_id": "0188326a891a4e1c97c856ceebe6e857",
                          "module": "backfill",
                          "host": "172.17.0.2",
                          "intent_score": 0,
                          "score": 0,
                          "port": 33930,
                          "@version": "1"
                        },
                        "sort": [
                          1531393692000
                        ]
                      }
                    ]
                  }
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
