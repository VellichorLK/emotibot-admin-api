# Vip Backend 2.0

取代原有 Houta 的唯品會管理介面 API, 每一個 package 為一個 module

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

