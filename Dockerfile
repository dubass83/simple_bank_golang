FROM golang:1.21.7-alpine3.19 AS builder

WORKDIR /app
COPY .  /app/
RUN go build -o main main.go

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/main /app
COPY --from=builder /app/app.env /app
EXPOSE 8080

CMD ["main"]
