PLATFORM=$(shell uname -s | tr '[:upper:]' '[:lower:]')
VERSION := $(shell grep -Eo '(v[0-9]+[\.][0-9]+[\.][0-9]+([-a-zA-Z0-9]*)?)' version.go)

.PHONY: build generate docker release

build:
	go fmt ./...
	@mkdir -p ./bin/
	CGO_ENABLED=0 go build -o bin/odfw github.com/adamdecaf/odfw

clean:
	@rm -rf bin/

docker: clean
# ACH docker image
	docker build --pull -t adamdecaf/odfw:$(VERSION) -f Dockerfile .
	docker tag adamdecaf/odfw:$(VERSION) adamdecaf/odfw:latest

release-push:
	docker push adamdecaf/odfw:$(VERSION)
	docker push adamdecaf/odfw:latest
