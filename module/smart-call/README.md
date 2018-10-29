# Vip Backend 2.0

取代原有 Houta 的唯品會管理介面 API, 每一個 package 為一個 module

## How to build & run

``` bash
# build image base on current date & tag
# also create a latest tag for local debug
./docker/build.sh

# run standalone admin-api at 8182 port
# if no tag is given, use latest
./run.sh [tag_name]
```

開發完成後, 部署 docker image 版本會將環境變數透過 entrypoint.sh 以及 env.template 轉換成

```
# Add prefix for every line of env
# So you can copy & paste to real production env
./add_prefix.sh ../local.env

# Remember also add new env into template file to active variable in production
vim ./docker/env.template
```

## Module List

* Dictionary 詞庫
* FAQ 問答 問答庫API(unimplemented)
* QA 問答 問答檔案上下傳/ QATest 問答測試
* Stats 数据管理
* switch 開關管理
* UI UI專用的 API 提供伺服器設定
* Robot 機器人設置
* initFiles 後台 Houta init 那包檔案專用

## Contribution


### How to Add a module

1. create your own package
1. create controller.go
1. Defined your moduleInfo (Remember your module name should match the privilege db)
1. Update server.go setRoute > add your module to modules.

### How to write Unit Test

1. New an iris app
1. Setup Routing
1. Get a httptest instance
1. Simulate http request

[Example](https://github.com/kataras/iris/blob/master/_examples/testing/httptest/main_test.go)

