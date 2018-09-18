# OpenAPI adapter

## OpenAPI 1.0
    將 OpenAPI 1.0 的 api 導向 BFOP controller (OpenAPI 2.0) 的規格:
    openapi v1 的[文件](http://console.developer.emotibot.com/api/ApiKey/documentation.php)

## OpenAPI 2.0
    作為一反向代理伺服器，直接將 request 轉發給 BFOP controller (OpenAPI 2.0)

## Required env

- OPENAPI_ADAPTER_SERVER_PORT: OpenAPI adapter server port (預設為 8080)
- OPENAPI_ADAPTER_EC_HOST: BFOP controller (OpenAPI 2.0) host url (需包含 http://xxx:yy prefix)
- OPENAPI_ADAPTER_DURATION: 限流器檢查頻率 (單位：秒)
- OPENAPI_ADAPTER_MAXREQUESTS: 同時最高可接受連線數
- OPENAPI_ADAPTER_BANPERIOD: 超過可接受連線數連線的拒絕時間 (單位：秒)
- OPENAPI_ADAPTER_LOGPERIOD: StatsD 採樣頻率 (單位：秒)
- OPENAPI_ADAPTER_STATSD_HOST: StatsD server IP
- OPENAPI_ADAPTER_STATSD_PORT: StatsD server port
- OPENAPI_LOG_LEVEL: Log 等級 (可用參數: ERROR, WARN, INFO, TRACE，預設為: INFO)

## Docker image

### Build
`./docker/build.sh`

### Run
`./docker/run.sh`
