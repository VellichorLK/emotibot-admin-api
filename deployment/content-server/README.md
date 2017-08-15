#### Commands to Run

* checkout this repo
* pull update

```
sh-agent bash -c 'ssh-add /home/deployer/.ssh/id_rsa_gitlab_deployer.secret.key; git pull --rebase'
```

* DEPLOY ORDER: database -> crawler -> webapi -> nginx

nginx 

```
nginx/run.sh {domain}

e.g. nginx/run.sh content-sh.emotibot.com
```
database

```
database/docker/run.sh database/docker/{ENV file}

e.g. database/docker/run.sh database/sta.env
```
crawler

```
crawler/docker/build.sh
crawler/docker/run.sh crawler/docker/{ENV file} {IMAGE TAG}
```

webapi

```
webapi/docker/build.sh
webapi/docker/run.sh webapi/docker/{ENV file} {IMAGE TAG}
```