FROM golang:1.14-alpine

ENV GO111MODULE=on
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
    
RUN apk --no-cache add git build-base ca-certificates \
  && go get github.com/go-delve/delve/cmd/dlv

WORKDIR $GOPATH/src/go.acpr.dev/ha-metrics/
COPY . .

RUN go mod download
