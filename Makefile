all: clean install test vet lint check-gofmt

check-gofmt:
	scripts/check_gofmt.sh

clean:
	go clean

install:
	go install

lint:
	golint -set_exit_status ./...

test:
	go test ./...

vet:
	go vet ./...
