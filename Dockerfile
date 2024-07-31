#build stage
FROM golang AS builder
RUN apt-get update && apt-get install -y --no-install-recommends git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -v -o /app main.go

#final stage
FROM golang:latest
RUN apt update && apt install -y --no-install-recommends ca-certificates
WORKDIR /app
COPY --from=builder /app .
COPY stations.json .
ENTRYPOINT ["./app"]
LABEL Name=radiotospotify Version=0.0.2