FROM golang:1.24.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=build /out/server /app/server
COPY web /app/web

EXPOSE 8081

USER appuser

ENTRYPOINT ["./server"]