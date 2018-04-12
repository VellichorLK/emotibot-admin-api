# VCA 模擬SDK 4s 實驗


实验环境:
nginx 不设定timeout 转发
nginx 设定timeout:
智能提示: 250ms
因为竹间机器上，无法复现4s环境，因此人为加入假設的sdk timeout机率如下:
智能提示:
40-200ms: 99%
4s: 1%

架构:
- fakeSDK: 模擬SDK的component 99% sleep 100 ms 回覆，1%的機率會sleep 4s 才回覆
- Nginx server:模擬的轉發設定
- ab benchmark tool: 測試結果

## 压测结果
10000 个request，跑5次实验，取平均

4s timeout个数
不设定timeout 转发
~12
设定timeout转发
0

可得知nginx 设定转发，可有效减少转发小i的机率。
 
Nginx 設定
location /vip/irobot/get-questions.action {
  proxy_pass http://vca-sdk-cluster-tip;
  proxy_read_timeout 250ms;
}
 
upstream vca-sdk-cluster-tip {
server 172.16.101.227:8080 max_fails=0;
server 172.16.101.227:8081 max_fails=0;
server 172.16.101.227:8082 max_fails=0;
}
