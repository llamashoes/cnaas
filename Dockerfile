FROM golang:1.7.1

ENV GOROOT=/usr/local/go
ADD ./.vendor .vendor
ADD ./.tools .tools
ADD ./make.go make.go
ADD ./src src
RUN go run make.go -v
EXPOSE 8080

ENTRYPOINT /go/fights
