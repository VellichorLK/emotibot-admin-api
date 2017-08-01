# Usage
### run docker
```
docker run -it -v /your/testfile/folder:/usr/src/app/testfile <docker image> /bin/sh
```

### run command in docker
single test
```
./tester -f testfile/your.wav -a Authorization:<appid> -h https://api-sh.emotibot.com
```
stress test
```
./tester -c 10 -p 300 -f testfile/your.wav -a Authorization:<appid> -h https://api-sh.emotibot.com -m stress
```

### arguments
```
 -a string
        authentication in header (default "Authorization:testappid")
  -c int
        concurrency, number of thread used to stress (default 1)
  -d int
        period of voice in second (default 143136)
  -f string
        the file to upload to server
  -h string
        url to send the task to. (default "https://api-sh.emotibot.com")
  -m string
        stress test or test single file (default "single")
  -p int
        second to continuely send the task (default 300)
  -t string
        file extension (default "wav")
```
