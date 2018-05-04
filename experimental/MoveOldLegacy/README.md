# MOL (MoveOldLegacy) Tool

這個小程式是用來scan 現有的/infrastructure/volumes/houta/Files/vip_answer_pic 圖片 將他插入到指定的MYSQL DB內
mol 參數有
```
    -p: 指定掃描的資料夾(ex: ./vip_answer_pic), 注意如果使用docker是mount的位址而非在host上的位址
    -db:插入的mysql db位址, 預設是localhost:3306
    -u: mysql user
    -pass: mysql password
    -loc: 找location 表上的URL value, 若找不到會報error.
```

ex:
docker run -i --rm docker-reg.emotibot.com.cn:55688/mol -p ./infrastructure/volumes/houta/Files/vip_answer_pic/ -db localhost:3306
-u root -pass password -loc vip.api.com