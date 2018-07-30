# Active Users (活躍用戶)

### 依時間統計：

#### 各個時間區段的活躍用戶數

此數據無法直接透過 ElasticSearch 計算，因此先統計 `enterprise_id` 為 **`emotibot`**、`app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，結果依照 **`day`** 分群。在各時間分群的結果中，再依各使用者 **`(group_by_users)`** 分群。在各使用者分群中，再篩選出記錄超過 *`activeUsersThreshold (10)`* 的使用者。最後再解析 ElasticSearch 所統計出來的使用者列表，個別計算各個時間分群中的使用者個數而得到活躍用戶數：

```
POST /records/_search
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
      },
      "aggs": {
        "group_by_users": {
          "terms": {
            "field": "user_id",
            "size": 3000000
          },
          "aggs": {
            "active_user_filter": {
              "bucket_selector": {
                "buckets_path": {
                  "DocCount": "_count"
                },
                "script": "params.DocCount > 10"
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
          "doc_count": 286,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 285
              }
            ]
          }
        },
        {
          "key_as_string": "2018-06-02 00:00:00",
          "key": 1527897600000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-03 00:00:00",
          "key": 1527984000000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-04 00:00:00",
          "key": 1528070400000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-05 00:00:00",
          "key": 1528156800000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-06 00:00:00",
          "key": 1528243200000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-07 00:00:00",
          "key": 1528329600000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-08 00:00:00",
          "key": 1528416000000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-09 00:00:00",
          "key": 1528502400000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-10 00:00:00",
          "key": 1528588800000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-11 00:00:00",
          "key": 1528675200000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-12 00:00:00",
          "key": 1528761600000,
          "doc_count": 1,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-13 00:00:00",
          "key": 1528848000000,
          "doc_count": 2,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-14 00:00:00",
          "key": 1528934400000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-15 00:00:00",
          "key": 1529020800000,
          "doc_count": 14,
          "group_by_users": {
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
          "key_as_string": "2018-06-16 00:00:00",
          "key": 1529107200000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-17 00:00:00",
          "key": 1529193600000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-18 00:00:00",
          "key": 1529280000000,
          "doc_count": 9,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-19 00:00:00",
          "key": 1529366400000,
          "doc_count": 555,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "123",
                "doc_count": 535
              },
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 20
              }
            ]
          }
        },
        {
          "key_as_string": "2018-06-20 00:00:00",
          "key": 1529452800000,
          "doc_count": 15,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-21 00:00:00",
          "key": 1529539200000,
          "doc_count": 85,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 67
              },
              {
                "key": "123123123",
                "doc_count": 12
              }
            ]
          }
        },
        {
          "key_as_string": "2018-06-22 00:00:00",
          "key": 1529625600000,
          "doc_count": 8,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-23 00:00:00",
          "key": 1529712000000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-24 00:00:00",
          "key": 1529798400000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-25 00:00:00",
          "key": 1529884800000,
          "doc_count": 178,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "1234",
                "doc_count": 50
              },
              {
                "key": "ed5a4d7c154a4efe9b7e1efa22a26743",
                "doc_count": 42
              },
              {
                "key": "123",
                "doc_count": 39
              },
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 30
              },
              {
                "key": "8d2d0adbb71f4c63ae7a276c1327db2d",
                "doc_count": 16
              }
            ]
          }
        },
        {
          "key_as_string": "2018-06-26 00:00:00",
          "key": 1529971200000,
          "doc_count": 89,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 31
              },
              {
                "key": "8d2d0adbb71f4c63ae7a276c1327db2d",
                "doc_count": 29
              },
              {
                "key": "ed5a4d7c154a4efe9b7e1efa22a26743",
                "doc_count": 27
              }
            ]
          }
        },
        {
          "key_as_string": "2018-06-27 00:00:00",
          "key": 1530057600000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-28 00:00:00",
          "key": 1530144000000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-29 00:00:00",
          "key": 1530230400000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key_as_string": "2018-06-30 00:00:00",
          "key": 1530316800000,
          "doc_count": 0,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        }
      ]
    }
  }
}
```

### 依維度統計：

#### 在所篩選的時間範圍內，各個維度的活躍用戶數
##### (以平台 (platform) 維度為例)

此數據無法直接透過 ElasticSearch 計算，因此先統計 `enterprise_id` 為 **`emotibot`**、`app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，且 `platform` 欄位不為 **`空字串`**，結果依各平台 **`(group_by_platforms)`** 分群後，再依各使用者 **`(group_by_users)`** 分群。在各使用者分群中，再篩選出記錄超過 *`activeUsersThreshold (10)`* 的使用者。最後再解析 ElasticSearch 所統計出來的使用者列表，個別計算各個平台分群中的使用者個數而得到活躍用戶數：

```
POST /records/_search
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
      ]
    }
  },
  "aggs": {
    "group_by_platforms": {
      "terms": {
        "field": "custom_info.platform.keyword",
        "shard_size": 3000000
      },
      "aggs": {
        "group_by_users": {
          "terms": {
            "field": "user_id",
            "size": 3000000
          },
          "aggs": {
            "active_user_filter": {
              "bucket_selector": {
                "buckets_path": {
                  "DocCount": "_count"
                },
                "script": "params.DocCount > 10"
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
  "took": 6,
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
    "group_by_platforms": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "android",
          "doc_count": 39,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "1234",
                "doc_count": 25
              },
              {
                "key": "4b21158a395311e88a710242ac110003",
                "doc_count": 14
              }
            ]
          }
        },
        {
          "key": "微信",
          "doc_count": 17,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "123",
                "doc_count": 11
              }
            ]
          }
        },
        {
          "key": "web",
          "doc_count": 2,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        },
        {
          "key": "ios",
          "doc_count": 1,
          "group_by_users": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": []
          }
        }
      ]
    }
  }
}
```
