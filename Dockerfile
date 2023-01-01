FROM golang:1.19.4-alpine3.17 AS build
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY *.go ./

RUN ls -l /app/
RUN go build -v ./cmd/transmission-exporter

FROM alpine:3.17
RUN apk add --update ca-certificates

COPY --from=build /app/transmission-exporter /usr/bin/transmission-exporter

EXPOSE 19091

ENTRYPOINT ["/usr/bin/transmission-exporter"]
