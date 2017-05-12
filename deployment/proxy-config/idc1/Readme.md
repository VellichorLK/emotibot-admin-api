# Proxy ports on idc1

### FIXME:
Somehow we need to restart the goproxy-uid once a while.
Put this to crontab
```
# crontab -e

45 03 * * *  docker restart goproxy-uid >/dev/null 2>&1 #Runs at 3:45 am every day
```

### Dir description
 - goproxy-uid: a proxy that extracts the uid from the incoming request
 - debug-proxy-server: a debug server to find out the userid -> server mapping
 - haproxy-idc1: a haproxy server that does the user mapping
 - lesports-backup-nginx: a dummy server that used to reply overflow requests
 - ssl_frontend: ssl termination haproxy
 - tests: some tests, work in progress...

### BotID
 - shadow (Our main cluster) : 0
 - lesports(lele) : 1
 - TBD

### Port assignment rules:
 - 9000 + BotID * 10 : bot cluster entry
 - 9000 + BotID * 10 + 1: bot backup webserver (if any)
 - TBD

| Port | Service | Ext mapping |
| ---- | ------- | ---- |
| 9000 | golang uid header adder | idc.emotibot.com:80 |
| 9001 | shadow cluster entry| N/A |
| 9010 | lele bot cluster| bot.emotibot.com:80 |
| 9011 | lele backup webserver | N/A |
| TBD | | |
| 9443 | ssl termination proxy | idc.emotibot.com:443 |
| 9527 | haproxy stats | haproxy monitoring page |
