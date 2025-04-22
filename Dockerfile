FROM golang:1.24-alpine

LABEL org.opencontainers.image.authors="s2410455005@fhooe.at"

WORKDIR /src

COPY . .

RUN go mod tidy
RUN go build -o app

EXPOSE 8010

ENTRYPOINT ["./app"]
