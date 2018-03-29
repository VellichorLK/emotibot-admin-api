# RF手動匯入版本

透過csv 檔案來匯入, csv只取第一個column內的值
預設Csv已打包到docker image裡面, 但可以透過 `-v xxx.csv:/usr/bin/data.csv` 來取代.

*原本的RFQuestions 會被清除掉後再匯入!*

example
```bash
docker run -i docker-reg.emotibot.com.cn:55688/RFImporter --address root:password@tcp\(172.16.101.47:3306\) --consul http://172.16.101.47:8500/v1/kv/idc/
```