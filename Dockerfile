FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /ssm-sync ./cmd/ssm-sync

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /ssm-sync /usr/local/bin/ssm-sync
EXPOSE 8080
VOLUME /data
ENV DATA_DIR=/data
ENTRYPOINT ["ssm-sync"]
