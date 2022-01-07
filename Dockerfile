FROM golang:1.17.1-alpine3.14 AS builder
RUN export GOPATH=/go
RUN apk update
RUN apk upgrade
RUN apk add --update gcc=10.3.1_git20210424-r2 g++=10.3.1_git20210424-r2
RUN apk --no-cache add   \
        git              
RUN mkdir -p /app
WORKDIR /app/
COPY go.* /app/
COPY *.go /app/
COPY services /app/services
COPY models /app/models
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine:latest

RUN apk update
RUN apk upgrade

COPY --from=builder /app/rqlite-recover /bin/rqlite-recover
COPY ./run.sh /bin/run.sh
EXPOSE 22
RUN chmod 777 /bin/run.sh

CMD ["sh", "/bin/run.sh"]

