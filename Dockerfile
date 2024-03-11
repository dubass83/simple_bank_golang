FROM golang:1.22.0-alpine3.19 AS builder

WORKDIR /app
COPY .  /app/
RUN go build -o main main.go

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/main /app
COPY --from=builder /app/app.env /app
COPY --from=builder /app/db/migration /app/db/migration
EXPOSE 8080

CMD ["main"]
