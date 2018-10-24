# Conversations (總會話)

### 依時間統計：

#### 各個時間區段的總會話數

統計 `app_id` 為 **`csbot`** 且資料介於 **`2018-09-01 00:00:00`** 與 **`2018-09-30 23:59:59`**，結果依照 **`day`** 分群：

```
POST /emotibot-sessions-csbot-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "range": {
            "end_time": {
              "gte": "2018-09-01 00:00:00",
              "lte": "2018-09-30 23:59:59",
              "format": "yyyy-MM-dd HH:mm:ss",
              "time_zone": "+08:00"
            }
          }
        }
      ]
    }
  },
  "aggs": {
    "histogram": {
      "date_histogram": {
        "field":"end_time",
        "format":"yyyy-MM-dd HH:mm:ss",
        "interval":"day",
        "time_zone":"+08:00",
        "min_doc_count":0,
        "extended_bounds":{
          "min":"2018-09-01 00:00:00",
          "max":"2018-09-30 23:59:59"
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
    "took": 31,
    "timed_out": false,
    "_shards": {
        "total": 880,
        "successful": 880,
        "skipped": 692,
        "failed": 0
    },
    "hits": {
        "total": 13437,
        "max_score": 0,
        "hits": []
    },
    "aggregations": {
        "histogram": {
            "buckets": [
                {
                    "key_as_string": "2018-09-01 00:00:00",
                    "key": 1535731200000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-02 00:00:00",
                    "key": 1535817600000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-03 00:00:00",
                    "key": 1535904000000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-04 00:00:00",
                    "key": 1535990400000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-05 00:00:00",
                    "key": 1536076800000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-06 00:00:00",
                    "key": 1536163200000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-07 00:00:00",
                    "key": 1536249600000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-08 00:00:00",
                    "key": 1536336000000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-09 00:00:00",
                    "key": 1536422400000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-10 00:00:00",
                    "key": 1536508800000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-11 00:00:00",
                    "key": 1536595200000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-12 00:00:00",
                    "key": 1536681600000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-13 00:00:00",
                    "key": 1536768000000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-14 00:00:00",
                    "key": 1536854400000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-15 00:00:00",
                    "key": 1536940800000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-16 00:00:00",
                    "key": 1537027200000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-17 00:00:00",
                    "key": 1537113600000,
                    "doc_count": 14
                },
                {
                    "key_as_string": "2018-09-18 00:00:00",
                    "key": 1537200000000,
                    "doc_count": 25
                },
                {
                    "key_as_string": "2018-09-19 00:00:00",
                    "key": 1537286400000,
                    "doc_count": 9035
                },
                {
                    "key_as_string": "2018-09-20 00:00:00",
                    "key": 1537372800000,
                    "doc_count": 34
                },
                {
                    "key_as_string": "2018-09-21 00:00:00",
                    "key": 1537459200000,
                    "doc_count": 32
                },
                {
                    "key_as_string": "2018-09-22 00:00:00",
                    "key": 1537545600000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-23 00:00:00",
                    "key": 1537632000000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-24 00:00:00",
                    "key": 1537718400000,
                    "doc_count": 0
                },
                {
                    "key_as_string": "2018-09-25 00:00:00",
                    "key": 1537804800000,
                    "doc_count": 37
                },
                {
                    "key_as_string": "2018-09-26 00:00:00",
                    "key": 1537891200000,
                    "doc_count": 2831
                },
                {
                    "key_as_string": "2018-09-27 00:00:00",
                    "key": 1537977600000,
                    "doc_count": 1380
                },
                {
                    "key_as_string": "2018-09-28 00:00:00",
                    "key": 1538064000000,
                    "doc_count": 43
                },
                {
                    "key_as_string": "2018-09-29 00:00:00",
                    "key": 1538150400000,
                    "doc_count": 5
                },
                {
                    "key_as_string": "2018-09-30 00:00:00",
                    "key": 1538236800000,
                    "doc_count": 1
                }
            ]
        }
    }
}

```

### 依維度統計：

#### 在所篩選的時間範圍內，各個維度的總會話數
##### (以平台 (platform) 維度為例)

統計 `app_id` 為 **`csbot`** 且資料介於 **`2018-09-01 00:00:00`** 與 **`2018-09-30 23:59:59`**，且 `platform` 不為 **`空字串`**，結果依照平台 **`(group_by_platform)`** 分群：

```
POST /emotibot-sessions-csbot-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "range": {
            "end_time": {
              "gte": "2018-09-01 00:00:00",
              "lte": "2018-09-30 23:59:59",
              "format": "yyyy-MM-dd HH:mm:ss",
              "time_zone": "+08:00"
            }
          }
        },
        {
          "exists": {
            "field": "custom_info.platform.keyword"
          }
        }
      ]
    }
  },
  "aggs": {
    "group_by_platform": {
      "terms": {
        "field": "custom_info.platform.keyword",
        "shard_size": 3000000
      }
    }
  },
  "size": 0
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
    "total": 2,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "group_by_platform": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "ios",
          "doc_count": 2
        },
        {
          "key": "微信",
          "doc_count": 5
        }
      ]
    }
  }
}
```
