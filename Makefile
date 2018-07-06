all: clean install test vet lint check-gofmt

check-gofmt:
	scripts/check_gofmt.sh

clean:
	go clean

install:
	go install

lint:
	golint -set_exit_status ./...

# Builds a package for upload to AWS Lambda.
package:
	GOOS=linux go build .
	zip perpetual.zip ./perpetual

test:
	go test ./...

vet:
	go vet ./...
