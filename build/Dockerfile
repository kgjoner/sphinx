#################################################
# BUILDER
#################################################
FROM golang:1.25.5 AS builder
WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN export GOPRIVATE=github.com/kgjoner/*
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url.ssh://git@github.com/.insteadOf https://github.com/

COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download

COPY . ./
RUN swag init -g server.go --dir internal/server,$(find internal/domains -type d -path "*/*http" | paste -sd ",") --parseDependency --parseInternal

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/sphinx
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/migrate

############################
# FINAL
############################
FROM alpine:3.17 AS final

ARG APP_VERSION=unknown
ENV APP_VERSION=${APP_VERSION}

WORKDIR /

COPY --from=builder /app/sphinx /sphinx
COPY --from=builder /app/migrate /migrate

EXPOSE 8080

CMD ["/sphinx"]