# Total Asks (總提問)

### 依時間統計：

#### 各個時間區段的總提問數

統計 `app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，結果依照 **`day`** 分群：

```
POST /emotibot-records-csbot-*/_search
{
  "query": {
    "bool": {
      "filter": [
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
    "histogram": {
      "date_histogram": {
        "field": "log_time",
        "format": "yyyy-MM-dd HH:mm:ss",
        "interval": "day",
        "time_zone": "+08:00",
        "min_doc_count": 0, 
        "extended_bounds": {
          "min": "2018-06-01 00:00:00",
          "max": "2018-06-30 23:59:59"
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
  "took": 2,
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
    "histogram": {
      "buckets": [
        {
          "key_as_string": "2018-06-01 00:00:00",
          "key": 1527811200000,
          "doc_count": 286
        },
        {
          "key_as_string": "2018-06-02 00:00:00",
          "key": 1527897600000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-03 00:00:00",
          "key": 1527984000000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-04 00:00:00",
          "key": 1528070400000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-05 00:00:00",
          "key": 1528156800000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-06 00:00:00",
          "key": 1528243200000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-07 00:00:00",
          "key": 1528329600000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-08 00:00:00",
          "key": 1528416000000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-09 00:00:00",
          "key": 1528502400000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-10 00:00:00",
          "key": 1528588800000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-11 00:00:00",
          "key": 1528675200000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-12 00:00:00",
          "key": 1528761600000,
          "doc_count": 1
        },
        {
          "key_as_string": "2018-06-13 00:00:00",
          "key": 1528848000000,
          "doc_count": 2
        },
        {
          "key_as_string": "2018-06-14 00:00:00",
          "key": 1528934400000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-15 00:00:00",
          "key": 1529020800000,
          "doc_count": 14
        },
        {
          "key_as_string": "2018-06-16 00:00:00",
          "key": 1529107200000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-17 00:00:00",
          "key": 1529193600000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-18 00:00:00",
          "key": 1529280000000,
          "doc_count": 9
        },
        {
          "key_as_string": "2018-06-19 00:00:00",
          "key": 1529366400000,
          "doc_count": 555
        },
        {
          "key_as_string": "2018-06-20 00:00:00",
          "key": 1529452800000,
          "doc_count": 15
        },
        {
          "key_as_string": "2018-06-21 00:00:00",
          "key": 1529539200000,
          "doc_count": 85
        },
        {
          "key_as_string": "2018-06-22 00:00:00",
          "key": 1529625600000,
          "doc_count": 8
        },
        {
          "key_as_string": "2018-06-23 00:00:00",
          "key": 1529712000000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-24 00:00:00",
          "key": 1529798400000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-25 00:00:00",
          "key": 1529884800000,
          "doc_count": 178
        },
        {
          "key_as_string": "2018-06-26 00:00:00",
          "key": 1529971200000,
          "doc_count": 89
        },
        {
          "key_as_string": "2018-06-27 00:00:00",
          "key": 1530057600000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-28 00:00:00",
          "key": 1530144000000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-29 00:00:00",
          "key": 1530230400000,
          "doc_count": 0
        },
        {
          "key_as_string": "2018-06-30 00:00:00",
          "key": 1530316800000,
          "doc_count": 0
        }
      ]
    }
  }
}
```

### 依維度統計：

#### 在所篩選的時間範圍內，各個維度的總提問數
##### (以平台 (platform) 維度為例)

統計 `app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，且 `platform` 欄位不為 **`空字串`**，結果依照平台 **`(group_by_platform)`** 分群：

```
POST /emotibot-records-csbot-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "range": {
            "log_time": {
              "gte": "2018-06-01 00:00:00",
              "lte": "2018-06-30 23:59:59",
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
  "took": 1,
  "timed_out": false,
  "_shards": {
    "total": 5,
    "successful": 5,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 59,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "group_by_platform": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "android",
          "doc_count": 39
        },
        {
          "key": "微信",
          "doc_count": 17
        },
        {
          "key": "web",
          "doc_count": 2
        },
        {
          "key": "ios",
          "doc_count": 1
        }
      ]
    }
  }
}
```
