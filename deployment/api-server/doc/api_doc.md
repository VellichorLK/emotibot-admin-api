# 语音情绪质检云服务系统
## 文件修改记录
日期 | 版本 | 说明
----|------|-----
2017.06.26 | 0.1 | 初版记录，新增音档上传及查询端口
2017.06.27 | 0.2 | 增加认证机制说明，修改上传的端口，加入服務編號


## 认证机制
用户于注册后，会取得一组认证码，其后执行任何端口，必须将该认证码罝入标头中。
若未放置该认证码，则伺服器会返回 401 的错误。

```
Authorization: D8FI9WDYR7UYTWIZOUNWNMCU3Z38ZR3U
```

## 音档上傳
### 端口
```
POST https://cloud.emotibot.com/api/voice/v1/upload
```

### 内容类型 (Content-Type)
**multipart/form-data**

### 请求内容 (Requests Body)
Name | Description | Type | Data Type | Required | Default | Limitation 
-----| ----------- | ---- | ----------| -------- | --------| -----------
serviceId | 服务编号 | FormData | FormData | 是 | 英文+数字+dot，长度最多128
fileName | 音档档名 | FormData | File | 是 | 
fileType | 音档副档名，全英文小写 | FormData | String | 是 | 
duration | 音档总时长，以秒计 | FormData | Integer | 是   
createTime | 音档发生时间，时间戳格式 | FormData | Integer | 是
checksum | md5 检查码 | FormData | String | 否 
tag	 | 自定义标记值 | FormData | String | 否
priority | 优先级，笵围 0-4，越高的音档会优先被处理 | FormData | Integer | 否 | 2

**注：服务编号(serviceId) 为该音档的唯一识别码，若上传相同服务编号的音档，则后来的会将之前的音档覆写，统计资料也将会被覆写。**

### 响应 (Response)
无

### 响应状态码
Status Code | Description 
------------|------------ 
200 | OK
400 | Bad Request
401 | Unauthorized
403 | Forbidden
429 | Too Many Requests
503 | Service Unavailable

## 音档查询
### 端口
```
GET https://cloud.emotibot.com/api/voice/v1/emotion/<服务编号>
```
### 响应 (Response)
返回内容为该查询服务编号的属性值与分析结果。

Fields | Type | Description | Default
-------|------|------------ | -------
serviceId | String | 服务编号
fileName | String | 音档档名
fileType | String | 音档副档名
duration | Integer | 音档总时长，以秒计
size | Integer | 音档总大小，以 bytes 计
createdTime | Integer | 音档发生时间，时间戳格式
checksum | String | 检查码
tag | String | 自定义标记值
priority | Integer | 优先级
score | Float | 情绪分析分數值| -1 
manual | Json String | 人工检验结果 | {}


**人工检验结果说明 (manual 栏位)**

Field | Type | Description | Default  
------|------|-------------|--------
comment| String | 意见     | empty string
tag | String | 人工检验结果 | empty string

**Example**

```
{
	"serviceId": "20170626180005005",
	"fileName": "customer.20170626.mp3",
	"createTime": 1498475041
	...
	"score": 0.75,
	"manual": {
		"tag": "有情绪",
		"comment": ""
	}
}
```

### 响应状态码
Status Code | Description
------------|------------
200 | OK
400 | Bad Request
401 | Unauthorized
404 | Not Found
429 | Too Many Request
503 | Service Unavailable o b