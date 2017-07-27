# 语音情绪质检云服务系统 API 
日期 | 版本 | 说明
----|------|-----
2017.07.04 | 0.1 | 查询音档、上传音档 API
2017.07.14 | 0.2 | 新增音档列表、上传音档 API

## 认证机制
用户于注册后，会取得一组认证码，其后执行任何端口，必须将该认证码罝入标头中。
若未放置该认证码，则伺服器会返回 401 的错误。

```
Authorization: D8FI9WDYR7UYTWIZOUNWNMCU3Z38ZR3U
```

## API 列表
API | 端口 | 说明
--------|-----|----
音档列表 | GET /voice/emotion/v1/files | 列出符合查询条件下的音档
音档列表(续) | POST /voice/emotion/v1/files_continue | 继续取得后续的资料 (列表过大时)
音档上传 | POST /voice/emotion/v1/upload | 上传音档


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
ch1\_anger_score | 查询声道1憤怒值最低分
ch2\_anger_score | 查询声道2憤怒值最低分


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
