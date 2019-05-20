ARG baseImageName
FROM ${baseImageName}

COPY vendors ${GOPATH}/src/
RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/kataras/iris
RUN go get github.com/kataras/golog
RUN go get github.com/kataras/pio
RUN go get github.com/tealeg/xlsx
RUN go get github.com/hashicorp/go-version
RUN go get github.com/hashicorp/consul/api
RUN go get github.com/bitly/go-simplejson
RUN go get github.com/gonum/blas
RUN go get github.com/gonum/internal/asm/f32
RUN go get github.com/gonum/internal/asm/f64
RUN go get github.com/gorilla/mux
RUN go get github.com/dgrijalva/jwt-go
RUN go get github.com/siongui/gojianfan
RUN go get github.com/satori/go.uuid
RUN go get gopkg.in/olivere/elastic.v6
RUN go get github.com/olivere/elastic/config
RUN go get github.com/olivere/elastic/uritemplates
RUN go get gopkg.in/DATA-DOG/go-sqlmock.v1
RUN go get github.com/rubenv/sql-migrate/...
RUN go get github.com/coreos/go-semver/semver
RUN go get gopkg.in/yaml.v2
RUN go get github.com/tidwall/gjson
RUN go get github.com/tidwall/sjson
RUN go get github.com/sahilm/fuzzy
RUN go get github.com/anknown/ahocorasick
RUN go get github.com/robfig/cron
RUN go get github.com/mediocregopher/radix
# rabbitmq client
RUN go get github.com/streadway/amqp
# Unit-test binary
RUN go get github.com/axw/gocov/...
RUN go get github.com/AlekSi/gocov-xml
RUN go get -u github.com/jstemmer/go-junit-report
RUN go get github.com/stretchr/testify
RUN go get github.com/mojocn/base64Captcha

RUN go get github.com/minio/minio-go
