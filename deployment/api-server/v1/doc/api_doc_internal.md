# 语音情绪质检云服务系统 API 
## 文件修改记录
日期 | 版本 | 说明
----|------|-----
2017.06.27 | 0.1 | 初版记录，新增音档上传及查询端口
2017.06.28 | 0.2 | 增加认证机制说明，修改上传的端口，加入服務編號
2017.06.29 | 0.3 | 增加统计报表端口，回传值未定；修改属性端口回传值，加入频道以及三种分数呈现机制；新增上传端口回传值；新增音档列表端口；
2017.06.30 | 0.4 | 修正参数用词，所有参数改为全小写，复合字词用 "_" 隔开
2017.07.03 | 0.5 | 加入时间栏位说明
2017.07.04 | 0.6 | 加入统计报表 API 说明；查询 API 加入 cursor 值定义
2017.07.05 | 0.7 | 修正查询音档接口 by 14:30 会议
2017.07.12 | 0.8 | 調整查詢音档的 API, 分离详细/简要；cloud.emotibot.com 改成 api.emotibot.com；
2017.07.14 | 0.9 | 修正查詢音檔的接口，增加 tag2 的欄位；取消优先级设计；



## 认证机制
用户于注册后，会取得一组认证码，其后执行任何端口，必须将该认证码罝入标头中。
若未放置该认证码，则伺服器会返回 401 的错误。

```
Authorization: D8FI9WDYR7UYTWIZOUNWNMCU3Z38ZR3U
```

## API 列表
API | endpoint | description 
----|----------|--------------------
查询音档 | GET /voice/emotion/v1/files/<file_id> | 除了简要的输出以外，亦有每个音频单句的结果
音档列表 | GET /voice/emotion/v1/files | 列出时间内的音档，只含有简要内容
音档列表(续) | POST /voice/emotion/v1/files_continue | 继续取得后续的资料 (列表过大时)
音档上传 | POST /voice/emotion/v1/upload | 上传音档
统计报表(未定) | GET /voice/emotion/v1/report | 查询一定时间区间内的统计资料 


## 查询音档
### 端口
```
GET https://api.emotibot.com/voice/emotion/v1/files/<file_id>
```


### 响应 (Response)

Fields | Type | Description | Default
-------|------|------------ | -------
file_id | String | 音档识别码 
file_name | String | 音档档名
file_type | String | 音档副档名
duration | Integer | 音档总时长，以秒计
size | Integer | 音档总大小，以 bytes 计
created_time | Integer | 音档发生时间，时间戳格式 (epoch: 1970/01/01 (UTC))
checksum | String | 检查码
tag1 | String | 自定义标记值
tag2 | String | 自定义标记值
channels | List | 音档中每个音频分析結果 | []

**情緒分析結果說明 (channels 栏位)**

Field | Type | Description | Default 
------|------|-------------|---------
channel_id | Integer | 音档频道索引 | -1
result | List | 该音频情绪标签分数值 | []

**Example**

```
GET https://api.emotibot.com/voice/emotion/v1/files/ecovacs_customer.20170626180000

{
	"file_id": "ecovacs_customer.20170626180000",
	"file_name": "customer.20170626180000.mp3",
	"created_time": 1498475041
	...
	"channels": [
		{
			"channel_id": 0,
			"result": [
				{
					"label": "anger",
					"score": 50
				}
			],
			"vad_result": [
				{
					"index": 1,
					"start": 23,
					"end": 25,
					"result": [
						{
							"label": "anger",
							"score": 12
						}
					]
				},
				{
					"index": 2,
					...
				}
			]
		},
		{
			"channel_id": 1,
			"result": [
				{
					"label": "anger",
					"score": 73
				}
			]
		}
	]
}
```

## 音档列表
### 端口1
```
GET https://api.emotibot.com/voice/emotion/v1/files
```
**查询参数**

栏位 | 说明 | 预设值
----|-----|-------
t1  | 查询起始时间区间，时间截格式 (epoch: 1970/01/01 (UTC)) | 
t2  | 查询结束时间区间，时间截格式 (epoch: 1970/01/01 (UTC)) | 
file_name | 欲查询的档案名称，模糊比对 | 无 
status | 音档状态，支援 done、wait、all | all
tag1 | 音档上传使用的自订标签
tag2 | 音档上传使用的自订标签
ch1\_anger_score | 查询声道1憤怒值最低分，大于等于该值的音档资讯会返回
ch2\_anger_score | 查询声道2憤怒值最低分，大于等于该值的音档资讯会返回


*status 栏位值尚待确定

### 响应 (Response)

```
curl -XPOST --header "Authentication: D8FI9WDYR7UYTWIZOUNWNMCU3Z38ZR3U" \
	  -d '{"t1": 1471234567, "t2": 1471239087}' \
	  https://api.emotibot.com/voice/emotion/v1/files

{
	"total": 12345,
	"cursor": "1awx3drv4bgqki7ehf=ksdf",
	"result": [
		{
			"file_id": "ecovacs_20170625175859_sn20170625175859",
			"file_name": "20170625175859_sn20170625175859"
			"created_time": 1499064712
		},
		...
	]
}
```


## 端口2 (取得长列表资料)
* 若是响应过长(多于500笔资料)，则会切分返回的数据，以 cursor 值来作为取值的笵园的坐标

```
POST https://api.emotibot.com/voice/emotion/v1/files_continue
```

**Body 查询参数**

栏位 | 说明 | 预设值
----|------|------
cursor | 若是列表多于500笔，则会返回 cursor 值，实作者可将 cursor 代入 /voice/emotion/v1/files_continue API，则可得到后续500笔资料 | 无

### 响应 (Response)
```
curl -XPOST --header "Authentication: D8FI9WDYR7UYTWIZOUNWNMCU3Z38ZR3U" \
	  -d '{"cursor": "1awx3drv4bgqki7ehf=ksdf"}' \
	  https://api.emotibot.com/voice/emotion/v1/files_continue
	  
{
	"total": 12345,
	"cursor": "9ijn1qazxsw23edc-=jka1k",
	"result": [
		{
			"file_id": "ecovacs_20170625175859_sn1231488137",
			"file_name": "20170625175859_sn1231488137",
			"created_time": 1499064845
		},
		...
	]
}
```


## 音档上傳
### 端口
```
POST https://api.emotibot.com/voice/emotion/v1/upload
```

### 内容类型 (Content-Type)
**multipart/form-data**

### 请求内容 (Requests Body)
Name | Description | Type | Data Type | Required | Default | Limitation 
-----| ----------- | ---- | ----------| -------- | --------| -----------
file_name | 音档档名 | FormData | File | 是 | 
file_type | 音档副档名，全英文小写 | FormData | String | 是 | 
duration | 音档总时长，以秒计 | FormData | Integer | 是   
created_time | 音档发生时间，时间戳格式 (epoch: 1970/01/01 (UTC)) | FormData | Integer | 是
checksum | md5 检查码 | FormData | String | 否 
tag1 | 自定义标记值 | FormData | String | 否 | 長度 128 bytes
tag2 | 自定义标记值 | FormData | String | 否 | 长度 128 bytes


**请求示例**

```
# 上传本地端 target.mp3 至云端分析

curl --header "Authentication: D8FI9WDYR7UYTWIZOUNWNMCU3Z38ZR3U" \
  -F "file_name=target.mp3" \
  -F "file_type=mp3" \
  -F "duration=60" \
  -F "created_time=1499066627" \
  -F "file=@/path/to/your/local/target.mp3" \
  https://api.emotibot.com/voice/emotion/v1/upload
```

### 响应 (Response)
返回同查询返回值，但音频分析结果 channels 为空值或预设值

###示例代码

```
curl --header "Authentication: D8FI9WDYR7UYTWIZOUNWNMCU3Z38ZR3U" \
  -F "file_name=target.mp3" \
  -F "file_type=mp3" \
  -F "duration=60" \
  -F "created_time=1499066627" \
  -F "file=@/path/to/your/local/target.mp3" \
  https://api.emotibot.com/voice/emotion/v1/upload
  
# response:
{
	"file_id": "ecovacs_file_1499066627",
	"file_name": "target.mp3",
	"create_time": 1499066627
	...
	"channels": []
}
```



## 统计报表 (尚未開放)
### 端口
```
GET https://api.emotibot.com/voice/emotion/v1/report
```

**查询参数**

栏位 | 说明 | 预设值
----|-----|-------
t1  | 查询起始时间区间，时间戳格式 (epoch: 1970/01/01 (UTC)) |   
t2  | 查询结束时间区间，时间戳格式 (epoch: 1970/01/01 (UTC)) | 
export | 汇出的格式，目前只支援 csv ，未给定则输出 json | 无


### 响应(Response)

栏位 | 说明 | 预设值
----|------|-------
total | 音频总数
total_length | 音频总时长，以秒計 | 
total\_ana\_length | 意频总分析时长
result | 每日个别统计资料

**Example**

```
TBD
```

## 响应状态码
Status Code | Description
------------|------------
200 | OK
400 | Bad Request
401 | Unauthorized
404 | Not Found
410 | Gone | 该链结已过期，通常在查询介面，给定的 cursor 值已过期时回复
429 | Too Many Request
503 | Service Unavailable

