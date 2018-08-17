# OPEN API adapter

為了將 openAPI 1.0 的 api 導向 BFOP openAPI 的規格:
openapi v1的[文件](http://console.developer.emotibot.com/api/ApiKey/documentation.php)

## require env

- OPENAPI_URL: string, 轉打的 BFOP OPEN API 網址與端口，需要是完整的 url schema 格式
- MODULE_LEVEL: int, LOG 等級(0:印出, >0:不印出)，若無也不印出

## Docker image

use ./docker/build.sh to build latest tag
TAG: latest tag can be found in ./docker/VERSION
image name: docker-reg.emotibot.com.cn:55688/openapi-adapter:<TAG>
