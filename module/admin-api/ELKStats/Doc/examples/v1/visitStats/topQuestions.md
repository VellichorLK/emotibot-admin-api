# Top Questions (熱點問題)

#### 前 N 個熱點問題
##### (以 N = 20 為例)

統計 `app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，數量總和前 **20** 個的 FAQ 標準問題：

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
          "term": {
            "module": "faq"
          }
        },
        {
          "range": {
            "log_time": {
              "gte": "2018-10-01 00:00:00",
              "lte": "2018-10-31 23:59:59",
              "format": "yyyy-MM-dd HH:mm:ss",
              "time_zone": "+08:00"
            }
          }
        }
      ],
      "must_not": {
        "term": {
          "std_q.keyword": ""
        }
      }
    }
  },
  "aggs": {
    "top_questions": {
      "terms": {
        "field": "std_q.keyword",
        "size": 20,
        "shard_size": 10000,
        "order": {
          "_count": "desc"
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
  "took": 8,
  "timed_out": false,
  "_shards": {
    "total": 190,
    "successful": 190,
    "skipped": 122,
    "failed": 0
  },
  "hits": {
    "total": 37,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "top_questions": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "APP无法付款怎么办",
          "doc_count": 15
        },
        {
          "key": "test",
          "doc_count": 12
        },
        {
          "key": "e家保理赔",
          "doc_count": 2
        },
        {
          "key": "一年期综合意外险可以预留被保险人信息吗",
          "doc_count": 2
        },
        {
          "key": "你好",
          "doc_count": 2
        },
        {
          "key": "一年期综合意外险保障时间",
          "doc_count": 1
        },
        {
          "key": "一直显示提现结算中怎么办",
          "doc_count": 1
        },
        {
          "key": "交强险保费费率实施时间",
          "doc_count": 1
        },
        {
          "key": "你在哪里",
          "doc_count": 1
        }
      ]
    }
  }
}
```
