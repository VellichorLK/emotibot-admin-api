
# Stage 1: Build go binary
FROM golang:1.8-alpine AS build-env

# Create the workspace
ENV GOPREFIX /go/src/emotibot.com/emotigo
RUN mkdir -p ${GOPREFIX}

# Add entire emotigo repo to gopath
ADD . ${GOPREFIX}

# Build binary
ARG PROJECT
RUN cd ${GOPREFIX}/${PROJECT} && go install

# Stage 2: A small container that holds the binary
FROM alpine:3.5

COPY --from=build-env /go/bin/proxy /usr/local/bin/proxy
WORKDIR /usr/local/bin
CMD ["/usr/local/bin/proxy"]
