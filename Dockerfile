FROM golang:1.23-alpine as build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/gateway        ./cmd/gateway && \
    CGO_ENABLED=0 go build -o /out/worker         ./cmd/worker && \
    CGO_ENABLED=0 go build -o /out/usage-consumer ./cmd/usage-consumer && 

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 app

WORKDIR /app
COPY --from=build /out/ /app/
USER app

EXPOSE 8080
CMD ["/cmd/gateway"]


