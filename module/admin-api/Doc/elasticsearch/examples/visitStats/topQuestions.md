# Top Questions (熱點問題)

#### 前 N 個熱點問題
##### (以 N = 20 為例)

統計 `app_id` 為 **`csbot`** 且資料介於 **`2018-06-01 00:00:00`** 與 **`2018-06-30 23:59:59`**，數量總和前 **20** 個的使用者問題：

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
    "top_questions": {
      "terms": {
        "field": "user_q.keyword",
        "size": 20,
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
    "total": 1242,
    "max_score": 0,
    "hits": []
  },
  "aggregations": {
    "top_questions": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 271,
      "buckets": [
        {
          "key": "我要退货",
          "doc_count": 545
        },
        {
          "key": "welcome_tag",
          "doc_count": 157
        },
        {
          "key": "订酒店",
          "doc_count": 28
        },
        {
          "key": "1+1",
          "doc_count": 27
        },
        {
          "key": "p2p专家信仔为您解读",
          "doc_count": 25
        },
        {
          "key": "123",
          "doc_count": 24
        },
        {
          "key": "报数",
          "doc_count": 21
        },
        {
          "key": "e融视讯面签时客户联网核查不通过原因",
          "doc_count": 16
        },
        {
          "key": "app无法付款怎么办",
          "doc_count": 15
        },
        {
          "key": "是",
          "doc_count": 15
        },
        {
          "key": "美元兑欧元汇率",
          "doc_count": 15
        },
        {
          "key": "退货",
          "doc_count": 13
        },
        {
          "key": "上海",
          "doc_count": 10
        },
        {
          "key": "台北",
          "doc_count": 10
        },
        {
          "key": "北京",
          "doc_count": 9
        },
        {
          "key": "我要订酒店",
          "doc_count": 9
        },
        {
          "key": "p2p专区有什么用",
          "doc_count": 8
        },
        {
          "key": "一个敏感词",
          "doc_count": 8
        },
        {
          "key": "你好",
          "doc_count": 8
        },
        {
          "key": "换货",
          "doc_count": 8
        }
      ]
    }
  }
}
```
