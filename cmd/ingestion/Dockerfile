FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY pkg/ ./pkg

COPY api/ ./api

COPY cmd/ingestion/ ./cmd/ingestion

COPY internal/ingestion ./internal/ingestion

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/service ./cmd/ingestion

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/service .

CMD [ "./service" ]