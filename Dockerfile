FROM golang:1.23

WORKDIR /usr/app

COPY . .

RUN go mod tidy
RUN go install github.com/air-verse/air@latest
RUN go mod download
