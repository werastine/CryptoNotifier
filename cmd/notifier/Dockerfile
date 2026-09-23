FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY pkg/ ./pkg

COPY api/ ./api

COPY cmd/notifier ./cmd/notifier

COPY internal/notifier ./internal/notifier

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/service ./cmd/notifier

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/service .

CMD [ "./service" ]