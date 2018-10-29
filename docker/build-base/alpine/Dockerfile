ARG VERSION
FROM golang:${VERSION}-alpine


RUN apk update && apk add --no-cache git tzdata 
# Need this line, because there are packages import c & require gcc 
# source: https://github.com/mattn/go-sqlite3/issues/212
RUN apk add --no-cache --update gcc alpine-sdk
