# Answer Categories (問答類別統計)

#### 所選時間範圍內各問答類別的問題數
- 業務類
- 聊天類
- 其他類

統計 `enterprise_id` 為 **`emotibot`**、`app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，其中：
- 業務類 = 標準回覆數
    - (`module` 為 **`faq`** 或 **`task_engine`**)
- 聊天類 = 聊天數
    - (`module` 為 **`chat`**)
- 其他類 = 全部 - (業務類 + 聊天類)
    - (`module` 不為 **`faq`**、**`task_engine`** 及 **`chat`**)

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
              "gte": "2018-06-01 00:00:00",
              "lte": "2018-06-30 23:59:59",
              "format": "yyyy-MM-dd HH:mm:ss",
              "time_zone": "+08:00"
            }
          }
        }
      ]
    }
  },
  "aggs": {
    "answer_categories": {
      "filters": {
        "filters": {
          "business": {
            "terms": {
              "module": [
                "faq",
                "task_engine"
              ]
            }
          },
          "chat": {
            "term": {
              "module": "chat"
            }
          },
          "other": {
            "bool": { 
              "must_not": [
                {
                  "terms": {
                    "module": [
                      "faq",
                      "task_engine",
                      "chat"
                    ]
                  }
                }
              ]
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
  "took": 5,
  "timed_out": false,
  "_shards": {
    "total": 5,
    "successful": 5,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 1242,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "answer_categories": {
      "buckets": {
        "business": {
          "doc_count": 282
        },
        "chat": {
          "doc_count": 36
        },
        "other": {
          "doc_count": 109
        }
      }
    }
  }
}
```
