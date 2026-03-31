FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /ssm-server ./cmd/ssm-server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /ssm-server /usr/local/bin/ssm-server
EXPOSE 8080
VOLUME /data
ENV DATA_DIR=/data
ENTRYPOINT ["ssm-server"]
