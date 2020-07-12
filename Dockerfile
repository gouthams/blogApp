FROM golang:1.14.4-alpine
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.6/main' >> /etc/apk/repositories
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.6/community' >> /etc/apk/repositories
RUN apk update
RUN apk add --no-cache mongodb-tools mongodb

RUN mkdir -p /app
COPY /server /app/server
COPY go.mod /app
COPY start.sh /app
WORKDIR /app

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN go get /app/server/... && \
    go build -o blog /app/server/main.go

VOLUME /data/db
ENV GIN_MODE=release
EXPOSE 8080

RUN chmod +x /app/start.sh
CMD ["sh", "-c", "/app/start.sh"]

