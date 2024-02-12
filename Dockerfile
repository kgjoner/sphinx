# syntax=docker/dockerfile:experimental
FROM golang:1.19.0 as development

WORKDIR /app

RUN export GOPRIVATE=github.com/kgjoner/cornucopia

COPY go.mod go.sum ./
RUN --mount=type=secret,id=mynetrc,dst=/root/.netrc go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /sphinx ./cmd

EXPOSE 8080

CMD ["/sphinx"]