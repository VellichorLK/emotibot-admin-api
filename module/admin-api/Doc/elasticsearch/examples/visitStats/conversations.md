# Conversations (總會話)

### 依時間統計：

#### 各個時間區段的總會話數

統計 `enterprise_id` 為 **`emotibot`**、`app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，且 `session_id` 不為 **`空字串`**，結果依照 **`day`** 分群後，再依 **`(group_by_sessions)`** 分群：

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
      ],
      "must_not": [
        {
          "term": {
            "session_id": ""
          }
        }
      ]
    }
  },
  "aggs": {
    "histogram": {
      "date_histogram": {
        "field":"log_time",
        "format":"yyyy-MM-dd HH:mm:ss",
        "interval":"day",
        "time_zone":"+08:00",
        "min_doc_count":0,
        "extended_bounds":{
          "min":"2018-06-01 00:00:00",
          "max":"2018-06-30 23:59:59"
        }
      },
      "aggs": {
        "group_by_sessions": {
          "terms":{
            "field":"session_id",
            "size":3000000
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
    "total": 69,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "histogram": {
      "buckets": [
        {
          "key_as_string": "2018-06-01 00:00:00",
          "key": 1530374400000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-02 00:00:00",
          "key": 1530460800000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-03 00:00:00",
          "key": 1530547200000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-04 00:00:00",
          "key": 1530633600000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-05 00:00:00",
          "key": 1530720000000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-06 00:00:00",
          "key": 1530806400000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-07 00:00:00",
          "key": 1530892800000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-08 00:00:00",
          "key": 1530979200000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-09 00:00:00",
          "key": 1531065600000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-10 00:00:00",
          "key": 1531152000000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-11 00:00:00",
          "key": 1531238400000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-12 00:00:00",
          "key": 1531324800000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-13 00:00:00",
          "key": 1531411200000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-14 00:00:00",
          "key": 1531497600000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-15 00:00:00",
          "key": 1531584000000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-16 00:00:00",
          "key": 1531670400000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-17 00:00:00",
          "key": 1531756800000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-18 00:00:00",
          "key": 1531843200000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-19 00:00:00",
          "key": 1531929600000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-20 00:00:00",
          "key": 1532016000000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-21 00:00:00",
          "key": 1532102400000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-22 00:00:00",
          "key": 1532188800000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-23 00:00:00",
          "key": 1532275200000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-24 00:00:00",
          "key": 1532361600000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-25 00:00:00",
          "key": 1532448000000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-26 00:00:00",
          "key": 1532534400000,
          "doc_count": 12,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 7
              },
              {
                "key": "b6b0fdaf364647ae8445ea509007e05b",
                "doc_count": 5
              }
            ]
          }
        },
        {
          "key_as_string": "2018-06-27 00:00:00",
          "key": 1532620800000,
          "doc_count": 14,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 14
              }
            ]
          }
        },
        {
          "key_as_string": "2018-06-28 00:00:00",
          "key": 1532707200000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-29 00:00:00",
          "key": 1532793600000,
          "doc_count": 0,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-30 00:00:00",
          "key": 1532880000000,
          "doc_count": 43,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 43
              }
            ]
          }
        }
      ]
    }
  }
}
```

### 依維度統計：

#### 在所篩選的時間範圍內，各個維度的總會話數
##### (以平台 (platform) 維度為例)

統計 `enterprise_id` 為 **`emotibot`**、`app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，且 `session_id` 及 `platform` 不為 **`空字串`**，結果依照平台 **`(group_by_platform)`** 分群後，再依 **`(group_by_sessions)`** 分群：

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
        },
        {
          "exists": {
            "field": "custom_info.platform.keyword"
          }
        }
      ],
      "must_not": [
        {
          "term": {
            "session_id": ""
          }
        }
      ]
    }
  },
  "aggs": {
    "group_by_platform": {
      "terms": {
        "field": "custom_info.platform.keyword"
      },
      "aggs": {
        "group_by_sessions": {
          "terms":{
            "field":"session_id",
            "size":3000000
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
          "doc_count": 2,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 2
              }
            ]
          }
        },
        {
          "key": "微信",
          "doc_count": 5,
          "group_by_sessions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "5956818c88e811e882253d00999cd970",
                "doc_count": 5
              }
            ]
          }
        }
      ]
    }
  }
}
```
