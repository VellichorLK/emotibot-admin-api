
# Usage
### Prerequsite
Docker version has to support multi-stage build.
Version 17.05.0 or later.  

### Build docker
```
./build.sh
```
### Run docker
```
./run.sh
```
### Swagger UI
http://\<ip\>:\<port\>  
default port is 7878

# Structure
The service composes of three components, web service, queue (rabbitmq) and workers. Each component has its dockerfile, build script, run script and enviroment file in docker directory. User can change the value in test.env to adhere to the enviroment setting.  

Web service use default handler to receive http requests and format it into json format as below, then push into assigned queue, which is defined in serviceCfg.yaml, in rabbitmq. The worker receives the task and does the relative job, then return the result.
```
task json format

{
    method:"GET",
    path:"/doGolang",
    query:"n=10",
    body:""
}

```

```
directory tree

├── build.sh
├── rabbitmq
│   └── docker
│       ├── build.sh
│       ├── Dockerfile
│       ├── run.sh
│       └── test.env
├── README.md
├── run.sh
├── vendor
│   ├── github.com
│   │   └── streadway
│   └── gopkg.in
│       └── yaml.v2
├── web_service
│   ├── docker
│   │   ├── build.sh
│   │   ├── Dockerfile
│   │   ├── run.sh
│   │   └── test.env
│   ├── handlers
│   │   ├── channelController.go
│   │   ├── default.go
│   │   ├── rabbitmq.go
│   │   ├── register.go
│   │   ├── register_test.go
│   │   └── typedef.go
│   ├── html
│   │   ├── favicon-16x16.png
│   │   ├── favicon-32x32.png
│   │   ├── index.html
│   │   ├── oauth2-redirect.html
│   │   ├── serviceCfg.yaml
│   │   ├── swagger-ui-bundle.js
│   │   ├── swagger-ui-bundle.js.map
│   │   ├── swagger-ui.css
│   │   ├── swagger-ui.css.map
│   │   ├── swagger-ui.js
│   │   ├── swagger-ui.js.map
│   │   ├── swagger-ui-standalone-preset.js
│   │   └── swagger-ui-standalone-preset.js.map
│   └── httpd.go
└── workers
    ├── golang_worker
    │   ├── docker
    │   ├── goworker.go
    │   └── RabbitMQTool
    ├── java_worker
    │   ├── docker
    │   ├── entrypoint.sh
    │   └── rabbitmq
    └── python_worker
        ├── docker
        └── pika_server.py
```

# How to add a new worker
1.  Add your service information to service configuration file, serviceCfg.yaml, which conforms to  [swagger spec](http://swagger.io/docs/specification/what-is-swagger/).  
    a.  open the serviceCfg.yaml, located in web_service/html/ folder.  
    b.  add a new tag which would group your api into a block in swagger UI.    
    ```
    Add new tag example.

    tags:
        -name: test_service
        -description: this is api for test.
    ```

    c.  add a new path for your service and its supported method and relative information, including return type (produces), input paramters, queue name, handler etc.
    ```
    Add new service path example.

    Define your service url path, /doTest.
    Define your supported method, get.
    x-queue is queue name which web service puts the request from /doTest to in rabbitmq. So the worker has to wait in this queue to receive task.
    x-handler defines which handler is used to handle the /doTest request. Using default handler would simply put the method, path, query, body into json format.

    paths:
        /doTest:
            get:
                tags:
                    - test_service
                summary: do test service
                description: this api will call test service
                parameters:
                    - name: teststring
                    type: string
                    in: query
                produces:
                    - text/plain
                responses:
                    200:
                        $ref: '#/responses/Description200'
                    400:
                        $ref: '#/responses/Description400'
                    404:
                        $ref: '#/responses/Description404'
                    408:
                        $ref: '#/responses/Description408'
                    429:
                        $ref: '#/responses/Description429'
                    503:
                        $ref: '#/responses/Description503'
            x-handler: default
            x-queue: test_task
    ```
2. Writing your worker code, of course. There are three types of program language in the workers folder to demonstrate how to commnuicate with rabbitmq. User can use wrapped SDK in the foler or use original (lower level) SDK provided by [official](https://www.rabbitmq.com/getstarted.html). Either way is ok. 

    *Notice! When using SDK, the queue name must be identical to the name (x-queue) you defined in serviceCfg.yaml*  

# How to defined your handler
As we mention above, the *default* handler put path, method, query, body into the json format. It may not suit to every one. One may want more information, like http header or just doesn't like json format.  

Write your handler with package name *handlers* and put it in the web_service/handlers folder. The handler must follow the golang http handler function format like below:
```golang
package handlers

func YourHandler(w http.ResponseWriter, r *http.Request) {
    //do your customize thing ...
}
```

Add your handler to *RegisterAllHandlers* function in the file web_service/handlers/register.go.

```golang
//RegisterAllHandlers register all usable module handler
func RegisterAllHandlers() {
	handlers = make(ModuleHandleFunc)
	funcSupport = make(FuncSupportMethod)

	handlers["default"] = DefaultHandler

	handlers["key_of_your_handler"] = YourHandler
	//add your handler here

	setHTTPPath()
}

```

Finally, declare to use your handler in the serviceCfg.yaml.
```
    x-handler: key_of_your_handler
```
