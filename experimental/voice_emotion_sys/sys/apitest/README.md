# Usage
### run docker
run.sh will mount the test file folder that user assigned to the /usr/src/app/testfile in docker
```
./run.sh /path/to/your/testfile/folder
```

### run command in docker
single test
```
./tester -f testfile/test.wav -a Authorization:<appid> -h https://api-sh.emotibot.com -t wav -r 1501656592 -d 143 -t1 test -s testfile/result.json
```
stress test
```
./tester -c 10 -p 300 -f testfile/test.wav -a Authorization:<appid> -h https://api-sh.emotibot.com -m stress
```

### arguments
```
  -a string
    	authentication in header (default "Authorization:testappid")
  -c int
    	concurrency, number of thread used to stress (default 1)
  -d int
    	period of voice in second (default 143)
  -f string
    	the file to upload to server
  -h string
    	url to send the task to. (default "https://api-sh.emotibot.com")
  -m string
    	stress test or test single file (default "single")
  -p int
    	second to continuely send the task (default 300)
  -s string
    	save the resutl in json format to the assigned file. Only used at single mode
  -t string
    	file extension (default "wav")
```
