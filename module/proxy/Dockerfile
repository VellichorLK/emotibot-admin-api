
# Stage 1: Build go binary
FROM golang:1.9-alpine AS build-env

# Create the workspace
WORKDIR /go/src/emotibot.com/emotigo/module/proxy

# Build binary
COPY ./ ./
RUN go install

# Stage 2: A small container that holds the binary
FROM alpine:3.5

COPY --from=build-env /go/bin/proxy /usr/local/bin/proxy
ADD k8slist /usr/local/bin/k8slist
WORKDIR /usr/local/bin
CMD ["/usr/local/bin/proxy"]
