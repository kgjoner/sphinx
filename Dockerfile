# syntax=docker/dockerfile:experimental
FROM golang:1.22.0 as development

WORKDIR /app

RUN export GOPRIVATE=github.com/kgjoner/cornucopia
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url.ssh://git@github.com/.insteadOf https://github.com/

COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /sphinx ./cmd

EXPOSE 8080

CMD ["/sphinx"]