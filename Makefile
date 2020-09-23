binary_name=migratestore
version=latest

# Build the tool binary
build: clean
	@CGO_ENABLED=0 go build -v -a -ldflags "-X main.version=${version}" -o $$GOPATH/bin/${binary_name} ./cmd/migratestore/main.go

# Delete all artifacts generated by the build routines.
clean:
	@rm -f $$GOPATH/bin/${binary_name}

# Build the tool docker image.
docker:
	docker build -t migrate-file-store:${version} .

# Build and execute the tool binary.
run: build
	$$GOPATH/bin/${binary_name}