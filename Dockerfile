# Build stage
FROM golang:1.25-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /self-ca .

# Runtime stage
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /self-ca /usr/local/bin/self-ca
COPY config.yml /etc/self-ca/config.yml
WORKDIR /data
VOLUME ["/data"]
EXPOSE 8080 8443
ENTRYPOINT ["self-ca", "-config", "/etc/self-ca/config.yml"]
