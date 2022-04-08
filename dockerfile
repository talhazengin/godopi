FROM golang:1.17-alpine AS builder

RUN apk update
RUN apk add git

WORKDIR /app
COPY . .

RUN go mod download

# Build Swagger files.
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN swag init -g internal/app/api/server/router.go

RUN go build -o godopi

FROM alpine:3.15.3 AS prod

COPY --from=builder /app/godopi /usr/local/bin/
COPY --from=builder /app/docs /usr/local/bin/docs

CMD ["godopi"]