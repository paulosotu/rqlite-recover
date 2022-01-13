FROM golang:1.17.1-alpine3.14 AS builder
RUN export GOPATH=/go
RUN apk update
RUN apk upgrade
RUN apk add --update gcc=10.3.1_git20210424-r2 g++=10.3.1_git20210424-r2
RUN apk --no-cache add   \
        git              
RUN mkdir -p /app
WORKDIR /app
COPY go.* /app/
COPY *.go /app/
COPY pkg /app/pkg
COPY cmd /app/cmd
COPY bin /app/bin
COPY vendor /app/vendor

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/rqlite-recover -ldflags '-linkmode external -w -extldflags "-static"' /app/cmd/rqlite-recover/main.go

FROM alpine:latest

RUN apk update
RUN apk upgrade

COPY --from=builder /app/bin/rqlite-recover /bin/rqlite-recover
COPY ./scripts/run.sh /bin/run.sh
EXPOSE 22
RUN chmod 777 /bin/run.sh

CMD ["sh", "/bin/run.sh"]

