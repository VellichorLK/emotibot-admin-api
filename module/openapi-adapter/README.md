# OpenAPI adapter

## OpenAPI 1.0
    將 OpenAPI 1.0 的 api 導向 BFOP controller (OpenAPI 2.0) 的規格:
    openapi v1 的[文件](http://console.developer.emotibot.com/api/ApiKey/documentation.php)

## OpenAPI 2.0
    作為一反向代理伺服器，直接將 request 轉發給 BFOP controller (OpenAPI 2.0)

## Required env

- OPENAPI_ADAPTER_SERVER_PORT: OpenAPI adapter server port (預設為 8080)
- OPENAPI_ADAPTER_EC_HOST: BFOP controller (OpenAPI 2.0) host url (需包含 http:// prefix)
- OPENAPI_LOG_LEVEL: Log 等級 (0: 印出, >0: 不印出)，若無給值也不會印出

## Docker image

### Build
`./docker/build.sh`

### Run
`./docker/run.sh`
