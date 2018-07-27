# Top Unmatched Questions (未匹配問題)

#### 前 N 個未匹配問題
##### (以 N = 20 為例)

統計 `enterprise_id` 為 **`emotibot`**、`app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，`module` 為 **`backfill`**，數量總和前 **20** 個的使用者問題。另外再透過 **`max_log_time`** 及 **`min_log_time`** 分群取得各問題第一次及最後一次被詢問的時間：

```
POST /records/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "term": {
            "module": "backfill"
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
    "group_by_userq": {
      "terms": {
        "field": "user_q.keyword",
        "size": 20,
        "shard_size": 3000000
      },
      "aggs": {
        "max_log_time": {
          "max": {
            "field": "log_time",
            "format": "yyyy-MM-dd HH:mm:ss"
          }
        },
        "min_log_time": {
          "min": {
            "field": "log_time",
            "format": "yyyy-MM-dd HH:mm:ss"
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
  "took": 3,
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
    "group_by_userq": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "西雅图天气如何",
          "doc_count": 4,
          "max_log_time": {
            "value": 1531394128000,
            "value_as_string": "2018-07-12T11:15:28.000+0000"
          },
          "min_log_time": {
            "value": 1531393692000,
            "value_as_string": "2018-07-12T11:08:12.000+0000"
          }
        }
      ]
    }
  }
}
```
