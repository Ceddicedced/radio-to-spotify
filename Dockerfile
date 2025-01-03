#build stage
FROM golang AS builder
RUN apt-get update && apt-get install -y --no-install-recommends git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -v -o /app main.go

#final stage
FROM debian:stable-slim
RUN apt update && apt install -y --no-install-recommends ca-certificates curl
WORKDIR /app
COPY --from=builder /app .
COPY stations.json /app/stations.json
ENTRYPOINT ["./app"]
LABEL Name=radiotospotify Version=0.0.3
ENV ENABLE_HEALTHCHECK=true
HEALTHCHECK --interval=3m --timeout=3s --start-period=1m \
  CMD curl -s http://localhost:8585/health | grep -q '"status":"healthy"' || exit 1
