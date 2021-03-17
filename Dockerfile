FROM golang:1.14-alpine AS backend

RUN apk add --no-cache ca-certificates git
WORKDIR /go/src/github.com/RocketChat/filestore-migrator
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o filestore-migrator ./cmd/filestore-migrator/

FROM scratch as runtime

WORKDIR /usr/local/filestore-migrator

COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /go/src/github.com/RocketChat/filestore-migrator/filestore-migrator .

ENTRYPOINT [ "./filestore-migrator" ]
