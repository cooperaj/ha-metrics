FROM golang:1.18-alpine

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
    
RUN apk --no-cache add git build-base ca-certificates \
  && go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR $GOPATH/src/go.acpr.dev/ha-metrics/
COPY . .

RUN go mod download
