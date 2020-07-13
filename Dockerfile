FROM golang:1.14.4-alpine AS base
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.6/main' >> /etc/apk/repositories
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.6/community' >> /etc/apk/repositories
RUN apk update
RUN apk add --no-cache mongodb-tools mongodb

FROM base AS builder
RUN mkdir -p /app
COPY /server /app/server
COPY go.mod /app
COPY *.sh /app/
WORKDIR /app
ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on
VOLUME /data/db
RUN go get /app/server/... && \
    go build -o /app/blog /app/server/main.go

FROM builder AS test
COPY --from=builder /app /app
RUN chmod +x /app/startTest.sh
CMD ["sh", "-c", "/app/startTest.sh"]

FROM base AS prod
ENV GIN_MODE=release
EXPOSE 8080
COPY --from=builder /app/blog /app/
COPY start.sh /app
RUN chmod +x /app/start.sh
CMD ["sh", "-c", "/app/start.sh"]
