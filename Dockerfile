# stage1 builds bigbot STATICALLY in a fatter development environment
FROM golang:latest AS build

LABEL maintainer="duck. <me@duck.moe>"

WORKDIR /go/src/github.com/thebiggame/bigbot

ADD . /go/src/github.com/thebiggame/bigbot

RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/bigbot

# stage2 moves the static binary into a super ultra lean image
FROM scratch

WORKDIR /app

COPY --from=build /go/src/github.com/thebiggame/bigbot/main ./bigbot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["/app/bigbot"]