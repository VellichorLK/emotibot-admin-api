# New Users (新增用戶)

### 依時間統計：

#### 各個時間區段的新增用戶數

統計 `app_id` 為 **`csbot`**，且資料介於 **`2018-10-01 00:00:00`** 與 **`2018-10-31 23:59:59`**，結果依照 **`day`** 分群。但由於為了做到基於 `維度` 的統計，因此 `emotibot-users-*` 的 document ID 其實是透過 **`user_id` + `custom_info`** hash 過後得來的。因此有可能有同一個使用者，但維度不同而導致產生多個 documents，所以最後還必須再做一次 `cardinality aggregation` 排除重複的 `user_id` 後方可計算出正確的新增用戶數：

```
POST /emotibot-users-csbot-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "range": {
            "first_log_time": {
              "gte": "2018-10-01 00:00:00",
              "lte": "2018-10-31 23:59:59",
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
        "field":"first_log_time",
        "format":"yyyy-MM-dd HH:mm:ss",
        "interval":"day",
        "time_zone":"+08:00",
        "min_doc_count":0,
        "extended_bounds":{
          "min":"2018-10-01 00:00:00",
          "max":"2018-10-31 23:59:59"
        }
      },
      "aggs": {
        "unique_users_count": {
          "cardinality": {
            "field": "user_id"
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
  "took": 2,
  "timed_out": false,
  "_shards": {
    "total": 10,
    "successful": 10,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 12,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "histogram": {
      "buckets": [
        {
          "key_as_string": "2018-10-01 00:00:00",
          "key": 1538323200000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-02 00:00:00",
          "key": 1538409600000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-03 00:00:00",
          "key": 1538496000000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-04 00:00:00",
          "key": 1538582400000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-05 00:00:00",
          "key": 1538668800000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-06 00:00:00",
          "key": 1538755200000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-07 00:00:00",
          "key": 1538841600000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-08 00:00:00",
          "key": 1538928000000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-09 00:00:00",
          "key": 1539014400000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-10 00:00:00",
          "key": 1539100800000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-11 00:00:00",
          "key": 1539187200000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-12 00:00:00",
          "key": 1539273600000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-13 00:00:00",
          "key": 1539360000000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-14 00:00:00",
          "key": 1539446400000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-15 00:00:00",
          "key": 1539532800000,
          "doc_count": 6,
          "unique_users_count": {
            "value": 2
          }
        },
        {
          "key_as_string": "2018-10-16 00:00:00",
          "key": 1539619200000,
          "doc_count": 6,
          "unique_users_count": {
            "value": 4
          }
        },
        {
          "key_as_string": "2018-10-17 00:00:00",
          "key": 1539705600000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-18 00:00:00",
          "key": 1539792000000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-19 00:00:00",
          "key": 1539878400000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-20 00:00:00",
          "key": 1539964800000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-21 00:00:00",
          "key": 1540051200000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-22 00:00:00",
          "key": 1540137600000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-23 00:00:00",
          "key": 1540224000000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-24 00:00:00",
          "key": 1540310400000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-25 00:00:00",
          "key": 1540396800000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-26 00:00:00",
          "key": 1540483200000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-27 00:00:00",
          "key": 1540569600000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-28 00:00:00",
          "key": 1540656000000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-29 00:00:00",
          "key": 1540742400000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-30 00:00:00",
          "key": 1540828800000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
          }
        },
        {
          "key_as_string": "2018-10-31 00:00:00",
          "key": 1540915200000,
          "doc_count": 0,
          "unique_users_count": {
            "value": 0
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

統計 `app_id` 為 **`csbot`**，且資料介於 **`2018-10-01 00:00:00`** 與 **`2018-10-31 23:59:59`**，且 `platform` 欄位不為 **`空字串`**，結果依照平台 **`(group_by_platform)`** 分群。但由於為了做到基於 `維度` 的統計，因此 `emotibot-users-*` 的 document ID 其實是透過 **`user_id` + `custom_info`** hash 過後得來的。因此有可能有同一個使用者，但維度不同而導致產生多個 documents，所以最後還必須再做一次 `cardinality aggregation` 排除重複的 `user_id` 後方可計算出正確的新增用戶數：

```
POST /emotibot-users-csbot-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "exists": {
            "field": "custom_info.platform.keyword"
          }
        },
        {
          "range": {
            "first_log_time": {
              "gte": "2018-10-01 00:00:00",
              "lte": "2018-10-31 23:59:59",
              "format": "yyyy-MM-dd HH:mm:ss",
              "time_zone": "+08:00"
            }
          }
        }
      ]
    }
  },
  "aggs": {
    "group_by_platform": {
      "terms": {
        "field": "custom_info.platform.keyword",
        "size": 3000000
      },
      "aggs": {
        "unique_users_count": {
          "cardinality": {
            "field": "user_id"
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
    "total": 10,
    "successful": 10,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 12,
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
          "doc_count": 6,
          "unique_users_count": {
            "value": 2
          }
        },
        {
          "key": "web",
          "doc_count": 5,
          "unique_users_count": {
            "value": 5
          }
        },
        {
          "key": "android",
          "doc_count": 1,
          "unique_users_count": {
            "value": 1
          }
        }
      ]
    }
  }
}
```
