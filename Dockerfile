FROM golang:1.14-alpine AS backend

RUN apk add --no-cache ca-certificates git
WORKDIR /go/src/github.com/RocketChat/MigrateFileStore
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o migrate

FROM scratch as runtime

WORKDIR /usr/local/MigrateFileStore

COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /go/src/github.com/RocketChat/MigrateFileStore/migrate .

CMD ["./migrate"]
