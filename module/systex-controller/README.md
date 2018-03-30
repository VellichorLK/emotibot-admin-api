# 精誠DEMO controller

精誠打了沒案件, 將Voice轉文字 再轉給TaskEngine處理。

Component:

- ASR: 語音組模組, 將聲音轉成文字. wav 檔案要求8k sample rate, api: /client/dynamic/recognize.
- TaskEngine 負責接收文字轉成任務. 
- Controller 

## How to Run

controller以外目前由各組自行部署。
controller example:

```bash
./docker/run.sh $PORT
```
