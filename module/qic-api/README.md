# QIC back-end API

取代原本 python 的 qic-controller back-end API, 每一個 package 為一個 module

## How to build & run

``` bash
# build image base on current date & tag
# also create a latest tag for local debug
./docker/build.sh

# run standalone admin-api at 8182 port
# if no tag is given, use latest
./run.sh [tag_name]
```

## Module List


## Contribution


### How to Add a module

1. create your own package
1. create controller.go
1. Defined your moduleInfo (Remember your module name should match the privilege db)
1. Update server.go setRoute > add your module to modules.

### TODO: How to write Unit Test
