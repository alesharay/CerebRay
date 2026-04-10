FROM golang:1.26-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app
COPY --from=build /bin/server /bin/server
COPY db/migrations /migrations
USER app
EXPOSE 8080
ENTRYPOINT ["/bin/server"]
