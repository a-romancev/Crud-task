FROM golang:1.17-alpine

WORKDIR /app

RUN go install -tags 'mongodb' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1
