# Elasticsearch + Logstash + Filebeat + Grafana Deployment
## Goal
1. View backend logs with visualization tools
2. Measure backend logs with specific terms.

## Flow
```
IDC hosts(filebeat - log collector)
        |
        |       (log analyze)   (full-text search db)
        +-----> logstash -----> elasticsearch
                   |      <IDC>       ^ 
                   |                  |
         <dev/sta> |                  |
                   |                  |
                   |               Kibana/Grafana (visualization)
                   V                  ^
                 mysql                |
                                      |
                                    Nginx (Authentication only for Kibana)
                                      ^
                                      |
                                      |
                                    Users

```

## Description
1. Curator is a tool to manage elasticsearch indice, here use to close index in 30 days.

## Code Tree
```
├── Readme.md
├── curator
│   ├── action_backup.yml
│   ├── action_close_index.yml
│   ├── action_delete_index.yml
│   ├── action_snapshot.yml
│   ├── config.yml
│   └── requirements.txt
├── database
│   ├── docker
│   │   ├── dev.env
│   │   ├── docker-compose.yml
│   │   ├── idc.env
│   │   ├── prod.env
│   │   ├── run.sh
│   │   └── sta.env
│   └── mysql
│       ├── backend_log.sql
│       ├── docker-entrypoint.sh
│       └── my.cnf
├── deploy_filebeat_idc.sh
├── elasticsearch
│   └── restart.sh
├── filebeat
│   ├── filebeat.dev.yml
│   ├── filebeat.idc.yml
│   ├── filebeat.yml
│   └── restart.sh
├── grafana
│   ├── grafana.ini
│   └── run.sh
├── kibana
│   ├── kibana.yml
│   └── restart.sh
├── logstash
│   ├── config-dir
│   │   ├── logstash.dev.conf
│   │   └── logstash.idc.conf
│   ├── docker
│   │   ├── Dockerfile
│   │   ├── build.sh
│   │   └── run.sh
│   └── plugin_jar
│       └── jdbc
│           └── mysql-connector-java-5.1.36-bin.jar
├── nginx
│   ├── kibana_htpasswd
│   ├── nginx.conf
│   └── run.sh
└── pull_image_from_SH.sh
```
## How to setup
* For IDC
1. Install order: *Elasticsearch/Curator -> Logstash -> Filebeat -> Grafana*
2. Install Elasticsearch
	
	```
	elasticsearch/restart.sh
	```
	
3. Deploy Logstash
	
	```
	logstash/restart.sh
	```
4. Deploy Filebeat 

	```
	# login deployment machine like 10.0.0.45
	deploy_filebeat_idc.sh
	```

5. Install Grafana
	
	```
	grafaha/run.sh
	```

6. Install Curator
	
	```
	virtualenv venv.curator
	source venv.curator/bin/activate
	pip install -r requirements.txt
	curator [--dry-run] --config config.yml action_close_index.yml 
	```

* For dev/sta
** Install order:  database -> logstash
1. database/docker/run.sh database/docker/dev.env
2. logstash/docker/build.sh ; logstash/docker/run.sh dev {IMAGE_TAG}
