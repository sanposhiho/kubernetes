FROM golang:1.16.6 AS build-env

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/simulator-server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o ./bin/simulator simulator.go

FROM alpine:3.14.0

#RUN apk add --no-cache bash ca-certificates curl

COPY --from=build-env /go/src/simulator-server/bin/simulator /simulator
RUN chmod a+x /simulator

EXPOSE 8080
CMD ["/simulator"]
