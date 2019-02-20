# QIC back-end API

取代原本 python 的 qic-controller back-end API, 每一個 package 為一個 module

## How to build & run

### Docker version

``` bash
# build image base on current date & tag
# also create a latest tag for local debug
./docker/build.sh

# run standalone admin-api at 8182 port
# if no tag is given, use latest
./docker/run.sh [tag_name]
```
### Native go version

should support go v1.10~11
先將docker/docekr-compose.yaml內的環境變數 copy 成一個local .env 檔案, 並把env內變數填值
``` bash
cd ./module/qic-api/
# retrive all dependencies
go get
go run server.go your.env
```

## CI FLow

all CI script is in ./jenkins folder
jenkins link [here]()


## Module List

- manual 人工質檢
- model 共用的db model
- cu 舊流程質檢功能
- qi 資源設定&上傳音檔&質檢功能
- sensitive 敏感詞
- setting 設定
- util 共用package(請勿直接在util package 下直接寫code)

## Contribution


### How to Add a module

1. create your own package
1. create controller.go
1. Defined your moduleInfo (Remember your module name should match the privilege db)
1. Update server.go setRoute > add your module to modules.

### TODO: How to write Unit Test
